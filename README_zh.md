<p align="center">
  <img src="https://img.shields.io/badge/DeepTrace-AI%20深度诊断-blue?style=for-the-badge" alt="DeepTrace"/>
</p>

<h1 align="center">DeepTrace</h1>

<p align="center">
  <strong>AI 驱动的生产系统深度诊断引擎</strong>
</p>

<p align="center">
  <a href="#核心特性">核心特性</a> •
  <a href="#架构">架构</a> •
  <a href="#安装">安装</a> •
  <a href="#快速开始">快速开始</a> •
  <a href="#配置">配置</a>
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

## 为什么选择 DeepTrace？

传统监控告诉你**发生了什么**，DeepTrace 告诉你**为什么**以及**在哪里**。

```
传统告警:    "CPU 使用率 92%"
            → 你需要 SSH 登录、运行 top、排查日志...

DeepTrace:  "CPU 92% → myapp进程 (PID 12345)
             → json.Marshal 阻塞在 handler.go:45
             → 处理 10MB JSON 响应
             → API 接口缺少分页参数
             → 建议添加分页，限制响应大小"
            → 根因自动定位
```

## 核心特性

### 诊断能力

| 特性 | 说明 |
|------|------|
| **五层深度诊断** | 现象 → 直接原因 → 根本原因 → 关联影响 → 预防措施 |
| **代码级可见性** | 支持 Go/Java/Python/Node.js 堆栈追踪 |
| **AI Agent** | 自主调用诊断工具进行根因分析 |
| **交互式 Chat** | 通过自然语言进行故障排查 |
| **MCP 集成** | 连接 Prometheus、Jaeger、CMDB 丰富诊断上下文 |

### 内置可观测性

- **31 个监控插件** — 覆盖系统、网络、存储、服务、安全
- **32 个诊断工具** — 按需深度检查工具集
- **多渠道告警** — 飞书、钉钉、PagerDuty、Webhook

### AI 集成

- **OpenAI 兼容 API** — 支持 GPT-4o、DeepSeek、Claude、Ollama
- **自动故障转移** — 限流或服务端错误时自动切换模型
- **Token 预算控制** — 每日限额和费用估算
- **多语言报告** — 支持中英文输出

---

## 架构

```
┌─────────────────────────────────────────────────────────────────┐
│                          DeepTrace CLI                           │
├─────────────┬─────────────┬─────────────┬───────────────────────┤
│   run       │   chat      │   inspect   │   diagnose            │
│   监控代理   │   AI 交互   │   健康检查   │   诊断记录            │
└──────┬──────┴─────────────┴──────┬──────┴───────────────────────┘
       │                           │
       ▼                           ▼
┌──────────────┐           ┌──────────────┐
│   Plugins    │           │   AI Engine  │
│   插件(31个)  │           │   AI 引擎    │
│              │           │              │
│  cpu, mem    │           │  ┌────────┐  │
│  disk, net   │           │  │LLM API │  │
│  redis, ...  │           │  └────┬───┘  │
└──────┬───────┘           │       │      │
       │                   │  ┌────▼───┐  │
       ▼                   │  │ MCP    │  │
┌──────────────┐           │  │ 工具   │  │
│   Engine     │           │  └────────┘  │
│   引擎       │           └──────┬───────┘
│              │                  │
│  阈值检测    │                  │
│  事件聚合    │                  │
└──────┬───────┘                  │
       │                          │
       ▼                          ▼
┌─────────────────────────────────────────┐
│              Notify Channels             │
│         Console │ 飞书 │ 钉钉 │ Webhook  │
└─────────────────────────────────────────┘
```

### 核心组件

