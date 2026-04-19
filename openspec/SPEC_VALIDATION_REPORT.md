# OpenSpec 格式验证报告

**生成时间**: 2026-04-19  
**扫描范围**: `/openspec/specs/` 下所有 spec 文件  
**总文件数**: 49 个

---

## 📊 执行摘要

| 指标 | 数值 | 百分比 |
|------|------|--------|
| **总 spec 数** | 49 | - |
| **✅ 格式完全正确** | 16 | 33% |
| **❌ 格式有问题** | 33 | 67% |
| **⚠️ 轻微问题** | ~6 | 12% |
| **❌ 严重错误** | ~27 | 55% |

---

## ✅ 格式完全正确的 Spec (16 个)

这些文件完全符合规范要求：

```
✅ agent-chat
✅ ai-context-enhancement
✅ cli-config-management
✅ command-autocomplete
✅ config-validation
✅ dynamic-prompts
✅ local-config
✅ memory-knowledge
✅ memory-search
✅ memory-session
✅ provider-registry
✅ skill-cli
✅ skill-content-tool
✅ skill-management
✅ skill-selector
✅ tui-memory-browser
```

**共性特征**:
- ✓ 以 `# 规格：` 开头
- ✓ 有完整的 `## Purpose` 章节
- ✓ 有 `## Requirements` 章节
- ✓ 包含多个 `### Requirement:` 小节
- ✓ 每个需求有 `#### Scenario:` 场景
- ✓ 场景包含 WHEN/THEN 或 当/那么 格式

---

## ⚠️ 有轻微问题的 Spec (6 个)

这些文件有内容但格式需要调整：

### 1. cli-chat
```
问题: 
  ❌ 缺少 '## Purpose' 章节
  ⚠️ 使用旧格式 '## 增加需求' 而非 '## Requirements'
  需求: 6 | 场景: 16

修复: 
  1. 添加 ## Purpose 章节
  2. 将 '## 增加需求' 改为 '## Requirements'
  3. 重新格式化需求为 '### Requirement:' 标题
```

### 2. knowledge-read
```
问题: 
  ❌ 缺少 '## Purpose' 章节
  ⚠️ 使用旧格式 '## 增加需求' 而非 '## Requirements'
  需求: 5 | 场景: 12

修复: 同上
```

### 3. knowledge-write
```
问题: 
  ❌ 缺少 '## Purpose' 章节
  ⚠️ 使用旧格式 '## 增加需求' 而非 '## Requirements'
  需求: 7 | 场景: 18

修复: 同上
```

### 4. session-persist
```
问题: 
  ❌ 缺少 '## Purpose' 章节
  ⚠️ 使用旧格式 '## 增加需求' 而非 '## Requirements'
  ⚠️ 场景内容不完整 (WHEN:10 THEN:20 vs Scenarios:10)
  需求: 5 | 场景: 10

修复: 
  1. 添加 ## Purpose 章节
  2. 将 '## 增加需求' 改为 '## Requirements'
  3. 审查场景的 WHEN-THEN 对应关系
```

### 5. session-resume
```
问题: 
  ❌ 缺少 '## Purpose' 章节
  ⚠️ 使用旧格式 '## 增加需求' 而非 '## Requirements'
  ⚠️ 场景内容不完整 (WHEN:13 THEN:26 vs Scenarios:13)
  需求: 5 | 场景: 13

修复: 同上
```

### 6. tool-error-protection
```
问题: 
  ❌ 缺少 '## Purpose' 章节
  ⚠️ 使用旧格式 '## 增加需求' 而非 '## Requirements'
  ⚠️ 场景内容不完整 (WHEN:9 THEN:21 vs Scenarios:9)
  需求: 7 | 场景: 9

修复: 同上
```

---

## ❌ 格式严重错误的 Spec (27 个)

### 按错误类型分类

#### 1️⃣ 缺少规格标题 `# 规格：` (10 个)

```
❌ cli-version          标题: # cli-version 规范
❌ cmd-structure        标题: # Command Structure
❌ position-management  标题: # Position Management
❌ release-automation   标题: # Release Automation
❌ self-update          标题: # Self-update
❌ session-tagging      标题: # 会话标签规格
❌ skill-asset-tool     标题: # Skill Asset Tool
❌ skill-reference-tool 标题: # Skill Reference Tool
❌ tool-response-format 标题: # Tool Response Format
❌ trade-analysis       标题: # Trade Analysis
```

