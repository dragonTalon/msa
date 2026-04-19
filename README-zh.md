# MSA - My Stock Agent

> [English](README.md) | 中文文档

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)

MSA 是一个轻量级、灵活的开源股票智能代理工具，为投资者和开发者设计。

目前已实现 AI 智能对话、股票数据查询、市场资讯搜索、模拟交易管理等核心功能，支持多种 LLM 提供商。未来计划扩展技术分析、策略回测等高级功能。采用模块化架构，MSA 降低了使用量化股票工具的门槛：既满足个人投资者对智能股票分析的需求，又为开发者提供了可扩展的开源框架。

## ✨ 功能特性

### 💬 AI 智能聊天
- 自然语言交互，支持股票查询和市场分析
- 集成多种 LLM 提供商（OpenAI、Claude、SiliconFlow 等）
- 流式输出，实时响应用户输入
- **动态 Skills 系统** - 模块化、可复用的领域知识提示模板
- **CLI 单次模式** - `msa -q "问题"` 无需进入 TUI 直接提问

### 📊 股票分析
- A 股及港股股票代码查询
- 公司信息获取
- K 线数据展示

### 🔍 智能搜索
- 多搜索引擎支持（Google、Bing 等）
- 自动降级机制，确保搜索可用性
- 网页内容抓取

### 💼 交易管理
- SQLite 本地数据库存储
- 账户管理（创建、查询、状态更新）
- 持仓查询与盈亏计算
- 交易记录管理
- 自动成交机制

### 🎯 精美 TUI 界面
- 基于 Bubble Tea 的终端界面
- Markdown 渲染支持
- 流畅的交互体验

## 🚀 快速开始

### 安装

#### 一键安装（推荐）

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/dragonTalon/msa/main/install.sh | sh
```

安装脚本会自动检测系统架构，下载最新版本并安装到 `~/.local/bin` 目录。

#### 手动安装

```bash
# 从 GitHub Releases 下载
# https://github.com/dragonTalon/msa/releases

# 或使用 go install 安装
go install github.com/dragonTalon/msa@latest

# 或克隆源码编译
git clone https://github.com/dragonTalon/msa.git
cd msa
go build
```

### 更新

```bash
# 检查是否有新版本
msa update --check

# 更新到最新版本
msa update
```

更新命令会自动从 GitHub Releases 下载最新版本并替换当前二进制文件。

### 配置

```bash
# 配置 API 密钥和模型
msa config
```

### 开始聊天

```bash
# TUI 多轮对话模式
msa

# CLI 单次对话模式
msa -q "腾讯的股票代码是什么？"

# 指定模型的单次对话
msa -q "分析一下茅台" -m "deepseek-r1"
```

## 📸 演示

> 截图和 GIF 即将添加...

## 📖 使用方式

### 聊天模式

```bash
# TUI 多轮对话（默认）
msa

# CLI 单次对话
msa -q "你的问题"
msa -q "你的问题" -m "模型名称"
```

在聊天模式下，你可以：
- 查询股票信息："帮我查一下腾讯的股票代码"
- 获取公司信息："腾讯控股的公司信息"
- 搜索市场资讯："搜索最新的 A 股市场新闻"
- 管理交易账户："创建一个初始资金 10 万元的账户"
- 查询持仓："查看我当前的所有持仓"
- 查询账户总览："显示我的账户总览和盈亏"
- 提交交易："买入 100 股贵州茅台，价格 1850 元"
- **访问记忆浏览器**：`/remember`

### 🧠 记忆系统 (NEW!)

MSA 现在包含强大的记忆系统，可以从你的对话中自动学习：

**功能特性：**
- 📝 **自动记录** - 所有对话自动保存
- 🧠 **AI 知识提取** - 从聊天中提取偏好、概念、策略
- 🤖 **智能记忆注入** - AI 使用你的记忆提供个性化回复
- 🔍 **记忆浏览器** - 查看历史、搜索对话和知识

**打开记忆浏览器：**
```
/remember
```

**记忆浏览器功能：**
1. **历史会话** - 浏览所有过去的对话
2. **知识库** - 查看提取的知识（用户画像、关注列表、概念、策略、问答）
3. **搜索** - 全文搜索所有会话和知识
4. **统计** - 查看使用统计

**恢复之前的会话：**
当你退出聊天会话时，MSA 会显示会话 ID，可以用于稍后恢复：
```
────────────────────────────────────────
会话已保存: abc-123-def-456
提示: 使用 "msa --resume abc-123-def-456" 恢复此会话
────────────────────────────────────────
```

要恢复之前的会话：
```bash
# 使用退出消息中的会话 ID
msa --resume abc-123-def-456

