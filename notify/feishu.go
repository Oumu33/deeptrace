package notify

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cprobe/catpaw/config"
	"github.com/cprobe/catpaw/logger"
	"github.com/cprobe/catpaw/types"
)

type FeishuNotifier struct {
	cfg    *config.FeishuConfig
	client *http.Client
}

func NewFeishuNotifier(cfg *config.FeishuConfig) *FeishuNotifier {
	return &FeishuNotifier{
		cfg: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout),
		},
	}
}

func (f *FeishuNotifier) Name() string { return "feishu" }

func (f *FeishuNotifier) Forward(event *types.Event) bool {
	card := f.buildCard(event)
	payload, err := json.Marshal(map[string]interface{}{
		"msg_type": "interactive",
		"card":     card,
	})
	if err != nil {
		logger.Logger.Errorw("feishu: marshal fail", "event_key", event.AlertKey, "error", err.Error())
		return false
	}

	for attempt := 0; attempt <= f.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
			logger.Logger.Infow("feishu: retrying", "event_key", event.AlertKey, "attempt", attempt+1)
		}

		ok, retryable := f.doRequest(event.AlertKey, payload)
		if ok {
			return true
		}
		if !retryable {
			return false
		}
	}

	logger.Logger.Errorw("feishu: all retries exhausted", "event_key", event.AlertKey, "max_retries", f.cfg.MaxRetries)
	return false
}

func (f *FeishuNotifier) doRequest(alertKey string, payload []byte) (ok bool, retryable bool) {
	url := f.cfg.Webhook
	if f.cfg.Secret != "" {
		timestamp := time.Now().Unix()
		sign := f.sign(timestamp, f.cfg.Secret)
		url = fmt.Sprintf("%s&timestamp=%d&sign=%s", f.cfg.Webhook, timestamp, sign)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		logger.Logger.Errorw("feishu: new request fail", "event_key", alertKey, "error", err.Error())
		return false, false
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := f.client.Do(req)
	if err != nil {
		logger.Logger.Errorw("feishu: do request fail", "event_key", alertKey, "error", err.Error())
		return false, true
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(io.LimitReader(res.Body, 4096))
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		var resp struct {
			StatusCode int    `json:"StatusCode"`
			Msg        string `json:"msg"`
		}
		json.Unmarshal(body, &resp)
		if resp.StatusCode == 0 {
			logger.Logger.Infow("feishu: forward completed", "event_key", alertKey)
			return true, false
		}
		logger.Logger.Errorw("feishu: api error", "event_key", alertKey, "msg", resp.Msg)
		return false, false
	}

	if res.StatusCode == 429 || res.StatusCode >= 500 {
		logger.Logger.Errorw("feishu: retryable error", "event_key", alertKey, "status", res.StatusCode)
		return false, true
	}

	logger.Logger.Errorw("feishu: non-retryable error", "event_key", alertKey, "status", res.StatusCode, "body", string(body))
	return false, false
}

func (f *FeishuNotifier) sign(timestamp int64, secret string) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	h.Write(nil)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (f *FeishuNotifier) buildCard(event *types.Event) map[string]interface{} {
	template := "orange"
	if event.EventStatus == types.EventStatusCritical {
		template = "red"
	} else if event.EventStatus == types.EventStatusOk {
		template = "green"
	} else if event.EventStatus == types.EventStatusInfo {
		template = "blue"
	}

	plugin := event.Labels["plugin"]
	target := event.Labels["target"]
	host := event.Labels["from_hostname"]
	hostIP := event.Labels["from_hostip"]

	var elements []map[string]interface{}

	// 告警状态和当前值
	fields1 := []map[string]interface{}{
		{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**告警级别**\n%s", statusEmoji(event.EventStatus)),
			},
		},
	}
	currentValue := event.Attrs[types.AttrCurrentValue]
	if currentValue != "" {
		fields1 = append(fields1, map[string]interface{}{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**当前值**\n%s", currentValue),
			},
		})
	}
	elements = append(elements, map[string]interface{}{
		"tag":    "div",
		"fields": fields1,
	})

	// 主机和时间
	fields2 := []map[string]interface{}{
		{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**主机**\n%s (%s)", host, hostIP),
			},
		},
		{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**告警时间**\n%s", time.Unix(event.EventTime, 0).Format("2006-01-02 15:04:05")),
			},
		},
	}
	elements = append(elements, map[string]interface{}{
		"tag":    "div",
		"fields": fields2,
	})

	// 分隔线
	elements = append(elements, map[string]interface{}{"tag": "hr"})

	// 插件和目标信息
	pluginInfo := fmt.Sprintf("**插件**: %s", plugin)
	if target != "" {
		pluginInfo = fmt.Sprintf("**插件**: %s | **目标**: %s", plugin, target)
	}
	elements = append(elements, map[string]interface{}{
		"tag": "div",
		"text": map[string]interface{}{
			"tag":     "lark_md",
			"content": pluginInfo,
		},
	})

	// 描述 / AI 诊断报告
	if event.Description != "" {
		elements = append(elements, map[string]interface{}{"tag": "hr"})

		// 检测是否为 AI 诊断报告
		if strings.Contains(event.Description, "AI 诊断报告") || strings.Contains(event.Description, "AI Diagnosis Report") {
			// AI 诊断报告：结构化展示
			elements = append(elements, f.buildAIReportElements(event.Description)...)
		} else {
			// 普通告警描述
			elements = append(elements, map[string]interface{}{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "lark_md",
					"content": fmt.Sprintf("**详情**\n%s", formatMarkdown(event.Description)),
				},
			})
		}
	}

	// 底部分隔线
	elements = append(elements, map[string]interface{}{"tag": "hr"})

	// 页脚
	elements = append(elements, map[string]interface{}{
		"tag": "note",
		"elements": []map[string]interface{}{
			{
				"tag":     "plain_text",
				"content": fmt.Sprintf("alert_key: %s | 由 DeepTrace 自动推送", event.AlertKey),
			},
		},
	})

	return map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
		"header": map[string]interface{}{
			"title": map[string]interface{}{
				"tag":     "plain_text",
				"content": fmt.Sprintf("🚨 %s告警 - %s", event.EventStatus, plugin),
			},
			"template": template,
		},
		"elements": elements,
	}
}

