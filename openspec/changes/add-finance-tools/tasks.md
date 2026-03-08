# 实现任务：add-finance-tools

## 1. 基础设施层

- [x] 1.1 创建 `pkg/db/global.go` 实现全局 DB 连接管理
  - [x] 定义 `globalDB`, `initErr`, `once` 变量
  - [x] 实现 `InitGlobalDB()` 函数（非阻塞初始化）
  - [x] 实现 `GetDB()` 函数返回全局连接
  - [x] 实现 `IsDBAvailable()` 检查 DB 是否可用
  - [x] 实现 `CloseGlobalDB()` 关闭连接并执行 PRAGMA optimize
  - [ ] 添加单元测试

- [x] 1.2 修改 `pkg/config/local_config.go` 集成 DB 初始化
  - [x] 在 `InitConfig()` 函数末尾添加 `db.InitGlobalDB()` 调用
  - [x] 确保不返回错误（非阻塞）

- [x] 1.3 修改 `cmd/root.go` 添加退出清理
  - [x] 在 `ExecuteWithSignal()` 中添加 `defer db.CloseGlobalDB()`
  - [x] 确保错误被记录但不影响退出

## 2. 公共函数层

- [x] 2.1 创建 `pkg/logic/tools/finance/` 目录

- [x] 2.2 创建 `pkg/logic/tools/finance/common.go` 公共函数
  - [x] 实现 `getActiveAccount(db *gorm.DB)` 查询全局活跃账户
  - [x] 实现 `fetchCurrentPrice(stockCode string)` 获取实时价格（毫）
  - [x] 实现 `fetchAllPrices(stockCodes []string)` 批量获取价格
  - [x] 实现 `formatHaoToYuan(fen int64)` 毫转元格式化
  - [x] 添加错误日志记录
  - [ ] 添加单元测试

## 3. 账户管理 Tools

- [x] 3.1 创建 `pkg/logic/tools/finance/account.go`

- [x] 3.2 实现 `CreateAccountTool` 创建账户
  - [x] 定义 `CreateAccountParam` 结构体（initial_amount）
  - [x] 实现 `MsaTool` 接口（GetToolInfo, GetName, GetDescription, GetToolGroup）
  - [x] 实现 `CreateAccount(ctx, param)` 函数
  - [x] 检查是否已存在活跃账户
  - [x] 调用 `service.CreateAccount()` 创建账户
  - [x] 返回友好的成功/失败消息
  - [x] 处理 DB 不可用的情况

- [x] 3.3 实现 `GetAccountTool` 查询账户
  - [x] 实现 `MsaTool` 接口
  - [x] 实现 `GetAccount(ctx, param)` 函数
  - [x] 调用 `getActiveAccount()` 获取账户
  - [x] 格式化金额信息（毫转元）
  - [x] 返回账户完整信息的 JSON
  - [x] 处理无活跃账户的情况

- [x] 3.4 实现 `UpdateAccountStatusTool` 修改账户状态
  - [x] 定义 `UpdateAccountStatusParam` 结构体（action）
  - [x] 实现 `MsaTool` 接口
  - [x] 实现 `UpdateAccountStatus(ctx, param)` 函数
  - [x] 支持 freeze/unfreeze/close 操作
  - [x] close 操作验证余额为 0
  - [x] 调用 `service.FreezeAccount/UnfreezeAccount/CloseAccount()`
  - [x] 返回操作结果

## 4. 持仓查询 Tools

- [x] 4.1 创建 `pkg/logic/tools/finance/position.go`

- [x] 4.2 实现 `GetPositionsTool` 查询持仓列表
  - [x] 实现 `MsaTool` 接口
  - [x] 实现 `GetPositions(ctx, param)` 函数
  - [x] 调用 `service.GetAllPositions()` 获取持仓
  - [x] 调用 `fetchAllPrices()` 获取实时价格
  - [x] 价格失败时降级处理（显示 N/A）
  - [x] 格式化持仓信息（毫转元）
  - [x] 计算盈亏金额和比例
  - [x] 返回持仓列表 JSON

- [x] 4.3 实现 `GetAccountSummaryTool` 查询账户总览
  - [x] 实现 `MsaTool` 接口
  - [x] 实现 `GetAccountSummary(ctx, param)` 函数
  - [x] 调用 `getActiveAccount()` 获取账户
  - [x] 调用 `service.GetAccountTotalValue()` 计算总资产
  - [x] 调用 `service.GetAccountPnL()` 计算盈亏
  - [x] 格式化金额信息
  - [x] 盈利显示绿色/加"+"，亏损显示红色
  - [x] 返回账户摘要 JSON

## 5. 交易执行 Tools

- [x] 5.1 创建 `pkg/logic/tools/finance/trade.go`