# 或使用简写形式
msa -r abc-123-def-456
```

恢复会话时：
- ✅ 历史消息加载到对话上下文
- ✅ AI 记住你之前的偏好和知识
- ✅ 你可以无缝继续对话
- ✅ 会话中提取的所有知识可用

**隐私与安全：**
- 所有数据本地存储在 `~/.msa/remember/`
- 自动过滤敏感信息（API 密钥、密码）
- 无云端同步 - 完全隐私保护

**禁用记忆系统：**
```bash
export MSA_MEMORY_ENABLED=false
msa chat
```

详细文档请参阅 [docs/memory-guide.md](docs/memory-guide.md)。

### 配置选项

MSA 支持多种配置方式，按优先级从高到低为：**命令行参数 > 环境变量 > 配置文件 > 默认值**

#### 交互式配置

```bash
# 启动 TUI 配置界面
msa config
```

在 TUI 配置界面中，你可以：
- 使用方向键选择配置项
- 按 `Enter` 进入编辑模式
- 选择 Provider 时自动填充 Base URL
- 按 `S` 保存配置
- 按 `R` 重置为默认值
- 按 `Q` 退出

#### 环境变量配置

```bash
# 设置 Provider
export MSA_PROVIDER=siliconflow

# 设置 API Key
export MSA_API_KEY=sk-xxxxxxxxxxxx

# 设置 Base URL（可选）
export MSA_BASE_URL=https://api.example.com/v1

# 设置日志级别（可选）
export MSA_LOG_LEVEL=debug

# 设置日志文件路径（可选）
export MSA_LOG_FILE=/path/to/msa.log
```

#### 命令行参数配置

```bash
# 恢复之前的会话
msa --resume <session-id>

# 使用配置文件
msa --config /path/to/config.json chat

# 使用 key=value 格式
msa --config apikey=sk-xxx --config loglevel=debug chat

# 混合使用
msa --config /path/to/config.json --config apikey=sk-xxx chat
```

**可用命令行参数：**

| 参数 | 简写 | 描述 | 示例 |
|------|------|------|------|
| `--question` | `-q` | 单次对话问题（不进入 TUI） | `msa -q "腾讯股票代码？"` |
| `--model` | `-m` | 本次运行指定模型 | `-m "deepseek-r1"` |
| `--resume` | - | 通过 ID 恢复之前的会话 | `--resume 2026-01-01_uuid` |
| `--config` | - | 设置配置（文件或 key=value） | `--config apikey=sk-xxx` |

#### TUI 中查看配置

在聊天界面中，可以使用以下命令：

```bash
/config
```

这会显示当前配置信息，包括 Provider、Model、Base URL、API Key（部分隐藏）、日志级别和日志文件路径。

#### 配置文件位置

配置文件保存在 `~/.msa/msa_config.json`，包含：
- Provider: LLM 提供商
- Model: 使用的模型
- Base URL: API 基础 URL
- API Key: API 密钥
- LogConfig: 日志配置（级别、格式、输出、文件路径）

### 配置安全注意事项

⚠️ **重要**：API Key 以明文形式存储在配置文件中。

为确保安全，建议：

```bash
# 设置配置文件权限，仅当前用户可读写
chmod 600 ~/.msa/msa_config.json

