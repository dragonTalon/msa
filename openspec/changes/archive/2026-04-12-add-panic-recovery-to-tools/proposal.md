# 提案：add-panic-recovery-to-tools

为所有工具入口函数添加统一的 panic recovery 保护，防止工具内部异常导致会话中断。

## 知识上下文

### 相关历史问题

知识库中记录了以下相关问题：

- **005**：Google CAPTCHA 检测
  - 类型：integration | 严重程度：high
  - 组件：search tools
  - 摘要：Google 检测到自动化行为导致搜索失败，需要优雅处理

- **004**：浏览器超时时间过短
  - 类型：performance | 严重程度：medium
  - 组件：search/browser
  - 摘要：超时配置不当可能导致请求中断

### 需避免的反模式

⚠️ **警告**：已为此类变更记录了以下反模式：

- 当前代码中 **没有任何 panic recovery 机制**，一旦工具内部发生 panic，整个会话会崩溃
- **直接返回 error 而非包装为 JSON**：部分底层函数返回 error，但上层已正确处理

### 推荐模式

✅ **推荐**：考虑使用以下模式：

- **003**：错误处理策略
  - 描述：分类错误处理，根据错误类型采取不同的处理策略
  - 使用时机：所有工具入口函数

- **004**：浏览器管理器单例模式
  - 描述：浏览器实例复用，独立 Tab 隔离并发任务
  - 使用时机：search/fetcher 工具已有实现，可作为参考

## 风险评估

基于历史问题，此变更具有以下风险因素：

- **medium**：修改所有工具入口函数，可能引入新的 bug
  - 相关问题：无（新变更）
  - 缓解措施：统一辅助函数，减少重复代码

- **low**：panic recovery 可能掩盖真正的错误
  - 相关问题：无
  - 缓解措施：记录完整的 panic 日志，便于调试

## 缓解策略

1. 在 `pkg/logic/tools/basic.go` 中添加统一的 `SafeExecute` 辅助函数
2. 所有 panic 都记录完整日志（包括堆栈信息）
3. 返回统一的错误 JSON 格式，让 LLM 能理解并继续处理
4. 保持向后兼容：工具签名不变，只是增加保护层

## 为什么

当前问题：
- 搜索工具（web_search, fetch_page_content）涉及 HTML 解析和网页抓取，容易出现意外异常
- 其他工具虽然风险较低，但缺乏统一的保护机制
- 任何工具内部的 panic 都会导致整个会话中断，用户体验差

解决目标：
- 防止 panic 导致会话崩溃
- 提供统一的错误处理模式
- 增强系统稳定性

## 变更内容

- 添加 `SafeExecute` 辅助函数到 `pkg/logic/tools/basic.go`
- 包装所有工具入口函数（约 24 个）
- 统一的 panic 日志记录
- 统一的错误 JSON 返回格式

## 能力

### 新增能力

- `tool-error-protection`：统一的工具错误保护机制，防止 panic 中断会话

### 修改能力

无（此变更不改变现有工具的行为规范，只是增加保护层）

## 影响

### 受影响的文件

| 模块 | 文件 | 函数数量 |
|------|------|---------|
| basic | `pkg/logic/tools/basic.go` | +1 (新增辅助函数) |
| search | `search.go`, `fetcher.go` | 2 |
| stock | `company.go`, `company_info.go`, `company_k.go` | 3 |
| finance | `account.go`, `trade.go`, `position.go` | 8 |
| skill | `skill.go` | 3 |
| todo | `create.go`, `update.go`, `check.go`, `verify.go`, `summary.go` | 5 |
| knowledge | `read.go`, `write.go` | 2 |

**总计约 24 个入口函数需要添加保护**

### 依赖和系统影响

- 无外部依赖变化
- 无 API 变化（函数签名保持不变）
- 对 LLM 透明：返回格式不变

---

**知识备注**：
- 此提案使用 `openspec/knowledge/` 中的知识生成
- 最后知识同步：2026-02-23
- 知识库总问题数：5
- 总模式数：6