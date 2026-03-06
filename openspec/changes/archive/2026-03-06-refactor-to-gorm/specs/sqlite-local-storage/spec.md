# 规格：sqlite-local-storage

> 📚 **知识上下文**：基于 `add-sqlite-local-db` 变更
>
> **原有实现**：使用 `database/sql` 和原生 SQL 语句进行数据库操作

## 修改需求

### [R-1] 数据库初始化（MODIFIED）

**原实现**：使用 `sql.Open()` 和 `sql.DB`

**新实现**：使用 GORM 的 `gorm.Open()`

#### 场景：初始化数据库连接
- **当** 调用 `InitDB()`
- **那么** 返回 `*gorm.DB` 而非 `*sql.DB`
- **并且** 使用 `gorm.Open()` 配置 SQLite driver
- **并且** 配置连接池参数以优化并发性能

#### 场景：测试环境初始化
- **当** 调用 `InitDBWithPath(dbPath)`
- **那么** 使用指定的数据库路径
- **并且** 返回 `*gorm.DB` 实例

---

### [R-2] 数据库迁移（MODIFIED）

**原实现**：手写 CREATE TABLE 和 CREATE INDEX SQL 语句

**新实现**：使用 GORM 的 `AutoMigrate()` 功能

#### 场景：执行迁移
- **当** 调用 `Migrate(db *gorm.DB)`
- **那么** 使用 `db.AutoMigrate(&model.Account{}, &model.Transaction{})`
- **并且** 自动创建表和索引
- **并且** 基于模型标签定义约束

#### 场景：迁移失败处理
- **当** 迁移过程中发生错误
- **那么** 返回错误信息
- **并且** 不创建不完整的表结构

---

### [R-3] 数据库连接管理（MODIFIED）

**原实现**：手动管理 `*sql.DB` 连接，需要手动配置 WAL 模式

**新实现**：GORM 自动管理连接池

#### 场景：关闭数据库连接
- **当** 调用 `CloseDB(db *gorm.DB)`
- **那么** 使用 `db.Exec("PRAGMA optimize")` 优化数据库
- **并且** 调用 `sqlDB.Close()` 关闭底层连接

#### 场景：获取底层 sql.DB
- **当** 需要访问底层 `*sql.DB`（如 Ping）
- **那么** 使用 `db.DB()` 获取
- **并且** 可以配置连接池参数

---

> 📚 **迁移说明**：
>
> **Breaking Changes**:
> - `InitDB()` 返回类型从 `*sql.DB` 改为 `*gorm.DB`
> - `Migrate()` 参数类型从 `*sql.DB` 改为 `*gorm.DB`
> - `CloseDB()` 参数类型从 `*sql.DB` 改为 `*gorm.DB`
>
> **Migration Path**:
> 1. 更新所有调用 `InitDB()` 的代码，使用 `*gorm.DB` 类型
> 2. 更新所有数据库操作函数的参数类型
> 3. 更新服务层代码以使用 GORM API
