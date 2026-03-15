# 设计：add-knowledge-tools

## 上下文

MSA 系统已实现 SKILL 系统用于自动化交易决策。当前创建了 4 个交易相关 SKILL（morning-analysis、afternoon-trade、market-close-summary、read-error），这些 SKILL 需要读写知识库文件和查询 Session 记录。

### 当前状态

- 知识库目录结构已建立：`.msa/knowledge/`
- 工具系统已成熟：`pkg/logic/tools/` 下有 stock、search、finance、skill 子包
- Session 模型已定义：`pkg/model/memory.go`，包含 Tags 字段
- 工具注册机制已完善：`MsaTool` 接口 + `RegisterTool` 函数

### 约束

- 知识文件使用 Markdown + YAML frontmatter 格式
- Session 数据存储在内存系统（文件或数据库）
- 工具必须符合 `MsaTool` 接口规范

## 目标 / 非目标

**目标：**

1. 实现 5 个知识库管理工具，支持 SKILL 系统运行
2. 确保文件操作安全，防止路径遍历攻击
3. 与现有工具系统无缝集成

**非目标：**

1. 不实现知识文件的版本控制
2. 不实现知识文件的全文搜索
3. 不实现复杂的 Session 关联查询

## 推荐模式

### 统一错误处理模式

**模式ID**：003
**来源问题**：add-web-search-tool

**描述**：
分类错误处理，根据错误类型采取不同策略。参数错误立即返回，临时错误可重试。

**适用场景**：
- 所有工具实现
- 文件操作错误处理

**实现**：
```go
// 参数校验错误 - 立即返回
if path == "" {
    return "", fmt.Errorf("路径不能为空")
}

// 路径安全错误 - 立即返回
if strings.Contains(path, "..") {
    return "", fmt.Errorf("路径遍历不允许")
}

// 文件不存在错误 - 立即返回
if !os.Exists(fullPath) {
    return "", fmt.Errorf("文件不存在: %s", path)
}

// 临时错误（如文件锁定）- 可考虑重试
```

### 资源管理单例模式

**模式ID**：004
**来源问题**：add-web-search-tool

**描述**：
共享资源使用单例模式管理，避免重复初始化和资源泄漏。

**适用场景**：
- 知识库根目录路径管理
- Session 存储管理

**实现**：
```go
var (
    knowledgeBasePath string
    knowledgeBaseOnce sync.Once
)

func GetKnowledgeBasePath() string {
    knowledgeBaseOnce.Do(func() {
        // 从配置或默认路径初始化
        knowledgeBasePath = ".msa/knowledge"
    })
    return knowledgeBasePath
}
```

## 避免的反模式

### 路径拼接不校验

**反模式ID**：security-001
**问题所在**：直接拼接用户输入路径可能导致目录遍历攻击

**影响**：可能读取或写入系统敏感文件

**不要这样做**：
```go
// 不要这样做
fullPath := knowledgeBasePath + "/" + userPath
content, _ := os.ReadFile(fullPath) // 危险！
```

**应该这样做**：
```go
// 应该这样做
func safePath(basePath, userPath string) (string, error) {
    // 清理路径
    cleanPath := filepath.Clean(userPath)

    // 检查路径遍历
    if strings.Contains(cleanPath, "..") {
        return "", fmt.Errorf("路径遍历不允许")
    }

    // 构建完整路径
    fullPath := filepath.Join(basePath, cleanPath)

    // 验证最终路径在允许范围内
    absBase, _ := filepath.Abs(basePath)
    absFull, _ := filepath.Abs(fullPath)
    if !strings.HasPrefix(absFull, absBase) {
        return "", fmt.Errorf("访问路径超出允许范围")
    }

    return fullPath, nil
}
```

### 忽略 YAML Frontmatter 解析错误

**反模式ID**：parse-001
**问题所在**：不处理 YAML 解析错误，导致元数据丢失

**影响**：知识文件的元数据无法正确读取

**应该这样做**：
```go
// 解析 frontmatter
metadata := make(map[string]interface{})
content := fileContent

if strings.HasPrefix(fileContent, "---\n") {
    parts := strings.SplitN(fileContent, "---\n", 3)
    if len(parts) >= 3 {
        if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
            // 记录警告，继续处理正文
            log.Warnf("解析 YAML frontmatter 失败: %v", err)
        }
        content = parts[2]
    }
}
```

## 设计决策

### 决策 1：文件操作限制在知识库目录

**理由**：安全性优先，防止任意文件读写

**方案**：
- 所有路径参数必须在 `.msa/knowledge/` 下
- 使用 `safePath` 函数校验所有路径
- 拒绝包含 `..` 的路径

### 决策 2：Session 查询使用日期索引

**理由**：SKILL 主要按日期查询 Session（今日、昨日）

**方案**：
- 在 Session 存储层添加日期索引
- 支持 ISO 日期格式和相对日期（`today`、`yesterday`）
- 支持按标签过滤

### 决策 3：原子写入避免数据损坏

**理由**：防止写入过程中断导致文件损坏

**方案**：
```go
func atomicWrite(path string, content []byte) error {
    // 先写入临时文件
    tmpPath := path + ".tmp"
    if err := os.WriteFile(tmpPath, content, 0644); err != nil {
        return err
    }
    // 原子重命名
    return os.Rename(tmpPath, path)
}
```

### 决策 4：工具独立文件组织

**理由**：遵循现有代码结构，便于维护

**方案**：
```
pkg/logic/tools/knowledge/
├── common.go          # 公共函数（路径安全、frontmatter 解析）
├── read.go            # read_knowledge_file
├── write.go           # write_knowledge_file
├── list.go            # list_knowledge_files
├── session.go         # query_sessions_by_date, add_session_tag
└── register.go        # 工具注册
```

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                        SKILL System                         │
│  (morning-analysis, afternoon-trade, market-close-summary)  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Knowledge Tools                           │
├─────────────────────────────────────────────────────────────┤
│  read_knowledge_file    │ write_knowledge_file              │
│  list_knowledge_files   │ query_sessions_by_date            │
│  add_session_tag        │                                    │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
┌─────────────────────────┐     ┌─────────────────────────┐
│   .msa/knowledge/       │     │    Session Storage      │
│   ├── user-profile/     │     │    (Memory System)      │
│   ├── strategies/       │     │                         │
│   ├── insights/         │     │    ┌─────────────────┐  │
│   ├── notes/            │     │    │ Session{        │  │
│   ├── summaries/        │     │    │   Tags: []      │  │
│   └── errors/           │     │    │ }               │  │
└─────────────────────────┘     │    └─────────────────┘  │
                                └─────────────────────────┘
```

## 风险 / 权衡

| 风险 | 缓解措施 |
|------|----------|
| 路径遍历攻击 | `safePath` 函数强制校验，单元测试覆盖边界情况 |
| 并发写入冲突 | 使用原子写入，文件锁（可选） |
| YAML 解析失败 | 容错处理，保留原始内容 |
| Session 存储依赖 | 确认 Session 存储接口，添加抽象层 |

## 迁移计划

1. **创建工具包**：`pkg/logic/tools/knowledge/`
2. **实现公共函数**：路径安全、frontmatter 解析
3. **实现 5 个工具**：按规格文件逐个实现
4. **注册工具**：修改 `pkg/logic/tools/basic.go`
5. **测试验证**：单元测试 + 集成测试

### 回滚策略

- 删除 `pkg/logic/tools/knowledge/` 目录
- 恢复 `pkg/logic/tools/basic.go` 中的注册代码
- 无数据库迁移，无需数据回滚