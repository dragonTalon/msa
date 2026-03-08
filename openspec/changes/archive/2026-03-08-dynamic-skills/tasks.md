# 实现任务：dynamic-skills

## 进度概览

**完成进度**: 86/91 任务 (94.5%)
- ✅ 阶段 1: 基础框架实现 (全部完成)
- ✅ 阶段 2: LLM 选择器实现 (全部完成)
- ✅ 阶段 3: CLI 命令实现 (全部完成)
- ✅ 阶段 4: 测试和优化 (核心完成，待更新 README)

---

## ⚠️ 风险警报

基于历史问题识别的风险：

### 中等：LLM 调用开销增加

- **相关问题**：设计文档 - 风险 1
- **缓解措施**：使用简洁的选择 Prompt（< 500 tokens），允许用户手动指定跳过选择

### 低：并发访问冲突

- **相关问题**：设计文档 - 风险 3
- **缓解措施**：使用 `sync.RWMutex` 保护缓存，确保线程安全

### 低：配置文件损坏

- **相关问题**：设计文档 - 风险 4
- **缓解措施**：解析失败时使用默认配置，记录错误并继续运行

---

## 1. 基础框架实现

### 1.1 创建包结构和核心数据结构

- [x] 1.1.1 创建 `pkg/logic/skills/` 目录
- [x] 1.1.2 创建 `skill.go`，定义 `Skill` 结构
  - [x] Name, Description, Version, Priority 字段
  - [x] Source (builtin/user) 枚举
  - [x] path, content, loaded 私有字段
  - [x] mu sync.RWMutex 并发保护
- [x] 1.1.3 创建 `SkillSource` 枚举类型
- [x] 1.1.4 添加 `GetContent()` 懒加载方法
  - [x] 双重检查锁定模式
  - [x] 读取文件并缓存
  - [x] 跳过 YAML frontmatter

### 1.2 实现 Registry（注册表）

- [x] 1.2.1 创建 `registry.go`
- [x] 1.2.2 定义 `Registry` 结构
  - [x] skillsByName map[string]*Skill
  - [x] mu sync.RWMutex
- [x] 1.2.3 实现 `Register(skill *Skill)` 方法
- [x] 1.2.4 实现 `Get(name string) (*Skill, error)` 方法
- [x] 1.2.5 实现 `ListAll() []*Skill` 方法
- [x] 1.2.6 实现 `ListAvailable() []*Skill` 方法（过滤禁用的）
- [x] 1.2.7 添加单元测试

### 1.3 实现 Loader（扫描和加载）

> 📚 **知识上下文**：基于模式 #003（错误处理和重试策略）
> 来源：错误处理和重试策略
> 经验：分类错误处理，跳过无效文件，继续处理其他文件

- [x] 1.3.1 创建 `loader.go`
- [x] 1.3.2 定义 `Loader` 结构
- [x] 1.3.3 实现 `scanBuiltinSkills()` 方法
  - [x] 扫描 `pkg/logic/skills/plugs/` 目录
  - [x] 查找所有 `SKILL.md` 文件
  - [x] 跳过无效文件，记录警告
- [x] 1.3.4 实现 `scanUserSkills()` 方法
  - [x] 扫描 `~/.msa/skills/` 目录
  - [x] 目录不存在时忽略（不报错）
  - [x] 跳过无效文件，记录警告
- [x] 1.3.5 实现 `parseSkillMetadata(path string)` 方法
  - [x] 读取 YAML frontmatter
  - [x] 解析 name, description, version, priority
  - [x] 使用默认值 5 作为 priority
  - [x] 返回 error 如果解析失败
- [x] 1.3.6 实现 `extractBody(data []byte) string` 方法
  - [x] 跳过 YAML frontmatter
  - [x] 去除开头换行符
- [x] 1.3.7 添加错误处理测试
  - [x] 测试无效 YAML 文件
  - [x] 测试缺少必需字段
  - [x] 测试目录权限错误

### 1.4 创建内置 Skills

- [x] 1.4.1 创建 `pkg/logic/skills/plugs/` 目录
- [x] 1.4.2 创建 `base/SKILL.md`
  - [x] 从现有 `BasePromptTemplate` 提取核心规则
  - [x] 工具调用规则
  - [x] 时间敏感性准则
  - [x] 搜索优化策略
  - [x] 特殊场景处理
- [x] 1.4.3 创建 `stock-analysis/SKILL.md`
  - [x] 数据处理流程
  - [x] 分析框架（技术面、基本面、市场面）
- [x] 1.4.4 创建 `account-management/SKILL.md`
  - [x] 账户管理流程
  - [x] 持仓登记流程
- [x] 1.4.5 创建 `output-formats/SKILL.md`
  - [x] 投资建议模板
  - [x] 分析报告格式
  - [x] 账户报告格式

### 1.5 实现 Config（配置管理）

- [x] 1.5.1 创建 `config.go`
- [x] 1.5.2 定义 `SkillsConfig` 结构
  - [x] DisableSkills []string 字段
