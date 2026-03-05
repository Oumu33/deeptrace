<p align="center">
  <img src="https://img.shields.io/badge/DeepTrace-AI%20Deep%20Diagnosis-blue?style=for-the-badge" alt="DeepTrace"/>
</p>

<h1 align="center">DeepTrace</h1>

<p align="center">
  <strong>AI-Powered Deep Diagnostic Engine for Production Systems</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#installation">Installation</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#configuration">Configuration</a>
</p>

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go" alt="Go Version"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green?style=flat" alt="License"/></a>
  <a href="https://github.com/Oumu33/deeptrace/releases"><img src="https://img.shields.io/github/v/release/Oumu33/deeptrace?style=flat" alt="Release"/></a>
</p>

<p align="center">
  <a href="README.md">English</a> | <a href="README_zh.md">中文</a>
</p>

---

## Why DeepTrace?

Traditional monitoring tells you **what** happened. DeepTrace tells you **why** and **where**.

```
Traditional Alert:    "CPU usage is 92%"
                      → You SSH in, run top, investigate logs...

DeepTrace Diagnosis:  "CPU 92% → myapp (PID 12345)
                       → json.Marshal at handler.go:45
                       → Processing 10MB JSON response
                       → API endpoint missing pagination
                       → Add pagination, limit response size"
                      → Root cause identified automatically
```

## Features

### Core Capabilities

| Feature | Description |
|---------|-------------|
| **5-Layer Deep Diagnosis** | Symptom → Direct Cause → Root Cause → Impact → Prevention |
| **Code-Level Visibility** | Stack traces for Go, Java, Python, Node.js |
| **AI Agent** | Autonomous tool invocation for root cause analysis |
| **Interactive Chat** | Natural language troubleshooting via CLI |
| **MCP Integration** | Connect Prometheus, Jaeger, CMDB for enriched context |

### Built-in Observability

- **31 Monitoring Plugins** — System, Network, Storage, Services, Security
- **32 Diagnostic Tools** — On-demand deep inspection utilities
- **Multi-Channel Alerts** — Feishu, DingTalk, PagerDuty, Webhook

### AI Integration

- **OpenAI-Compatible API** — Works with GPT-4o, DeepSeek, Claude, Ollama
- **Automatic Failover** — Switch models on rate limits or server errors
- **Token Budget Control** — Daily limits and cost estimation
- **Multi-Language Reports** — English and Chinese output

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          DeepTrace CLI                           │
├─────────────┬─────────────┬─────────────┬───────────────────────┤
│   run       │   chat      │   inspect   │   diagnose            │
│   Agent     │   AI REPL   │   Health    │   Records             │
└──────┬──────┴─────────────┴──────┬──────┴───────────────────────┘
       │                           │
       ▼                           ▼
┌──────────────┐           ┌──────────────┐
│   Plugins    │           │   AI Engine  │
│   (31)       │           │              │
│              │           │  ┌────────┐  │
│  cpu, mem    │           │  │LLM API │  │
│  disk, net   │           │  └────┬───┘  │
│  redis, ...  │           │       │      │
└──────┬───────┘           │  ┌────▼───┐  │
       │                   │  │ MCP    │  │
       ▼                   │  │ Tools  │  │
┌──────────────┐           │  └────────┘  │
│   Engine     │           └──────┬───────┘
│              │                  │
│  Thresholds  │                  │
│  Aggregation │                  │
└──────┬───────┘                  │
       │                          │
       ▼                          ▼
┌─────────────────────────────────────────┐
│              Notify Channels             │
│   Console │ Feishu │ DingTalk │ Webhook │
└─────────────────────────────────────────┘
```

### Components

| Component | Path | Responsibility |
|-----------|------|----------------|
| **CLI** | `main.go` | Command routing and argument parsing |
| **Agent** | `agent/` | Plugin lifecycle, configuration loading |
| **Engine** | `engine/` | Event processing, threshold detection |
| **Diagnose** | `diagnose/` | AI-powered root cause analysis |
| **Chat** | `chat/` | Interactive troubleshooting REPL |
| **MCP** | `mcp/` | External data source integration |
| **Notify** | `notify/` | Multi-channel alert dispatch |
| **Plugins** | `plugins/` | 31 built-in monitoring collectors |

---

## Installation

### Binary Download

```bash
# Linux AMD64
wget https://github.com/Oumu33/deeptrace/releases/latest/download/deeptrace-linux-amd64
chmod +x deeptrace-linux-amd64
sudo mv deeptrace-linux-amd64 /usr/local/bin/deeptrace

