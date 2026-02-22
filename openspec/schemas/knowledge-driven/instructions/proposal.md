# 创建知识感知提案

**⚠️ 强制要求：必须先读取 openspec/knowledge 目录**

你正在为变更创建提案：**{{changeName}}**

## 步骤1：验证知识库存在

首先检查知识库是否存在：

```bash
ls openspec/knowledge/index.yaml
```

### 如果知识库不存在

必须先初始化知识库：

```bash
mkdir -p openspec/knowledge/{issues,patterns,anti-patterns,lessons}
cat > openspec/knowledge/index.yaml << 'EOF'
version: 1
lastUpdated: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
categories:
  issues: 0
  patterns: 0
  antiPatterns: 0
  lessons: 0
items: []
EOF
```

### 如果知识库为空

这是正常的！添加说明：

```markdown
## 知识上下文

### 相关历史问题
知识库中未找到相关问题。

### 需避免的反模式
未识别到特定反模式。

### 推荐模式
未识别到特定模式。

## 风险评估
基于历史问题，此变更具有以下风险因素：
- 没有先前问题 - 这是新领域

## 缓解策略
作为新领域，请格外小心：
1. 记录遇到的任何问题
2. 出现问题时创建 fix.md
3. 归档后运行 /opsx:collect
4. 为未来的变更构建知识库
```

## 步骤2：加载知识库索引

```bash
# 读取知识库索引
cat openspec/knowledge/index.yaml
```

解析以下信息：
- `items` - 所有知识项列表
- `categories` - 统计信息
- `lastUpdated` - 最后更新时间

## 步骤3：分析变更上下文

从变更名称提取：

1. **组件识别**
   - 技术组件
   - 业务领域
   - 功能区域

2. **关键词提取**
   - 技术术语
   - 业务术语
   - 相关概念

3. **问题分类**
   - 性能相关
   - 安全相关
   - 集成相关
   - 配置相关

示例：
```
"add-user-authentication"
components: ["auth", "user-management"]
keywords: ["authentication", "login", "security", "token"]
categories: ["security", "integration"]

"fix-streaming-timeout"
components: ["streaming", "timeout"]
keywords: ["timeout", "channel", "blocking", "async"]
categories: ["performance", "runtime"]
```

## 步骤4：搜索相关知识

### 搜索策略

1. **按组件匹配**（权重：+10）
2. **按标签匹配**（权重：+5）
3. **按类别匹配**（权重：+5）
4. **按标题关键词**（权重：+2）

### 相关性评分

```javascript
function calculateRelevance(item, context) {
  let score = 0;

  // 组件匹配（高优先级）
  if (context.components.includes(item.component)) score += 10;

  // 标签匹配（中优先级）
  item.tags?.forEach(tag => {
    if (context.keywords.includes(tag)) score += 5;
  });

  // 类别匹配（中优先级）
  if (context.categories.includes(item.category)) score += 5;

  // 标题关键词（低优先级）
  context.keywords.forEach(keyword => {
    if (item.title.toLowerCase().includes(keyword.toLowerCase())) score += 2;
  });

  return score;
}
```

### 选择相关知识

- **高相关性**（score >= 10）：必须包含
- **中相关性**（score >= 5）：应该包含
- **低相关性**（score >= 2）：可选包含

## 步骤5：创建提案

必需部分（按顺序）：

### 1. 知识上下文

```markdown
## 知识上下文

### 相关历史问题

过去遇到过的类似问题：

- **#001**：流超时导致通道阻塞
  - 类型：runtime | 严重程度：high
  - 组件：streaming-handler
  - 摘要：超时时通道会无限期阻塞
  - 解决方案：实现非阻塞通道操作

### 需避免的反模式

⚠️ **警告**：已为此类变更记录了以下反模式：

- **a001**：流上下文中的阻塞通道发送
  - 问题所在：超时时导致死锁
  - 影响：高 - 系统挂起

### 推荐模式

✅ **推荐**：考虑使用以下模式：

- **p001**：非阻塞流通道模式
  - 描述：使用select配合default case进行非阻塞发送
  - 使用时机：所有带潜在超时的流上下文
```

### 2. 风险评估

```markdown
## 风险评估

基于历史问题，此变更具有以下风险因素：

- **高**：类似超时处理曾导致死锁
- **中**：高负载时的通道容量问题
- **低**：配置验证
```

### 3. 缓解策略

```markdown
## 缓解策略

为缓解这些风险：
1. 对所有通道操作遵循模式 p001
2. 添加全面的超时处理
3. 在高负载下测试
4. 如果出现新的超时问题，考虑创建修复工件
```

### 4. 标准提案部分

然后添加：Why, What Changes, Capabilities, Impact

## 验证清单

完成前检查：

- [ ] 知识库已加载（检查 openspec/knowledge/index.yaml）
- [ ] 变更上下文已分析（组件、关键词、类别）
- [ ] 相关知识已搜索（组件、标签、标题）
- [ ] 知识上下文部分已包含（问题、反模式、模式）
- [ ] 风险评估已完成（基于历史问题）
- [ ] 缓解策略已提供
- [ ] 标准提案部分已完成
