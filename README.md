# msa (My Stock Agent)

##  Description
MSA (My Stock Agent) is a lightweight and flexible open-source stock intelligence agent tool designed for investors and developers.
It integrates core capabilities including stock data collection, multi-dimensional market analysis, strategy backtesting,
and automated trading assistance, supporting custom strategy configuration and secondary development. With a modular architecture,
MSA lowers the barrier to using quantitative stock tools: 
it meets individual investors' needs for automated trading while providing developers with an extensible open-source framework.
You can optimize strategies based on existing modules, integrate new data sources, or contribute featured functions to build a collaborative community ecosystem. 
Whether you're a quantitative novice or an experienced developer, MSA enables efficient access to the stock market and facilitates strategy implementation 
and value accumulation.

## Project Structure 
```
msa/ (Project Root | 项目根目录)
├── go.mod (Go Module Dependency Configuration | Go模块依赖配置)
├── go.sum (Go Module Dependency Verification | Go模块依赖校验文件)
├── main.go (Project Entry File, Initializes Cobra CLI | 项目入口文件，初始化Cobra CLI)
├── cmd/ (Cobra CLI Command Routing Layer | Cobra CLI 命令路由层)
│   └── root.go (Root Command Definition | 根命令定义，仅负责路由)
└── pkg/ (Business Implementation Layer | 业务实现层)
    ├── app/ (Application Core Module | 应用核心模块)
    │   └── app.go (App Startup  | 应用启动)
    ├── tui/ (Terminal UI Module | 终端界面模块)
    │   ├── style.go (UI Style & Logo Rendering | 界面样式与LOGO渲染)
    │   └── chat.go (Chat Model - Bubble Tea Implementation | 聊天模型 - Bubble Tea实现)
    ├── config/ (Configuration Management Module | 配置管理模块)
    │   ├── local_config.go (Local Store Configuration | 本地存储配置)
    │   └── logger.go (Logger Configuration | 日志配置)
    └── utils/ (Utility Functions | 工具函数)
        └── file.go (File Utilities | 文件工具)
```

## Architecture Design | 架构设计

- **cmd/** - Command routing only, keeps code minimal | 仅存放命令路由，代码精简
- **pkg/** - All business implementations with clear responsibilities | 存放所有具体实现，职责清晰
  - **app/** - Application lifecycle management | 应用生命周期管理
  - **tui/** - Terminal user interface (Bubble Tea) | 终端用户界面
  - **config/** - Configuration and logging | 配置与日志
  - **utils/** - Common utilities | 通用工具函数

## Roadmap | 迭代规划

### Phase 1: Foundation | 基础架构 (v0.1.x)
- [x] CLI framework setup (Cobra) | CLI框架搭建
- [x] TUI interface (Bubble Tea) | TUI终端界面
- [x] Project structure refactoring | 项目结构重构
- [x] Configuration management | 配置管理模块
- [x] Logging system | 日志系统
- [ ] Unit test framework | 单元测试框架
- [ ] CI/CD pipeline (GitHub Actions) | CI/CD流水线

### Phase 2: Core Features | 核心功能 (v0.2.x)
- [ ] **Data Module** | 数据模块
  - [ ] Stock data fetching API integration | 股票数据API接入
  - [ ] Real-time quotes subscription | 实时行情订阅
  - [ ] Historical data storage (SQLite/PostgreSQL) | 历史数据存储
  - [ ] Data caching layer (Redis) | 数据缓存层
- [ ] **Analysis Module** | 分析模块
  - [ ] Technical indicators (MA, MACD, RSI, KDJ) | 技术指标计算
  - [ ] K-line pattern recognition | K线形态识别
  - [ ] Trend analysis | 趋势分析
- [ ] **CLI Commands** | CLI命令扩展
  - [ ] `msa quote <symbol>` - Get stock quote | 获取股票报价
  - [ ] `msa history <symbol>` - Historical data | 历史数据查询
  - [ ] `msa analyze <symbol>` - Technical analysis | 技术分析

### Phase 3: Intelligence | 智能化 (v0.3.x)
- [ ] **AI/LLM Integration** | AI/大模型集成
  - [ ] LLM API integration (OpenAI/Claude/Local) | 大模型API接入
  - [ ] Natural language stock query | 自然语言股票查询
  - [ ] AI-powered market analysis | AI驱动市场分析
  - [ ] Intelligent Q&A assistant | 智能问答助手
- [ ] **Strategy Module** | 策略模块
  - [ ] Strategy DSL definition | 策略DSL定义
  - [ ] Backtesting engine | 回测引擎
  - [ ] Performance metrics | 绩效指标统计
  - [ ] Strategy templates | 策略模板库

### Phase 4: Advanced | 高级特性 (v0.4.x)
- [ ] **Notification System** | 通知系统
  - [ ] Price alert | 价格预警
  - [ ] Strategy signal notification | 策略信号通知
  - [ ] Multi-channel support (Email/Webhook/Telegram) | 多渠道推送
- [ ] **Portfolio Management** | 组合管理
  - [ ] Portfolio tracking | 持仓跟踪
  - [ ] P&L calculation | 盈亏计算
  - [ ] Risk assessment | 风险评估
- [ ] **Plugin System** | 插件系统
  - [ ] Plugin architecture design | 插件架构设计
  - [ ] Custom data source plugins | 自定义数据源插件
  - [ ] Custom strategy plugins | 自定义策略插件

### Future Vision | 远景规划
- [ ] Web dashboard (optional) | Web控制台（可选）
- [ ] Mobile companion app | 移动端辅助应用
- [ ] Community strategy marketplace | 社区策略市场
- [ ] Multi-market support (US/HK/A-shares) | 多市场支持

## Tech Stack | 技术栈

| Category | Technology |
|----------|------------|
| Language | Go 1.21+ |
| CLI Framework | [Cobra](https://github.com/spf13/cobra) |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| UI Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| Logging | [Logrus](https://github.com/sirupsen/logrus) |
| Database | SQLite / PostgreSQL (planned) |
| Cache | Redis (planned) |
| AI/LLM | OpenAI / Claude / Ollama (planned) |

## Contributing | 贡献指南

Contributions are welcome! Please feel free to submit issues and pull requests.

欢迎贡献代码！请随时提交 Issue 和 Pull Request。

## License | 许可证

MIT License