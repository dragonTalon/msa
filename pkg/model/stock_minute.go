package model

// StockMinuteBar 单根分钟K线数据
type StockMinuteBar struct {
	Time     string  `json:"time"`     // 时间, 格式 HHmm (如 "0930")
	Price    float64 `json:"price"`    // 每分钟成交价(元)
	Volume   int64   `json:"volume"`   // 每分钟成交量(手, 已差分)
	Turnover float64 `json:"turnover"` // 每分钟成交额(元, 已差分)
}

// StockMinuteKResp 分钟K线响应
type StockMinuteKResp struct {
	StockCode string           `json:"stock_code"`
	Date      string           `json:"date"`  // 日期 YYYYMMDD
	Bars      []StockMinuteBar `json:"bars"`  // 分钟K线数组
	Count     int              `json:"count"` // 数据条数
}
