# OpenSpec 规范文件修复指南

## 概述

本指南提供分步骤的修复说明，帮助将不符合规范的 spec 文件转换为标准格式。

---

## 快速参考：问题映射表

| 文件名 | 主要问题 | 修复难度 | 参考章节 |
|--------|---------|--------|---------|
| account-management | 旧格式 + 缺 Purpose | 中等 | §2 + §3 |
| account-summary | 缺 Requirements + Purpose | 中等 | §2 + §3 |
| add-session-tag | 旧格式 + 缺 Purpose + 无内容 | 高 | §2 + §3 + §5 |
| agent-prompt | 缺 Requirements + Purpose | 中等 | §2 + §3 |
| amount-locking | 旧格式 + 缺 Purpose + 无内容 | 高 | §2 + §3 + §5 |
| cli-chat | 缺 Purpose + 旧格式 | 中等 | §2 + §3 + §4 |
| cli-version | 缺标题 + Purpose 空 | 低 | §1 + §2 |
| cmd-structure | 缺标题 | 低 | §1 |
| knowledge-read | 缺 Purpose + 旧格式 | 中等 | §2 + §3 + §4 |
| knowledge-write | 缺 Purpose + 旧格式 | 中等 | §2 + §3 + §4 |
| list-knowledge-files | 旧格式 + 缺 Purpose + 无内容 | 高 | §2 + §3 + §5 |
| position-calculation | 旧格式 + 缺 Purpose + 无内容 | 高 | §2 + §3 + §5 |
| position-management | 缺标题 + Purpose 空 | 低 | §1 + §2 |
| position-tracking | 无内容 | 高 | §5 |
| query-sessions-by-date | 旧格式 + 缺 Purpose + 无内容 | 高 | §2 + §3 + §5 |
| read-knowledge-file | 旧格式 + 缺 Purpose + 无内容 | 高 | §2 + §3 + §5 |
| release-automation | 缺标题 | 低 | §1 |
| self-update | 缺标题 + 缺 Purpose | 低 | §1 + §2 |
| session-persist | 缺 Purpose + 旧格式 + 内容不完 | 中等 | §2 + §3 + §4 + §6 |
| session-resume | 缺 Purpose + 旧格式 + 内容不完 | 中等 | §2 + §3 + §4 + §6 |
| session-tagging | 缺标题 + Purpose 空 | 低 | §1 + §2 |
| skill-asset-tool | 缺标题 + 缺 Requirements + 缺 Purpose | 中等 | §1 + §2 + §3 |
| skill-reference-tool | 缺标题 + 缺 Requirements + 缺 Purpose | 中等 | §1 + §2 + §3 |
| skill-system | 缺 Requirements + 缺 Purpose | 中等 | §2 + §3 |
| todo-management | 缺 Requirements + 缺 Purpose | 中等 | §2 + §3 |
| tool-error-protection | 缺 Purpose + 旧格式 + 内容不完 | 中等 | §2 + §3 + §4 + §6 |
| tool-response-format | 缺标题 + 缺 Requirements + 缺 Purpose | 中等 | §1 + §2 + §3 |
| trade-analysis | 缺标题 + Purpose 空 | 低 | §1 + §2 |
| trade-execution | 无内容 | 高 | §5 |
| transaction-recording | 旧格式 + 缺 Purpose + 无内容 | 高 | §2 + §3 + §5 |
| web-scraping | 旧格式 + 缺 Purpose + 无内容 | 高 | §2 + §3 + §5 |
| web-search | 旧格式 + 缺 Purpose + 无内容 | 高 | §2 + §3 + §5 |
| write-knowledge-file | 旧格式 + 缺 Purpose + 无内容 | 高 | §2 + §3 + §5 |

---

## §1: 修复规格标题格式

**问题**: 标题格式不符合 `# 规格：<name>` 规范

**受影响文件** (10 个):
- cli-version, cmd-structure, position-management, release-automation
- self-update, session-tagging, skill-asset-tool, skill-reference-tool
- tool-response-format, trade-analysis

### 修复步骤

#### 步骤 1: 打开文件

```bash
# 例如修复 cli-version
nano /openspec/specs/cli-version/spec.md
```

