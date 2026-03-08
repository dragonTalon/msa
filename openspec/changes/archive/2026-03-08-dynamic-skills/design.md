# 设计：dynamic-skills

## 上下文

当前 MSA 的 Prompt 模板硬编码在 `pkg/logic/agent/prompt.go` 中，包含超过 100 行的单一大模板。这导致修改 Prompt 需要重新编译部署，无法动态扩展能力，也难以维护。

**当前状态：**
- 单一的 `BasePromptTemplate` 常量
- 硬编码的系统提示词
- 无法根据场景切换能力
- 用户无法自定义 Prompt

**约束：**
- 必须保持向后兼容
- 不能破坏现有对话功能
- 需要最小化 LLM 调用开销
- 配置文件使用 `~/.msa/config.json`

**利益相关者：**
- 终端用户：希望更灵活的分析能力
- 开发者：希望更易于维护和扩展
- 系统稳定性：需要保证降级策略

## 目标 / 非目标

**目标：**
1. 实现动态 Skill 加载系统，支持内置和用户自定义 Skills
2. 使用 LLM 自动选择合适的 Skills
3. 提供懒加载机制，优化内存和性能
4. 提供 CLI 命令管理 Skills
5. 保持完全向后兼容

**非目标：**
1. 不实现 Skill 版本管理（预留字段，暂不使用）
2. 不实现 Skill 热更新（需要重启才能加载新 Skills）
3. 不实现 Skill 依赖关系（Skills 独立，无依赖）
4. 不实现远程 Skills（仅支持本地文件）

## 推荐模式

### 错误处理和重试策略

**模式ID**：003
**来源**：add-web-search-tool

**描述**：
分类错误处理，根据错误类型采取不同的处理策略：
- 参数错误：立即返回，不重试
- 永久错误：不重试（如文件不存在、格式错误）
- 临时错误：重试（如网络超时）

**适用场景：**
- Skill 文件加载失败
- LLM 调用超时
- 配置文件读取错误

**实现策略：**
1. **Skill 加载失败**：记录警告，跳过该 Skill，继续加载其他 Skills
2. **LLM 选择失败**：降级到 base skill，不中断对话
3. **配置文件错误**：使用默认配置，记录警告日志

---

### 配置变更时的自动处理

**模式ID**：006
**来源**：add-config-management

**描述**：
当配置变更时自动触发相关操作，保持系统状态一致。

**适用场景：**
- 禁用/启用 Skill 后自动重新加载
- 配置文件变更后刷新 Skill 注册表

**实现：**
- CLI 命令修改配置后自动触发重新加载
- 不需要手动重启应用

## 避免的反模式

### 过度设计

**问题所在：**
实现太多不必要的功能，增加复杂度。

**影响：**
- 代码难以维护
- 增加出错概率
- 延迟交付时间

**不要这样做：**
```go
// 不要实现复杂的版本依赖检查
type SkillDependency struct {
    Name    string
    Version string
}
```

**应该这样做：**
```go
// 简单的结构，够用即可
type Skill struct {
    Name        string
    Description string
    Priority    int
    Content     string
}
```

---

### 每次 LLM 调用都选择 Skills

**问题所在：**
如果缓存策略不当，每次对话都会产生额外的 LLM 调用。

**影响：**
- 增加 API 成本
- 增加响应延迟
- 降低用户体验

**不要这样做：**
```go
// 不要每次都调用 LLM 选择
func SelectSkills(input string) []string {
    // 每次都调用 LLM
    return llm.Select(input)
}
```

**应该这样做：**
```go
// 按需选择，考虑缓存
func SelectSkills(input string, manualOverride []string) []string {
    if len(manualOverride) > 0 {
        return manualOverride
    }
    return llm.Select(input)
}
```

---

### 阻塞式 Skill 加载

**问题所在：**
启动时加载所有 Skill 内容会延迟启动时间。

**影响：**
- 应用启动慢
- 内存占用高
- 用户体验差

**不要这样做：**
```go
// 启动时加载所有内容
func LoadAllSkills() {
    for _, skill := range skills {
        skill.Content = readFile(skill.Path) // 阻塞
    }
}
```

**应该这样做：**
```go
// 懒加载，按需读取
func (s *Skill) GetContent() string {
    if s.loaded {
        return s.content
    }
    s.content = readFile(s.path)
    s.loaded = true
    return s.content
}
```

## 设计决策

### 1. Skill 文件格式：Markdown with YAML Frontmatter

**决策：** 使用 Markdown 文件配合 YAML frontmatter 存储元数据。

**理由：**
- 易读易写，用户可以直接编辑
- YAML frontmatter 是标准做法，解析库成熟
- 支持 Markdown 格式的内容展示

**为什么不选 JSON：**
- JSON 不支持注释
- 用户编辑体验差
- 不适合长文本内容

### 2. LLM 自动选择 vs 用户指定

**决策：** 两者都支持。默认 LLM 自动选择，允许用户手动覆盖。

**理由：**
- LLM 自动选择提供最佳用户体验
- 手动覆盖满足高级用户需求
- 保持灵活性

### 3. Priority 范围：0-10

