# MSA - My Stock Agent

> [English](README.md) | 中文文档

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)

MSA（My Stock Agent）是一个轻量级、灵活的开源股票智能代理工具，为投资者和开发者设计。

目前支持 AI 智能对话、股票数据查询、市场资讯搜索、模拟交易管理等核心功能，兼容多种 LLM 提供商。采用模块化架构，MSA 降低了量化股票工具的使用门槛：既满足个人投资者对智能股票分析的需求，又为开发者提供了可扩展的开源框架。

## ✨ 功能特性

- **AI 智能聊天** - 自然语言交互，支持股票查询和市场分析，流式输出，集成多种 LLM 提供商（OpenAI、Claude、SiliconFlow 等）
- **动态 Skills 系统** - 模块化、可复用的领域知识提示模板，可根据对话上下文自动选用
- **股票分析** - A 股及港股代码查询、公司信息获取、K 线数据展示
- **智能搜索** - 多搜索引擎支持（Google、Bing 等），自动降级确保搜索可用性
- **交易管理** - 基于 SQLite 的账户、持仓、交易记录管理，支持盈亏计算
- **记忆系统** - 自动记录会话、AI 知识提取、智能记忆注入，提供个性化回复
- **精美 TUI** - 基于 Bubble Tea 的终端界面，支持 Markdown 渲染

## 📸 演示

> 截图和 GIF 即将添加...

## 🚀 快速开始

### 安装

```bash
# 一键安装（macOS / Linux）
curl -fsSL https://raw.githubusercontent.com/dragonTalon/msa/main/install.sh | sh

# 或使用 go install
go install github.com/dragonTalon/msa@latest

# 或克隆源码编译
git clone https://github.com/dragonTalon/msa.git
cd msa
go build
```

### 更新

```bash
msa update          # 更新到最新版本
msa update --check  # 仅检查更新
```

### 配置

```bash
msa config    # TUI 交互式配置
```

或使用环境变量：

```bash
export MSA_PROVIDER=siliconflow
export MSA_API_KEY=sk-xxxxxxxxxxxx
```

### 开始聊天

```bash
msa                                    # TUI 多轮对话模式
msa -q "腾讯的股票代码是什么？"         # CLI 单次对话
msa -r <session-id>                    # 恢复之前的会话
```

更多用法请参见 `msa --help`。

## 📖 文档

| 文档 | 说明 |
|------|------|
| [Skills 系统](docs/skills.md) | 管理和创建自定义技能 |
| [记忆系统](docs/memory-guide.md) | 会话记录、知识提取、隐私设置 |
| [配置指南](docs/configuration.md) | 配置方式、安全建议、环境变量 |
| [Provider 扩展](docs/provider-extension.md) | 添加新的 LLM 提供商 |
| [数据库](docs/database.md) | SQLite 存储、备份、数据模型 |

## 🗄️ 本地数据库

MSA 使用 SQLite 存储账户、持仓和交易记录。数据库文件位于 `~/.msa/msa.sqlite`。

备份数据：
```bash
cp ~/.msa/msa.sqlite ~/.msa/msa.sqlite.backup.$(date +%Y%m%d)
```

## 🧪 测试

```bash
go test ./pkg/...             # 运行所有测试
go test -cover ./pkg/...      # 查看覆盖率
```

## 🏗️ 架构设计

```
msa/
├── main.go                   # 项目入口，初始化 Cobra CLI
├── cmd/                      # CLI 命令路由 (Cobra)
└── pkg/
    ├── app/                  # TUI 模式入口
    ├── extcli/               # CLI 单次对话模式
    ├── core/                 # 核心 pipeline (无 UI 依赖)
    │   ├── event/            # 事件类型系统
    │   ├── agent/            # Agent: LLM + 工具调用 (Eino)
    │   └── runner/           # Agent → Renderer 编排
    ├── renderer/             # 输出适配层 (CLI / TUI)
    ├── db/                   # 数据库层 (SQLite + GORM)
    ├── tui/                  # 终端界面 (Bubble Tea)
    ├── config/               # 配置管理
    ├── logic/                # 业务逻辑
    │   ├── skills/           # 动态 Skills 系统
    │   ├── provider/         # LLM 提供商
    │   ├── finsvc/           # 金融服务
    │   └── tools/            # 股票、搜索等工具
    ├── model/                # 数据模型
    ├── session/              # 会话持久化与恢复
    └── utils/                # 工具函数
```

### 核心数据流

```
用户输入
  ↓
Runner.Ask(ctx, input)
  ↓
Agent.Run(ctx, messages) → <-chan Event (流式)
  ↓  ReAct 循环: LLM → tool → LLM → ...
Runner 消费 eventCh
  ↓
Renderer.Handle(ctx, event)
  ↓
CLI: fmt.Print    |    TUI: bubbletea msg
```

## 🗺️ 迭代规划

- **Phase 1 (v0.1.x)**: CLI/TUI 框架、配置管理、日志、测试 — ✅ 基本完成
- **Phase 2 (v0.2.x)**: 股票数据 API、交易管理 — ✅ 基本完成；技术指标、实时行情 — 🚧 计划中
- **Phase 3 (v0.3.x)**: AI/LLM 集成 — ✅ 已完成；策略 DSL、回测引擎 — 🚧 计划中
- **Phase 4 (v0.4.x)**: 通知系统、组合管理、插件系统 — 🚧 计划中

## 🛠️ 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.21+ |
| CLI 框架 | [Cobra](https://github.com/spf13/cobra) |
| TUI 框架 | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| UI 样式 | [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| 日志 | [Logrus](https://github.com/sirupsen/logrus) |
| ORM | [GORM](https://github.com/go-gorm/gorm) |
| 数据库 | SQLite (纯 Go 实现) |
| AI/LLM | OpenAI / Claude / SiliconFlow / Ollama |

## 🤝 贡献指南

欢迎贡献代码！请随时提交 Issue 和 Pull Request。

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [Cobra](https://github.com/spf13/cobra) - 强大的 CLI 框架
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - 精美的 TUI 框架
- [Cloudwego Eino](https://github.com/cloudwego/eino) - AI 应用开发框架