# 确保配置目录权限正确
chmod 700 ~/.msa/
```

**安全建议**：
- 不要将配置文件提交到版本控制系统
- 不要在共享环境中使用包含 API Key 的配置
- 定期轮换 API Key
- 使用不同的 API Key 用于不同环境

### Provider 扩展指南

MSA 支持扩展新的 LLM Provider。以下是添加步骤：

#### 1. 在 ProviderRegistry 中注册

编辑 `pkg/model/comment.go`，在 `ProviderRegistry` 中添加新的 Provider：

```go
var ProviderRegistry = map[LlmProvider]ProviderInfo{
    Siliconflow: {
        ID:             Siliconflow,
        DisplayName:    "SiliconFlow (硅基流动)",
        Description:    "国内 LLM API 提供商，兼容 OpenAI 格式",
        DefaultBaseURL: "https://api.siliconflow.cn/v1",
        KeyPrefix:      "sk-",
    },
    // 添加新的 Provider
    YourProvider: {
        ID:             YourProvider,
        DisplayName:    "Your Provider Name",
        Description:    "Provider description",
        DefaultBaseURL: "https://api.yourprovider.com/v1",
        KeyPrefix:      "custom-",
    },
}
```

#### 2. 添加 Provider 常量

在 `pkg/model/comment.go` 中添加 Provider 常量：

```go
const (
    Siliconflow LlmProvider = "siliconflow"
    YourProvider LlmProvider = "yourprovider"
)
```

#### 3. 更新 GetDisplayName() 和 GetDefaultBaseURL() 方法（如需要）

如果新 Provider 有特殊的显示名称或默认 URL 逻辑，可以在相应方法中添加特殊处理。

#### 4. 验证配置

```bash
# 重新编译
go build

# 测试配置
msa config

# 选择新 Provider 并验证配置保存
```

## 🗄️ 本地数据库

MSA 使用 SQLite 本地数据库存储：
- **账户**：用户账户及余额追踪
- **交易记录**：买卖订单及状态追踪
- **持仓**：持仓计算和盈亏统计

### 数据库位置

数据库文件存储在：
```
~/.msa/msa.sqlite
```

### 数据备份

备份数据：
```bash
cp ~/.msa/msa.sqlite ~/.msa/msa.sqlite.backup.$(date +%Y%m%d)
```

从备份恢复：
```bash
cp ~/.msa/msa.sqlite.backup.YYYYMMDD ~/.msa/msa.sqlite
```

### 金额单位

**重要**：所有金额字段以"毫"为单位存储（1元 = 10000毫），避免浮点数精度问题。
- `10000` = 1.00 元
- 显示转换：`amount / 10000 = 显示值`

## 🧪 测试

项目包含单元测试，覆盖核心业务逻辑和工具函数。

### 运行测试

```bash
# 运行所有测试
go test ./pkg/...

# 运行测试并显示覆盖率
go test -cover ./pkg/...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out
```

### 当前覆盖率

| 模块 | 覆盖率 |
|------|--------|
| pkg/utils | 83.7% |
| pkg/logic/tools | 100.0% |
| pkg/logic/provider | 76.9% |
| pkg/config | 53.4% |
| pkg/logic/command | 51.1% |

## 🏗️ 架构设计

```
msa/ (项目根目录)
├── go.mod                    # Go 模块依赖配置
├── go.sum                    # Go 模块依赖校验
├── main.go                   # 项目入口文件，初始化 Cobra CLI
├── cmd/                      # Cobra CLI 命令路由层
│   └── root.go              # 根命令定义，仅负责路由
└── pkg/                      # 业务实现层
    ├── app/                 # 应用启动（TUI 模式入口）
    ├── extcli/              # CLI 单次对话模式（msa -q）
    ├── core/                # 核心 pipeline（无 UI 依赖）
    │   ├── event/           # Event 类型系统（pipeline 数据流）
    │   ├── logger/          # 结构化日志，requestID 注入
    │   ├── agent/           # Agent：Run() 返回 <-chan Event
    │   └── runner/          # Runner：协调 agent → renderer
    ├── renderer/            # 输出适配层
    │   ├── renderer.go      # Renderer 接口
    │   ├── cli.go           # CLI 实现：写 stdout
    │   └── tui.go           # TUI 实现：转发给 Bubble Tea
    ├── db/                  # 数据库模块（SQLite + GORM）
    │   ├── db.go            # 数据库连接与初始化
    │   ├── global.go        # 全局数据库管理
    │   ├── migrate.go       # 数据库迁移
    │   ├── account.go       # 账户数据访问
    │   └── transaction.go   # 交易数据访问
    ├── model/               # 数据模型
    ├── session/             # 会话持久化与恢复
    ├── tui/                 # 终端界面（Bubble Tea，纯 UI）
    │   ├── style/           # UI 样式
    │   ├── config/          # 配置 TUI
    │   └── chat.go          # 聊天界面
    ├── config/              # 配置管理模块
    ├── logic/               # 业务逻辑
    │   ├── command/         # 命令处理
    │   ├── provider/        # LLM 提供商
    │   ├── skills/          # 动态 Skills 系统
    │   ├── finsvc/          # 金融服务（交易、持仓）
    │   └── tools/           # 工具（股票、搜索、金融等）
    └── utils/               # 工具函数
