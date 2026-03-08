# 设计：add-finance-tools

## 上下文

### 当前状态

项目已实现完整的 SQLite 本地数据库能力：
- **数据层** (`pkg/db/`): Account、Transaction 模型，CRUD 操作，迁移脚本
- **服务层** (`pkg/service/`): AccountService，交易处理函数，持仓计算函数
- **工具层** (`pkg/logic/tools/`): 现有 stock、search tools 使用 MsaTool 接口

**问题**：这些能力从未被集成到应用启动流程，AI Agent 无法调用。

### 约束条件

- **非阻塞初始化**：DB 初始化失败不应阻止程序启动
- **全局单账户**：数据库中只允许一个 ACTIVE 状态的账户
- **事务安全**：所有金额变更操作必须在事务中执行
- **复用现有代码**：充分利用已实现的 `pkg/db/` 和 `pkg/service/`

### 利益相关者

- **终端用户**：需要通过 AI Agent 对话进行账户、交易、持仓管理
- **系统**：需要保证数据一致性和完整性
- **开发者**：需要清晰的模块划分和可维护的代码结构

---

## 目标 / 非目标

**目标：**
- 集成 SQLite 到 CLI 启动流程
- 实现 8 个 finance tools（账户 3 个，持仓 2 个，交易 3 个）
- 复用现有的数据层和服务层代码
- 提供友好的错误处理和降级策略

**非目标：**
- 不支持多账户并发（全局单账户）
- 不实现自动交易（tools 需要用户显式调用）
- 不添加新的外部依赖
- 不修改现有的 Account、Transaction 数据模型

---

## 设计决策

### 1. 全局 DB 连接管理

**决策**：使用全局变量 + `sync.Once` 实现单例模式

**理由**：
- 简单高效，与现有 `configCache` 模式一致
- 确保整个应用生命周期内只有一个 DB 连接
- 延迟初始化，首次调用时才创建连接

**实现位置**：`pkg/db/global.go`

```go
var (
    globalDB  *gorm.DB
    initErr   error
    once      sync.Once
)

func InitGlobalDB() {
    once.Do(func() {
        globalDB, initErr = InitDB()
        if initErr != nil {
            log.Warnf("SQLite 初始化失败: %v", initErr)
            return
        }
        if err := Migrate(globalDB); err != nil {
            log.Warnf("数据库迁移失败: %v，关闭连接", err)
            CloseDB(globalDB)
            globalDB = nil
            initErr = err
            return
        }
        log.Info("SQLite 初始化成功")
    })
}

func GetDB() *gorm.DB {
    return globalDB
}

func IsDBAvailable() bool {
    return globalDB != nil
}
```

### 2. 非阻塞初始化时机

**决策**：在 `config.InitConfig()` 中调用 `db.InitGlobalDB()`

**理由**：
- 与配置初始化语义一致（都是设置基础设施）
- 日志系统已初始化，可以正常记录初始化日志
- 目录 `~/.msa/` 已创建，DB 文件可以正常存放

**修改位置**：`pkg/config/local_config.go`

```go
func InitConfig() error {
    // ... 现有代码：创建目录、加载配置、初始化日志 ...

    // 新增：初始化 SQLite（非阻塞）
    db.InitGlobalDB()

    return nil
}
```

### 3. 程序退出时清理

**决策**：在 `cmd/root.go` 的 `ExecuteWithSignal()` 中添加 defer

**理由**：
- 确保程序正常退出或信号中断时都能关闭 DB
- 执行 `PRAGMA optimize` 优化数据库
- 使用 defer 确保在任何退出路径都会执行

**修改位置**：`cmd/root.go`

```go
func ExecuteWithSignal(rootCmd *cobra.Command) {
    if err := config.InitConfig(); err != nil {
        log.Warnf("初始化配置失败: %v", err)
    }

    // 新增：注册清理函数
    defer func() {
        if err := db.CloseGlobalDB(); err != nil {
            log.Warnf("关闭数据库时出错: %v", err)
        }
    }()

    ctx, cancel := NotifySignal(...)
    defer cancel()

    if err := rootCmd.ExecuteContext(ctx); err != nil {
        // ...
    }
}
```