// buildAIReportElements 解析 AI 诊断报告并构建结构化的飞书卡片元素
func (f *FeishuNotifier) buildAIReportElements(desc string) []map[string]interface{} {
	var elements []map[string]interface{}

	// 解析报告结构
	sections := parseAIReport(desc)

	// 状态徽章
	if sections.Status != "" {
		statusIcon := "✅"
		if sections.Status == "failed" {
			statusIcon = "❌"
		}
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("%s **AI 诊断状态**: %s", statusIcon, sections.Status),
			},
		})
	}

	// 诊断信息
	if sections.Meta != "" {
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": sections.Meta,
			},
		})
	}

	// 诊断摘要（如果有）
	if sections.Summary != "" {
		elements = append(elements, map[string]interface{}{"tag": "hr"})
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**📋 诊断摘要**\n%s", formatMarkdown(sections.Summary)),
			},
		})
	}

	// 根因分析（如果有）
	if sections.RootCause != "" {
		elements = append(elements, map[string]interface{}{"tag": "hr"})
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**🔍 根因分析**\n%s", formatMarkdown(sections.RootCause)),
			},
		})
	}

	// 建议操作（如果有）
	if sections.Actions != "" {
		elements = append(elements, map[string]interface{}{"tag": "hr"})
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**✅ 建议操作**\n%s", formatMarkdown(sections.Actions)),
			},
		})
	}

	// 原始报告（折叠展示）
	if len(desc) > 500 {
		elements = append(elements, map[string]interface{}{"tag": "hr"})
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**📝 完整报告**\n%s", formatMarkdown(truncateText(desc, 2000))),
			},
		})
	}

	// 查看命令
	if sections.ViewCommand != "" {
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("`%s`", sections.ViewCommand),
			},
		})
	}

	return elements
}