| 组件 | 路径 | 职责 |
|------|------|------|
| **CLI** | `main.go` | 命令路由和参数解析 |
| **Agent** | `agent/` | 插件生命周期管理、配置加载 |
| **Engine** | `engine/` | 事件处理、阈值检测 |
| **Diagnose** | `diagnose/` | AI 驱动的根因分析 |
| **Chat** | `chat/` | 交互式故障排查 REPL |
| **MCP** | `mcp/` | 外部数据源集成 |
| **Notify** | `notify/` | 多渠道告警分发 |
| **Plugins** | `plugins/` | 31 个内置监控采集器 |

---

## 安装

### 二进制下载

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

### 从源码构建

```bash
git clone https://github.com/Oumu33/deeptrace.git
cd deeptrace
go build -o deeptrace .
```

### 环境要求

- Go 1.21+ (构建需要)
- Linux (推荐) / Windows / macOS

---

## 快速开始

### 1. 基础监控

```bash
# 创建配置目录
mkdir -p conf.d

# 使用默认插件启动
./deeptrace run
```

### 2. 交互式排查

```bash
# 启动 AI 交互会话
./deeptrace chat

> 服务器为什么变慢了？
> 检查 Redis 连接 10.0.0.1:6379
> 分析 5 分钟前的 CPU 飙升
```

### 3. 健康巡检

```bash
# 检查指定组件
./deeptrace inspect cpu
./deeptrace inspect redis 10.0.0.1:6379
./deeptrace inspect http https://api.example.com/health
```

### 4. 诊断记录

```bash
# 列出所有诊断记录
./deeptrace diagnose list

# 查看指定诊断详情
./deeptrace diagnose show <record-id>
```

---

## 配置

### 主配置 (`conf.d/config.toml`)

```toml
[global]
interval = "30s"                    # 采集间隔

[global.labels]
env = "production"
region = "cn-east-1"

[log]
level = "info"                      # debug, info, warn, error
format = "json"                     # json, text
output = "stdout"                   # stdout, file
```

### AI 配置

```toml
[ai]
enabled = true
model_priority = ["gpt4o", "deepseek"]  # 故障转移顺序
language = "zh"                         # zh, en
report_style = "professional"           # professional, casual
max_rounds = 15                         # 单次诊断最大工具调用轮次

[ai.models.gpt4o]
base_url = "https://api.openai.com/v1"
api_key = "${OPENAI_API_KEY}"           # 支持环境变量
model = "gpt-4o"
max_tokens = 4000

[ai.models.deepseek]
base_url = "https://api.deepseek.com"
api_key = "${DEEPSEEK_API_KEY}"
model = "deepseek-chat"
```

### 通知渠道

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

### MCP 集成

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

## 插件列表 (31 个内置)

### 系统资源

| 插件 | 说明 |
|------|------|
| `cpu` | CPU 使用率、负载 |
| `mem` | 内存使用率、Swap |
| `disk` | 磁盘使用率、inode |
| `diskio` | 磁盘 I/O 统计 |
| `uptime` | 系统运行时间 |
| `procnum` | 进程数量 |
| `filefd` | 文件描述符使用 |
| `zombie` | 僵尸进程检测 |

### 网络

| 插件 | 说明 |
|------|------|
| `net` | 网络流量 |
| `netif` | 网络接口状态 |
| `tcpstate` | TCP 连接状态 |
| `sockstat` | Socket 统计 |
| `dns` | DNS 解析监控 |
| `ping` | ICMP 延迟检测 |
| `conntrack` | 连接跟踪表 |
| `neigh` | ARP 邻居表 |

### 服务

| 插件 | 说明 |
|------|------|
| `redis` | Redis 健康检查 |
| `docker` | Docker 容器监控 |
| `systemd` | Systemd 服务/定时器 |
| `http` | HTTP 端点探测 |
| `ntp` | NTP 时钟同步 |

### 存储

| 插件 | 说明 |
|------|------|
| `mount` | 挂载点监控 |
| `filecheck` | 文件存在性检查 |

### 安全

| 插件 | 说明 |
|------|------|
| `cert` | SSL 证书过期检测 |
| `secmod` | 安全模块状态 |

