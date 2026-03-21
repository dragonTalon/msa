# 探索记录：Skill 系统重构

**日期**: 2026-03-21
**主题**: 基于 Google ADK 五种设计模式重构 Skill 系统

---

## 背景

用户分享了一篇关于 Google ADK Agent Skill 设计模式的文章，强调结构化的 Skill 定义，包括：
- 明确的触发条件
- 工具声明
- 输入输出规范

文章提出了五种设计模式：
1. **Tool Wrapper** - 让 agent 成为领域专家
2. **Generator** - 从模板生成结构化输出
3. **Reviewer** - 按检查清单评分
4. **Inversion** - Agent 先采访用户再行动
5. **Pipeline** - 严格的顺序工作流

---

## 问题诊断

### 当前设计差距

| 方面 | 原实现 | 问题 |
|------|--------|------|
| 元数据 | 只有 name, description, version, priority | 缺少 pattern, triggers, tools, dependencies |
| 目录结构 | 单一 SKILL.md 文件 | 无 references/ 和 assets/ 目录 |
| 内容组织 | 所有内容内嵌在一个文件 | 关注点未分离，token 消耗大 |
| 设计模式 | 无模式标识 | 无法让系统知道 Skill 的行为模式 |

### 原 SKILL.md 格式
```yaml
---
name: morning-analysis
description: 早盘开市分析...
version: 1.0.0
priority: 8
---
# Morning Analysis
## 执行流程
...（所有内容都在一个文件）
```

---

## 重构方案

### 新的 YAML 元数据格式
```yaml
---
name: morning-analysis
description: 早盘开市操作...
version: 2.0.0
priority: 8
pattern: pipeline          # 新增：设计模式类型
steps: 4                   # 新增：Pipeline 步骤数
triggers:                  # 新增：触发条件
  - time: "9:30-11:30"
    session: morning-session
tools:                     # 新增：依赖的工具
  - get_account_summary
  - submit_buy_order
dependencies:              # 新增：依赖的其他 Skill
  - trading-common
  - read-error
---
```

### 新的目录结构
```
skills/morning-analysis/
├── SKILL.md                 # 核心指令和工作流定义
├── references/              # 按需加载的知识
│   ├── sell-conditions.md   # 卖出条件规则
│   └── buy-conditions.md    # 买入条件规则
└── assets/                  # 模板和静态资源
    └── report-template.md   # 输出格式模板
```

### Skill 与设计模式映射

| Skill | 设计模式 | 说明 |
|-------|----------|------|
| trading-common | Tool Wrapper | 交易基础知识库 |
| stock-analysis | Tool Wrapper | 股票分析专家 |
| output-formats | Generator | 输出模板库 |
| read-error | Reviewer | 错误审查器 |
| account-management | Inversion | 用户引导流程 |
| morning-analysis | Pipeline | 早盘交易流程 |
| afternoon-trade | Pipeline | 午盘交易流程 |
| market-close-summary | Pipeline + Reviewer | 闭市复盘流程 |

---

## 执行记录

### 代码改动

#### 1. skill.go - 扩展 Skill 结构
- 新增 `SkillPattern` 枚举类型
- 新增 `SkillTrigger` 触发条件结构
- 新增 `SkillMetadata` 扩展元数据结构
- 支持 `references/` 和 `assets/` 目录的按需加载
- 新增 `GetReference()`, `GetAsset()`, `GetDirPath()`, `HasReferences()`, `HasAssets()` 方法

#### 2. loader.go - 扩展加载器
- 新增 `skillMetadataYAML` 结构支持新字段解析
- 修改 `loadSkill()` 接受目录路径而非文件路径
- 支持解析 pattern, triggers, tools, dependencies 等新字段

#### 3. manager.go - 扩展管理器
- 新增 `GetSkillContent()` 方法
- 新增 `GetSkillReference()` 方法
- 新增 `GetSkillAsset()` 方法
- 新增 `GetSkillMetadata()` 方法

#### 4. selector.go - 重命名避免冲突
- `SkillMetadata` → `SkillInfo`（避免与 skill.go 中的新结构冲突）

#### 5. cmd/skill/show.go - 适配新接口
- `GetPath()` → `GetDirPath()`

### SKILL 文件重构

#### trading-common/ (Tool Wrapper)
```
trading-common/
├── SKILL.md
└── references/
    ├── position-rules.md      # 仓位管理规则
    ├── execution-rules.md     # 交易执行规范
    └── file-formats.md        # 文件格式定义
```

#### stock-analysis/ (Tool Wrapper)
```
stock-analysis/
├── SKILL.md
└── references/
    ├── technical-analysis.md   # 技术面分析框架
    ├── fundamental-analysis.md # 基本面分析框架
    └── market-analysis.md      # 市场面分析框架
```

#### output-formats/ (Generator)
```
output-formats/
├── SKILL.md
└── assets/
    ├── investment-advice.md    # 投资建议模板
    ├── analysis-report.md      # 分析报告模板
    └── account-report.md       # 账户报告模板
```

#### read-error/ (Reviewer)
```
read-error/
├── SKILL.md
└── references/
    └── error-checklist.md      # 错误类型检查清单
```

#### account-management/ (Inversion)
```
account-management/
└── SKILL.md                    # 包含分阶段用户引导流程
```

#### morning-analysis/ (Pipeline)
```
morning-analysis/
├── SKILL.md
├── references/
│   ├── sell-conditions.md      # 卖出条件规则
│   └── buy-conditions.md       # 买入条件规则
└── assets/
    └── report-template.md      # 早盘报告模板
```

#### afternoon-trade/ (Pipeline)
```
afternoon-trade/
├── SKILL.md
├── references/
│   └── buy-conditions.md       # 午盘买入条件规则
└── assets/
    └── report-template.md      # 午盘报告模板
```

#### market-close-summary/ (Pipeline + Reviewer)
```
market-close-summary/
├── SKILL.md
├── references/
│   ├── review-criteria.md      # 复盘标准
│   └── error-checklist.md      # 错误检查清单
└── assets/
    └── summary-template.md     # 总结模板
```

---

## 验证结果

```bash
# 编译验证
go build ./pkg/logic/skills/...  # 成功
go build ./...                   # 成功

# 最终文件结构
find pkg/logic/skills/plugs -type f -name "*.md" | wc -l
# 结果: 26 个文件（8 个 SKILL.md + 18 个 references/assets 文件）
```

---

## 主要优势

| 改进 | 说明 |
|------|------|
| **关注点分离** | 知识和模板从主文件中分离 |
| **按需加载** | references/ 和 assets/ 按需加载，节省 context |
| **结构化元数据** | pattern, triggers, tools, dependencies 声明式定义 |
| **设计模式标识** | 每个 Skill 标识其设计模式类型 |
| **门控条件** | Pipeline 模式有明确的 GATE 检查点 |
| **易于维护** | 修改规则只需更新 references 文件 |

---

## 后续建议

1. **运行测试** - 验证功能是否正常
2. **更新 Agent 逻辑** - 让 Agent 能按需加载 references/assets
3. **添加更多 references** - 根据实际需要扩展知识库
4. **文档更新** - 更新项目文档说明新的 SKILL 规范

---

## 参考资源

- [Google ADK Skill Design Patterns](https://ofox.ai/zh/blog/google-adk-agent-skill-design-patterns-2026/)
- 项目原 design.md: `openspec/changes/archive/2026-03-08-dynamic-skills/design.md`
