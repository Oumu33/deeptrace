# DeepTrace

> **AI-Powered Deep Diagnostic Engine for Production Systems**

[English](README.md) | [中文](README_zh.md)

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen?style=flat)](CONTRIBUTING.md)

---

## What is DeepTrace?

DeepTrace is an intelligent diagnostic engine that goes beyond traditional monitoring alerts. When something breaks, it doesn't just tell you **what** happened — it tells you **why** and **where** in your code.

```
Traditional Alert:  "CPU usage is 92%"
DeepTrace Report:   "CPU 92% → myapp process (PID 12345) → json.Marshal at handler.go:45
                     → Processing 10MB JSON response → API missing pagination"
```

## 🚀 Features

### 🧠 5-Layer Deep Diagnosis

| Layer | What We Find | Example |
|-------|--------------|---------|
| **1. Symptom** | What triggered the alert | CPU 92%, threshold 80% |
| **2. Direct Cause** | Which process/component | myapp (PID 12345), 89% CPU |
| **3. Root Cause** | Why it happened | Large JSON serialization, no pagination |
| **4. Impact** | What else is affected | API latency 50ms → 3200ms, connection pool 85% |
| **5. Prevention** | How to prevent recurrence | Add pagination, limit response size |

### 🔧 79+ Built-in Diagnostic Tools

**System Layer**: CPU, Memory, Disk, Network, Process, File Descriptors

**Network Layer**: Connections, TCP states, Retransmission, Latency, DNS, Firewall

**Storage Layer**: I/O latency, Block devices, LVM, Mount points

**Security Layer**: SELinux, AppArmor, Audit logs, Conntrack

**Application Layer**: Stack traces for Go/Java/Python/Node.js

### 🎯 Code-Level Visibility

```go
// DeepTrace can capture stack traces and show you exactly where code is stuck
goroutine 1 [running]:
encoding/json.Marshal()
    /usr/local/go/src/encoding/json/encode.go:161
main.handleRequest()
    /opt/myapp/handler.go:45  ← Problem is here
main.main()
    /opt/myapp/main.go:23
```

### 💬 Dual-Mode Reports

**For Ops** — Quick summary + actionable commands:
- Problem: myapp CPU 89%
- Action: `top -p 12345` or `systemctl restart myapp`

**For Devs** — Stack traces + code location + root cause:
- File: `handler.go:45`
- Cause: Large JSON serialization without pagination

### 📡 Multi-Channel Notifications

- DingTalk (钉钉) — Native Markdown support
- Feishu (飞书) — Via FlashDuty integration
- WeChat Work (企业微信) — Via FlashDuty integration
- PagerDuty — For international teams
- Generic Webhook — Any HTTP endpoint

## 📦 Installation

```bash
# Download from releases
wget https://github.com/Oumu33/deeptrace/releases/latest/download/deeptrace-linux-amd64
chmod +x deeptrace-linux-amd64
sudo mv deeptrace-linux-amd64 /usr/local/bin/deeptrace

# Or build from source
git clone https://github.com/Oumu33/deeptrace.git
cd deeptrace
go build -o deeptrace .
```

## ⚡ Quick Start

### 1. Basic Monitoring

Create `conf.d/config.toml`:

```toml
[global]
interval = "30s"

[notify.console]
enabled = true
```

Create CPU monitor `conf.d/p.cpu/cpu.toml`:

```toml
[[instances]]
targets = ["localhost"]

[[instances.alerts]]
check = "cpu_usage"
warn_ge = 80
crit_ge = 95
```

Run:

```bash
./deeptrace run
```

### 2. Enable AI Diagnosis

Add to `conf.d/config.toml`:

```toml
[ai]
enabled = true
model_priority = ["gpt4o"]
report_style = "professional"  # professional / casual / humorous

[ai.models.gpt4o]
base_url = "https://api.openai.com/v1"
api_key = "${OPENAI_API_KEY}"
model = "gpt-4o"
```

Now when alerts fire, DeepTrace automatically:
1. Collects diagnostic data using 79+ tools
2. Analyzes root cause with AI
3. Generates 5-layer deep diagnosis report
4. Pushes report to your notification channel

### 3. Interactive Troubleshooting

```bash
./deeptrace chat
```

Ask anything:
- "Why is CPU high on this server?"
- "Check Redis memory usage and find big keys"
- "Show me network connections to 10.0.0.1"

### 4. Proactive Health Inspection

```bash
# Inspect Redis
./deeptrace inspect redis 10.0.0.1:6379

# Inspect local system
./deeptrace inspect cpu
./deeptrace inspect mem
./deeptrace inspect disk
```

## 🎨 Report Example

Here's what you'll see in DingTalk when CPU alert fires:

```
🚨 CPU告警 - 生产服务器

告警级别: ⚠️ Warning
主机: prod-server-01 (192.168.1.10)
当前值: CPU 92.3%（阈值 ≥80%）

━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 问题概述
CPU 使用率持续升高，由 myapp 进程导致

━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔍 根因分析（深度诊断）

Layer 1 - 现象确认 ✅
CPU 使用率 92.3%，已超阈值

Layer 2 - 直接原因 ✅
进程: myapp (PID: 12345)
CPU 占用: 89%，内存: 2.1GB

Layer 3 - 根本原因 ✅
堆栈显示阻塞在 json.Marshal
根因: API 返回全量数据，缺少分页

Layer 4 - 关联影响 ✅
API 响应时间: 50ms → 3200ms
下游调用方出现超时告警

Layer 5 - 预防措施 ✅
后端添加分页，限制单次返回数量

━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ 建议操作

运维立即执行:
1. top -p 12345 确认进程状态
2. systemctl restart myapp

转交开发处理:
1. 检查 /api/users/list 接口
2. 添加分页参数

━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📎 代码位置
handler.go:45 ← 问题代码位置

诊断深度: 5/5 层
诊断耗时: 3.2s
```

## 🛠️ Supported Languages for Stack Trace

| Language | Tool | Requirement |
|----------|------|-------------|
| Go | pprof endpoint | Program exposes pprof |
| Java | jstack | JDK installed |
| Python | py-spy | `pip install py-spy` |
| Node.js | llnode/lldb | llnode installed |
| Generic | gdb | gdb installed + debug symbols |

## 📋 Plugin List

| Plugin | Description |
|--------|-------------|
| `cpu` | CPU utilization and load average |
| `mem` | Memory and swap usage |
| `disk` | Disk space and I/O |
| `network` | Network interfaces and connections |
| `redis` | Redis monitoring and diagnosis |
| `docker` | Docker container status |
| `systemd` | Service health check |
| `http` | HTTP endpoint probing |
| `dns` | DNS resolution check |
| `ping` | ICMP reachability |
| ... | 25+ plugins total |

## 🔌 MCP Integration

Connect external data sources:

```toml
[[ai.mcp.servers]]
name = "prometheus"
command = "/usr/local/bin/mcp-prometheus"
[ai.mcp.servers.env]
PROMETHEUS_URL = "http://localhost:9090"
```

AI can now query Prometheus for historical metrics during diagnosis.

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md).

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

<p align="center">
  <b>Stop guessing. Start tracing.</b><br>
  <sub>Built with ❤️ for ops and devs who hate 3 AM debugging sessions</sub>
</p>