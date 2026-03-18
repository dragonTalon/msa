# 实现任务：enhance-trading-system

## 1. Session 标签自动添加

> 📚 **知识上下文**：基于设计决策 1，在 `Manager.Initialize()` 中自动添加时间标签

- [x] 1.1 在 `pkg/logic/memory/manager.go` 中添加时间常量
  - 定义 `MorningStart = "09:30"`, `MorningEnd = "11:30"` 等
  - 验证：常量格式为 "15:04"

- [x] 1.2 实现 `getAutoTag()` 辅助函数
  - 输入：`time.Time`
  - 输出：对应的时间标签字符串
  - 验证：各时间段返回正确标签

- [x] 1.3 修改 `Initialize()` 方法
  - 在创建 SessionManager 后调用 `getAutoTag()`
  - 自动添加标签到 `m.currentSession.Tags`
  - 验证：日志输出"自动添加会话标签: xxx"

- [x] 1.4 编写单元测试
  - 测试早盘、午盘、收盘、非交易时段
  - 验证：`go test ./pkg/logic/memory/...`

## 2. 交易分析工具

> 📚 **知识上下文**：基于 Pattern #003 错误处理和重试策略

- [x] 2.1 创建 `pkg/logic/tools/knowledge/analyze.go` 文件
  - 定义 `AnalyzeTodayTradesParam` 参数结构
  - 定义 `PotentialError` 和 `AnalysisResult` 结果结构
  - 验证：结构体包含必要的 JSON 标签

- [x] 2.2 实现 `AnalyzeTodayTradesTool` 工具接口
  - 实现 `GetToolInfo()` 方法
  - 实现 `GetName()` 返回 "analyze_today_trades"
  - 实现 `GetDescription()` 返回中文描述
  - 实现 `GetToolGroup()` 返回 `model.KnowledgeToolGroup`
  - 验证：工具注册正确

- [x] 2.3 实现错误识别逻辑
  - 识别追高买入（涨幅 > 3%）
  - 识别仓位过重（单次 > 50%）
  - 识别未及时止损（亏损 > 20%）
  - 识别同日过度买入（累计 > 70%）
  - 验证：每种错误类型有对应测试用例

- [x] 2.4 实现结果输出
  - 输出结构化的 `AnalysisResult`
  - 包含错误模板供模型使用
  - 验证：输出可被 JSON 序列化

- [x] 2.5 注册工具到工具组
  - 在 `pkg/logic/tools/knowledge/` 的注册文件中添加
  - 验证：`go vet ./...` 无错误

- [x] 2.6 编写单元测试
  - 测试正常情况、无交易、数据获取失败
  - 验证：`go test ./pkg/logic/tools/knowledge/...`

## 3. SKILL 文档更新

### 3.1 trading-common SKILL

- [x] 3.1.1 新增"仓位管理规则"章节
  - 计划仓位计算公式
  - 买入限制参数表格
  - 验证：文档格式正确

- [x] 3.1.2 新增"分段买入策略"章节
  - 三种分段方式详细说明
  - 选择依据说明
  - 验证：表格格式清晰

- [x] 3.1.3 新增"分段卖出策略"章节
  - 正常分批卖出规则
  - 止损例外规则
  - 验证：止损条件明确（亏损 > 20%）

### 3.2 morning-analysis SKILL

- [x] 3.2.1 移除手动标签说明
  - 删除"为当前Session添加标签：morning-session"
  - 验证：文档中无手动标签说明

- [x] 3.2.2 引用仓位策略
  - 添加对 trading-common 的引用
  - 验证：引用格式正确

### 3.3 afternoon-trade SKILL

- [x] 3.3.1 移除手动标签说明
  - 删除"为当前Session添加标签：afternoon-session"
  - 验证：文档中无手动标签说明

- [x] 3.3.2 强化早盘读取说明
  - 明确使用 `query_sessions_by_date(today, "morning-session")`
  - 验证：说明清晰易懂

### 3.4 market-close-summary SKILL

- [x] 3.4.1 移除手动标签说明
  - 删除"为当前Session添加标签：close-session"
  - 验证：文档中无手动标签说明

- [x] 3.4.2 新增"错误判断标准"章节
  - 5 种必须记录的错误类型
  - 验证：每种错误有明确判断标准

- [x] 3.4.3 新增"错误检查清单"
  - 5 项检查项
  - 验证：清单格式为复选框

- [x] 3.4.4 引用 analyze_today_trades 工具
  - 说明工具用途和调用方式
  - 验证：工具名称正确

## 4. 验证和测试

- [x] 4.1 代码编译验证
  - 运行 `go vet ./...`
  - 验证：无错误输出

- [x] 4.2 单元测试
  - 运行 `go test ./pkg/logic/memory/... ./pkg/logic/tools/knowledge/...`
  - 验证：所有测试通过

- [ ] 4.3 集成测试（手动）
  - 启动应用，验证 Session 标签自动添加
  - 验证：日志显示"自动添加会话标签"

## 基于知识的测试

### 时间判断回归测试

> 📚 **来源**：设计文档 - 避免的反模式

**历史问题**：
硬编码时间判断可能导致时区问题。

**测试场景**：
1. 在不同时区运行程序
2. 检查 Session 标签是否正确
3. 验证边界时间（9:30, 11:30, 13:00, 14:30, 16:00）

**预期行为**：
- 9:30-11:30 → `morning-session`
- 13:00-14:30 → `afternoon-session`
- 16:00+ → `close-session`
- 其他时间 → 无标签

### 工具输出回归测试

> 📚 **来源**：设计文档 - 避免的反模式

**历史问题**：
工具返回信息过于简单，模型无法理解。

**测试场景**：
1. 调用 analyze_today_trades
2. 检查返回结果结构
3. 验证包含详细错误信息

**预期行为**：
- 返回结构化的 `AnalysisResult`
- 每个错误包含 type, stock_code, stock_name, detail, suggestion
- 包含 error_template 字段