**修复**: 将标题改为 `# 规格：<capability-name>` 格式

#### 2️⃣ 缺少 Purpose 章节 (25 个)

```
❌ account-management      ❌ agent-prompt            ❌ amount-locking
❌ account-summary         ❌ cli-chat                ❌ knowledge-read
❌ add-session-tag         ❌ knowledge-write         ❌ list-knowledge-files
❌ position-calculation    ❌ query-sessions-by-date  ❌ read-knowledge-file
❌ self-update             ❌ session-persist         ❌ session-resume
❌ skill-asset-tool        ❌ skill-reference-tool    ❌ skill-system
❌ todo-management         ❌ tool-error-protection   ❌ tool-response-format
❌ transaction-recording   ❌ web-scraping            ❌ web-search
❌ write-knowledge-file
```

**修复**: 添加 2-3 句话的 `## Purpose` 章节

#### 3️⃣ 缺少 Requirements 标题但有需求 (13 个)

```
❌ account-summary         需求: 6  场景: 10
❌ agent-prompt            需求: 4  场景: 8
❌ cli-chat                需求: 6  场景: 16
❌ knowledge-read          需求: 5  场景: 12
❌ knowledge-write         需求: 7  场景: 18
❌ session-persist         需求: 5  场景: 10
❌ session-resume          需求: 5  场景: 13
❌ skill-asset-tool        需求: 1  场景: 4
❌ skill-reference-tool    需求: 1  场景: 4
❌ skill-system            需求: 4  场景: 8
❌ todo-management         需求: 6  场景: 15
❌ tool-error-protection   需求: 7  场景: 9
❌ tool-response-format    需求: 4  场景: 7
```

**修复**: 在所有需求前添加 `## Requirements` 标题

#### 4️⃣ 使用旧格式 `## 需求:` 或 `## 增加需求` (17 个)

```
❌ account-management      使用: ## 需求：创建账户
❌ add-session-tag         使用: ## 需求：为 Session 添加标签
❌ amount-locking          使用: ## 需求：买入订单锁定金额
❌ cli-chat                使用: ## 增加需求
❌ knowledge-read          使用: ## 增加需求
❌ knowledge-write         使用: ## 增加需求
❌ list-knowledge-files    使用: ## 需求：列出知识库文件
❌ position-calculation    使用: ## 需求：持仓数量计算
❌ query-sessions-by-date  使用: ## 需求：按日期查询 Session
❌ read-knowledge-file     使用: ## 需求：安全读取知识文件
❌ session-persist         使用: ## 增加需求
❌ session-resume          使用: ## 增加需求
❌ tool-error-protection   使用: ## 增加需求
❌ transaction-recording   使用: ## 需求：创建交易记录
❌ web-scraping            使用: ## 需求：抓取网页内容
❌ web-search              使用: ## 需求：执行网页搜索
❌ write-knowledge-file    使用: ## 需求：安全写入知识文件
```

**修复**: 
1. 将 `## 需求：xxx` 改为 `### Requirement: xxx`
2. 将 `## 增加需求` 改为 `## Requirements`
3. 将场景格式改为 `#### Scenario:`

#### 5️⃣ 完全缺少需求声明 (13 个)

```
❌ account-management       需求: 0  场景: 0
❌ add-session-tag          需求: 0  场景: 0
❌ amount-locking           需求: 0  场景: 0
❌ list-knowledge-files     需求: 0  场景: 0
❌ position-calculation     需求: 0  场景: 0
❌ position-tracking        需求: 0  场景: 0
❌ query-sessions-by-date   需求: 0  场景: 0
❌ read-knowledge-file      需求: 0  场景: 0
❌ trade-execution          需求: 0  场景: 0
❌ transaction-recording    需求: 0  场景: 0
❌ web-scraping             需求: 0  场景: 0
❌ web-search               需求: 0  场景: 0
❌ write-knowledge-file     需求: 0  场景: 0
```

**问题**: 这些文件使用了旧格式（## 需求），检查工具无法正确识别

**修复**: 按 4️⃣ 中的旧格式转换说明处理

#### 6️⃣ 场景内容不完整 (6-7 个)

