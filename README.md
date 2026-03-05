# DeepTrace

> **AI-Powered Deep Diagnostic Engine for Production Systems**

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat)](LICENSE)

[English](README.md) | [中文](README_zh.md)

---

## What is DeepTrace?

DeepTrace goes beyond traditional monitoring alerts. When something breaks, it tells you **why** and **where** in your code.

```
Traditional:  "CPU usage is 92%"
DeepTrace:    "CPU 92% → myapp (PID 12345) → json.Marshal at handler.go:45
               → Processing 10MB JSON → API missing pagination"
```

## Core Features

- **5-Layer Deep Diagnosis** — Symptom → Direct Cause → Root Cause → Impact → Prevention
- **Code-Level Visibility** — Stack traces for Go/Java/Python/Node.js
- **31 Built-in Plugins** — CPU, Memory, Disk, Network, Redis, Docker, DNS, and more
- **AI-Powered Analysis** — Automatic root cause analysis with LLM integration
- **Multi-Channel Alerts** — DingTalk, PagerDuty, Webhook

## Quick Start

### Install

```bash
# Download binary
wget https://github.com/Oumu33/deeptrace/releases/latest/download/deeptrace-linux-amd64
chmod +x deeptrace-linux-amd64 && sudo mv deeptrace-linux-amd64 /usr/local/bin/deeptrace

# Or build from source
git clone https://github.com/Oumu33/deeptrace.git && cd deeptrace && go build
```

### Basic Usage

```bash
# Run monitoring with AI diagnosis
./deeptrace run

# Interactive troubleshooting
./deeptrace chat

# Health inspection
./deeptrace inspect cpu
./deeptrace inspect redis 10.0.0.1:6379
```

### Configuration

`conf.d/config.toml`:

```toml
[global]
interval = "30s"

[notify.console]
enabled = true

[ai]
enabled = true
model_priority = ["gpt4o"]

[ai.models.gpt4o]
base_url = "https://api.openai.com/v1"
api_key = "${OPENAI_API_KEY}"
model = "gpt-4o"
```

## 5-Layer Diagnosis

| Layer | Purpose | Example |
|-------|---------|---------|
| Symptom | What triggered | CPU 92% |
| Direct Cause | Which process | myapp (PID 12345) |
| Root Cause | Why it happened | Large JSON, no pagination |
| Impact | Side effects | API latency 50ms → 3200ms |
| Prevention | How to fix | Add pagination |

## Plugin Categories

| Category | Plugins |
|----------|---------|
| System | `cpu`, `mem`, `disk`, `diskio`, `uptime`, `procnum`, `filefd` |
| Network | `net`, `netif`, `tcpstate`, `sockstat`, `dns`, `ping`, `conntrack`, `neigh` |
| Storage | `mount`, `filecheck` |
| Service | `docker`, `systemd`, `redis`, `http`, `ntp` |
| Security | `secmod`, `cert` |
| Observability | `logfile`, `journaltail`, `exec`, `scriptfilter` |

Full list: [plugins/](plugins/)

## Stack Trace Support

| Language | Tool | Requirement |
|----------|------|-------------|
| Go | pprof | Program exposes pprof endpoint |
| Java | jstack | JDK installed |
| Python | py-spy | `pip install py-spy` |
| Node.js | llnode | llnode installed |

## Report Example

```
🚨 CPU Alert - Production Server

Host: prod-server-01 (192.168.1.10)
Current: CPU 92.3% (threshold ≥80%)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 Summary
CPU spike caused by myapp process

🔍 Root Cause Analysis

Layer 1 - Symptom
CPU 92.3%, exceeds threshold

Layer 2 - Direct Cause
Process: myapp (PID: 12345), CPU: 89%

Layer 3 - Root Cause
Stack trace shows json.Marshal blocking
Cause: API returns full dataset, missing pagination

Layer 4 - Impact
API latency: 50ms → 3200ms
Downstream timeout alerts triggered

Layer 5 - Prevention
Add pagination, limit response size

━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Actions

For Ops:
1. top -p 12345
2. systemctl restart myapp

For Devs:
1. Check /api/users/list endpoint
2. Add pagination parameters

📎 Code Location: handler.go:45

Diagnosis: 5/5 layers | Time: 3.2s
```

## MCP Integration

Connect external data sources for enhanced diagnosis:

```toml
[[ai.mcp.servers]]
name = "prometheus"
command = "/usr/local/bin/mcp-prometheus"
[ai.mcp.servers.env]
PROMETHEUS_URL = "http://localhost:9090"
```

## License

MIT License - see [LICENSE](LICENSE)

---

<p align="center"><b>Stop guessing. Start tracing.</b></p>