// AIReportSections 表示解析后的 AI 诊断报告结构
type AIReportSections struct {
	Status     string
	Meta       string
	Summary    string
	RootCause  string
	Actions    string
	ViewCommand string
}

// parseAIReport 解析 AI 诊断报告文本，提取关键部分
func parseAIReport(desc string) AIReportSections {
	var sections AIReportSections

	lines := strings.Split(desc, "\n")

	var currentSection *strings.Builder
	var metaLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 提取状态
		if strings.Contains(trimmed, "[success]") || strings.Contains(trimmed, "[failed]") {
			if strings.Contains(trimmed, "success") {
				sections.Status = "success"
			} else {
				sections.Status = "failed"
			}
			continue
		}

		// 提取元信息 (插件, 目标, 时间等)
		if strings.HasPrefix(trimmed, "插件:") || strings.HasPrefix(trimmed, "Plugin:") ||
			strings.HasPrefix(trimmed, "诊断时间:") || strings.HasPrefix(trimmed, "Time:") ||
			strings.HasPrefix(trimmed, "目标:") || strings.HasPrefix(trimmed, "Target:") {
			metaLines = append(metaLines, trimmed)
			continue
		}

		// 检测章节标题
		if strings.Contains(trimmed, "诊断摘要") || strings.Contains(trimmed, "Diagnostic Summary") {
			currentSection = &strings.Builder{}
			continue
		}
		if strings.Contains(trimmed, "根因分析") || strings.Contains(trimmed, "Root Cause") {
			if currentSection != nil && sections.Summary == "" {
				sections.Summary = currentSection.String()
			}
			currentSection = &strings.Builder{}
			continue
		}
		if strings.Contains(trimmed, "建议操作") || strings.Contains(trimmed, "Recommended Actions") {
			if currentSection != nil && sections.RootCause == "" {
				sections.RootCause = currentSection.String()
			}
			currentSection = &strings.Builder{}
			continue
		}

		// 提取查看命令
		if strings.HasPrefix(trimmed, "查看命令:") || strings.HasPrefix(trimmed, "View command:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				sections.ViewCommand = strings.TrimSpace(parts[1])
			}
			continue
		}

		// 收集当前章节内容
		if currentSection != nil && trimmed != "" && !strings.HasPrefix(trimmed, "---") {
			currentSection.WriteString(trimmed)
			currentSection.WriteString("\n")
		}
	}

	// 保存最后一个章节
	if currentSection != nil {
		content := currentSection.String()
		if sections.Actions == "" && (strings.Contains(desc, "建议操作") || strings.Contains(desc, "Actions")) {
			sections.Actions = content
		} else if sections.RootCause == "" {
			sections.RootCause = content
		}
	}

	// 组装元信息
	if len(metaLines) > 0 {
		sections.Meta = strings.Join(metaLines, " | ")
	}

	return sections
}

// formatMarkdown 将 Markdown 格式转换为飞书 lark_md 兼容格式
func formatMarkdown(text string) string {
	// 飞书 lark_md 支持的格式：
	// **粗体**、*斜体*、`代码`、<font color="red">彩色文字</font>
	// 不支持：### 标题、> 引用、- 列表（需用数字或符号）

	// 保留粗体和斜体
	// 处理标题：转换为粗体
	text = strings.ReplaceAll(text, "### ", "**")
	text = strings.ReplaceAll(text, "## ", "**")
	text = strings.ReplaceAll(text, "# ", "**")

	// 处理列表：添加换行和符号
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 数字列表保持原样
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			// 无序列表转换为符号
			line = "• " + strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// truncateText 截断文本到指定长度
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "\n...[已截断]"
}

func statusEmoji(status string) string {
	switch status {
	case types.EventStatusCritical:
		return "🔴 Critical"
	case types.EventStatusWarning:
		return "🟡 Warning"
	case types.EventStatusInfo:
		return "🔵 Info"
	case types.EventStatusOk:
		return "🟢 OK"
	default:
		return status
	}
}