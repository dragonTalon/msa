## 1. 基础设施

- [x] 1.1 创建 `pkg/model/tool_result.go`，定义 `ToolResult` 结构体
- [x] 1.2 实现 `NewSuccessResult(data interface{}, message string) string` 函数
- [x] 1.3 实现 `NewErrorResult(message string) string` 函数
- [x] 1.4 编写 `tool_result.go` 的单元测试

## 2. Search 模块改造

- [x] 2.1 修改 `pkg/model/search.go`，将 `WebSearchResponse` 改为使用 `ToolResult` 格式
- [x] 2.2 修改 `pkg/model/search.go`，将 `FetchPageResponse` 改为使用 `ToolResult` 格式
- [x] 2.3 修改 `pkg/logic/tools/search/search.go`，使用新的响应结构
- [x] 2.4 修改 `pkg/logic/tools/search/fetcher.go`，使用新的响应结构
- [x] 2.5 运行 search 模块测试验证

## 3. Finance 模块改造

- [x] 3.1 修改 `pkg/logic/tools/finance/account.go`（CreateAccount, GetAccount, UpdateAccountStatus）
- [x] 3.2 修改 `pkg/logic/tools/finance/trade.go`（SubmitBuyOrder, SubmitSellOrder, GetTransactions）
- [x] 3.3 修改 `pkg/logic/tools/finance/position.go`（GetPositions, GetAccountSummary）
- [x] 3.4 运行 finance 模块测试验证

## 4. Stock 模块改造

- [x] 4.1 修改 `pkg/logic/tools/stock/company.go`（GetStockCompanyCode）
- [x] 4.2 修改 `pkg/logic/tools/stock/company_info.go`（GetStockCurrentQuote）
- [x] 4.3 修改 `pkg/logic/tools/stock/company_k.go`（GetStockCompanyK）
- [x] 4.4 运行 stock 模块测试验证

## 5. Knowledge 模块改造

- [x] 5.1 修改 `pkg/logic/tools/knowledge/read.go`（ReadKnowledgeFile）
- [x] 5.2 修改 `pkg/logic/tools/knowledge/write.go`（WriteKnowledgeFile）
- [x] 5.3 修改 `pkg/logic/tools/knowledge/list.go`（ListKnowledgeFiles）
- [x] 5.4 修改 `pkg/logic/tools/knowledge/session.go`（QuerySessionsByDate, AddSessionTag）
- [x] 5.5 运行 knowledge 模块测试验证

## 6. Skill 模块改造

- [x] 6.1 修改 `pkg/logic/tools/skill/skill.go`（GetSkillContent）
- [x] 6.2 运行 skill 模块测试验证

## 7. 验证与清理

- [x] 7.1 运行 `go vet ./...` 验证编译问题
- [x] 7.2 运行所有单元测试
- [ ] 7.3 手动测试 Agent 调用 tool 的错误处理流程