**决策：** 使用 0-10 的整数，默认 5，base skill 为 10。

**理由：**
- 范围小，简单直观
- 足够区分优先级
- 便于排序

**为什么不选 1-100：**
- 过多的选项没有实际意义
- 增加决策成本

### 4. 懒加载策略

**决策：** 只缓存 metadata，content 按需加载。

**理由：**
- 减少内存占用
- 加快启动速度
- 降低初始 I/O

### 5. 降级策略

**决策：** 任何失败都降级到 base skill + 硬编码 Prompt。

**理由：**
- 保证基本功能可用
- 用户不会完全无法使用
- 符合渐进增强原则

## 架构设计

### 模块结构

```
pkg/logic/skills/
├── skill.go              # Skill 结构定义
├── loader.go             # 扫描和加载 Skills
├── registry.go           # Skill 注册表
├── selector.go           # LLM 自动选择器
├── builder.go            # Prompt 构建器
├── manager.go            # 对外管理 API
├── config.go             # 配置管理
└── plugs/                # 内置 Skills
    ├── base/SKILL.md
    ├── stock-analysis/SKILL.md
    ├── account-management/SKILL.md
    └── output-formats/SKILL.md
```

### 数据流

```
用户输入
   │
   ▼
Skill Selector (LLM)
   │
   ├─▶ 失败 → 降级到 base
   │
   ├─▶ 成功 → 返回技能列表
   │        │
   │        ▼
   │   Skill Content Loader (懒加载)
   │        │
   │        ▼
   │   Prompt Builder
   │        │
   ▼        ▼
    Main Agent
```

### 核心接口

```go
// Skill 表示一个技能单元
type Skill struct {
    Name        string
    Description string
    Version     string
    Priority    int
    Source      SkillSource
    path        string
    content     string
    loaded      bool
    mu          sync.RWMutex
}

// Registry 技能注册表
type Registry struct {
    skillsByName map[string]*Skill
    mu           sync.RWMutex
}

// Selector LLM 选择器
type Selector struct {
    llm      *openai.ChatModel
    registry *Registry
}

// Builder Prompt 构建器
type Builder struct {
    registry *Registry
}

// Manager 对外管理 API
type Manager struct {
    registry *Registry
    selector *Selector
    builder  *Builder
}
```

## 风险 / 权衡

### 风险 1：LLM 调用开销

**风险描述：**
每次对话增加一次 LLM 调用，增加延迟和成本。

**缓解措施：**
- 使用简洁的选择 Prompt（< 500 tokens）
- 考虑未来实现选择结果缓存
- 允许用户手动指定跳过选择

### 风险 2：Skill 文件格式错误

**风险描述：**
用户创建的 Skill 文件格式错误导致解析失败。

**缓解措施：**
- 跳过无效文件，不中断启动
- 记录详细的错误日志
- 提供 `msa skills validate` 命令（未来）

### 风险 3：并发访问冲突

**风险描述：**
多个 goroutine 同时访问 Skill 缓存可能导致数据竞争。

**缓解措施：**
- 使用 `sync.RWMutex` 保护缓存
- 懒加载时加锁
- 确保线程安全

### 风险 4：配置文件损坏

**风险描述：**
`~/.msa/config.json` 损坏或格式错误。

**缓解措施：**
- 解析失败时使用默认配置
- 记录错误并继续运行
- 提供配置恢复建议

## 迁移计划

### 阶段 1：实现基础框架

1. 创建 `pkg/logic/skills/` 包结构
2. 实现 `Skill` 结构和 `Registry`
3. 实现 `Loader` 扫描和加载 Skills
4. 创建内置 Skills（迁移现有 Prompt）

### 阶段 2：实现 LLM 选择

1. 实现 `Selector` 调用 LLM
2. 实现降级策略
3. 集成到 `chat.go`

### 阶段 3：实现 CLI 命令

1. 添加 `msa skills list` 命令
2. 添加 `msa skills show` 命令
3. 添加 `msa skills enable/disable` 命令

### 阶段 4：测试和优化

1. 单元测试
2. 集成测试
3. 性能优化

### 回滚策略

如果实现出现问题：
1. 保留 `prompt.go` 作为兜底
2. 配置项关闭 Skill 系统
3. 完全回退到原有实现

## 性能考虑

### 启动性能

- 只扫描目录，不加载内容
- YAML 解析使用流式解析器
- 并行扫描多个目录

### 运行时性能

- Skill 内容缓存，不重复读取
- LLM 选择结果不缓存（每次都是新输入）
- 使用读写锁优化并发

### 内存占用

- 只缓存元数据（~100 bytes per skill）
- 内容按需加载并缓存
- 预估 10 个 Skills ~1KB 元数据 + 10KB 内容

## 安全考虑

### 文件访问

- 只读取指定目录的文件
- 不支持路径遍历
- 用户目录权限检查

### LLM 调用

- 不发送敏感内容给 LLM
- 只发送 metadata（name, description, priority）
- 不发送 Skill 内容

### 配置文件

- 验证 JSON 格式
- 不执行配置中的代码
- 沙盒化配置解析
