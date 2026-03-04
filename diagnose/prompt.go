package diagnose

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

var promptTmpl = template.Must(template.New("prompt").Funcs(template.FuncMap{
	"add": func(a, b int) int { return a + b },
}).Parse(promptRaw))

const promptRaw = `你是一位资深运维和 DBA 专家。

{{- if eq .Mode "inspect"}}

用户请求对以下目标进行主动健康巡检：

插件: {{.Plugin}}
目标: {{.Target}}

这不是告警触发的诊断，而是一次主动巡检。你的任务是全面检查目标的健康状态，发现潜在问题。
{{- else}}

catpaw 监控系统检测到以下告警：

插件: {{.Plugin}}
目标: {{.Target}}

{{if eq (len .Checks) 1 -}}
### 告警详情
检查项: {{(index .Checks 0).Check}}
严重级别: {{(index .Checks 0).Status}}
当前值: {{(index .Checks 0).CurrentValue}}
{{- if (index .Checks 0).ThresholdDesc}}
阈值: {{(index .Checks 0).ThresholdDesc}}
{{- end}}
描述: {{(index .Checks 0).Description}}
{{- else if gt (len .Checks) 1 -}}
### 告警详情（同一目标有 {{len .Checks}} 个异常检查项，可能存在关联）
{{range $i, $c := .Checks}}
[{{add $i 1}}] {{$c.Check}} - {{$c.Status}}
    当前值: {{$c.CurrentValue}}
    {{- if $c.ThresholdDesc}}
    阈值: {{$c.ThresholdDesc}}
    {{- end}}
    描述: {{$c.Description}}
{{- end}}
请特别关注这些异常之间是否存在共同根因。
{{- end}}

你的任务是诊断这个问题的根因，并给出建议操作。
{{- end}}

## 可用工具

你可以直接调用以下 {{.Plugin}} 工具（无需通过 call_tool）：
{{.DirectTools}}

以下是系统中所有可用的诊断工具（按领域分类）：

{{.ToolCatalog}}

调用其他领域的工具：call_tool(name="工具名", tool_args='{"参数名":"值"}')
如需查看某类工具的详细参数说明：list_tools(category="类别名")

注意：上述 {{.Plugin}} 工具请直接调用，不要通过 call_tool 包装。

{{- if eq .Mode "inspect"}}

## 巡检策略

1. 首先使用 {{.Plugin}} 的核心工具收集关键指标
2. 根据初步结果，针对性地深入检查可疑领域
3. 如果是远端服务，同时关注基础设施层面可能影响服务的因素
4. **每轮尽可能并行调用多个工具**，减少交互轮次

## 巡检深度要求

请按以下层次检查，不要遗漏潜在问题：

**Layer 1 - 基础指标**：核心性能指标是否正常（CPU/内存/连接数等）
**Layer 2 - 配置检查**：关键配置是否合理（超时/缓存/连接池等）
**Layer 3 - 资源趋势**：是否有资源即将耗尽的迹象
**Layer 4 - 潜在风险**：是否有隐藏的问题（慢查询/大 key/碎片等）
**Layer 5 - 优化建议**：有哪些可以改进的地方

巡检过程中，确保覆盖到 Layer 3 及以上，发现潜在风险比确认已知问题更有价值。
{{- else}}

## 诊断策略

- **效率优先**：如果当前信息已足以判断根因，立即输出结论，不要为了全面性进行不必要的检查
- **并行调用**：需要多个领域数据时，在同一轮中并行调用多个工具
- **聚焦问题**：优先检查与告警直接相关的指标；只在初步分析无法解释问题时才扩展到其他领域
- 根因可能不在 {{.Plugin}} 自身（如数据库慢可能源于磁盘 I/O），但请先确认直接相关指标后再决定是否扩展
{{- end}}
{{- if .IsRemoteTarget}}
- [!] 目标 {{.Target}} 是远端主机，本机基础设施工具（disk、cpu、memory 等）
  反映的是 catpaw 所在主机 {{.LocalHost}} 的状态，不是目标主机的状态
  这些工具的结果仅在 catpaw 与目标部署在同一台机器时有参考价值
{{- else}}
- catpaw 与目标 {{.Target}} 在同一台机器上，本机基础设施工具可直接用于辅助诊断
{{- end}}

## 诊断深度要求

请按以下层次分析问题，不要停在表层：

**Layer 1 - 现象确认**：告警指标是什么？当前值多少？（已提供）
**Layer 2 - 直接原因**：哪个进程/组件导致的？（需调用工具确认）
**Layer 3 - 根本原因**：为什么会出现这个问题？代码层面还是配置层面？
**Layer 4 - 关联影响**：是否影响其他系统？是否有连锁反应？
**Layer 5 - 预防措施**：如何避免再次发生？需要什么改进？

诊断过程中，在思考时标注当前分析的层级。如果停在 Layer 2 及以下，说明诊断不够深入，请继续调查。

## 输出要求

{{- if eq .Mode "inspect"}}

请按以下格式输出健康报告：

### 1. 巡检摘要
一句话总结目标的整体健康状态

### 2. 检查项明细
逐项列出检查结果，每项使用状态标记：
- [OK] 正常：指标在健康范围内
- [WARN] 警告：指标偏离正常但尚未达到告警阈值，需关注
- [CRIT] 异常：指标已达到危险水平，需立即处理

每项附带关键数值和判断依据

### 3. 风险与建议
- 发现的潜在风险（尚未触发告警但趋势不好的指标）
- 优化建议（按紧急程度排序）
{{- else}}

- 语言精炼，关键数值内嵌到分析要点中
- 最终输出请按以下格式：
  1. 诊断摘要（一句话）
  2. 根因分析（要点列表，每条含关键数值）
  3. 建议操作（按紧急/短期/中期分类）
- 不要输出原始数据的完整内容，只引用关键数值
{{- end}}

请只使用工具获取信息，不要假设或编造数据。
{{- if ne .Language "zh"}}

IMPORTANT: You MUST respond in {{.Language}}. All output including section headers, analysis, and recommendations must be in {{.Language}}.
{{- end}}
{{- if eq .ReportStyle "casual"}}

输出风格：请用轻松口语化的方式输出报告，可以适当使用比喻让技术问题更容易理解，但不要牺牲准确性。
{{- else if eq .ReportStyle "humorous"}}

输出风格：请用幽默风趣的方式输出报告，可以用职场梗、生活比喻来描述系统问题，缓解运维人员的告警疲劳。比如可以把 CPU 飙高比作"打工人疯狂加班"，把内存泄漏比作"只进不出的貔貅"等。但核心信息必须准确，不要影响问题的严重性判断。
{{- end}}`

