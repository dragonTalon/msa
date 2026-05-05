# MSA - My Stock Agent

> English | [中文文档](README-zh.md)

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)

MSA (My Stock Agent) is a lightweight, flexible open-source stock intelligence agent tool designed for investors and developers.

It features AI-powered chat, stock data query, market news search, and simulated trading management, with support for multiple LLM providers. With a modular architecture, MSA lowers the barrier to quantitative stock tools: it serves individual investors' needs for intelligent stock analysis while providing developers with an extensible open-source framework.

## ✨ Features

- **AI-Powered Chat** - Natural language interaction for stock queries and market analysis, with streaming output and support for multiple LLM providers (OpenAI, Claude, SiliconFlow, etc.)
- **Dynamic Skills System** - Modular, reusable prompt templates for domain-specific expertise, auto-selected based on conversation context
- **Stock Analysis** - A-share and Hong Kong stock code lookup, company information retrieval, K-line data display
- **Smart Search** - Multi-engine search (Google, Bing, etc.) with automatic fallback and web content fetching
- **Trading Management** - SQLite-based account, position, and transaction management with P&L tracking
- **Memory System** - Automatic session recording, AI knowledge extraction, and smart memory injection for personalized responses
- **Beautiful TUI** - Terminal interface built with Bubble Tea, supporting Markdown rendering

## 📸 Demo

> Screenshots and GIFs coming soon...

## 🚀 Quick Start

### Installation

```bash
# One-line install (macOS / Linux)
curl -fsSL https://raw.githubusercontent.com/dragonTalon/msa/main/install.sh | sh

# Or via go install
go install github.com/dragonTalon/msa@latest

# Or build from source
git clone https://github.com/dragonTalon/msa.git
cd msa
go build
```

### Update

```bash
msa update          # Update to latest
msa update --check  # Check for updates only
```

### Configuration

```bash
msa config    # Interactive TUI configuration
```

Or via environment variables:

```bash
export MSA_PROVIDER=siliconflow
export MSA_API_KEY=sk-xxxxxxxxxxxx
```

### Start Chatting

```bash
msa                                    # TUI interactive mode
msa -q "What's Tencent's stock code?"  # Single-round CLI mode
msa -r <session-id>                    # Resume previous session
```

For detailed usage, see `msa --help`.

## 📖 Documentation

| Document | Description |
|----------|-------------|
| [Skills System](docs/skills.md) | Managing and creating custom skills |
| [Memory System](docs/memory-guide.md) | Session recording, knowledge extraction, privacy |
| [Configuration](docs/configuration.md) | Config methods, security, environment variables |
| [Provider Extension](docs/provider-extension.md) | Adding new LLM providers |
| [Database](docs/database.md) | SQLite storage, backup, data model |

## 🗄️ Local Database

MSA uses SQLite for storing accounts, positions, and transactions. The database file is at `~/.msa/msa.sqlite`.

Backup your data:
```bash
cp ~/.msa/msa.sqlite ~/.msa/msa.sqlite.backup.$(date +%Y%m%d)
```

## 🧪 Testing

```bash
go test ./pkg/...             # Run all tests
go test -cover ./pkg/...      # With coverage
```

## 🏗️ Architecture

```
msa/
├── main.go                   # Entry point, initializes Cobra CLI
├── cmd/                      # CLI command routing (Cobra)
└── pkg/
    ├── app/                  # TUI mode entry
    ├── extcli/               # CLI single-round mode
    ├── core/                 # Core pipeline (no UI dependency)
    │   ├── event/            # Event type system
    │   ├── agent/            # Agent: LLM + tool calls (Eino)
    │   └── runner/           # Agent → Renderer orchestration
    ├── renderer/             # Output adapters (CLI / TUI)
    ├── db/                   # Database layer (SQLite + GORM)
    ├── tui/                  # Terminal UI (Bubble Tea)
    ├── config/               # Configuration management
    ├── logic/                # Business logic
    │   ├── skills/           # Dynamic Skills system
    │   ├── provider/         # LLM providers
    │   ├── finsvc/           # Finance services
    │   └── tools/            # Stock, search, and other tools
    ├── model/                # Data models
    ├── session/              # Session persistence and resume
    └── utils/                # Utility functions
```

### Core Data Flow

```
User Input
  ↓
Runner.Ask(ctx, input)
  ↓
Agent.Run(ctx, messages) → <-chan Event (streaming)
  ↓  ReAct loop: LLM → tool → LLM → ...
Runner consumes eventCh
  ↓
Renderer.Handle(ctx, event)
  ↓
CLI: fmt.Print    |    TUI: bubbletea msg
```

## 🗺️ Roadmap

- **Phase 1 (v0.1.x)**: CLI/TUI framework, config management, logging, testing — ✅ mostly complete
- **Phase 2 (v0.2.x)**: Stock data API, trading management — ✅ mostly complete; technical indicators, real-time quotes — 🚧 planned
- **Phase 3 (v0.3.x)**: AI/LLM integration — ✅ done; strategy DSL, backtesting engine — 🚧 planned
- **Phase 4 (v0.4.x)**: Notification system, portfolio management, plugin system — 🚧 planned

## 🛠️ Tech Stack

| Category | Technology |
|----------|------------|
| Language | Go 1.21+ |
| CLI Framework | [Cobra](https://github.com/spf13/cobra) |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| UI Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| Logging | [Logrus](https://github.com/sirupsen/logrus) |
| ORM | [GORM](https://github.com/go-gorm/gorm) |
| Database | SQLite (pure Go) |
| AI/LLM | OpenAI / Claude / SiliconFlow / Ollama |

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - Powerful CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Beautiful TUI framework
- [Cloudwego Eino](https://github.com/cloudwego/eino) - AI application development framework
