## ADDED Requirements

### Requirement: 盘前准备 Skill 模块

系统必须新增 `pre-market-analysis` skill，在 8:30-9:25 盘前窗口执行。

Skill 元数据:
- `name`: pre-market-analysis
- `priority`: 8
- `pattern`: pipeline
- `steps`: 3
- `triggers`: time "8:30-9:25", session `pre-market-session`
- `dependencies`: trading-common, stock-analysis

#### Scenario: 早盘触发盘前分析

- **GIVEN** 当前时间在 8:30-9:25 之间，session 标识为 pre-market-session
- **WHEN** 用户触发交易流程
- **THEN** 应优先加载 pre-market-analysis skill，执行 3 步流程

---

### Requirement: Step 1 — 板块热度与市场全景

盘前分析第一步必须收集板块热度和市场全景数据。

使用工具:
- `get_board_rank(type=01, order=0)` 获取申万行业涨幅排行
- `get_board_rank(type=02, order=0)` 获取概念板块涨幅排行
- `get_board_rank(type=01, order=1)` 获取跌幅最大板块（风险识别）

#### Scenario: 获取板块热度前20

- **GIVEN** 盘前分析启动
- **WHEN** 执行 Step 1 信息收集
- **THEN** 应返回概念板块 TOP20 涨幅榜和行业板块跌幅榜，传递给 Step 2 分析

---

### Requirement: Step 2 — 隔夜外盘与新闻

盘前分析第二步必须分析隔夜外盘影响。

- 查询美股三大指数收盘情况 (道琼斯/纳斯达克/标普500)
- 查询 A50 期货走势
- 查询港股开盘表现
- 通过 web_search 获取财经新闻头条

#### Scenario: 外盘大跌预示低开

- **GIVEN** 隔夜美股纳斯达克下跌 3%，A50 期货下跌 1.5%
- **WHEN** 盘前分析执行 Step 2
- **THEN** 应在分析报告中标注"外盘利空，预计 A 股低开 1-2%，今日操作偏谨慎"

---

### Requirement: Step 3 — 操作预判

盘前分析第三步必须生成当日操作预案，传递给 morning-analysis。

- 基于板块热度、外盘影响、新闻事件生成预判
- 标注今日重点关注板块和个股
- 标注今日市场风险等级（低/中/高）
- 输出格式使用 output-formats skill 的 analysis-report 模板

#### Scenario: 生成操作预判

- **GIVEN** Step 1 发现半导体、航天板块强势，Step 2 外盘平稳
- **WHEN** 盘前分析执行 Step 3
- **THEN** 应输出包含"关注半导体、航天板块 + 今日风险：低"的预判报告
