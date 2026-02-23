## Context

当前项目 pkg/ 目录下约有 40+ 个 Go 源文件，但没有任何测试文件。项目处于活跃开发阶段，功能快速迭代，缺乏测试增加了代码质量风险。

## Goals / Non-Goals

**Goals:**
- 遵循 Go 标准约定创建单元测试
- 覆盖核心业务逻辑和工具函数
- 达到 60% 以上的代码覆盖率
- 建立可持续的测试文化

**Non-Goals:**
- 集成测试和 E2E 测试（后续阶段考虑）
- 性能测试和基准测试
- 100% 覆盖率（优先核心模块）

## Decisions

| 决策 | 理由 | 替代方案 |
|------|------|----------|
| Go 标准约定 `*_test.go` | 工具链原生支持，IDE 友好 | 独立 test 目录 |
| 表格驱动测试 | Go 惯用法，易于扩展和维护 | 手写多个测试用例 |
| testify 辅助库 | 简化断言和 mock 编写 | 纯标准库 |
| 覆盖率 60% | 实用目标，平衡投入产出 | 100%（成本高） |

## 测试模块优先级

```
┌─────────────────────────────────────────────────────────────────┐
│                      测试实现优先级                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Phase 1: 纯函数 (高 ROI)                                       │
│  ├── pkg/utils/format.go    - JSON/字符串处理                   │
│  ├── pkg/utils/http.go      - HTTP 客户端                       │
│  └── pkg/logic/command/cmd.go - 命令注册/查找                   │
│                                                                  │
│  Phase 2: 核心逻辑                                              │
│  ├── pkg/config/              - 配置管理                        │
│  ├── pkg/logic/tools/basic.go - 工具注册                        │
│  └── pkg/logic/provider/      - LLM provider 接口               │
│                                                                  │
│  Phase 3: 业务模块                                              │
│  ├── pkg/logic/agent/         - AI agent 逻辑                   │
│  ├── pkg/logic/tools/stock/   - 股票工具                        │
│  └── pkg/logic/tools/search/  - 搜索工具                        │
│                                                                  │
│  Phase 4: UI 层 (可选)                                          │
│  └── pkg/tui/                  - 终端界面                        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| 涉及文件系统/HTTP 的测试难以 mock | 使用接口抽象，依赖注入 |
| 单例模式影响测试隔离 | 添加重置函数（如已存在的 ResetRestyClient）|
| 配置缓存影响测试结果 | 每个测试前清理状态 |
| 覆盖率目标难以达成 | 分阶段实现，优先核心模块 |

## Testing Guidelines

### 文件组织
```
pkg/
├── utils/
│   ├── format.go
│   ├── format_test.go      ← 新增
│   ├── file.go
│   └── file_test.go        ← 新增
├── config/
│   ├── local_config.go
│   └── local_config_test.go ← 新增
```

### 命名约定
- 测试文件: `xxx_test.go`
- 测试函数: `Test<FunctionName>`
- 子测试: `Test<FunctionName>/<Scenario>`
- 示例: `TestPrettyJSON`, `TestPrettyJSON/WithValidInput`

### Mock 策略
- HTTP 调用: 使用 `httptest` 或 mock server
- 文件系统: 使用临时目录
- 外部 API: 接口抽象 + mock 实现

## Migration Plan

1. 安装测试依赖（testify）
2. 从 utils/ 开始实现测试（ROI 最高）
3. 逐模块扩展到 config/ 和 logic/
4. 运行覆盖率检查
5. 调整直至达到 60% 目标