#### 步骤 2: 查找当前标题

查看第一行，应该是类似这样的格式：
```
# cli-version 规范           ❌ 错误
# Command Structure           ❌ 错误
# Position Management         ❌ 错误
# 会话标签规格                ❌ 错误
```

#### 步骤 3: 替换为标准格式

将第一行改为：
```
# 规格：cli-version
# 规格：cmd-structure
# 规格：position-management
# 规格：session-tagging
# 规格：skill-asset-tool
# 规格：skill-reference-tool
# 规格：tool-response-format
# 规格：trade-analysis
# 规格：release-automation
# 规格：self-update
```

#### 步骤 4: 保存文件

```bash
# Nano: Ctrl+X → Y → Enter
# Vim: :wq
```

---

## §2: 添加或修复 Purpose 章节

**问题**: 缺少 `## Purpose` 章节或 Purpose 为空

**受影响文件** (25+ 个)

### 修复步骤

#### 步骤 1: 理解 Purpose 的含义

Purpose 应该清晰地说明：
- **是什么**: 这个能力是什么
- **为什么**: 为什么需要这个能力
- **解决什么**: 这个能力解决什么问题

**良好的 Purpose 示例** (来自 agent-chat):
```markdown
## Purpose

定义 agent 对话功能如何与 Skill 系统集成：通过 system prompt 
注入 skill 元数据，由 agent 自主判断并按需调用 `get_skill_content` 
获取完整流程。
```

**良好的 Purpose 示例** (来自 config-validation):
```markdown
## Purpose

定义配置验证的规范和实现方式，确保所有配置参数都经过有效性检查
并符合系统需求。
```

#### 步骤 2: 定位插入位置

打开文件后找到：
```
# 规格：capability-name

[这里可能有一句话的概述]

## Purpose  ← 如果存在，跳到步骤 4
或
## Requirements  ← 如果在这之前没有 Purpose，需要插入
```

#### 步骤 3: 编写 Purpose

基于能力名称和需求，编写 2-3 句话的 Purpose。参考示例：

**对于 account-management**:
```markdown
## Purpose

提供完整的用户账户管理能力，包括账户创建、查询、状态管理和资金
管理功能，为交易系统的多用户支持奠定基础。
```

**对于 web-search**:
```markdown
## Purpose

集成网页搜索功能，使 AI Agent 能够主动查询网络信息并基于实时数据
进行决策，提升 AI 助手的知识广度。
```

**对于 position-tracking**:
```markdown
## Purpose

实现持仓跟踪系统，记录用户的持仓变化和历史，支持风险管理和业绩
分析。
```

#### 步骤 4: 添加/修复 Purpose 章节

在规格标题之后（如果已有其他内容，将其移除或合并）添加：

```markdown
# 规格：capability-name

## Purpose

[您编写的 Purpose 文本]

## Requirements
```

#### 步骤 5: 检查 Purpose 是否为空

如果当前 Purpose 章节为空（下面直接是 ## Requirements），例如：
```markdown
## Purpose

## Requirements
```

填入 2-3 句话的说明。

---

## §3: 修复/添加 Requirements 章节

**问题**: 缺少 `## Requirements` 标题或位置不对

**受影响文件** (13 个): 
account-summary, agent-prompt, cli-chat, knowledge-read, knowledge-write, 
session-persist, session-resume, skill-asset-tool, skill-reference-tool, 
skill-system, todo-management, tool-error-protection, tool-response-format

### 修复步骤

#### 步骤 1: 查找需求位置

打开文件，查找类似的模式：
```markdown
## Purpose

...

### Requirement: 需求 1    ← 缺少中间的 ## Requirements
#### Scenario: ...
```

或：
```markdown
## Purpose

...

### ADDED Requirements   ← 错误的标题格式
### Requirement: 需求 1
```

#### 步骤 2: 添加标准的 Requirements 标题

在第一个 `### Requirement:` 之前插入 `## Requirements`：

```markdown
## Purpose

...

## Requirements          ← 添加这一行

### Requirement: 需求 1
#### Scenario: ...
```

#### 步骤 3: 清理旧的或错误的标题

