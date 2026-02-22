# 创建修复工件

当遇到意外错误、bug、性能问题、配置问题或设计缺陷时创建 `fix.md`。

## 快速开始

```bash
/opsx:continue
# 选择创建 fix.md
```

## 前置信息

```yaml
---
id: "003"
title: "描述性标题"
type: "runtime"              # runtime|compile|logic|config
severity: "high"             # critical|high|medium|low
component: "component-name"
tags:
  - "keyword1"
  - "component-name"
created_at: "2025-02-23T10:00:00Z"
resolved_at: "2025-02-23T11:30:00Z"
change: "changes/your-change"
status: "resolved"           # open|investigating|resolved|cancelled
---
```

## 必需部分

### 问题

- 发生了什么？
- 应该发生什么？
- 重现步骤
- 错误消息和堆栈跟踪
- 影响评估（用户/系统/业务）

### 根本原因

- 调查过程
- 根本原因陈述
- 促成因素
- 带行号的相关文件

### 解决方案

- 解决方案方法
- 考虑的替代方案
- 代码变更（使用diff格式）

```markdown
#### 添加超时上下文

**文件**：`src/handlers/streaming.ts`

```diff
+ export function createStreamingContext(timeout: number) {
+   const controller = new AbortController();
+   setTimeout(() => controller.abort(), timeout);
+   return { signal: controller.signal };
+ }
```

**说明**：添加了超时处理的abort controller
**原因**：允许超时时适当清理
```

### 验证

- 如何验证（步骤）
- 测试结果

## 关键：模式提取

这构建知识库！

### 可复用模式（如适用）

如果解决方案可泛化：

```markdown
## 可复用模式

**模式名称**：带超时的非阻塞通道发送
**模式ID**：p005

**描述**：
使用非阻塞操作配合超时处理以防止死锁。

**适用场景**：
- 所有流式操作
- 任何异步上下文中的通道发送

**实现**：
```typescript
function sendWithTimeout<T>(
  channel: Channel<T>,
  value: T,
  timeout: number
): boolean {
  const result = channel.trySend(value);
  if (!result) {
    throw new Error(`通道发送在 ${timeout}ms 后超时`);
  }
  return result;
}
```
```

### 反模式（如适用）

```markdown
## 反模式

**反模式名称**：异步上下文中的阻塞通道发送
**反模式ID**：a003

**问题所在**：
异步上下文中的阻塞发送在缓冲区满时会导致死锁。

**避免**：
```typescript
// 不要这样做
channel.send(value);  // 可能无限期阻塞
```
```

## 经验教训

```markdown
## 经验教训

### 出了什么问题
- 假设通道永远不会满
- 没有考虑超时与阻塞发送的交互
- 缺少超时场景的测试

### 学到了什么
- 在异步代码中总是考虑通道容量
- 超时处理需要非阻塞操作
- 需要边缘情况的全面测试

### 预防措施
1. 添加通道容量假设的验证
2. 在所有异步上下文中使用非阻塞操作
3. 添加超时和溢出场景的测试
4. 记录通道行为模式
```

## 质量清单

- [ ] 清晰的描述性标题
- [ ] 完整的前置信息
- [ ] 清楚描述问题
- [ ] 重现步骤
- [ ] 错误消息/日志
- [ ] 影响评估
- [ ] 识别根本原因
- [ ] 链接相关文件
- [ ] 记录解决方案
- [ ] 验证步骤
- [ ] 测试结果
- [ ] 提取模式（如适用）
- [ ] 记录反模式（如适用）
- [ ] 记录经验教训

## 创建修复后

1. 完成修复
2. 彻底验证
3. 归档：`/opsx:archive`
4. 收集知识：`/opsx:collect`

修复将自动处理到知识库！

## 常见错误

❌ **太模糊**："修复了一个bug"
✅ **具体**："通过添加非阻塞发送修复通道阻塞问题"

❌ **没有根本原因**："不工作"
✅ **根本原因**："Channel.send() 阻塞因为缓冲区已满"

❌ **没有验证**："现在应该工作"
✅ **已验证**："测试了1000个并发请求，无死锁"

❌ **没有模式**："只是修复了它"
✅ **提取模式**："为非阻塞发送创建模式 p005"
