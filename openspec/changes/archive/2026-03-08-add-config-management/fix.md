# 修复：add-config-management

---
id: "fix-001"
title: "EditorModel 悬垂指针导致编辑器卡死"
type: "runtime"
severity: "critical"
component: "pkg/tui/config"
tags:
  - "dangling-pointer"
  - "memory-safety"
  - "tui"
created_at: "2026-03-08"
resolved_at: "2026-03-08"
change: "openspec/changes/add-config-management"
status: "resolved"
---

## 问题

### 问题描述

在 TUI 配置界面中编辑 API Key 等字段时，键盘输入完全无效，编辑器无法接收任何按键。

### 重现步骤

1. 运行 `./msa config`
2. 选择任意配置项（如 API Key）
3. 按 `Enter` 进入编辑模式
4. 尝试输入任何字符
5. **观察到**：编辑器没有任何响应，光标不移动

### 错误消息

无明确错误消息，程序不崩溃但无法正常工作。

### 影响评估

- **用户影响**：用户无法通过 TUI 配置界面编辑配置，核心功能完全不可用
- **系统影响**：整个配置管理 TUI 功能失效
- **业务影响**：严重阻碍配置管理功能的交付

## 根本原因

### 调查过程

1. 检查 `Update()` 方法是否正确调用
2. 检查消息传递链：ConfigModel → updateEditMode → EditorModel
3. 添加日志发现 `updateEditMode` 先检查 `tea.QuitMsg`，阻止了正常的 KeyMsg 处理
4. 进一步检查发现 `EditorModel` 的指针问题

### 根本原因

**Dangling Pointer（悬垂指针）问题**：

在 `enterEditMode()` 函数中：
```go
var editor *EditorModel
if item.Key == "apikey" {
    e := NewEditorModel(...)  // e 是局部变量（栈上分配）
    editor = &e             // 取局部变量的地址
}
m.state.EditorModel = editor  // 保存指针
```

当 `enterEditMode()` 函数返回后，局部变量 `e` 被销毁，但 `editor` 指针仍然指向那个已失效的内存地址。后续对该指针的任何操作都是未定义行为。

### 相关文件

- `pkg/tui/config/config.go:258-294`
  - 行：260-294 `enterEditMode()` 函数
  - 说明：创建了局部变量然后取地址存储

- `pkg/tui/config/editor.go:55-75`
  - 行：56 `NewEditorModel()` 返回值类型
  - 说明：返回 `EditorModel` 值类型而非指针

## 解决方案

### 解决方案方法

1. 修改 `NewEditorModel()` 返回指针类型（`*EditorModel`）
2. 移除中间局部变量，直接使用返回的指针
3. 统一使用指针接收器更新方法

### 代码变更

#### 变更 1：修改 NewEditorModel 返回指针

**文件**：`pkg/tui/config/editor.go`

```diff
  // NewEditorModel 创建新的编辑模型
- func NewEditorModel(title, label, value string, isSecret bool) EditorModel {
+ func NewEditorModel(title, label, value string, isSecret bool) *EditorModel {
      input := textinput.New()
      input.Placeholder = "输入新值..."
      input.SetValue(value)
      input.Focus()

      if isSecret {
          input.EchoMode = textinput.EchoPassword
      }

-     return EditorModel{
+     return &EditorModel{
          title:     title,
          label:     label,
          value:     value,
          input:     input,
          isSecret:  isSecret,
          quitting:  false,
          aborted:   false,
      }
  }
```

#### 变更 2：简化 enterEditMode 使用

**文件**：`pkg/tui/config/config.go`

```diff
  // 文本编辑器
  var editor *EditorModel
  isSecret := item.IsSecret

  if item.Key == "apikey" {
-     e := NewEditorModel("编辑 API Key", "API Key", item.Value, isSecret)
-     e.SetValidator(func(v string) []*ValidationError {
+     editor = NewEditorModel("编辑 API Key", "API Key", item.Value, isSecret)
+     editor.SetValidator(func(v string) []*ValidationError {
          cfg := config.GetLocalStoreConfig()
          return validateAPIKey(v, string(cfg.Provider))
      })
-     editor = &e
  }
```

#### 变更 3：统一 Update 方法为指针接收器

**文件**：`pkg/tui/config/editor.go`

```diff
  // Init bubbletea.Model 接口实现
- func (m EditorModel) Init() tea.Cmd {
+ func (m *EditorModel) Init() tea.Cmd {
      return textinput.Blink
  }

  // Update bubbletea.Model 接口实现
- func (m EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
+ func (m *EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
      // ... 实现
  }

  // View bubbletea.Model 接口实现
- func (m EditorModel) View() string {
+ func (m *EditorModel) View() string {
      // ... 实现
  }

  // GetValue 获取编辑后的值
- func (m EditorModel) GetValue() string {
+ func (m *EditorModel) GetValue() string {
      return m.input.Value()
  }

  // IsAborted 是否被中止
- func (m EditorModel) IsAborted() bool {
+ func (m *EditorModel) IsAborted() bool {
      return m.aborted
  }

+ // IsDone 是否已完成
+ func (m *EditorModel) IsDone() bool {
+     return m.quitting
+ }
```

