## ADDED Requirements

### Requirement: 板块排行工具

系统必须新增 `get_board_rank` Go 工具，封装腾讯财经 mktHs/rank 接口获取板块排行。

工具规格:
- `name`: get_board_rank
- `ToolGroup`: stock
- 参数: `board_type`(string: 01=申万行业/02=概念板块/03=地域板块), `order`(int: 0=降序涨幅最大/1=升序跌幅最大), `count`(int)
- 返回: 板块名称、代码、指数、涨跌幅、领涨股、5日/20日涨跌幅

#### Scenario: 获取概念板块涨幅排行

- **GIVEN** 需要了解当日热点概念
- **WHEN** 调用 `get_board_rank(board_type="02", order=0, count=20)`
- **THEN** 应返回涨幅最大的 20 个概念板块，每个含板块名称、代码、涨跌幅、领涨股

#### Scenario: 获取行业板块跌幅排行用于风险识别

- **GIVEN** 需要识别当日弱势行业
- **WHEN** 调用 `get_board_rank(board_type="01", order=1, count=10)`
- **THEN** 应返回跌幅最大的 10 个申万行业板块

#### Scenario: 获取地域板块排行

- **GIVEN** 需要了解地域资金流向
- **WHEN** 调用 `get_board_rank(board_type="03", order=0, count=10)`
- **THEN** 应返回涨幅最大的 10 个地域板块（如青海、海南、上海等）

---

### Requirement: API 响应解析

工具必须正确解析 mktHs/rank API 的 JSON 响应。

API 端点: `https://proxy.finance.qq.com/ifzqgtimg/appstock/app/mktHs/rank?l={count}&p=1&t={board_type}/averatio&o={order}`

#### Scenario: 解析板块排行数据

- **GIVEN** API 返回 `{"code":0,"data":[{"bd_name":"半导体","bd_code":"pt01801081","bd_zdf":"4.71","nzg_name":"寒武纪","nzg_zdf":"20.00",...}]}`
- **WHEN** 工具解析响应
- **THEN** 应正确提取 bd_name(板块名称)、bd_code(板块代码)、bd_zdf(涨跌幅)、nzg_name(领涨股) 等字段

---

### Requirement: 板块成分股查询(可选增强)

工具应支持通过板块代码查询成分股，复用现有 `getBoardRankList` API。

- 使用 `board_code={板块代码}` 参数查询成分股
- 支持按涨跌幅/换手率/主力资金等排序

#### Scenario: 查询半导体板块成分股

- **GIVEN** 已获取半导体板块代码 pt01801081
- **WHEN** 调用板块成分股查询 `board_code=pt01801081&sort_type=priceRatio&direct=down&count=10`
- **THEN** 应返回半导体板块涨幅前 10 的成分股
