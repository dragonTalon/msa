# 修复：add-finance-tools

---
id: "001"
title: "CGO 链接错误 - macOS 13.2.1 SDK 不兼容"
type: "compile"
severity: "critical"
component: "pkg/db"
tags:
  - "cgo"
  - "sqlite"
  - "macos"
created_at: "2026-03-07"
resolved_at: "2026-03-07"
change: "openspec/changes/add-finance-tools"
status: "resolved"
---

## 问题

### 问题描述
使用 `gorm.io/driver/sqlite` (底层使用 `mattn/go-sqlite3`) 编译时出现链接错误：

```
Undefined symbols for architecture x86_64:
  "_SecTrustCopyCertificateChain", referenced from:
    _crypto/x509/internal/macos.x509_SecTrustCopyCertificateChain_trampoline.abi0 in go.o
```

### 重现步骤
1. macOS 13.2.1 + Go 1.25.5
2. 运行 `go build`
3. 观察到链接错误

### 错误消息
```
ld: symbol(s) not found for architecture x86_64
clang: error: linker command failed with exit code 1
```

### 影响评估
- **用户影响**：无法编译项目
- **系统影响**：项目无法构建
- **业务影响**：阻塞开发进度

## 根本原因

### 调查过程
1. 确认错误来自 `mattn/go-sqlite3` 的 CGO 依赖
2. 该驱动需要链接 macOS Security Framework
3. macOS 13.2.1 的 Command Line Tools SDK 缺少 `_SecTrustCopyCertificateChain` 符号
4. 尝试设置 `SDKROOT` 环境变量无效
5. `CGO_ENABLED=0` 可编译但驱动变为 stub

### 根本原因
`mattn/go-sqlite3` 需要 CGO，而当前 macOS SDK 与 Go 1.25.5 不兼容，缺少必需的 Security Framework 符号。

### 相关文件
- `pkg/db/db.go`
  - 行：9
  - 说明：使用了需要 CGO 的驱动

## 解决方案

### 解决方案方法
使用纯 Go 实现的 SQLite 驱动 `github.com/glebarez/sqlite`，该驱动：
- 不需要 CGO
- 与 GORM 完全兼容
- API 与 `gorm.io/driver/sqlite` 相同

### 代码变更

#### 变更描述
**文件**：`pkg/db/db.go`

```diff
- "gorm.io/driver/sqlite"
+ "github.com/glebarez/sqlite"
```

```diff
- // 打开数据库连接，如果不存在会自动创建
  database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
+ // 打开数据库连接，如果不存在会自动创建
+ // 使用纯 Go SQLite 驱动 glebarez/sqlite，不需要 CGO
  database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
```

**说明**：替换驱动，保留相同的使用方式
**原因**：消除 CGO 依赖，解决链接错误

## 验证

### 验证步骤
1. 运行 `go get github.com/glebarez/sqlite`
2. 运行 `go build`
3. 运行 `./msa --help`

### 测试结果
- 单元测试：N/A
- 集成测试：PASS (编译成功，程序正常运行)

## 可复用模式

**模式名称**：CGO-free SQLite
**模式ID**：p001

**描述**：在可能遇到 CGO 兼容性问题的平台（如 macOS），优先使用纯 Go 实现的 SQLite 驱动。

**适用场景**：
- macOS 开发环境
- 交叉编译场景
- 容器化部署

**实现**：
```go
import (
    "github.com/glebarez/sqlite"
    "gorm.io/gorm"
)

database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
```

## 反模式

**反模式名称**：使用 CGO 依赖的 SQLite 驱动
**反模式ID**：a001

**问题所在**：`mattn/go-sqlite3` 需要 CGO，在某些平台可能导致链接失败。

**避免**：
```go
// 不要这样做（除非有特殊需求）
import "gorm.io/driver/sqlite"  // 底层使用 mattn/go-sqlite3

// 应该这样做
import "github.com/glebarez/sqlite"  // 纯 Go 实现
```

## 经验教训

### 出了什么问题
- 选择了默认的 `gorm.io/driver/sqlite`，未考虑 CGO 兼容性问题

### 学到了什么
- 纯 Go 实现的 SQLite 驱动性能已足够好
- CGO 依赖会带来额外的部署复杂性

### 预防措施
1. 优先选择纯 Go 实现的库
2. 在选型时考虑跨平台兼容性
3. 使用 `CGO_ENABLED=0` 测试编译兼容性

---

---
id: "002"
title: "SQLite 数据库文件名后缀错误"
type: "config"
severity: "low"
component: "pkg/db"
tags:
  - "sqlite"
  - "naming"
