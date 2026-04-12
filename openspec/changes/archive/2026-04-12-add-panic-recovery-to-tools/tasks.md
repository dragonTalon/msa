# 实现任务：add-panic-recovery-to-tools

## ⚠️ 风险警报

基于历史问题识别的风险：

### medium：修改 24 个工具入口函数可能引入 bug

- **相关问题**：无（新变更）
- **缓解措施**：统一辅助函数减少重复代码；逐模块修改并测试

### low：Panic recovery 可能掩盖真正的错误

- **相关问题**：知识库反模式
- **缓解措施**：记录完整堆栈信息；使用 ERROR 级别日志

### high：高风险工具（search/fetcher）已知会出现异常

- **相关问题**：#005 Google CAPTCHA 检测、#004 浏览器超时
- **缓解措施**：优先处理这些工具；添加针对性测试

---

## 1. 基础设施

### [T-1] 添加 SafeExecute 辅助函数

> 📚 **知识上下文**：
> 来源：模式 #003 - 错误处理策略
> 经验：分类错误处理，统一格式返回

**文件**：`pkg/logic/tools/basic.go`

- [x] 1.1 添加 `SafeExecute` 泛型辅助函数
- [x] 1.2 添加 `SafeExecuteWithContext` 版本（带 context）
- [x] 1.3 使用 `runtime/debug.Stack()` 获取堆栈
- [x] 1.4 调用 `message.BroadcastToolEnd` 通知 UI
- [x] 1.5 返回 `model.NewErrorResult` 格式化错误

**验收标准**：
- [x] 函数签名兼容所有工具
- [x] Panic 被正确捕获
- [x] 日志包含完整堆栈

---

## 2. 高优先级工具（搜索模块）

### [T-2] 保护 WebSearch 工具

> 📚 **知识上下文**：
> 来源：问题 #005 - Google CAPTCHA 检测
> 经验：搜索工具容易出现异常，需要优雅处理

**文件**：`pkg/logic/tools/search/search.go`

- [x] 2.1 使用 SafeExecute 包装 WebSearch 函数
- [x] 2.2 保持原有参数校验逻辑
- [x] 2.3 确保错误消息对 LLM 友好

### [T-3] 保护 FetchPageContent 工具

> 📚 **知识上下文**：
> 来源：问题 #004 - 浏览器超时配置
> 经验：HTML 解析和网页抓取容易触发异常

**文件**：`pkg/logic/tools/search/fetcher.go`

- [x] 3.1 使用 SafeExecute 包装 FetchPageContent 函数
- [x] 3.2 保持原有错误处理逻辑

---

## 3. 中优先级工具（股票模块）

### [T-4] 保护 stock 工具

**文件**：`pkg/logic/tools/stock/company.go`, `company_info.go`, `company_k.go`

- [x] 4.1 保护 GetStockCompanyCode
- [x] 4.2 保护 GetStockCompanyInfo
- [x] 4.3 保护 GetStockCompanyK

---

## 4. 低优先级工具（金融模块）

### [T-5] 保护 finance 工具

**文件**：`pkg/logic/tools/finance/account.go`, `trade.go`, `position.go`

- [x] 5.1 保护 CreateAccount
- [x] 5.2 保护 GetAccount
- [x] 5.3 保护 UpdateAccountStatus
- [x] 5.4 保护 SubmitBuyOrder
- [x] 5.5 保护 SubmitSellOrder
- [x] 5.6 保护 GetTransactions
- [x] 5.7 保护 GetPositions
- [x] 5.8 保护 GetAccountSummary

---

## 5. 低优先级工具（技能模块）

### [T-6] 保护 skill 工具

**文件**：`pkg/logic/tools/skill/skill.go`

- [x] 6.1 保护 GetSkillContent
- [x] 6.2 保护 GetSkillReference
- [x] 6.3 保护 GetSkillAsset

---

## 6. 低优先级工具（TODO 模块）

### [T-7] 保护 todo 工具

**文件**：`pkg/logic/tools/todo/*.go`

- [x] 7.1 保护 CreateTodo
- [x] 7.2 保护 UpdateTodoStep
- [x] 7.3 保护 CheckSkillTodo
- [x] 7.4 保护 VerifyTodoCompletion
- [x] 7.5 保护 FillTodoSummary

---

## 7. 低优先级工具（知识库模块）

### [T-8] 保护 knowledge 工具

**文件**：`pkg/logic/tools/knowledge/read.go`, `write.go`

- [x] 8.1 保护 ReadKnowledge
- [x] 8.2 保护 WriteKnowledge

---

## 8. 测试与验证

### [T-9] 添加单元测试

- [x] 9.1 为 SafeExecute 添加单元测试
- [x] 9.2 测试 panic 恢复场景
- [x] 9.3 测试正常执行场景
- [x] 9.4 测试日志输出

### [T-10] 验证编译

- [x] 10.1 运行 `go vet ./...` 验证编译问题
- [x] 10.2 运行现有测试确保无回归
- [ ] 10.3 手动测试 search 工具

---

## 基于知识的测试

### 问题 #005 回归测试

> 📚 **来源**：问题 #005 - Google CAPTCHA 检测

**历史问题**：
Google 检测到自动化行为导致搜索失败，需要优雅处理。

**测试场景**：
1. 模拟搜索工具内部 panic
2. 验证会话不崩溃
3. 验证返回错误 JSON
4. 验证日志记录完整

**预期行为**：
- Panic 被捕获
- 返回 `{"success": false, "error_msg": "工具内部错误: ..."}`
- 会话继续运行

### 问题 #004 回归测试

> 📚 **来源**：问题 #004 - 浏览器超时配置

**历史问题**：
超时配置不当可能导致请求中断。

**测试场景**：
1. 模拟 fetcher 工具超时导致的 panic
2. 验证 panic recovery 生效
3. 验证用户能看到友好的错误提示

**预期行为**：
- Panic 被捕获
- 用户看到工具失败提示
- 会话可以继续

---

## 验收清单

- [x] 所有 24 个工具入口函数都有 panic recovery 保护（实际 23 个）
- [x] SafeExecute 辅助函数实现完成
- [x] 单元测试覆盖 panic 恢复场景
- [x] `go vet ./...` 无警告
- [x] 现有测试全部通过
- [ ] 手动测试 search 工具正常工作
