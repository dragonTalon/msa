# MSA - My Stock Agent

> English | [中文文档](README-zh.md)

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)

MSA (My Stock Agent) is a lightweight and flexible open-source stock intelligence agent tool designed for investors and developers.

It currently implements core features including AI-powered chat, stock data query, market news search, and simulated trading management, with support for multiple LLM providers. Future plans include technical analysis, strategy backtesting, and other advanced features. With a modular architecture, MSA lowers the barrier to using quantitative stock tools: it meets individual investors' needs for intelligent stock analysis while providing developers with an extensible open-source framework.

## ✨ Features

### 💬 AI-Powered Chat
- Natural language interaction for stock queries and market analysis
- Multiple LLM provider support (OpenAI, Claude, SiliconFlow, etc.)
- Streaming output with real-time responses
- **Dynamic Skills System** - Modular, reusable prompt templates for domain-specific expertise
- **🧠 Memory System** - AI remembers your preferences and conversation history

### 📊 Stock Analysis
- A-share & Hong Kong stock code lookup
- Company information retrieval
- K-line data display

### 🔍 Smart Search
- Multi-engine support (Google, Bing, etc.)
- Automatic fallback mechanism ensuring search availability
- Web content fetching

### 💼 Trading Management
- SQLite local database storage
- Account management (create, query, status update)
- Position tracking with P&L calculation
- Transaction record management
- Automatic execution mechanism

### 🎯 Beautiful TUI
- Terminal interface built with Bubble Tea
- Markdown rendering support
- Smooth and intuitive user experience

### 🧠 Memory System (NEW!)
- **Automatic Session Recording** - All conversations are automatically saved
- **AI Knowledge Extraction** - Extracts user preferences, concepts, strategies from conversations
- **Smart Memory Injection** - AI uses your memory to provide personalized responses
- **Memory Browser** - View history sessions, knowledge base, and search memories
- **Privacy Protection** - Automatically filters sensitive information (API keys, passwords)
- **Local Storage** - All data stored locally in `~/.msa/remember/`

## 🚀 Quick Start

### Installation

#### One-line Install (Recommended)

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/dragonTalon/msa/main/install.sh | sh
```

The install script automatically detects your system architecture, downloads the latest release, and installs to `~/.local/bin`.

#### Manual Installation

```bash
# Download from GitHub Releases
# https://github.com/dragonTalon/msa/releases

# Or install via go install
go install github.com/dragonTalon/msa@latest

# Or build from source
git clone https://github.com/dragonTalon/msa.git
cd msa
go build
```

### Update

```bash
# Check for updates
msa update --check

# Update to the latest version
msa update
```

The update command automatically downloads the latest release from GitHub Releases and replaces the current binary.

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

### 🎯 Skills System

MSA features a dynamic Skills system that allows modular, reusable prompt templates for domain-specific expertise. Skills are automatically selected based on conversation context or can be manually specified.

#### Managing Skills

```bash
# List all available skills
msa skills list

# Show skill details
msa skills show base

# Disable a skill
msa skills disable stock-analysis

# Enable a skill
msa skills enable stock-analysis
```

#### Using Skills in Conversation

**Manual skill selection (CLI parameter):**
```bash
# Use specific skills for the entire session
msa --skills=base,stock-analysis,output-formats
```

**Manual skill selection (in chat):**
```
/skills: base,stock-analysis
What's the technical analysis of AAPL?
```

**Automatic skill selection:**
The system automatically selects relevant skills based on your question:
```
What's the current price of Tencent?
# Automatically uses: base, stock-analysis
```

#### Creating Custom Skills

Create your own skills in `~/.msa/skills/your-skill/SKILL.md`:

```markdown
---
name: my-custom-skill
description: My specialized analysis framework
version: 1.0.0
priority: 7
---

# My Custom Skill