created_at: "2026-03-07"
resolved_at: "2026-03-07"
change: "openspec/changes/add-finance-tools"
status: "resolved"
---

## 问题

### 问题描述
数据库文件使用 `.sqlite` 后缀，而 SQLite 标准后缀是 `.db`。

### 重现步骤
1. 查看 `pkg/db/db.go` 第 24 行
2. 发现文件名为 `msa.sqlite`

### 影响评估
- **用户影响**：文件名不符合惯例，可能造成混淆
- **系统影响**：无
- **业务影响**：无

## 根本原因

### 根本原因
开发时使用了不标准的 `.sqlite` 后缀。

### 相关文件
- `pkg/db/db.go`
  - 行：24
  - 说明：`dbPath := filepath.Join(dbDir, "msa.sqlite")`

## 解决方案

### 解决方案方法
1. 将文件名从 `msa.sqlite` 改为 `msa.db`
2. 重命名已存在的数据库文件

### 代码变更

#### 变更描述
**文件**：`pkg/db/db.go`

```diff
- // 路径为 ~/.msa/msa.sqlite
```

```diff
- dbPath := filepath.Join(dbDir, "msa.sqlite")
+ dbPath := filepath.Join(dbDir, "msa.db")
```

**说明**：使用标准的 SQLite 数据库文件后缀
**原因**：符合行业惯例

## 验证

### 验证步骤
1. 重命名 `~/.msa/msa.db` 为 `~/.msa/msa.sqlite`
2. 运行 `go build`
3. 运行 `./msa --help`

### 测试结果
- 手动验证：PASS (数据库文件正确重命名，程序正常运行)

## 可复用模式

**模式名称**：SQLite 数据库文件命名规范
**模式ID**：p002

**描述**：SQLite 数据库文件应使用 `.db` 作为文件后缀。

**适用场景**：
- 所有 SQLite 数据库文件

**实现**：
```go
dbPath := filepath.Join(dbDir, "app.db")  // 而不是 app.sqlite
```

## 经验教训

### 出了什么问题
- 使用了非标准的文件命名约定

### 学到了什么
- SQLite 的标准后缀是 `.db`，不是 `.sqlite`

### 预防措施
1. 遵循行业标准惯例
2. 参考主流项目的命名方式
---
>
> ---
> id: "001"
> title: "问题标题"
> type: "runtime|compile|logic|config"
> severity: "critical|high|medium|low"
> component: "pkg/logic/tools/finance"
> tags:
>   - "tag1"
>   - "finance"
> created_at: "2026-03-07"
> resolved_at: "2026-03-07"
> change: "openspec/changes/add-finance-tools"
> status: "open|investigating|resolved|cancelled"
> ---
>
> ## 问题
>
> ### 问题描述
> <!-- 发生了什么？应该发生什么？实际行为是什么？ -->
>
> ### 重现步骤
> 1. ...
> 2. ...
> 3. 观察到错误
>
> ### 错误消息
> ```
> <!-- 粘贴确切的错误消息和堆栈跟踪 -->
> ```
>
> ### 影响评估
> - **用户影响**：...
> - **系统影响**：...
> - **业务影响**：...
>
> ## 根本原因
>
> ### 调查过程
> <!-- 你如何调查的 -->
>
> ### 根本原因
> <!-- 明确陈述根本原因 -->
>
> ### 相关文件
> - `path/to/file`
>   - 行：123
>   - 说明：...
>
> ## 解决方案
>
> ### 解决方案方法
> <!-- 采取的方法、替代方案 -->
>
> ### 代码变更
>
> #### 变更描述
> **文件**：`path/to/file`
>
> ```diff
> + // 新增代码
> - // 删除代码
>   // 修改代码
> ```
>
> **说明**：...
> **原因**：...
>
> ## 验证
>
> ### 验证步骤
> 1. ...
> 2. ...
>
> ### 测试结果
> - 单元测试：PASS/FAIL
> - 集成测试：PASS/FAIL
>
> ## 可复用模式
>
> <!-- 如发现可复用模式，在此记录 -->
>
> **模式名称**：...
> **模式ID**：pXXX
>
> **描述**：...
>
> **适用场景**：
> - ...
>
> **实现**：
> ```go
> ...
> ```
>
> ## 反模式
>
> <!-- 如发现需要避免的反模式，在此记录 -->
>
> **反模式名称**：...
> **反模式ID**：aXXX
>
> **问题所在**：...
>
> **避免**：
> ```go
> // 不要这样做
> ...
> ```
>
> ## 经验教训
>
> ### 出了什么问题
> - ...
>
> ### 学到了什么
> - ...
>
> ### 预防措施
> 1. ...
> 2. ...
>
> ---