如果存在 `### ADDED Requirements` 或 `### MODIFIED Requirements`，删除这些行。

#### 步骤 4: 保存文件

---

## §4: 转换旧格式为标准格式

**问题**: 使用 `## 需求:` 或 `## 增加需求` 代替标准格式

**受影响文件** (17 个):
account-management, add-session-tag, amount-locking, cli-chat, 
knowledge-read, knowledge-write, list-knowledge-files, 
position-calculation, query-sessions-by-date, read-knowledge-file, 
session-persist, session-resume, tool-error-protection, 
transaction-recording, web-scraping, web-search, write-knowledge-file

### 修复步骤

#### 步骤 1: 识别旧格式

打开文件，查找：
```markdown
## 需求：创建账户           ← 旧中文格式
### 场景：创建新账户

## 增加需求                 ← 旧格式
### 场景：场景名称

## ADDED Requirements        ← 旧英文格式
### Scenario: ...
```

#### 步骤 2: 转换层次结构

使用以下映射转换：

| 旧格式 | 新格式 |
|--------|--------|
| `## 需求：xxx` | `### Requirement: xxx` |
| `## 增加需求` | `## Requirements` |
| `## ADDED Requirements` | `## Requirements` |
| `## MODIFIED Requirements` | `## Requirements` |
| `### 场景：xxx` | `#### Scenario: xxx` |
| `### Scenario: xxx` (在三级标题位置) | `#### Scenario: xxx` |

**转换前**:
```markdown
## 需求：创建账户
系统必须支持创建新账户。

### 场景：创建新账户
- **当** 用户提交创建请求
- **那么** 系统创建账户
```

**转换后**:
```markdown
### Requirement: 创建账户
系统必须支持创建新账户。

#### Scenario: 创建新账户
- **WHEN** 用户提交创建请求
- **THEN** 系统创建账户
```

#### 步骤 3: 添加 Requirements 父标题

如果转换后没有 `## Requirements` 标题，在第一个 `### Requirement:` 前添加：

```markdown
## Purpose
...

## Requirements

### Requirement: 创建账户
...
```

#### 步骤 4: 规范化 WHEN/THEN

检查场景中的条件语句，使用统一格式：

**中文格式**:
```markdown
- **当** ... 
- **那么** ...
- **并且** ...
```

**英文格式**:
```markdown
- **WHEN** ...
- **THEN** ...
- **AND** ...
```

选择一种格式并保持一致（不要混用）。

#### 步骤 5: 逐个检查和转换

对于受影响的每个文件：
1. 打开文件
2. 使用查找替换（Ctrl+H）或手动编辑
3. 按步骤 2-4 转换
4. 保存文件

---

## §5: 处理无需求的文件

**问题**: 文件中完全没有需求声明（需求数=0）

**受影响文件** (13 个):
account-management, add-session-tag, amount-locking, list-knowledge-files, 
position-calculation, position-tracking, query-sessions-by-date, 
read-knowledge-file, trade-execution, transaction-recording, 
web-scraping, web-search, write-knowledge-file

### 判断与处理

#### 情况 A: 文件使用了旧格式

**识别方式**: 查看原始文件，如果有 `## 需求：xxx` 但验证脚本没有检测到

**原因**: 验证脚本只寻找 `### Requirement:` 格式，无法识别旧的 `## 需求:` 格式

**解决方法**: 按 §4 的说明转换旧格式

**示例** (account-management):
```markdown
# 规格：account-management

## 需求：创建账户      ← 旧格式，脚本无法识别
...

### 场景：创建新账户
...
```

按 §4 转换后：
```markdown
# 规格：account-management

## Requirements

### Requirement: 创建账户
...

#### Scenario: 创建新账户
...
```

#### 情况 B: 文件确实是空的（待填充）

**识别方式**: 文件只有标题，没有任何需求内容

**示例**:
```markdown
# 规格：position-tracking

或

# 规格：trade-execution

[只有标题和可能的 Purpose，但没有任何需求]
```

**选项 1: 编写完整的需求** (推荐)
1. 根据能力名称理解功能
2. 参考相关的需求文档或设计文档
3. 编写至少 2-3 个需求和 5-10 个场景
4. 按标准格式组织内容