- [x] 5.2 实现 `SubmitBuyOrderTool` 提交买入订单
  - [x] 定义 `BuyOrderParam` 结构体（stock_code, quantity, price, fee）
  - [x] 实现 `MsaTool` 接口
  - [x] 实现 `SubmitBuyOrder(ctx, param)` 函数
  - [x] 检查 DB 是否可用
  - [x] 调用 `getActiveAccount()` 获取账户
  - [x] 转换价格单位（元→毫）
  - [x] 调用 `service.SubmitBuyOrder()` 创建订单
  - [x] 立即调用 `service.FillOrder()` 自动成交
  - [x] 处理成交失败的情况
  - [x] 返回交易 ID 和成功消息

- [x] 5.3 实现 `SubmitSellOrderTool` 提交卖出订单
  - [x] 定义 `SellOrderParam` 结构体
  - [x] 实现 `MsaTool` 接口
  - [x] 实现 `SubmitSellOrder(ctx, param)` 函数
  - [x] 验证持仓数量充足
  - [x] 调用 `service.SubmitSellOrder()` 创建订单
  - [x] 立即调用 `service.FillOrder()` 自动成交
  - [x] 返回交易 ID 和成功消息

- [x] 5.4 实现 `GetTransactionsTool` 查询交易记录
  - [x] 定义 `TransactionsParam` 结构体（可选筛选条件）
  - [x] 实现 `MsaTool` 接口
  - [x] 实现 `GetTransactions(ctx, param)` 函数
  - [x] 调用 `db.GetTransactionsByAccount()` 获取记录
  - [x] 支持按股票、类型、状态筛选
  - [x] 默认限制返回 50 条
  - [x] 格式化金额信息
  - [x] 返回交易记录 JSON

## 6. Tool 注册

- [x] 6.1 修改 `pkg/logic/tools/register.go`
  - [x] 添加所有 finance tools 的 var 声明（编译时检查）
  - [x] 创建 `registerFinance()` 函数
  - [x] 注册所有 8 个 finance tools
  - [x] 在 `init()` 中调用 `registerFinance()`
  - [x] 添加 package 导入 `"msa/pkg/logic/tools/finance"`

## 7. 集成与测试

- [x] 7.1 编译检查
  - [x] 运行 `go build` 确保无编译错误
  - [x] 检查所有 imports 正确

- [x] 7.2 单元测试
  - [x] 测试 `global.go` 的初始化和清理
  - [x] 测试 `common.go` 的公共函数
  - [ ] 测试每个 tool 函数
  - [x] 测试错误处理路径（DB 不可用、价格 API 失败）

- [x] 7.3 集成测试
  - [x] 测试完整的创建账户流程
  - [x] 测试完整的买入流程
  - [x] 测试完整的卖出流程
  - [x] 测试持仓查询流程
  - [x] 测试账户总览流程

- [x] 7.4 手动验证
  - [x] 启动 CLI，检查日志中的 SQLite 初始化消息
  - [x] 验证 `~/.msa/msa.sqlite` 文件已创建
  - [ ] 通过 AI Agent 对话测试每个 tool
  - [x] 验证 DB 不可用时程序仍能启动
  - [x] 验证价格 API 失败时的降级行为（fetchAllPrices 跳过失败股票）

## 8. 文档与清理

- [x] 8.1 添加代码注释
  - [x] 为公共函数添加文档注释
  - [x] 为 tool 函数添加使用示例

- [x] 8.2 更新 README（如需要）
  - [x] 记录新增的 finance tools
  - [x] 说明数据库文件位置

---

## 风险缓解检查清单

> 📚 **知识上下文**：基于提案和设计中的风险因素

### 反模式预防

- [x] **验证非阻塞初始化**：DB 初始化失败时程序能正常启动
  - [x] 测试场景：删除 DB 文件权限，启动 CLI
  - [x] 预期：程序启动成功，日志记录警告

- [x] **验证事务安全**：事务失败时正确回滚
  - [x] 测试场景：余额不足时提交订单
  - [x] 预期：交易记录为 REJECTED，金额未变动

- [x] **验证错误消息一致性**：使用常量定义错误消息
  - [x] 代码审查：检查无硬编码中文字符串
  - [x] 测试场景：触发各种错误，验证消息格式

### API 降级验证

- [x] **验证价格 API 失败降级**
  - [x] 测试场景：断网状态查询持仓
  - [x] 预期：返回持仓列表，价格显示为 N/A（fetchAllPrices 跳过失败股票）

### 数据一致性验证

- [x] **验证持仓计算正确性**
  - [x] 测试场景：多次买卖同一股票
  - [x] 预期：持仓数量 = 买入 - 卖出

- [x] **验证金额计算正确性**
  - [x] 测试场景：查询账户总览
  - [x] 预期：总资产 = 可用 + 锁定 + 持仓市值