# Linux ARM64
wget https://github.com/Oumu33/deeptrace/releases/latest/download/deeptrace-linux-arm64
chmod +x deeptrace-linux-arm64
sudo mv deeptrace-linux-arm64 /usr/local/bin/deeptrace
```

### Build from Source

```bash
git clone https://github.com/Oumu33/deeptrace.git
cd deeptrace
go build -o deeptrace .
```

### Requirements

- Go 1.21+ (for building)
- Linux (recommended) / Windows / macOS

---

## Quick Start

### 1. Basic Monitoring

```bash
# Create configuration directory
mkdir -p conf.d

# Run with default plugins
./deeptrace run
```

### 2. Interactive Troubleshooting

```bash
# Start AI chat session
./deeptrace chat

> Why is my server slow?
> Check Redis connection to 10.0.0.1:6379
> Analyze the CPU spike from 5 minutes ago
```

### 3. Health Inspection

```bash
# Inspect specific component
./deeptrace inspect cpu
./deeptrace inspect redis 10.0.0.1:6379
./deeptrace inspect http https://api.example.com/health
```

### 4. Diagnosis Records

```bash
# List all diagnosis records
./deeptrace diagnose list

# View specific diagnosis
./deeptrace show <record-id>
```

---

## Configuration

### Main Config (`conf.d/config.toml`)

```toml
[global]
interval = "30s"                    # Collection interval

[global.labels]
env = "production"
region = "us-east-1"

[log]
level = "info"                      # debug, info, warn, error
format = "json"                     # json, text
output = "stdout"                   # stdout, file
```

### AI Configuration

```toml
[ai]
enabled = true
model_priority = ["gpt4o", "deepseek"]  # Failover order
language = "en"                         # en, zh
report_style = "professional"           # professional, casual
max_rounds = 15                         # Max tool calls per diagnosis

[ai.models.gpt4o]
base_url = "https://api.openai.com/v1"
api_key = "${OPENAI_API_KEY}"           # Environment variable
model = "gpt-4o"
max_tokens = 4000

[ai.models.deepseek]
base_url = "https://api.deepseek.com"
api_key = "${DEEPSEEK_API_KEY}"
model = "deepseek-chat"
```

### Notification Channels

```toml
[notify.console]
enabled = true

[notify.feishu]
enabled = true
webhook = "https://open.feishu.cn/open-apis/bot/v2/hook/xxx"
secret = "${FEISHU_SECRET}"

[notify.dingtalk]
enabled = true
webhook = "https://oapi.dingtalk.com/robot/send?access_token=xxx"
secret = "${DINGTALK_SECRET}"

[notify.webapi]
enabled = true
url = "https://your-service.example.com/api/events"
method = "POST"
[notify.webapi.headers]
Authorization = "Bearer ${API_TOKEN}"
```

### MCP Integration

```toml
[ai.mcp]
enabled = true

[[ai.mcp.servers]]
name = "prometheus"
command = "/usr/local/bin/mcp-prometheus"
[ai.mcp.servers.env]
PROMETHEUS_URL = "http://localhost:9090"

