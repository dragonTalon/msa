# 创建知识感知设计

你正在为变更创建设计：**{{changeName}}**

## 上下文

你的设计应该：
1. 应用知识库中推荐的模式
2. 避免记录的反模式
3. 主动解决历史问题
4. 考虑过去导致问题的边缘情况

## 步骤1：应用推荐模式

```markdown
## 推荐模式

✅ 基于知识库，推荐以下模式：

### 非阻塞流通道模式

**模式ID**：p001
**来源问题**：#001 - 流超时导致通道阻塞

**描述**：
使用非阻塞通道操作配合select语句，优雅处理超时而不死锁。

**适用场景**：
- 所有流式操作
- 带超时的异步操作
- 生产者-消费者模式

**实现**：
```typescript
async function sendWithTimeout<T>(
  channel: Channel<T>,
  value: T,
  timeout: number
): Promise<boolean> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  try {
    const result = await select(
      [channel.send(value)],
      [controller.signal]
    );
    return result === 0;
  } finally {
    clearTimeout(timeoutId);
  }
}
```

**收益**：
- 防止死锁
- 优雅的超时处理
- 清晰的错误消息
- 可预测的行为
```

## 步骤2：避免反模式

```markdown
## 避免的反模式

⚠️ 已记录以下反模式：

### 流上下文中的阻塞通道发送

**反模式ID**：a001
**来源问题**：#001 - 流超时导致通道阻塞

**问题所在**：
当以下情况时阻塞发送会导致死锁：
- 通道缓冲区已满
- 发送期间超时
- 多个操作同时阻塞

**影响**：
- 高严重性 - 系统挂起
- 难以调试
- 影响所有连接的操作

**不要这样做**：
```typescript
// 不要这样做
function sendValue<T>(channel: Channel<T>, value: T) {
  channel.send(value);  // 如果满了会无限期阻塞！
}
```

**应该这样做**：
```typescript
// 应该这样做
async function sendValue<T>(channel: Channel<T>, value: T) {
  const sent = await channel.trySend(value);
  if (!sent) {
    throw new Error('通道已满 - 需要背压处理');
  }
}
```
```

## 步骤3：组件设计应用模式

```markdown
## 组件

### StreamingHandler

> 📚 **知识备注**：此组件解决问题 #001，其中超时处理
> 导致通道阻塞。

**关键决策**（来自模式 p001）：
1. 使用非阻塞通道操作
2. 实现取消的abort controller
3. 添加超时监控
4. 提供清晰的错误消息

**接口**：
```typescript
interface StreamingHandler {
  withTimeout(ms: number): StreamingHandler;
  stream<T>(source: DataSource<T>): Promise<void>;
  cancel(): void;
}
```
```

## 步骤4：错误处理

应用历史问题的教训：

```markdown
## 错误处理

### 超时错误

> 📚 **历史上下文**：来自问题 #001

**以前的问题**：
超时导致通道阻塞，引起级联故障。

**改进的错误处理**：
```typescript
class TimeoutError extends Error {
  constructor(
    public operation: string,
    public timeout: number,
    public elapsed: number
  ) {
    super(`操作 '${operation}' 在 ${elapsed}ms 后超时（限制：${timeout}ms）`);
  }
}

// 处理器确保清理
function handleTimeout(error: TimeoutError) {
  abortController.abort();
  channel.close();
  cleanup();
  logger.error('超时发生', { operation, timeout, elapsed });
  notify(error);
}
```
```

## 步骤5：测试策略

包含历史问题的测试：

```markdown
## 测试

### 单元测试

**测试模式应用**：
```typescript
describe('非阻塞发送（模式 p001）', () => {
  it('应该无阻塞地处理满通道', async () => {
    const channel = new Channel<string>(1);
    channel.send('first');

    await expect(
      sendWithTimeout(channel, 'second', 100)
    ).rejects.toThrow('通道已满');
  });
});
```

**测试反模式预防**：
```typescript
describe('反模式 a001 预防', () => {
  it('应该永远不无限期阻塞', async () => {
    const timeout = 100;
    const start = Date.now();

    try {
      await sendWithTimeout(blockingChannel, 'value', timeout);
    } catch (e) { /* 预期 */ }

    const elapsed = Date.now() - start;
    expect(elapsed).toBeLessThan(timeout * 2);
  });
});
```

### 回归测试

```typescript
describe('问题 #001 回归测试', () => {
  it('应该无死锁地处理超时', async () => {
    const handler = new StreamingHandler({ timeout: 100 });
    const result = await handler.stream(slowSource);
    expect(result.status).toBe('timeout');
    expect(handler.isBlocked()).toBe(false);
  });
});
```
```

## 验证清单

- [ ] 应用了推荐模式
- [ ] 明确避免了反模式
- [ ] 解决了历史问题
- [ ] 考虑了边缘情况
- [ ] 从过去改进了错误处理
- [ ] 包含历史问题的测试
- [ ] 为过去问题添加了监控
- [ ] 记录了知识引用
