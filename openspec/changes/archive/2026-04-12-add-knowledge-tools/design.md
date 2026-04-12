# 设计：add-knowledge-tools

## 上下文

MSA 系统需要为交易 Skill 提供一个简单的知识库系统，用于存储历史错误和每日总结。

**当前状态**：
- `pkg/logic/tools/knowledge/` 目录不存在
- `read-error` SKILL 存在但设计不够灵活
- `morning-analysis`、`afternoon-trade`、`market-close-summary` 依赖知识读取

**约束**：
- 文件存储在用户目录 `~/.msa/`
- 使用 Markdown 格式，便于用户查看和编辑
- 遵循现有工具接口规范（Eino 框架）

## 目标 / 非目标

**目标：**
- 实现 `read_knowledge` 和 `write_knowledge` 两个工具
- 简化文件结构：`errors.md` + `summaries/`
- 写入时强制验证 + 重试机制
- 更新依赖 SKILL 的配置

**非目标：**
- 不实现复杂的知识分类（策略、洞察等）
- 不使用数据库存储
- 不支持知识搜索或查询
- 不实现知识的自动清理或归档

## 推荐模式

### 错误处理和重试策略

**模式ID**：003
**来源**：openspec/knowledge/index.yaml

**描述**：
分类错误处理，根据错误类型采取不同的处理策略。写入失败时自动重试。

**适用场景**：
- `write_knowledge` 工具的写入验证
- 文件操作失败时的重试

**实现**：
```go
const MaxRetries = 3
const RetryDelay = 100 * time.Millisecond

for retryCount := 0; retryCount < MaxRetries; retryCount++ {
    if retryCount > 0 {
        time.Sleep(RetryDelay)
    }

    path, err = executeWrite(param)
    if err != nil {
        continue
    }

    // 验证写入成功
    if verifyWrite(path, param) {
        break
    }
}
```

### 工具接口设计模式

**模式ID**：002
**来源**：openspec/changes/add-todo-list-skill/design.md

**描述**：
TODO 工具遵循现有工具接口规范，使用 Eino 框架的 tool 接口，返回统一的 Result 格式。

**适用场景**：
- 所有知识库工具

**实现**：
```go
type ReadKnowledgeTool struct{}

func (t *ReadKnowledgeTool) GetToolInfo() (tool.BaseTool, error) {
    return utils.InferTool(t.GetName(), t.GetDescription(), ReadKnowledge)
}

func (t *ReadKnowledgeTool) GetName() string {
    return "read_knowledge"
}

func (t *ReadKnowledgeTool) GetToolGroup() model.ToolGroup {
    return model.KnowledgeToolGroup
}
```

## 避免的反模式

### 硬编码文件路径

**问题所在**：
在代码中硬编码文件路径会导致测试困难和灵活性降低。

**不要这样做**：
```go
errorsPath := "/Users/xxx/.msa/errors.md"
```

**应该这样做**：
```go
func GetErrorsPath() string {
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".msa", "errors.md")
}
```

### 忽略文件不存在的情况

**问题所在**：
知识库文件可能不存在，直接读取会报错。

**不要这样做**：
```go
content, err := os.ReadFile(path)
if err != nil {
    return nil, err
}
```

**应该这样做**：
```go
content, err := os.ReadFile(path)
if err != nil {
    if os.IsNotExist(err) {
        return []ErrorEntry{}, nil // 返回空列表，不报错
    }
    return nil, err
}
```

## 设计决策

### D1: 文件存储位置

**决策**：存储在 `~/.msa/` 用户目录

**理由**：
- 与现有 Session 文件位置 `~/.msa/memory/` 一致
- 用户可以手动查看和编辑
- 跨 Session 持久化

### D2: 错误文件格式

**决策**：单文件追加写入

**理由**：
- 简单直观，所有历史错误在一个文件中
- 便于浏览历史记录
- 避免多文件管理的复杂性

### D3: 总结文件格式

**决策**：每日一个文件 `summaries/YYYY-MM-DD.md`

**理由**：
- 便于按日期查找
- 避免单文件过大
- 支持覆盖更新（每日总结只保留最新版本）

### D4: 写入验证策略

**决策**：写入后重新读取验证 + 最多 3 次重试

**理由**：
- 确保数据不丢失
- 参考模式 #003 的错误处理策略
- 指数退避避免频繁重试

## 风险 / 权衡

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 写入失败导致数据丢失 | 高 | 验证 + 重试 + 返回验证状态 |
| errors.md 文件过大 | 低 | 用户可手动清理旧记录 |
| 总结文件格式不一致 | 中 | 解析时保留原始内容供 Agent 参考 |
| Agent 忽略验证失败 | 中 | 在 SKILL 中明确要求检查 verified 字段 |

## 文件结构

```
pkg/logic/tools/knowledge/
├── read.go           # read_knowledge 工具
├── write.go          # write_knowledge 工具（带验证和重试）
├── parser.go         # 文件解析器
├── paths.go          # 路径管理
└── register.go       # 工具注册

~/.msa/
├── errors.md         # 历史错误记录
├── summaries/        # 每日总结
│   ├── 2026-04-05.md
│   └── 2026-04-06.md
└── memory/           # Session 记录（已有）
```

## 迁移计划

1. **阶段一：创建工具**
   - 创建 `pkg/logic/tools/knowledge/` 目录
   - 实现 `read.go`、`write.go`、`parser.go`、`paths.go`
   - 注册工具到 `pkg/logic/tools/register.go`

2. **阶段二：更新 SKILL**
   - 更新 `morning-analysis/SKILL.md` 的依赖
   - 更新 `afternoon-trade/SKILL.md` 的依赖
   - 更新 `market-close-summary/SKILL.md` 的依赖

3. **阶段三：删除旧组件**
   - 删除 `pkg/logic/skills/plugs/read-error/` 目录

**无回滚需求**：此变更为增量添加，不影响现有功能。
