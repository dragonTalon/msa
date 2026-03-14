## MODIFIED Requirements

### Requirement: 输入处理增强

聊天输入框必须处理命令自动补全相关的按键事件。

#### Scenario: 处理 Tab 键补全

- **WHEN** 用户在命令模式下按下 Tab 键
- **THEN** 执行命令补全（如有匹配项）
- **AND** 更新输入框内容
- **AND** 退出命令模式

#### Scenario: 处理方向键导航

- **WHEN** 用户在命令模式下按下 `↑` 或 `↓` 键
- **THEN** 更新选中项索引
- **AND** 重新渲染建议列表
- **AND** 不影响输入框内容

#### Scenario: 处理 Enter 执行命令

- **WHEN** 用户在命令模式下按下 Enter 键
- **THEN** 执行当前选中的命令（如有选中）
- **OR** 执行输入框中的命令（如无选中但输入有效命令）

#### Scenario: 处理 Esc 取消

- **WHEN** 用户在命令模式下按下 Esc 键
- **THEN** 清空输入框
- **AND** 退出命令模式
- **AND** 关闭建议列表

### Requirement: 命令建议数据

Chat 模型必须维护命令建议的状态数据。

#### Scenario: 维护建议状态

- **WHEN** Chat 初始化
- **THEN** 包含以下命令建议相关字段：
  - `commandSelectedIndex` - 当前选中项索引
  - `commandSuggestions` - 建议列表（包含名称和描述）

#### Scenario: 更新建议列表

- **WHEN** 输入内容匹配命令前缀
- **THEN** 调用 `command.GetCommandSuggestions()` 获取带描述的建议
- **AND** 更新 `commandSuggestions` 字段
- **AND** 重置 `commandSelectedIndex` 为 0