- [x] 1.5.3 实现 `loadConfig()` 方法
  - [x] 读取 `~/.msa/config.json`
  - [x] 解析失败时使用默认配置
  - [x] 记录警告日志
- [x] 1.5.4 实现 `saveConfig()` 方法
  - [x] 保存到 `~/.msa/config.json`
  - [x] 文件不存在时创建

### 1.6 实现 Manager（对外 API）

- [x] 1.6.1 创建 `manager.go`
- [x] 1.6.2 定义 `Manager` 结构
  - [x] registry *Registry
  - [x] loader *Loader
  - [x] config *SkillsConfig
- [x] 1.6.3 实现 `GetManager() *Manager` 单例
- [x] 1.6.4 实现 `Initialize()` 方法
  - [x] 扫描内置 Skills
  - [x] 扫描用户 Skills
  - [x] 加载配置
- [x] 1.6.5 实现 `GetSkill(name string) (*Skill, error)` 方法
- [x] 1.6.6 实现 `ListSkills() []*Skill` 方法
- [x] 1.6.7 实现 `IsDisabled(name string) bool` 方法

---

## 2. LLM 选择器实现

### 2.1 实现 Selector（LLM 选择器）

> 📚 **知识上下文**：基于模式 #003（错误处理和重试策略）
> 来源：错误处理和重试策略
> 经验：LLM 调用失败时降级到 base skill，不中断对话

- [x] 2.1.1 创建 `selector.go`
- [x] 2.1.2 定义选择器 Prompt 模板
  - [x] 使用 Go 模板变量
  - [x] 包含 skills metadata
  - [x] 包含用户输入
- [x] 2.1.3 定义 `Selector` 结构
  - [x] llm *openai.ChatModel
  - [x] registry *Registry
- [x] 2.1.4 实现 `SelectSkills(ctx, userInput string) ([]string, error)` 方法
  - [x] 获取可用 skills（过滤禁用的）
  - [x] 构建选择 Prompt
  - [x] 调用 LLM
  - [x] 解析 JSON 结果
- [x] 2.1.5 实现 `ensureBaseSkill(skills []string) []string` 方法
- [x] 2.1.6 实现 `sortByPriority(skills []string) []string` 方法
- [x] 2.1.7 实现降级策略
  - [x] LLM 调用失败 → 返回 ["base"]
  - [x] JSON 解析失败 → 返回 ["base"]
  - [x] 超时 → 返回 ["base"]
- [x] 2.1.8 添加日志记录
  - [x] INFO：用户输入摘要、选中列表
  - [x] DEBUG：可用 skills、LLM reasoning

### 2.2 实现 Builder（Prompt 构建器）

- [x] 2.2.1 创建 `builder.go`
- [x] 2.2.2 定义 `Builder` 结构
  - [x] registry *Registry
- [x] 2.2.3 实现 `BuildSystemPrompt(selectedSkills []string) (string, error)` 方法
  - [x] 按顺序合并：基础角色 + 系统配置 + Skills + 基础规则
  - [x] 每个 Skill 前添加标题
- [x] 2.2.4 实现 `getSkillContent(name string) (string, error)` 方法
  - [x] 调用 Skill.GetContent() 懒加载
- [x] 2.2.5 添加错误处理
  - [x] Skill 不存在 → 跳过
  - [x] 内容为空 → 记录警告

### 2.3 集成到 chat.go

- [x] 2.3.1 修改 `GetDefaultTemplate()` 函数签名
  - [x] 添加 `selectedSkills []string` 参数
  - [x] 保持向后兼容（可选参数）
- [x] 2.3.2 修改 `Ask()` 函数
  - [x] 调用 Selector.SelectSkills()
  - [x] 获取 Skill 内容
  - [x] 构建完整 Prompt
  - [x] 错误时降级到原有实现
- [x] 2.3.3 添加用户手动覆盖支持
  - [x] 检查 CLI 参数 `--skills`
  - [x] 检查对话命令 `/skills:`
- [x] 2.3.4 添加集成测试

---

## 3. CLI 命令实现

### 3.1 实现 skills 命令框架

- [x] 3.1.1 创建 `cmd/skills.go`
- [x] 3.1.2 定义 `skillsCmd` Cobra 命令
- [x] 3.1.3 添加子命令：list, show, enable, disable

### 3.2 实现 list 命令

- [x] 3.2.1 创建 `cmd/skills_list.go`
- [x] 3.2.2 实现 `--builtin` 标志
- [x] 3.2.3 实现 `--user` 标志
- [x] 3.2.4 实现表格格式输出
  - [x] 列：Name, Priority, Source, Status
- [x] 3.2.5 实现 `--json` 输出格式
- [x] 3.2.6 添加过滤和排序

### 3.3 实现 show 命令