### 4. 活跃账户查询逻辑

**决策**：在 tools 层实现 `getActiveAccount()` 辅助函数

**理由**：
- 全局单账户逻辑是 tools 特定的，不应放在 db/service 层
- 方便所有 finance tools 复用
- 返回友好错误提示用户先创建账户

**实现位置**：`pkg/logic/tools/finance/common.go`

```go
func getActiveAccount(db *gorm.DB) (*model.Account, error) {
    var account model.Account
    err := db.Where("status = ?", model.AccountStatusActive).
        First(&account).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("未找到活跃账户，请先创建账户")
        }
        return nil, fmt.Errorf("查询账户失败: %w", err)
    }
    return &account, nil
}
```

### 5. 价格获取复用

**决策**：复用 `stock/fetchStockData()` 并封装为 `fetchCurrentPrice()`

**理由**：
- 避免重复实现 HTTP 请求逻辑
- 统一价格数据来源
- 封装单位转换（字符串 → 毫）

**实现位置**：`pkg/logic/tools/finance/common.go`

```go
func fetchCurrentPrice(stockCode string) (int64, error) {
    resp, err := stock.FetchStockData(stockCode)
    if err != nil {
        return 0, err
    }
    // resp.CurrentPrice 是字符串，如 "10.50"
    price, err := strconv.ParseFloat(resp.CurrentPrice, 64)
    if err != nil {
        return 0, fmt.Errorf("解析价格失败: %w", err)
    }
    return model.YuanToHao(price), nil
}
```

### 6. 自动成交流程

**决策**：在 tool 函数内依次调用 `SubmitBuyOrder()` 和 `FillOrder()`

**理由**：
- 保持服务层函数职责单一（Submit 提交，Fill 成交）
- 在 tool 层组合业务流程
- 失败时保持 PENDING 状态，数据不丢失

**实现位置**：`pkg/logic/tools/finance/trade.go`

```go
func SubmitBuyOrderTool(ctx context.Context, param *BuyOrderParam) (string, error) {
    db := msadb.GetDB()
    if db == nil {
        return "", fmt.Errorf("数据库未初始化")
    }

    account, err := getActiveAccount(db)
    if err != nil {
        return "", err
    }

    // 1. 提交订单 (PENDING)
    order := service.Order{...}
    transID, err := service.SubmitBuyOrder(db, account.ID, order)
    if err != nil {
        return "", err
    }

    // 2. 自动成交
    if err := service.FillOrder(db, transID); err != nil {
        return "", fmt.Errorf("订单创建成功但成交失败: %w", err)
    }

    return fmt.Sprintf("买入订单已成交，交易ID: %d", transID), nil
}
```

### 7. Tool 注册方式

**决策**：在 `pkg/logic/tools/register.go` 的 `init()` 中添加 `registerFinance()`

**理由**：
- 遵循现有注册模式（registerStock、registerSearch）
- 集中管理所有 tools 的注册
- 编译时检查接口实现（通过 var 声明）

**修改位置**：`pkg/logic/tools/register.go`

```go
var _ MsaTool = (*finance.CreateAccountTool)(nil)
// ... 其他 tools

func init() {
    registerStock()
    registerSearch()
    registerFinance() // 新增
}

func registerFinance() {
    RegisterTool(&finance.CreateAccountTool{})
    RegisterTool(&finance.GetAccountTool{})
    // ... 其他 finance tools
}
```

---

## 推荐模式

### 错误处理和降级策略

**模式ID**：003
**来源**：知识库

**描述**：
外部 API 调用（股票价格）失败时，不应阻止整个功能，应提供降级策略。

**适用场景**：
- 获取股票实时价格失败
- 数据库不可用时

**实现**：
```go
// 价格获取失败时跳过，返回部分数据
for _, stockCode := range stockCodes {
    price, err := fetchCurrentPrice(stockCode)
    if err != nil {
        log.Warnf("获取 %s 价格失败: %v", stockCode, err)
        // 设置价格为 0，后续计算时标注为 N/A
        price = 0
    }
    prices[stockCode] = price
}
```

### MsaTool 接口实现模式

**描述**：
所有 tools 必须实现 MsaTool 接口，使用 `utils.InferTool()` 自动推导工具信息。