```
⚠️ account-summary         WHEN:10  THEN:18  vs Scenarios:10
⚠️ agent-prompt            WHEN:8   THEN:7   vs Scenarios:8
⚠️ cli-chat                WHEN:19  THEN:33  vs Scenarios:16
⚠️ config-validation       WHEN:24  THEN:24  vs Scenarios:23
⚠️ session-persist         WHEN:10  THEN:20  vs Scenarios:10
⚠️ session-resume          WHEN:13  THEN:26  vs Scenarios:13
⚠️ tool-error-protection   WHEN:9   THEN:21  vs Scenarios:9
```

**问题**: WHEN/THEN 数量与 Scenario 数量不对应

**修复**: 审查每个场景，确保完整的 WHEN-THEN 对

---

## 🔧 修复优先级

### 🟥 P0 - 关键问题（影响文件结构理解）

| 优先级 | 任务 | 受影响文件 | 预期工作量 |
|--------|------|-----------|-----------|
| **P0-1** | 添加缺失的 `## Purpose` 章节 | 25 个 | 中等 |
| **P0-2** | 修正规格标题格式为 `# 规格：` | 10 个 | 低 |
| **P0-3** | 修正 Requirements 章节结构 | 13 个 | 中等 |

**预期工作量**: 约 4-6 小时

### 🟧 P1 - 中等问题（影响内容质量和一致性）

| 优先级 | 任务 | 受影响文件 | 预期工作量 |
|--------|------|-----------|-----------|
| **P1-1** | 转换旧格式为标准格式 | 17 个 | 中等 |
| **P1-2** | 检查并完善场景内容 | 6-7 个 | 中等 |
| **P1-3** | 调查和完成无需求的文件 | 13 个 | 高 |

**预期工作量**: 约 8-12 小时

---

## 📋 标准 Spec 文件结构

所有 spec 文件应遵循以下标准结构：

```markdown
# 规格：capability-name

[可选的知识上下文引用或简述]

## Purpose

[2-3 句话描述这个能力的核心目的]
- 这个功能/能力要解决什么问题
- 为什么这个能力很重要
- 这个能力的核心价值是什么

## Requirements

### Requirement: 需求标题 1

需求的详细描述。

#### Scenario: 场景描述 1

- **WHEN** / **当** [触发条件]
- **THEN** / **那么** [预期结果]
- **AND** / **并且** [可选的补充条件或结果]

#### Scenario: 场景描述 2

- **WHEN** / **当** [触发条件]
- **THEN** / **那么** [预期结果]

### Requirement: 需求标题 2

...

[可选的其他章节如 Test Plan, Implementation Notes 等]
```

---

## ✅ 验证检查清单

使用以下清单验证每个修复后的 spec 文件：

```
[ ] 1. 标题格式: # 规格：<capability-name>
[ ] 2. 有 ## Purpose 章节且非空（2-3 句话）
[ ] 3. 有 ## Requirements 章节
[ ] 4. 至少有一个 ### Requirement: 小节
[ ] 5. 每个需求至少有一个 #### Scenario: 场景
[ ] 6. 每个场景有 **WHEN**/当 和 **THEN**/那么 语句
[ ] 7. 没有使用旧格式（## 需求、## 增加需求、ADDED Requirements 等）
[ ] 8. 正确使用英文或中文（不混用，除非特殊情况）
[ ] 9. 场景中的 WHEN 数量 ≥ 1，THEN 数量 ≥ 1
[ ] 10. 所有链接和引用都有效
```

---

## 📞 后续步骤

### 第一阶段：修正结构 (P0)
1. ✅ 所有文件添加 `# 规格：` 标题
2. ✅ 所有文件添加 `## Purpose` 章节
3. ✅ 所有文件添加 `## Requirements` 章节

### 第二阶段：转换格式 (P1)
1. ✅ 将旧格式（`## 需求`）转换为新格式（`### Requirement`）
2. ✅ 验证 Requirements 章节的完整性
3. ✅ 检查 Purpose 章节的内容质量

### 第三阶段：内容验证
1. ✅ 审查所有场景的 WHEN-THEN 结构
2. ✅ 完成空白或缺失内容的 spec 文件
3. ✅ 再次运行验证脚本确保 100% 合规

---

## 🔍 验证命令

为了在修复后验证结果，可以使用以下命令：

```bash
# 重新运行验证脚本
python3 /tmp/validate_specs.py

# 查看具体文件
cat /Users/dragon/Documents/Code/goland/msa/.worktrees/arch-refactor/openspec/specs/<capability>/spec.md
```

---

**报告生成时间**: 2026-04-19  
**下次验证建议**: 完成所有 P0 级别问题后
