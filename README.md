# MSA - My Stock Agent

> English | [中文文档](README-zh.md)

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)

MSA (My Stock Agent) is a lightweight and flexible open-source stock intelligence agent tool designed for investors and developers.

It integrates core capabilities including stock data collection, multi-dimensional market analysis, strategy backtesting, and automated trading assistance, supporting custom strategy configuration and secondary development. With a modular architecture, MSA lowers the barrier to using quantitative stock tools: it meets individual investors' needs for automated trading while providing developers with an extensible open-source framework.

## ✨ Features

### 💬 AI-Powered Chat
- Natural language interaction for stock queries and market analysis
- Multiple LLM provider support (OpenAI, Claude, SiliconFlow, etc.)
- Streaming output with real-time responses

### 📊 Stock Analysis
- A-share & Hong Kong stock code lookup
- Company information retrieval
- K-line data display

### 🔍 Smart Search
- Multi-engine support (Google, Bing, etc.)
- Automatic fallback mechanism ensuring search availability
- Web content fetching

### 🎯 Beautiful TUI
- Terminal interface built with Bubble Tea
- Markdown rendering support
- Smooth and intuitive user experience

## 🚀 Quick Start

### Installation

```bash
# Install via go install
go install github.com/yourusername/msa@latest

# Or clone the repository
git clone https://github.com/yourusername/msa.git
cd msa
go build
```

### Configuration

```bash
# Configure API key
msa config

# Select model
msa chat
```

### Start Chatting

```bash
msa chat
```

## 📸 Demo

> Screenshots and GIFs coming soon...

## 📖 Usage

### Chat Mode

```bash
msa chat
```

In chat mode, you can:
- Query stock information: "What's Tencent's stock code?"
- Get company info: "Company information for Tencent"
- Search market news: "Search latest A-share market news"

### Configuration Options

```bash
# Show configuration
msa config show

# Update configuration
msa config set
```

## 🏗️ Architecture

```
msa/ (Project Root)
├── go.mod                    # Go module dependency configuration
├── go.sum                    # Go module dependency verification
├── main.go                   # Project entry, initializes Cobra CLI
├── cmd/                      # Cobra CLI command routing layer
│   └── root.go              # Root command definition, routing only
└── pkg/                      # Business implementation layer
    ├── app/                 # Application core module
    │   └── app.go           # Application startup
    ├── tui/                 # Terminal UI module
    │   ├── style/           # UI styling
    │   ├── chat.go          # Chat interface
    │   └── model_selector.go # Model selector
    ├── config/              # Configuration management
    │   ├── local_config.go  # Local storage configuration
    │   └── logger.go        # Logging configuration
    ├── logic/               # Business logic
    │   ├── agent/           # AI agent
    │   ├── command/         # Command handling
    │   ├── provider/        # LLM providers
    │   └── tools/           # Tools (stock, search, etc.)
    └── utils/               # Utility functions
        ├── file.go
        ├── http.go
        └── format.go
```

### Architecture Principles

- **cmd/** - Command routing only, keeps code minimal
- **pkg/** - All business implementations with clear responsibilities
  - **app/** - Application lifecycle management
  - **tui/** - Terminal user interface (Bubble Tea)
  - **config/** - Configuration and logging
  - **logic/** - Core business logic
  - **utils/** - Common utility functions

## 🗺️ Roadmap

### Phase 1: Foundation (v0.1.x)
- [x] CLI framework setup (Cobra)
- [x] TUI interface (Bubble Tea)
- [x] Project structure refactoring
- [x] Configuration management
- [x] Logging system
- [ ] Unit test framework
- [ ] CI/CD pipeline (GitHub Actions)

### Phase 2: Core Features (v0.2.x)
- [ ] **Data Module**
  - [ ] Stock data fetching API integration
  - [ ] Real-time quotes subscription
  - [ ] Historical data storage (SQLite/PostgreSQL)
  - [ ] Data caching layer (Redis)
- [ ] **Analysis Module**
  - [ ] Technical indicators (MA, MACD, RSI, KDJ)
  - [ ] K-line pattern recognition
  - [ ] Trend analysis
- [ ] **CLI Commands**
  - [ ] `msa quote <symbol>` - Get stock quote
  - [ ] `msa history <symbol>` - Historical data
  - [ ] `msa analyze <symbol>` - Technical analysis

### Phase 3: Intelligence (v0.3.x)
- [ ] **AI/LLM Integration**
  - [ ] LLM API integration (OpenAI/Claude/Local)
  - [ ] Natural language stock query
  - [ ] AI-powered market analysis
  - [ ] Intelligent Q&A assistant
- [ ] **Strategy Module**
  - [ ] Strategy DSL definition
  - [ ] Backtesting engine
  - [ ] Performance metrics
  - [ ] Strategy templates

### Phase 4: Advanced (v0.4.x)
- [ ] **Notification System**
  - [ ] Price alert
  - [ ] Strategy signal notification
  - [ ] Multi-channel support (Email/Webhook/Telegram)
- [ ] **Portfolio Management**
  - [ ] Portfolio tracking
  - [ ] P&L calculation
  - [ ] Risk assessment
- [ ] **Plugin System**
  - [ ] Plugin architecture design
  - [ ] Custom data source plugins
  - [ ] Custom strategy plugins

### Future Vision
- [ ] Web dashboard (optional)
- [ ] Mobile companion app
- [ ] Community strategy marketplace
- [ ] Multi-market support (US/HK/A-shares)

## 🛠️ Tech Stack

| Category | Technology |
|----------|------------|
| Language | Go 1.21+ |
| CLI Framework | [Cobra](https://github.com/spf13/cobra) |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| UI Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| Logging | [Logrus](https://github.com/sirupsen/logrus) |
| Database | SQLite / PostgreSQL (planned) |
| Cache | Redis (planned) |
| AI/LLM | OpenAI / Claude / SiliconFlow / Ollama |

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

### Development Workflow

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - Powerful CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Beautiful TUI framework
- [Cloudwego Eino](https://github.com/cloudwego/eino) - AI application development framework
