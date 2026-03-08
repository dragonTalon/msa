# MSA - My Stock Agent

> [English](README.md) | 中文文档

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)

MSA 是一个轻量级、灵活的开源股票智能代理工具，为投资者和开发者设计。

它集成了股票数据采集、多维市场分析、策略回测和自动交易辅助等核心能力，支持自定义策略配置和二次开发。采用模块化架构，MSA 降低了使用量化股票工具的门槛：既满足个人投资者对自动交易的需求，又为开发者提供了可扩展的开源框架。

## ✨ 功能特性

### 💬 AI 智能聊天
- 自然语言交互，支持股票查询和市场分析
- 集成多种 LLM 提供商（OpenAI、Claude、SiliconFlow 等）
- 流式输出，实时响应用户输入

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

```bash
# 使用 go install 安装
go install github.com/yourusername/msa@latest

# 或克隆源码
git clone https://github.com/yourusername/msa.git
cd msa
go build
```

### 配置

```bash
# 配置 API 密钥
msa config

# 选择模型
msa chat
```

### 开始聊天

```bash
msa chat
```

## 📸 演示

> 截图和 GIF 即将添加...

## 📖 使用方式

### 聊天模式

```bash
msa chat
```

在聊天模式下，你可以：
- 查询股票信息："帮我查一下腾讯的股票代码"
- 获取公司信息："腾讯控股的公司信息"
- 搜索市场资讯："搜索最新的 A 股市场新闻"
- 管理交易账户："创建一个初始资金 10 万元的账户"
- 查询持仓："查看我当前的所有持仓"
- 查询账户总览："显示我的账户总览和盈亏"
- 提交交易："买入 100 股贵州茅台，价格 1850 元"

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
# 使用配置文件
msa --config /path/to/config.json chat

# 使用 key=value 格式
msa --config apikey=sk-xxx --config loglevel=debug chat

# 混合使用
msa --config /path/to/config.json --config apikey=sk-xxx chat
```

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
    ├── app/                 # 应用核心模块
    │   └── app.go           # 应用启动
    ├── db/                  # 数据库模块
    │   ├── db.go            # 数据库连接与初始化
    │   ├── global.go        # 全局数据库管理
    │   ├── account.go       # 账户数据访问
    │   └── transaction.go   # 交易数据访问
    ├── model/               # 数据模型
    ├── service/             # 业务服务层
    ├── tui/                 # 终端界面模块
    │   ├── style/           # UI 样式
    │   ├── chat.go          # 聊天界面
    │   └── model_selector.go # 模型选择器
    ├── config/              # 配置管理模块
    │   ├── local_config.go  # 本地存储配置
    │   └── logger.go        # 日志配置
    ├── logic/               # 业务逻辑
    │   ├── agent/           # AI 代理
    │   ├── command/         # 命令处理
    │   ├── provider/        # LLM 提供商
    │   └── tools/           # 工具（股票、搜索、交易等）
    └── utils/               # 工具函数
        ├── file.go
        ├── http.go
        └── format.go
```

### 架构原则

- **cmd/** - 仅存放命令路由，代码精简
- **pkg/** - 存放所有具体实现，职责清晰
  - **app/** - 应用生命周期管理
  - **tui/** - 终端用户界面（Bubble Tea）
  - **config/** - 配置与日志
  - **logic/** - 核心业务逻辑
  - **utils/** - 通用工具函数

## 🗺️ 迭代规划

### Phase 1: 基础架构 (v0.1.x)
- [x] CLI 框架搭建（Cobra）
- [x] TUI 终端界面（Bubble Tea）
- [x] 项目结构重构
- [x] 配置管理模块
- [x] 日志系统
- [ ] 单元测试框架
- [ ] CI/CD 流水线

### Phase 2: 核心功能 (v0.2.x)
- [ ] **数据模块**
  - [ ] 股票数据 API 接入
  - [ ] 实时行情订阅
  - [ ] 历史数据存储（SQLite/PostgreSQL）
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
- [ ] **AI/大模型集成**
  - [ ] 大模型 API 接入
  - [ ] 自然语言股票查询
  - [ ] AI 驱动市场分析
  - [ ] 智能问答助手
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
| 数据库 | SQLite (github.com/glebarez/sqlite, 纯 Go 实现) |
| 数据存储 | ~/.msa/msa.db |
| 缓存 | Redis（计划中）|
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