**选项 2: 标记为 [WIP] (进行中)**
```markdown
# 规格：capability-name [WIP]

## Purpose
[描述]

## Requirements

### Requirement: [TODO]

#### Scenario: [TODO]

_本 spec 仍在进行中，等待详细需求_
```

#### 情况 C: 文件是占位符

**识别方式**: 能力目录存在但功能还未确定

**处理方法**:
1. 确认该功能是否真的需要
2. 如果需要，按情况 B 处理
3. 如果不需要，可以将整个目录删除（需要确认）

### 建议优先级

**高优先级** (应该有内容):
- account-management, add-session-tag, amount-locking
- list-knowledge-files, query-sessions-by-date, read-knowledge-file
- transaction-recording, web-scraping, web-search, write-knowledge-file
- position-calculation

**中优先级** (需要确认):
- position-tracking, trade-execution

### 修复模板

对于需要填充内容的文件，使用以下模板：

```markdown
# 规格：capability-name

## Purpose

[2-3 句话说明功能目的]

## Requirements

### Requirement: 功能需求 1

[需求描述]

#### Scenario: 场景 1
- **WHEN** [条件 1]
- **THEN** [结果 1]

#### Scenario: 场景 2
- **WHEN** [条件 2]
- **THEN** [结果 2]

### Requirement: 功能需求 2

[需求描述]

#### Scenario: 场景 3
- **WHEN** [条件 3]
- **THEN** [结果 3]
```

---

## §6: 修复场景内容不完整

**问题**: WHEN/THEN 数量与 Scenario 数量不对应

**受影响文件** (6-7 个):
account-summary, agent-prompt, cli-chat, config-validation, 
session-persist, session-resume, tool-error-protection

### 问题类型

#### 类型 A: THEN 数量多于 Scenario

**症状**: `THEN: 20 vs Scenarios: 10` (或类似)

**原因**: 每个场景有多个 AND 补充结果，验证脚本分别计算了每一行

**是否问题**: ❌ **不是问题** - AND 用于补充条件/结果是正确的用法

**处理**: 无需修改

**示例**:
```markdown
#### Scenario: 创建账户
- **WHEN** 用户提交创建请求
- **THEN** 系统创建账户
- **AND** 返回账户 ID              ← AND 也算一个 THEN
- **AND** 初始化金额               ← 又一个 AND
- **AND** 设置状态为 ACTIVE         ← 再一个 AND
```

#### 类型 B: WHEN/THEN 不匹配 Scenario

**症状**: `WHEN: 8 THEN: 7 vs Scenarios: 8`

**原因**: 某个场景缺少 THEN 或 WHEN

**是否问题**: ✅ **是问题** - 每个场景应该有完整的 WHEN-THEN 对

**处理**: 补充缺失的语句

**修复步骤**:

1. 打开文件，找到报告中指出的 Scenario

   ```bash
   # 例如找 agent-prompt 的问题
   grep -n "#### Scenario:" /openspec/specs/agent-prompt/spec.md
   ```

2. 逐个检查每个场景，确保有 WHEN 和 THEN

   ```markdown
   ✅ 正确:
   #### Scenario: 场景 1
   - **WHEN** 条件
   - **THEN** 结果
   
   ❌ 缺少 WHEN:
   #### Scenario: 场景 1
   - **THEN** 结果   ← 没有 WHEN
   
   ❌ 缺少 THEN:
   #### Scenario: 场景 1
   - **WHEN** 条件   ← 没有 THEN
   ```

3. 补充缺失的部分

   ```markdown
   # 修复前
   #### Scenario: 获取账户信息
   - **THEN** 返回账户对象
   
   # 修复后
   #### Scenario: 获取账户信息
   - **WHEN** 用户请求账户信息
   - **THEN** 返回账户对象
   ```

4. 保存并验证

   ```bash
   python3 /tmp/validate_specs.py
   ```

---

## 修复执行计划

### 第 1 天：修复低难度项 (估计 1-2 小时)

这些文件只需要修复标题或 Purpose：