- [x] 3.3.1 创建 `cmd/skills_show.go`
- [x] 3.3.2 实现 `<name>` 参数
- [x] 3.3.3 显示完整元数据
- [x] 3.3.4 显示 Skill 内容
- [x] 3.3.5 显示来源目录
- [x] 3.3.6 显示启用/禁用状态
- [x] 3.3.7 处理 Skill 不存在的情况

### 3.4 实现 disable 命令

- [x] 3.4.1 创建 `cmd/skills_disable.go`
- [x] 3.4.2 实现 `<name>` 参数
- [x] 3.4.3 禁止禁用 base skill
- [x] 3.4.4 更新 `~/.msa/config.json`
  - [x] 添加到 `disable_skills` 数组
  - [x] 文件不存在时创建
- [x] 3.4.5 显示确认信息

### 3.5 实现 enable 命令

- [x] 3.5.1 创建 `cmd/skills_enable.go`
- [x] 3.5.2 实现 `<name>` 参数
- [x] 3.5.3 从 `disable_skills` 数组移除
- [x] 3.5.4 更新 `~/.msa/config.json`
- [x] 3.5.5 显示确认信息

---

## 4. 测试和优化

### 4.1 单元测试

- [x] 4.1.1 Skill 结构测试
  - [x] 测试懒加载
  - [x] 测试并发访问
  - [x] 测试缓存机制
- [x] 4.1.2 Registry 测试
  - [x] 测试注册和查询
  - [x] 测试并发安全
- [x] 4.1.3 Loader 测试
  - [x] 测试 YAML 解析
  - [x] 测试错误处理
  - [x] 测试 frontmatter 提取
- [x] 4.1.4 Selector 测试
  - [x] 测试降级策略
  - [x] 测试 base skill 保证
  - [x] 测试 priority 排序
- [x] 4.1.5 Builder 测试
  - [x] 测试 Prompt 合并
  - [x] 测试变量填充

### 4.2 集成测试

- [x] 4.2.1 端到端对话测试
  - [x] 测试完整工作流程
  - [x] 测试无效 skill 文件处理
  - [x] 测试 skill 来源追踪
  - [x] 测试懒加载机制
- [x] 4.2.2 CLI 命令测试
  - [x] 测试 list/show/enable/disable
  - [x] 测试配置文件更新
- [x] 4.2.3 向后兼容性测试
  - [x] 测试现有对话流程
  - [x] 测试配置不存在的情况

### 4.3 性能优化

- [ ] 4.3.1 启动性能测试
  - [ ] 测试 10 个 Skills 的扫描时间
  - [ ] 确保 < 100ms
- [ ] 4.3.2 内存占用测试
  - [ ] 测试 10 个 Skills 的内存占用
  - [ ] 确保元数据 < 1KB
- [ ] 4.3.3 LLM 调用优化
  - [ ] 测试选择 Prompt 的 token 数
  - [ ] 确保 < 500 tokens

### 4.4 文档和示例

- [x] 4.4.1 创建用户文档
  - [x] 如何创建自定义 Skill
  - [x] SKILL.md 格式说明
  - [x] CLI 命令使用指南
- [x] 4.4.2 创建示例 Skills
  - [x] 简单示例
  - [x] 复杂示例
- [x] 4.4.3 更新 README.md

---

## 基于知识的测试

### 错误处理回归测试

> 📚 **来源**：模式 #003（错误处理和重试策略）

**历史问题**：
错误处理不当会导致系统崩溃或数据丢失。

**测试场景**：
1. 创建格式错误的 SKILL.md 文件
2. 启动系统并加载 Skills
3. 验证系统继续运行
4. 验证错误日志被记录
5. 验证无效 Skill 被跳过

**预期行为**：
- 系统正常启动
- 无效 Skill 被跳过
- 错误日志清晰说明问题

---

### LLM 调用失败回归测试

> 📚 **来源**：设计文档 - 风险 1（LLM 调用开销）

**历史问题**：
LLM 调用失败时系统没有降级策略，导致对话中断。

**测试场景**：
1. 模拟 LLM 调用超时
2. 验证系统降级到 base skill
3. 验证对话继续进行
4. 验证降级日志被记录

**预期行为**：
- 对话不中断
- 自动使用 base skill
- 记录降级原因

---

### 并发访问回归测试

> 📚 **来源**：设计文档 - 风险 3（并发访问冲突）

**历史问题**：
并发访问共享数据导致数据竞争。

**测试场景**：
1. 启动 100 个 goroutine 同时访问 Skill
2. 同时调用 GetContent()
3. 使用 go test -race 检测
4. 验证无数据竞争

**预期行为**：
- 无数据竞争
- 所有请求正常返回
- 缓存正确工作

---

### 配置文件损坏回归测试

> 📚 **来源**：设计文档 - 风险 4（配置文件损坏）

**历史问题**：
配置文件损坏导致系统无法启动。

**测试场景**：
1. 创建损坏的 config.json
2. 启动系统
3. 验证系统使用默认配置
4. 验证系统正常运行

**预期行为**：
- 系统正常启动
- 使用默认配置
- 记录警告日志