**适用场景**：
- 所有新增的 finance tools

**实现**：
```go
type CreateAccountTool struct{}

func (t *CreateAccountTool) GetToolInfo() (tool.BaseTool, error) {
    return utils.InferTool(t.GetName(), t.GetDescription(), CreateAccount)
}

func (t *CreateAccountTool) GetName() string {
    return "create account"
}

func (t *CreateAccountTool) GetDescription() string {
    return "创建新的交易账户（全局只能有一个活跃账户）| Create a new trading account (only one active account globally)"
}

func (t *CreateAccountTool) GetToolGroup() model.ToolGroup {
    return model.FinanceToolGroup
}

func CreateAccount(ctx context.Context, param *AccountParam) (string, error) {
    // 实现逻辑
}
```

---

## 避免的反模式

### 反模式 1：阻塞式 DB 初始化

**问题所在**：
如果 DB 初始化失败就阻塞程序启动，会导致整个 CLI 不可用。

**影响**：
用户无法使用其他与 DB 无关的功能（如股票查询、搜索）。

**不要这样做**：
```go
func InitConfig() error {
    db, err := db.InitDB()
    if err != nil {
        return err  // 阻止程序启动
    }
    // ...
}
```

**应该这样做**：
```go
func InitConfig() error {
    db.InitGlobalDB()  // 仅记录警告，不返回错误
    return nil
}
```

### 反模式 2：忽略事务错误

**问题所在**：
事务操作失败后没有正确回滚，可能导致数据不一致。

**影响**：
账户金额与实际交易不匹配，余额计算错误。

**不要这样做**：
```go
tx := db.Begin()
tx.Create(&trans)
tx.Commit()  // 忽略错误
```

**应该这样做**：
```go
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := tx.Create(&trans).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Commit().Error; err != nil {
    return err
}
```

### 反模式 3：硬编码错误消息

**问题所在**：
错误消息散落在代码中，难以维护和国际化。

**影响**：
修改提示语需要多处修改，容易遗漏。

**不要这样做**：
```go
return "", errors.New("未找到账户")
```

**应该这样做**：
```go
// 定义常量
const (
    ErrNoActiveAccount = "未找到活跃账户，请先创建账户"
)
return "", fmt.Errorf(ErrNoActiveAccount)
```

---

## 风险 / 权衡

### 已知风险

1. **[SQLite 写锁]** → 单用户场景写入不频繁，影响可控
2. **[价格 API 限流]** → 腾讯 API 可能有限流，需控制调用频率
3. **[持仓数量为负]** → 卖出验证不足可能导致负持仓，需加强校验
4. **[并发安全]** → GORM 连接池配置了 WAL 模式，支持一定并发

### 权衡

1. **全局单账户 vs 多账户**：选择单账户简化设计，满足当前需求
2. **自动成交 vs 手动成交**：选择自动成交简化流程，符合即时交易场景
3. **价格实时获取 vs 缓存**：选择实时获取确保数据最新
4. **集中注册 vs 分散注册**：选择集中注册便于管理

---

## 迁移计划

### 部署步骤

1. **新增文件**
   ```bash
   pkg/db/global.go
   pkg/logic/tools/finance/common.go
   pkg/logic/tools/finance/account.go
   pkg/logic/tools/finance/position.go
   pkg/logic/tools/finance/trade.go
   ```

2. **修改文件**
   ```bash
   pkg/config/local_config.go  # 添加 db.InitGlobalDB()
   cmd/root.go                 # 添加 defer db.CloseGlobalDB()
   pkg/logic/tools/register.go # 添加 registerFinance()
   ```

3. **编译测试**
   ```bash
   go build
   ./msa
   ```

4. **验证初始化**
   ```bash
   # 检查日志
   tail -f ~/.msa/msa.log | grep SQLite
   ```

### 回滚策略

- 删除新增的文件
- 恢复修改的文件
- DB 文件独立存储，不影响原有功能

### 测试策略

1. **单元测试**：每个 tool 函数的测试
2. **集成测试**：DB 初始化、迁移测试
3. **手动测试**：通过 AI Agent 对话测试完整流程
