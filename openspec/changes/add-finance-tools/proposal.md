# 提案：add-finance-tools

## 知识上下文

### 相关历史问题

知识库中未找到相关问题（当前知识库主要包含浏览器自动化和 web 搜索相关内容）。

### 需避免的反模式

未识别到特定反模式。

### 推荐模式

✅ **推荐**：考虑使用以下模式（从知识库中的通用模式）：

- **错误处理和重试策略** (003)
  - 描述：外部 API 调用失败时的处理策略
  - 使用时机：获取股票实时价格时

## 为什么

当前项目已实现 SQLite 本地数据库和完整的账户、交易、持仓管理功能（位于 `pkg/db/` 和 `pkg/service/`），但这些能力从未被实际使用。数据库连接在 CLI 启动时从未初始化，导致：

1. 已投入开发的代码闲置（Account、Transaction 模型，CRUD 操作，业务逻辑）
2. AI Agent 无法通过 tools 调用这些核心功能
3. 用户无法进行账户管理、交易记录、持仓查询等操作

通过将这些能力封装为 MsaTool，可以让 AI Agent 在对话中直接调用，实现真正的交互式股票交易管理。

## 变更内容

### 新增内容

- **pkg/logic/tools/finance/**: 新增 finance tools 模块
  - `common.go`: 公共函数（获取活跃账户、获取价格、格式化）
  - `account.go`: 账户管理 tools
  - `position.go`: 持仓查询 tools
  - `trade.go`: 交易执行 tools

- **pkg/db/global.go**: 新增全局 DB 连接管理
  - `InitGlobalDB()`: 初始化全局 DB 连接（非阻塞）
  - `GetDB()`: 获取全局 DB 连接
  - `IsDBAvailable()`: 检查 DB 是否可用
  - `CloseGlobalDB()`: 关闭 DB 连接

### 修改内容

- **pkg/config/local_config.go**: 在 `InitConfig()` 中调用 `db.InitGlobalDB()`
- **cmd/root.go**: 在 `ExecuteWithSignal()` 中添加 `defer db.CloseGlobalDB()`
- **pkg/logic/tools/register.go**: 注册 finance tools

## 能力

### 新增能力

- **account-management**: 账户创建、查询、状态管理（冻结/解冻/关闭）
  - 全局只允许一个活跃账户
  - 支持账户状态变更

- **position-tracking**: 持仓查询、市值计算、盈亏统计
  - 实时价格获取（复用 stock/fetchStockData）
  - 持仓列表、账户总览

- **trade-execution**: 订单提交、自动成交
  - 买入/卖出订单
  - 提交后自动成交（PENDING → FILLED）
  - 交易记录查询

### 修改能力

无现有能力的需求变更。

## 影响

### 代码影响

- **新增文件**:
  - `pkg/db/global.go`
  - `pkg/logic/tools/finance/*.go` (4个文件)

- **修改文件**:
  - `pkg/config/local_config.go`
  - `cmd/root.go`
  - `pkg/logic/tools/register.go`

- **新增 Tool Group**: 使用现有的 `FinanceToolGroup`

### API 影响

- 新增 8 个 MsaTool，通过 `GetAllTools()` 暴露给 AI Agent
- DB 初始化改为非阻塞模式，失败时仅记录警告

### 依赖影响

- 无新增外部依赖
- 复用现有的 `gorm.io/gorm`, `pkg/db`, `pkg/service`
- 复用 `stock/fetchStockData()` 获取实时价格

### 运行时影响

- CLI 启动时初始化 SQLite（~10-50ms）
- DB 文件位于 `~/.msa/msa.sqlite`
- 退出时执行 `PRAGMA optimize` 优化数据库

---

**知识备注**：
- 此提案使用 `openspec/knowledge` 中的知识生成
- 最后知识同步：2025-02-23
- 知识库总问题数：5
- 总模式数：3
- **注意**：这是新领域（finance tools），如遇问题请创建 fix.md 来构建知识库