Your specialized knowledge and rules here...
```

For detailed documentation, see [docs/SKILLS.md](docs/SKILLS.md).

### Chat Mode

```bash
msa chat
```

In chat mode, you can:
- Query stock information: "What's Tencent's stock code?"
- Get company info: "Company information for Tencent"
- Search market news: "Search latest A-share market news"
- Manage trading accounts: "Create an account with 100,000 yuan initial capital"
- Query positions: "Show all my current positions"
- View account summary: "Display my account overview and P&L"
- Submit trades: "Buy 100 shares of Kweichow Moutai at 1850 yuan"
- **Access memory browser**: `/remember`

### 🧠 Memory System

MSA now includes a powerful memory system that automatically learns from your conversations:

**Features:**
- 📝 **Automatic Recording** - All conversations are automatically saved
- 🧠 **AI Knowledge Extraction** - Extracts preferences, concepts, strategies from your chats
- 🤖 **Smart Memory Injection** - AI uses your memory to provide personalized responses
- 🔍 **Memory Browser** - View history, search through conversations and knowledge

**Open Memory Browser:**
```
/remember
```

**Memory Browser Features:**
1. **History Sessions** - Browse all past conversations
2. **Knowledge Base** - View extracted knowledge (user profile, watchlist, concepts, strategies, Q&A)
3. **Search** - Full-text search across all sessions and knowledge
4. **Statistics** - View usage statistics

**Resume Previous Session:**
When you exit a chat session, MSA displays a session ID that you can use to resume later:
```
────────────────────────────────────────
会话已保存: abc-123-def-456
提示: 使用 "msa --resume abc-123-def-456" 恢复此会话
────────────────────────────────────────
```

To resume a previous session:
```bash
# Using the session ID from exit message
msa --resume abc-123-def-456

# Or using the short form
msa -r abc-123-def-456
```

When you resume a session:
- ✅ Historical messages are loaded into the conversation context
- ✅ AI remembers your previous preferences and knowledge
- ✅ You can continue the conversation seamlessly
- ✅ All extracted knowledge from the session is available

**Privacy & Security:**
- All data stored locally in `~/.msa/remember/`
- Automatic filtering of sensitive information (API keys, passwords)
- No cloud synchronization - complete privacy

**Disable Memory System:**
```bash
export MSA_MEMORY_ENABLED=false
msa chat
```

For detailed documentation, see [docs/memory-guide.md](docs/memory-guide.md).

### Configuration Options

MSA supports multiple configuration methods with the following priority: **CLI parameters > Environment variables > Config file > Default values**

#### Interactive Configuration

```bash
# Start TUI configuration interface
msa config
```

In the TUI configuration interface, you can:
- Use arrow keys to select configuration items
- Press `Enter` to enter edit mode
- Auto-fill Base URL when selecting Provider
- Press `S` to save configuration
- Press `R` to reset to default values
- Press `Q` to exit

#### Environment Variables

```bash
# Set Provider
export MSA_PROVIDER=siliconflow

# Set API Key
export MSA_API_KEY=sk-xxxxxxxxxxxx

# Set Base URL (optional)
export MSA_BASE_URL=https://api.example.com/v1

# Set log level (optional)
export MSA_LOG_LEVEL=debug

# Set log file path (optional)
export MSA_LOG_FILE=/path/to/msa.log
```

#### CLI Parameters

```bash
# Resume a previous session
msa --resume <session-id>

# Use config file
msa --config /path/to/config.json chat

# Use key=value format
msa --config apikey=sk-xxx --config loglevel=debug chat

# Mixed usage
msa --config /path/to/config.json --config apikey=sk-xxx chat
```

**Available CLI Parameters:**

| Parameter | Short | Description | Example |
|-----------|-------|-------------|---------|
| `--config` | - | Set configuration (file or key=value) | `--config apikey=sk-xxx` |
| `--resume` | `-r` | Resume a previous session by ID | `--resume abc-123-def-456` |

#### View Configuration in TUI

In the chat interface, use the following command:

```bash
/config
```

This displays current configuration including Provider, Model, Base URL, API Key (partially hidden), log level, and log file path.

#### Configuration File Location

Configuration is saved at `~/.msa/msa_config.json`, containing:
- Provider: LLM provider
- Model: Model to use
- Base URL: API base URL
- API Key: API key
- LogConfig: Log configuration (level, format, output, file path)

### Configuration Security

⚠️ **Important**: API Keys are stored in plaintext in the configuration file.

For security, it is recommended to:

```bash
# Set config file permissions (readable/writable by owner only)
chmod 600 ~/.msa/msa_config.json