### 可观测性

| 插件 | 说明 |
|------|------|
| `logfile` | 日志模式匹配 |
| `journaltail` | Journal 日志监控 |
| `exec` | 自定义命令执行 |
| `scriptfilter` | 脚本输出过滤 |
| `hostident` | 主机标识 |

### 诊断工具

| 插件 | 说明 |
|------|------|
| `sysdiag` | 32+ 按需诊断工具集 |

---

## 五层诊断

DeepTrace 对每个事件进行五层深度分析：

| 层级 | 核心问题 | 示例 |
|------|----------|------|
| **1. 现象层** | 告警触发原因？ | CPU 使用率 92% |
| **2. 直接原因** | 哪个进程导致？ | myapp (PID 12345) |
| **3. 根本原因** | 为什么发生？ | 大 JSON 序列化，缺少分页 |
| **4. 关联影响** | 有什么副作用？ | API 延迟 50ms → 3200ms |
| **5. 预防措施** | 如何避免复发？ | 添加分页，限制响应大小 |

---

## 堆栈追踪支持

| 语言 | 工具 | 要求 |
|------|------|------|
| Go | pprof | 程序暴露 pprof 端点 |
| Java | jstack | 主机安装 JDK |
| Python | py-spy | `pip install py-spy` |
| Node.js | llnode | 安装 llnode |

---

## 报告示例

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
                    CPU 告警 - 生产服务器
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

主机: prod-server-01 (192.168.1.10)
时间: 2026-03-05 14:32:15 CST
告警: CPU 92.3% (阈值 ≥80%)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
概述
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

CPU 飙升由 myapp 进程执行大 JSON 序列化导致，
接口缺少分页机制。

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
根因分析
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

第一层 - 现象
  CPU 使用率 92.3%，超过阈值 80%

第二层 - 直接原因
  进程: myapp (PID: 12345)
  CPU: 占总使用率 89%
  命令: /opt/myapp/bin/server

第三层 - 根本原因
  堆栈追踪显示 json.Marshal 阻塞于:
    handler.go:45 → processUsers()

  分析: /api/users/list 接口返回全量数据，
  无分页机制，导致大 JSON 序列化。

第四层 - 关联影响
  API 延迟: 50ms → 3200ms (增长 64 倍)
  下游服务触发超时告警
  近 15 分钟用户投诉增加

第五层 - 预防措施
  1. 为 /api/users/list 添加分页参数
  2. 限制最大响应体 1MB
  3. 增加请求超时处理

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
操作建议
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

运维操作:
  $ top -p 12345
  $ systemctl restart myapp    # 紧急缓解

开发修复:
  文件: handler.go:45
  方案: 为列表接口添加分页参数

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
诊断深度: 5/5 层 | 工具调用: 12 次 | 耗时: 3.2s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## 命令参考

| 命令 | 说明 | 示例 |
|------|------|------|
| `run` | 启动监控代理 | `deeptrace run` |
| `chat` | 交互式 AI 排查 | `deeptrace chat -v` |
| `inspect <plugin>` | 健康检查 | `deeptrace inspect redis :6379` |
| `diagnose list` | 列出诊断记录 | `deeptrace diagnose list` |
| `diagnose show <id>` | 查看诊断详情 | `deeptrace show abc123` |
| `selftest` | 工具冒烟测试 | `deeptrace selftest` |
| `mcptest` | MCP 连接测试 | `deeptrace mcptest` |

### 全局参数

| 参数 | 说明 |
|------|------|
| `--configs <dir>` | 配置目录 (默认: `conf.d`) |
| `--loglevel <level>` | 日志级别: debug, info, warn, error |
| `--model <name>` | 强制使用指定 AI 模型 |
| `--version` | 显示版本号 |

---

## 许可证

[MIT License](LICENSE)

---

<p align="center"><b>停止猜测，开始追踪。</b></p>