```
1. cli-version        § 1 + § 2
2. cmd-structure      § 1
3. position-management § 1 + § 2
4. release-automation § 1
5. self-update        § 1 + § 2
6. session-tagging    § 1 + § 2
7. trade-analysis     § 1 + § 2
```

### 第 2 天：修复中等难度项 (估计 3-4 小时)

这些文件需要多步骤修复：

```
1. account-summary        § 2 + § 3 + § 6
2. agent-prompt           § 2 + § 3 + § 6
3. cli-chat               § 2 + § 3 + § 4 + § 6
4. knowledge-read         § 2 + § 3 + § 4
5. knowledge-write        § 2 + § 3 + § 4
6. session-persist        § 2 + § 3 + § 4 + § 6
7. session-resume         § 2 + § 3 + § 4 + § 6
8. skill-asset-tool       § 1 + § 2 + § 3
9. skill-reference-tool   § 1 + § 2 + § 3
10. skill-system          § 2 + § 3
11. todo-management       § 2 + § 3
12. tool-error-protection § 2 + § 3 + § 4 + § 6
13. tool-response-format  § 1 + § 2 + § 3
```

### 第 3 天：修复高难度项 (估计 4-6 小时)

这些文件需要补充内容或转换旧格式：

```
1. account-management      § 2 + § 3 + § 4 + § 5
2. add-session-tag         § 2 + § 3 + § 4 + § 5
3. amount-locking          § 2 + § 3 + § 4 + § 5
4. list-knowledge-files    § 2 + § 3 + § 4 + § 5
5. position-calculation    § 2 + § 3 + § 4 + § 5
6. position-tracking       § 5
7. query-sessions-by-date  § 2 + § 3 + § 4 + § 5
8. read-knowledge-file     § 2 + § 3 + § 4 + § 5
9. trade-execution         § 5
10. transaction-recording  § 2 + § 3 + § 4 + § 5
11. web-scraping           § 2 + § 3 + § 4 + § 5
12. web-search             § 2 + § 3 + § 4 + § 5
13. write-knowledge-file   § 2 + § 3 + § 4 + § 5
```

---

## 验证修复结果

### 使用验证脚本

```bash
# 在完成修复后重新运行
python3 /tmp/validate_specs.py

# 或检查特定文件
grep -A 10 "# 规格：account-management" /openspec/specs/account-management/spec.md
```

### 手动检查清单

对于每个修复的文件，确认：

- [ ] 标题格式: `# 规格：capability-name`
- [ ] 有 `## Purpose` 章节（非空）
- [ ] 有 `## Requirements` 章节
- [ ] 所有需求都是 `### Requirement:` 格式
- [ ] 所有场景都是 `#### Scenario:` 格式
- [ ] 每个场景都有 WHEN 和 THEN（或 当/那么）
- [ ] 没有使用旧格式标题
- [ ] 没有使用混合的中英文格式

### 目标

完成所有修复后，应该达到：
- ✅ 所有 49 个 spec 都符合格式规范
- ✅ 没有格式严重错误
- ✅ 没有缺失的关键章节
- ✅ 所有需求都有对应的场景
- ✅ 所有场景都有完整的 WHEN-THEN 结构

---

## 常见问题

### Q: 什么是 Purpose，应该写多长？

**A**: Purpose 说明这个能力的核心目的。应该是 1-3 句话（2-5 行）。

### Q: 能混用英文和中文吗？

**A**: 不建议混用。选择一种语言（或主要语言）并保持一致。
- WHEN/THEN vs 当/那么
- Requirement vs 需求
- Scenario vs 场景

### Q: 如果文件没有内容怎么办？

**A**: 参考 §5 的处理方法。可能需要编写完整的需求。

### Q: 场景中的 AND 会影响计数吗？

**A**: 是的。验证脚本分别计算 WHEN、THEN、AND 的出现次数。
多个 AND 是正常的，用于补充条件或结果。

### Q: 修复后需要创建 commit 吗？

**A**: 建议按能力或按难度分组创建 commit，便于追踪。例如：
```bash
git commit -m "fix: standardize spec format for account-management and related files"
```

---

**最后更新**: 2026-04-19
