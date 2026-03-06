# 提案：refactor-to-gorm

将现有的原生 SQLite 数据库操作重构为使用 GORM 框架，简化代码并提高可维护性。

## 知识上下文

### 相关历史问题

知识库中未找到与 GORM 或数据库重构相关的历史问题。

不过，在实现 `add-sqlite-local-db` 变更时遇到了以下问题：
- **测试超时**: 原生 SQL 配置导致服务层测试超时
- **并发问题**: SQLite 连接配置不当导致事务处理超时

### 需避免的反模式

未识别到特定反模式。

### 推荐模式

未识别到特定模式。

---

## 为什么

当前使用原生 SQL 操作 SQLite 数据库存在以下问题：

1. **代码冗长**: 需要手写大量 SQL 语句
2. **维护成本高**: 表结构变更需要修改多处 SQL 代码
3. **类型转换繁琐**: 需要手动处理数据库字段到 Go 结构体的映射
4. **事务处理复杂**: 需要手动管理事务的开始、提交和回滚
5. **测试困难**: 原生 SQL 配置导致并发测试超时

使用 GORM 可以解决这些问题，提供更简洁、更易维护的数据库操作方式。

## 变更内容

- 添加 GORM 依赖 (`gorm.io/gorm` 和 `gorm.io/driver/sqlite`)
- 重构 `pkg/db/` 包，使用 GORM 替代原生 SQL
- 更新 `pkg/model/` 包，添加 GORM 标签
- 移除原生 SQL 相关代码
- 更新 `pkg/service/` 包以使用 GORM 接口

## 能力

### 新增能力

- `gorm-orm`: GORM 框架集成，提供 ORM 数据库操作能力

### 修改能力

- `sqlite-local-storage`: 使用 GORM 实现本地存储能力
- `account-management`: 使用 GORM 实现账户管理
- `transaction-recording`: 使用 GORM 实现交易记录
- `position-calculation`: 使用 GORM 实现持仓计算
- `amount-locking`: 使用 GORM 实现金额锁定

## 影响

**新增依赖**:
```
gorm.io/gorm
gorm.io/driver/sqlite
```

**代码变更**:
- `pkg/db/` - 完全重构，使用 GORM
- `pkg/model/` - 添加 GORM 标签 (gorm.Model)
- `pkg/service/` - 更新以使用 GORM 接口

**好处**:
- 代码量减少约 40-50%
- 类型安全，编译时检查
- 自动处理数据库迁移
- 解决测试超时问题

---

**知识备注**:
- 此提案使用 `openspec/knowledge/index.yaml` 中的知识生成
- 这是新领域，如遇问题请创建 fix.md 来构建知识库
- 最后知识同步：2025-03-06
- 知识库总问题数：5
- 总模式数：3