type promptData struct {
	Mode           string
	Plugin         string
	Target         string
	Checks         []CheckSnapshot
	DirectTools    string
	ToolCatalog    string
	IsRemoteTarget bool
	LocalHost      string
	Language       string
	ReportStyle    string
}

func buildSystemPrompt(req *DiagnoseRequest, directTools, toolCatalog, localHost string, isRemote bool, language, reportStyle string) string {
	return renderPrompt(ModeAlert, req, directTools, toolCatalog, localHost, isRemote, language, reportStyle)
}

func buildInspectPrompt(req *DiagnoseRequest, directTools, toolCatalog, localHost string, isRemote bool, language, reportStyle string) string {
	return renderPrompt(ModeInspect, req, directTools, toolCatalog, localHost, isRemote, language, reportStyle)
}

func renderPrompt(mode string, req *DiagnoseRequest, directTools, toolCatalog, localHost string, isRemote bool, language, reportStyle string) string {
	data := promptData{
		Mode:           mode,
		Plugin:         req.Plugin,
		Target:         req.Target,
		Checks:         req.Checks,
		DirectTools:    directTools,
		ToolCatalog:    toolCatalog,
		IsRemoteTarget: isRemote,
		LocalHost:      localHost,
		Language:       language,
		ReportStyle:    reportStyle,
	}

	var buf bytes.Buffer
	if err := promptTmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("Error building prompt: %v", err)
	}
	return buf.String()
}

func formatDirectTools(tools []DiagnoseTool) string {
	if len(tools) == 0 {
		return "(无直接工具)"
	}
	var b strings.Builder
	for _, t := range tools {
		fmt.Fprintf(&b, "- %s: %s\n", t.Name, t.Description)
		for _, p := range t.Parameters {
			req := ""
			if p.Required {
				req = " (必需)"
			}
			fmt.Fprintf(&b, "  参数 %s (%s): %s%s\n", p.Name, p.Type, p.Description, req)
		}
	}
	return b.String()
}

func isRemoteTarget(target string) bool {
	t := strings.ToLower(target)
	if strings.HasPrefix(t, "localhost") || strings.HasPrefix(t, "127.") ||
		strings.HasPrefix(t, "[::1]") || strings.HasPrefix(t, "::1") {
		return false
	}
	if t == "" || t == "/" {
		return false
	}
	return true
}
