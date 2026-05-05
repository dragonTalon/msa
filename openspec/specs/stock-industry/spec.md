## ADDED Requirements

### Requirement: 个股行业分类工具

系统必须新增 `get_stock_industry` Go 工具，封装腾讯财经 jiankuang 接口获取行业分类。

工具规格:
- `name`: get_stock_industry
- `ToolGroup`: stock
- 参数: `stock_code`(string)
- 返回: 行业分类(一级+二级)、主要财务指标、公司基本信息

#### Scenario: 获取个股行业和财务数据

- **GIVEN** 股票代码 sh600519（贵州茅台）
- **WHEN** 调用 `get_stock_industry(stock_code="sh600519")`
- **THEN** 应返回:
  - 一级行业: "食品饮料" (id: 01801120)
  - 二级行业: "白酒Ⅱ" (id: 01801125)
  - 主要财务指标: 每股收益、净利润、营收、ROE
  - 公司信息: 名称、主营业务、地域、上市日期

#### Scenario: 行业分类用于集中度检查

- **GIVEN** 持仓列表包含 3 只股票
- **WHEN** 分别调用 `get_stock_industry` 获取行业
- **THEN** AI 应根据返回的行业 id 判断是否存在同一行业集中问题

---

### Requirement: API 响应解析

工具必须正确解析 jiankuang API 响应中的 `plate` 字段。

API 端点: `https://proxy.finance.qq.com/ifzqgtimg/appstock/app/stockinfo/jiankuang?code={market}{code}`

#### Scenario: 解析 plate 字段

- **GIVEN** API 返回 `{"code":0,"data":{"gsjj":{"plate":[{"name":"食品饮料","id":"01801120","level":"1"},{"name":"白酒Ⅱ","id":"01801125","level":"2"}]}}}`
- **WHEN** 工具解析响应
- **THEN** 应正确提取一级行业和二级行业的 name、id、level