```

### 核心数据流

```
用户输入
  ↓
Runner.Ask(ctx, input)
  ↓
Agent.Run(ctx, messages) → <-chan Event
  ↓  StreamAdapter 处理 Eino stream
  ↓  ReAct 循环：LLM → tool → LLM → ...
  ↓  发射结构化 Event
Runner 消费 eventCh
  ↓
Renderer.Handle(ctx, event)
  ↓
CLI: fmt.Print          TUI: bubbletea msg
```

### 架构原则

- **cmd/** - 仅存放命令路由，代码精简
- **pkg/core/** - 业务核心，零 UI 依赖，可独立测试
  - **event/** - 类型化事件系统，替代全局广播
  - **agent/** - Eino 封装，干净的 stream 处理（替代 toolCallChecker）
  - **runner/** - CLI 和 TUI 共用的编排层
- **pkg/renderer/** - UI 适配层，吸收 CLI/TUI 差异
- **pkg/tui/** - 纯 UI 组件，无业务逻辑
- **pkg/logic/tools/** - 工具实现，全部接收 `ctx`

## 🗺️ 迭代规划

### Phase 1: 基础架构 (v0.1.x)
- [x] CLI 框架搭建（Cobra）
- [x] TUI 终端界面（Bubble Tea）
- [x] 项目结构重构
- [x] 配置管理模块
- [x] 日志系统
- [x] 单元测试框架
- [ ] CI/CD 流水线

### Phase 2: 核心功能 (v0.2.x)
- [x] **数据模块**
  - [x] 股票数据 API 接入
  - [ ] 实时行情订阅
  - [x] 历史数据存储（SQLite）
  - [x] 账户管理
  - [x] 交易记录
  - [x] 持仓跟踪
  - [ ] 数据缓存层（Redis）
- [ ] **分析模块**
  - [ ] 技术指标计算（MA、MACD、RSI、KDJ）
  - [ ] K 线形态识别
  - [ ] 趋势分析
- [ ] **CLI 命令扩展**
  - [ ] `msa quote <symbol>` - 获取股票报价
  - [ ] `msa history <symbol>` - 历史数据查询
  - [ ] `msa analyze <symbol>` - 技术分析

### Phase 3: 智能化 (v0.3.x)
- [x] **AI/大模型集成**
  - [x] 大模型 API 接入
  - [x] 自然语言股票查询
  - [x] AI 驱动市场分析
  - [x] 智能问答助手
- [ ] **策略模块**
  - [ ] 策略 DSL 定义
  - [ ] 回测引擎
  - [ ] 绩效指标统计
  - [ ] 策略模板库

### Phase 4: 高级特性 (v0.4.x)
- [ ] **通知系统**
  - [ ] 价格预警
  - [ ] 策略信号通知
  - [ ] 多渠道推送（Email/Webhook/Telegram）
- [ ] **组合管理**
  - [ ] 持仓跟踪
  - [ ] 盈亏计算
  - [ ] 风险评估
- [ ] **插件系统**
  - [ ] 插件架构设计
  - [ ] 自定义数据源插件
  - [ ] 自定义策略插件

### 远景规划
- [ ] Web 控制台（可选）
- [ ] 移动端辅助应用
- [ ] 社区策略市场
- [ ] 多市场支持（美股/港股/A 股）

## 🛠️ 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.21+ |
| CLI 框架 | [Cobra](https://github.com/spf13/cobra) |
| TUI 框架 | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| UI 样式 | [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| 日志 | [Logrus](https://github.com/sirupsen/logrus) |
| ORM | [GORM](https://github.com/go-gorm/gorm) |
| 数据库 | SQLite ([github.com/glebarez/sqlite](https://github.com/glebarez/sqlite), 纯 Go 实现) |
| 数据存储 | ~/.msa/msa.sqlite |
| 缓存 | Redis（计划中） |
| AI/LLM | OpenAI / Claude / SiliconFlow / Ollama |

## 🤝 贡献指南

欢迎贡献代码！请随时提交 Issue 和 Pull Request。

### 开发流程

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

- [Cobra](https://github.com/spf13/cobra) - 强大的 CLI 框架
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - 精美的 TUI 框架
- [Cloudwego Eino](https://github.com/cloudwego/eino) - AI 应用开发框架
