# DeepTrace

> **AI 驱动的生产系统深度诊断引擎**

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat)](LICENSE)

[English](README.md) | [中文](README_zh.md)

---

## DeepTrace 是什么？

DeepTrace 超越传统监控告警。系统出问题时，它告诉你**为什么**以及**代码在哪里**。

```
传统告警:  "CPU 使用率 92%"
DeepTrace: "CPU 92% → myapp进程(PID 12345) → json.Marshal阻塞在 handler.go:45
           → 处理 10MB JSON → API 缺少分页参数"
```

## 核心特性

- **五层深度诊断** — 现象 → 直接原因 → 根本原因 → 关联影响 → 预防措施
- **代码级可见性** — 支持 Go/Java/Python/Node.js 堆栈追踪
- **31 个内置插件** — CPU、内存、磁盘、网络、Redis、Docker、DNS 等
- **AI 智能分析** — 集成 LLM 自动分析根因
- **多渠道告警** — 钉钉、PagerDuty、Webhook

## 快速开始

### 安装

```bash
# 下载二进制
wget https://github.com/Oumu33/deeptrace/releases/latest/download/deeptrace-linux-amd64
chmod +x deeptrace-linux-amd64 && sudo mv deeptrace-linux-amd64 /usr/local/bin/deeptrace

# 或从源码构建
git clone https://github.com/Oumu33/deeptrace.git && cd deeptrace && go build
```

### 基本用法

```bash
# 运行监控（启用 AI 诊断）
./deeptrace run

# 交互式排查
./deeptrace chat

# 健康巡检
./deeptrace inspect cpu
./deeptrace inspect redis 10.0.0.1:6379
```

### 配置

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

## 五层诊断

| 层级 | 目的 | 示例 |
|------|------|------|
| 现象层 | 告警触发原因 | CPU 92% |
| 直接原因 | 哪个进程导致 | myapp (PID 12345) |
| 根本原因 | 为什么发生 | 大 JSON，无分页 |
| 关联影响 | 有什么副作用 | API 延迟 50ms → 3200ms |
| 预防措施 | 如何解决 | 添加分页 |

## 插件分类

| 类别 | 插件 |
|------|------|
| 系统 | `cpu`, `mem`, `disk`, `diskio`, `uptime`, `procnum`, `filefd` |
| 网络 | `net`, `netif`, `tcpstate`, `sockstat`, `dns`, `ping`, `conntrack`, `neigh` |
| 存储 | `mount`, `filecheck` |
| 服务 | `docker`, `systemd`, `redis`, `http`, `ntp` |
| 安全 | `secmod`, `cert` |
| 可观测 | `logfile`, `journaltail`, `exec`, `scriptfilter` |

完整列表: [plugins/](plugins/)

## 堆栈追踪支持

| 语言 | 工具 | 要求 |
|------|------|------|
| Go | pprof | 程序暴露 pprof 端点 |
| Java | jstack | 安装 JDK |
| Python | py-spy | `pip install py-spy` |
| Node.js | llnode | 安装 llnode |

## 报告示例

```
🚨 CPU 告警 - 生产服务器

主机: prod-server-01 (192.168.1.10)
当前: CPU 92.3%（阈值 ≥80%）

━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 概述
CPU 飙升由 myapp 进程导致

🔍 根因分析

Layer 1 - 现象
CPU 92.3%，已超阈值

Layer 2 - 直接原因
进程: myapp (PID: 12345), CPU: 89%

Layer 3 - 根本原因
堆栈显示 json.Marshal 阻塞
原因: API 返回全量数据，缺少分页

Layer 4 - 关联影响
API 延迟: 50ms → 3200ms
下游触发超时告警

Layer 5 - 预防措施
添加分页，限制响应大小

━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ 操作建议

运维:
1. top -p 12345
2. systemctl restart myapp

开发:
1. 检查 /api/users/list 接口
2. 添加分页参数

📎 代码位置: handler.go:45

诊断深度: 5/5 层 | 耗时: 3.2s
```

## MCP 集成

连接外部数据源增强诊断能力：

```toml
[[ai.mcp.servers]]
name = "prometheus"
command = "/usr/local/bin/mcp-prometheus"
[ai.mcp.servers.env]
PROMETHEUS_URL = "http://localhost:9090"
```

## 许可证

MIT License - 详见 [LICENSE](LICENSE)

---

<p align="center"><b>停止猜测，开始追踪。</b></p>