**说明**：确保所有方法使用指针接收器，这样修改的是同一实例。

**原因**：Go 语言中，值类型的变量在函数返回后会被销毁，保存其地址会导致悬垂指针。

## 验证

### 验证步骤

1. 重新编译项目：`go build`
2. 运行配置界面：`./msa config`
3. 选择 API Key 配置项
4. 按 `Enter` 进入编辑
5. 输入测试字符：`test-key-12345`
6. 观察编辑器正常响应，字符显示在输入框中
7. 按 `Enter` 确认，值正确应用到配置项
8. 按 `S` 保存，配置成功保存到文件

### 测试结果

- 单元测试：**PASS** - `pkg/tui/config` 所有测试通过
- 集成测试：**PASS** - 手动测试编辑功能正常工作

---

## 可复用模式

**模式名称**：TUI 子模型指针管理
**模式ID**：p006

**描述**：

在 bubbletea TUI 框架中，当主模型需要管理子模型（如编辑器、选择器）时，子模型必须以指针形式存储，并且必须提供状态检查方法（如 `IsDone()`）。

**适用场景**：
- 所有使用 bubbletea 的 TUI 应用
- 需要动态切换子模型的场景
- 编辑器、选择器等交互组件

**实现**：

```go
// ✅ 正确：返回指针
func NewEditorModel() *EditorModel {
    return &EditorModel{
        // 初始化字段
    }
}

// ✅ 正确：提供状态检查方法
func (m *EditorModel) IsDone() bool {
    return m.quitting
}

// ✅ 正确：主模型中的处理
func (m *ConfigModel) updateEditMode(msg tea.Msg) (tea.Model, tea.Cmd) {
    var teaModel tea.Model = m.state.EditorModel
    teaModel, cmd = teaModel.Update(msg)
    m.state.EditorModel = teaModel

    // 检查子模型状态
    if editor, ok := teaModel.(*EditorModel); ok && editor.IsDone() {
        // 处理完成逻辑
    }

    return m, cmd
}
```

**收益**：
- 避免悬垂指针问题
- 确保状态更新的正确性
- 提供清晰的子模型生命周期管理

---

## 反模式

**反模式名称**：取局部变量的地址存储到接口
**反模式ID**：a006

**问题所在**：

在 Go 语言中，局部变量存储在栈上，函数返回后会被自动回收。取局部变量的地址并存储到接口或持久化结构中会导致悬垂指针，后续访问会产生未定义行为。

**避免**：

```go
// ❌ 错误：取局部变量地址
func createModel() *EditorModel {
    model := NewModel()  // 局部变量
    return &model         // 取地址 - 危险！
}

// ❌ 错误：在循环中取地址
for i := 0; i < 10; i++ {
    item := &Item{ID: i}   // item 在下次循环时失效
    items = append(items, item)
}

// ✅ 正确：直接返回指针
func createModel() *EditorModel {
    return &EditorModel{}  // 堆上分配
}

// ✅ 正确：在循环中取地址
for i := 0; i < 10; i++ {
    item := &Item{ID: i}
    items = append(items, item)  // item 指向有效内存
}
```

**关键规则**：
- 函数返回的指针必须指向堆分配的内存（`&T` 或 `new(T)`）
- 不要取局部变量的地址并长期存储
- 使用指针接收器修改状态

---

## 经验教训

### 出了什么问题

- **悬垂指针导致未定义行为**：程序可能卡死、崩溃或产生随机错误
- **消息处理逻辑错误**：`switch msg.(type)` 检查 `tea.QuitMsg` 阻止了正常的按键消息处理
- **类型不匹配**：值类型和指针类型混用导致状态更新丢失

### 学到了什么

- **Go 内存模型基础**：局部变量在栈上，函数返回后失效
- **bubbletea 子模型模式**：子模型应该是指针，提供 `IsDone()` 状态检查
- **TUI 消息处理流程**：先更新模型，再检查状态，而不是先检查消息类型

### 预防措施

1. **代码审查检查项**：
   - 是否有 `&localVariable` 模式
   - 是否正确使用指针接收器
   - 子模型是否提供状态检查方法

2. **单元测试覆盖**：
   - 测试模型的创建和销毁
   - 测试状态转换逻辑
   - 测试消息处理流程

3. **编译器警告**：
   - 启用 Go 1.18+ 的 `-d=checkptr=dangling` 检测器
   - 使用 `go vet` 检测常见问题

4. **设计模式**：
   - 使用工厂函数返回指针：`func New() *Type`
   - 避免在函数间传递局部变量的地址
   - 子模型实现状态接口：`IsDone()`, `IsAborted()` 等