# Ensure config directory permissions are correct
chmod 700 ~/.msa/
```

**Security recommendations**:
- Do not commit configuration files to version control
- Do not use configurations with API Keys in shared environments
- Rotate API Keys regularly
- Use different API Keys for different environments

### Provider Extension Guide

MSA supports extending new LLM Providers. Here are the steps to add one:

#### 1. Register in ProviderRegistry

Edit `pkg/model/comment.go` and add the new Provider to `ProviderRegistry`:

```go
var ProviderRegistry = map[LlmProvider]ProviderInfo{
    Siliconflow: {
        ID:             Siliconflow,
        DisplayName:    "SiliconFlow",
        Description:    "LLM API provider, OpenAI compatible",
        DefaultBaseURL: "https://api.siliconflow.cn/v1",
        KeyPrefix:      "sk-",
    },
    // Add new Provider
    YourProvider: {
        ID:             YourProvider,
        DisplayName:    "Your Provider Name",
        Description:    "Provider description",
        DefaultBaseURL: "https://api.yourprovider.com/v1",
        KeyPrefix:      "custom-",
    },
}
```

#### 2. Add Provider Constant

Add the Provider constant in `pkg/model/comment.go`:

```go
const (
    Siliconflow LlmProvider = "siliconflow"
    YourProvider LlmProvider = "yourprovider"
)
```

#### 3. Update GetDisplayName() and GetDefaultBaseURL() methods if needed

If the new Provider has special display name or default URL logic, add special handling in the corresponding methods.

#### 4. Verify Configuration

```bash
# Rebuild
go build

# Test configuration
msa config

# Select new Provider and verify configuration is saved
```

## 🗄️ Local Database

MSA includes SQLite local database support for storing:
- **Accounts**: User accounts with balance tracking
- **Transactions**: Buy/sell orders with status tracking
- **Positions**: Holding calculations and profit/loss

### Database Location

The database file is stored at:
```
~/.msa/msa.sqlite
```

### Data Backup

To backup your data:
```bash
cp ~/.msa/msa.sqlite ~/.msa/msa.sqlite.backup.$(date +%Y%m%d)
```

To restore from backup:
```bash
cp ~/.msa/msa.sqlite.backup.YYYYMMDD ~/.msa/msa.sqlite
```

### Amount Units

**Important**: All monetary amounts are stored in "毫" (1/10000 of a Yuan) as integers to avoid floating-point precision issues.
- `10000` = 1.00 元
- Display conversion: `amount / 10000 = displayed value`

## 🧪 Testing

The project includes unit tests covering core business logic and utility functions.

### Running Tests

```bash
# Run all tests
go test ./pkg/...

# Run tests with coverage
go test -cover ./pkg/...

# Generate coverage report
go test -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out
```

### Current Coverage

| Module | Coverage |
|--------|----------|
| pkg/utils | 83.7% |
| pkg/logic/tools | 100.0% |
| pkg/logic/provider | 76.9% |
| pkg/config | 53.4% |
| pkg/logic/command | 51.1% |

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
    ├── db/                  # Database layer
    │   ├── db.go            # Database initialization
    │   ├── global.go        # Global database management
    │   ├── migrate.go       # Schema migration
    │   ├── account.go       # Account operations
    │   └── transaction.go   # Transaction operations
    ├── model/               # Data models
    │   ├── account.go       # Account model
    │   ├── transaction.go   # Transaction model
    │   └── stock.go         # Stock data model
    ├── service/             # Business services
    │   ├── account_service.go    # Account service
    │   ├── trade_service.go      # Trading service
    │   └── position_service.go   # Position calculation
    ├── tui/                 # Terminal UI module
    │   ├── style/           # UI styling
    │   ├── config/          # Configuration TUI
    │   ├── chat.go          # Chat interface
    │   └── model_selector.go # Model selector
    ├── config/              # Configuration management
    │   ├── local_config.go  # Local storage configuration
    │   ├── env.go           # Environment variables
    │   ├── validator.go     # Configuration validator
    │   └── logger.go        # Logging configuration
    ├── logic/               # Business logic
    │   ├── agent/           # AI agent
    │   ├── command/         # Command handling
    │   ├── message/         # Message management
    │   ├── provider/        # LLM providers
    │   ├── skills/          # Dynamic Skills system
    │   └── tools/           # Tools
    │       ├── stock/       # Stock tools
    │       ├── search/      # Search tools
    │       └── finance/     # Finance tools
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
- [x] Unit test framework
- [ ] CI/CD pipeline (GitHub Actions)

### Phase 2: Core Features (v0.2.x)
- [x] **Data Module**
  - [x] Stock data fetching API integration
  - [ ] Real-time quotes subscription
  - [x] Historical data storage (SQLite)
  - [x] Account management
  - [x] Transaction recording
  - [x] Position tracking
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
- [x] **AI/LLM Integration**
  - [x] LLM API integration (OpenAI/Claude/Local)
  - [x] Natural language stock query
  - [x] AI-powered market analysis
  - [x] Intelligent Q&A assistant
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
| ORM | [GORM](https://github.com/go-gorm/gorm) |
| Database | SQLite ([github.com/glebarez/sqlite](https://github.com/glebarez/sqlite), pure Go) |
| Data Storage | ~/.msa/msa.sqlite |
| Skills Format | Markdown with YAML frontmatter |
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