[[ai.mcp.servers]]
name = "jaeger"
command = "/usr/local/bin/mcp-jaeger"
```

---

## Plugins (31 Built-in)

### System Resources

| Plugin | Description |
|--------|-------------|
| `cpu` | CPU usage, load average |
| `mem` | Memory usage, swap |
| `disk` | Disk usage, inode |
| `diskio` | Disk I/O statistics |
| `uptime` | System uptime |
| `procnum` | Process count |
| `filefd` | File descriptor usage |
| `zombie` | Zombie process detection |

### Network

| Plugin | Description |
|--------|-------------|
| `net` | Network traffic |
| `netif` | Network interface status |
| `tcpstate` | TCP connection states |
| `sockstat` | Socket statistics |
| `dns` | DNS resolution monitoring |
| `ping` | ICMP latency |
| `conntrack` | Connection tracking table |
| `neigh` | ARP neighbor table |

### Services

| Plugin | Description |
|--------|-------------|
| `redis` | Redis health check |
| `docker` | Docker container metrics |
| `systemd` | Systemd services/timers |
| `http` | HTTP endpoint probing |
| `ntp` | NTP clock sync |

### Storage

| Plugin | Description |
|--------|-------------|
| `mount` | Mount point monitoring |
| `filecheck` | File existence check |

### Security

| Plugin | Description |
|--------|-------------|
| `cert` | SSL certificate expiration |
| `secmod` | Security module status |

### Observability

| Plugin | Description |
|--------|-------------|
| `logfile` | Log pattern matching |
| `journaltail` | Journal log monitoring |
| `exec` | Custom command execution |
| `scriptfilter` | Script output filtering |
| `hostident` | Host identification |

### Diagnostic Tools

| Plugin | Description |
|--------|-------------|
| `sysdiag` | 32+ on-demand diagnostic utilities |

---

## 5-Layer Diagnosis

DeepTrace analyzes every incident through five layers:

| Layer | Question | Example |
|-------|----------|---------|
| **1. Symptom** | What triggered the alert? | CPU usage 92% |
| **2. Direct Cause** | Which process is responsible? | myapp (PID 12345) |
| **3. Root Cause** | Why did it happen? | Large JSON serialization, missing pagination |
| **4. Impact** | What are the side effects? | API latency 50ms → 3200ms |
| **5. Prevention** | How to prevent recurrence? | Add pagination, limit response size |

---

## Stack Trace Support

| Language | Tool | Requirement |
|----------|------|-------------|
| Go | pprof | Program exposes pprof endpoint |
| Java | jstack | JDK installed on host |
| Python | py-spy | `pip install py-spy` |
| Node.js | llnode | llnode installed |

---

## Sample Report

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
                    CPU Alert - Production Server
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Host: prod-server-01 (192.168.1.10)
Time: 2026-03-05 14:32:15 UTC
Alert: CPU 92.3% (threshold ≥80%)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SUMMARY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

CPU spike caused by myapp process performing large JSON
serialization without pagination.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ROOT CAUSE ANALYSIS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Layer 1 - Symptom
  CPU usage 92.3%, exceeding threshold of 80%

Layer 2 - Direct Cause
  Process: myapp (PID: 12345)
  CPU: 89% of total
  Command: /opt/myapp/bin/server

Layer 3 - Root Cause
  Stack trace shows json.Marshal blocking at:
    handler.go:45 → processUsers()

  Analysis: /api/users/list returns all records
  without pagination, causing large JSON serialization.

Layer 4 - Impact
  API latency: 50ms → 3200ms (64x increase)
  Downstream timeout alerts triggered
  User complaints in last 15 minutes

Layer 5 - Prevention
  1. Add pagination to /api/users/list
  2. Limit max response size to 1MB
  3. Add request timeout handling

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
RECOMMENDED ACTIONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

For Operations:
  $ top -p 12345
  $ systemctl restart myapp    # If immediate relief needed

For Development:
  File: handler.go:45
  Fix: Add pagination parameters to list endpoint

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Diagnosis: 5/5 layers | Tools used: 12 | Duration: 3.2s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## Commands Reference

| Command | Description | Example |
|---------|-------------|---------|
| `run` | Start monitoring agent | `deeptrace run` |
| `chat` | Interactive AI troubleshooting | `deeptrace chat -v` |
| `inspect <plugin>` | Health check | `deeptrace inspect redis :6379` |
| `diagnose list` | List diagnosis records | `deeptrace diagnose list` |
| `diagnose show <id>` | View diagnosis detail | `deeptrace show abc123` |
| `selftest` | Tool smoke test | `deeptrace selftest` |
| `mcptest` | MCP connection test | `deeptrace mcptest` |

### Global Flags

| Flag | Description |
|------|-------------|
| `--configs <dir>` | Configuration directory (default: `conf.d`) |
| `--loglevel <level>` | Log level: debug, info, warn, error |
| `--model <name>` | Force specific AI model |
| `--version` | Show version |

---

## License

[MIT License](LICENSE)

---

<p align="center"><b>Stop guessing. Start tracing.</b></p>