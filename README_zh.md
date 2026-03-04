# DeepTrace

> **AI 驱动的生产系统深度诊断引擎**

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat)](LICENSE)

---

## DeepTrace 是什么？

DeepTrace 是一个智能诊断引擎，超越传统监控告警。当系统出问题时，它不只告诉你**发生了什么**，还会告诉你**为什么**以及**代码在哪里**。

```
传统告警:  "CPU 使用率 92%"
DeepTrace: "CPU 92% → myapp进程(PID 12345) → json.Marshal阻塞在 handler.go:45
           → 正在处理 10MB JSON 响应 → API 缺少分页参数"
```

## 🚀 核心特性

### 🧠 五层深度诊断

| 层级 | 发现什么 | 示例 |
|------|---------|------|
| **1. 现象层** | 告警触发原因 | CPU 92%，阈值 80% |
| **2. 直接原因** | 哪个进程/组件 | myapp (PID 12345)，CPU 89% |
| **3. 根本原因** | 为什么发生 | 大 JSON 序列化，无分页 |
| **4. 关联影响** | 还影响了什么 | API 延迟 50ms → 3200ms |
| **5. 预防措施** | 如何避免复发 | 添加分页，限制响应大小 |

### 🔧 79+ 内置诊断工具

**系统层**：CPU、内存、磁盘、网络、进程、文件描述符

**网络层**：连接状态、TCP 状态、重传率、延迟、DNS、防火墙

**存储层**：I/O 延迟、块设备、LVM、挂载点

**安全层**：SELinux、AppArmor、审计日志、连接跟踪

**应用层**：Go/Java/Python/Node.js 堆栈追踪

### 🎯 代码级可见性

```go
// DeepTrace 可以捕获堆栈，精确定位代码卡在哪里
goroutine 1 [running]:
encoding/json.Marshal()
    /usr/local/go/src/encoding/json/encode.go:161
main.handleRequest()
    /opt/myapp/handler.go:45  ← 问题在这里
main.main()
    /opt/myapp/main.go:23
```

### 💬 双模式报告

**运维视角** — 快速摘要 + 可执行命令：
- 问题：myapp CPU 89%
- 操作：`top -p 12345` 或 `systemctl restart myapp`

**开发视角** — 堆栈追踪 + 代码位置 + 根因：
- 文件：`handler.go:45`
- 原因：大 JSON 序列化，缺少分页

### 📡 多渠道通知

- 钉钉 — 原生 Markdown 支持
- 飞书 — 通过 FlashDuty 集成
- 企业微信 — 通过 FlashDuty 集成
- PagerDuty — 国际化团队
- 通用 Webhook — 任意 HTTP 端点

## 📦 安装

```bash
# 从 Release 下载
wget https://github.com/Oumu33/deeptrace/releases/latest/download/deeptrace-linux-amd64
chmod +x deeptrace-linux-amd64
sudo mv deeptrace-linux-amd64 /usr/local/bin/deeptrace

# 或从源码构建
git clone https://github.com/Oumu33/deeptrace.git
cd deeptrace
go build -o deeptrace .
```

## ⚡ 快速开始

### 1. 基础监控

创建 `conf.d/config.toml`：

```toml
[global]
interval = "30s"

[notify.console]
enabled = true
```

创建 CPU 监控 `conf.d/p.cpu/cpu.toml`：

```toml
[[instances]]
targets = ["localhost"]

[[instances.alerts]]
check = "cpu_usage"
warn_ge = 80
crit_ge = 95
```

运行：

```bash
./deeptrace run
```

### 2. 启用 AI 诊断

添加到 `conf.d/config.toml`：

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

告警触发时，DeepTrace 自动：
1. 使用 79+ 工具收集诊断数据
2. AI 分析根因
3. 生成五层深度诊断报告
4. 推送到你的通知渠道

### 3. 交互式排查

```bash
./deeptrace chat
```

随便问：
- "为什么 CPU 这么高？"
- "检查 Redis 内存使用，找大 key"
- "显示到 10.0.0.1 的网络连接"

### 4. 主动健康巡检

```bash
# 巡检 Redis
./deeptrace inspect redis 10.0.0.1:6379

# 巡检本地系统
./deeptrace inspect cpu
./deeptrace inspect mem
./deeptrace inspect disk
```

## 🎨 报告示例

CPU 告警触发时，钉钉收到的消息：

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

## 🛠️ 支持堆栈追踪的语言

| 语言 | 工具 | 要求 |
|------|------|------|
| Go | pprof 端点 | 程序暴露 pprof |
| Java | jstack | 安装 JDK |
| Python | py-spy | `pip install py-spy` |
| Node.js | llnode/lldb | 安装 llnode |
| 通用 | gdb | 安装 gdb + 调试符号 |

## 🤝 参与贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE)

---

<p align="center">
  <b>停止猜测，开始追踪。</b><br>
  <sub>为讨厌凌晨 3 点调试的运维和开发而构建</sub>
</p>