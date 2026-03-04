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

	// Header fields: status and current value
	fields1 := []map[string]interface{}{
		{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":    "lark_md",
				"content": fmt.Sprintf("告警级别：%s", statusEmoji(event.EventStatus)),
			},
		},
	}
	currentValue := event.Attrs[types.AttrCurrentValue]
	if currentValue != "" {
		fields1 = append(fields1, map[string]interface{}{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":    "lark_md",
				"content": fmt.Sprintf("当前值：%s", currentValue),
			},
		})
	}
	elements = append(elements, map[string]interface{}{
		"tag":    "div",
		"fields": fields1,
	})

	// Host and time
	fields2 := []map[string]interface{}{
		{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":    "lark_md",
				"content": fmt.Sprintf("主机：%s (%s)", host, hostIP),
			},
		},
		{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":    "lark_md",
				"content": fmt.Sprintf("告警时间：%s", time.Unix(event.EventTime, 0).Format("2006-01-02 15:04:05")),
			},
		},
	}
	elements = append(elements, map[string]interface{}{
		"tag":    "div",
		"fields": fields2,
	})

	// Divider
	elements = append(elements, map[string]interface{}{"tag": "hr"})

	// Plugin and target info
	pluginInfo := fmt.Sprintf("📊 插件：%s", plugin)
	if target != "" {
		pluginInfo = fmt.Sprintf("📊 插件：%s | 目标：%s", plugin, target)
	}
	elements = append(elements, map[string]interface{}{
		"tag": "div",
		"text": map[string]interface{}{
			"tag":    "lark_md",
			"content": pluginInfo,
		},
	})

	// Description
	if event.Description != "" {
		elements = append(elements, map[string]interface{}{"tag": "hr"})
		desc := event.Description
		// Clean up description for better display
		desc = strings.ReplaceAll(desc, "**", "")
		desc = strings.ReplaceAll(desc, "`", "")
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":    "lark_md",
				"content": fmt.Sprintf("📝 详情：\n%s", desc),
			},
		})
	}

	// Divider before footer
	elements = append(elements, map[string]interface{}{"tag": "hr"})

	// Footer
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

func statusEmoji(status string) string {
	switch status {
	case types.EventStatusCritical:
		return "🔴 Critical"
	case types.EventStatusWarning:
		return "⚠️ Warning"
	case types.EventStatusInfo:
		return "ℹ️ Info"
	case types.EventStatusOk:
		return "✅ OK"
	default:
		return status
	}
}
