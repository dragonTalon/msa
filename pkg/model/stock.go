package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

type SearchResponse struct {
	Stock []StockInfo `json:"stock"`
}
type StockInfo struct {
	Code    string      `json:"code"`
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Suggest string      `json:"suggest"`
	Status  string      `json:"status"`
	Report  *ReportInfo `json:"report"`
}

type ReportInfo struct {
	MatchField     string `json:"match_field"`
	MatchLevel     string `json:"match_level"`
	RerankByZixuan string `json:"rerank_by_zixuan"`
}

// StockKLineDataWrapper K线数据包装器,处理动态股票代码key
type StockKLineDataWrapper struct {
	StockData map[string]*StockKLineDetail `json:"-"`
}

// UnmarshalJSON 自定义JSON解析,处理动态的股票代码key
func (s *StockKLineDataWrapper) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.StockData = make(map[string]*StockKLineDetail)
	for key, value := range raw {
		var detail StockKLineDetail
		if err := json.Unmarshal(value, &detail); err != nil {
			return fmt.Errorf("failed to unmarshal stock data for key %s: %w", key, err)
		}
		s.StockData[key] = &detail
	}
	return nil
}

// ==================== Python对齐的核心结构 ====================

// StockCurrentResp 股票当前行情响应 (与Python完全对齐)
type StockCurrentResp struct {
	Data              []string `json:"data"`                // 数据列表
	Date              string   `json:"date"`                // 日期
	Lot2Share         string   `json:"lot_2_share"`         // 每手股数
	CurrentPrice      string   `json:"current_price"`       // 当前价
	CurrentMaxPrice   string   `json:"current_max_price"`   // 最高价
	CurrentMinPrice   string   `json:"current_min_price"`   // 最低价
	CurrentStartPrice string   `json:"current_start_price"` // 开盘价
	PrevClose         string   `json:"prev_close"`          // 昨收
	PERatio           string   `json:"pe_ratio"`            // 市盈率
	Amplitude         string   `json:"amplitude"`           // 振幅
	WeekHighIn52      string   `json:"week_high_in_52"`     // 52周最高价
	WeekLowIn52       string   `json:"week_low_in_52"`      // 52周最低价
	VolumeByLot       string   `json:"volume_by_lot"`       // 成交量（手）
}

// ==================== 完整的详细数据结构 ====================

// StockKLineDetail 股票K线详细数据 (完整版)
type StockKLineDetail struct {
	Data    *StockKLineTimeData `json:"data"`     // 分时数据
	Qt      *StockQuoteData     `json:"qt"`       // 实时行情数据
	Market  *StockMarketInfo    `json:"market"`   // 市场状态信息
	MxPrice *StockMxPrice       `json:"mx_price"` // 价格相关信息
}

// StockKLineTimeData 分时数据
type StockKLineTimeData struct {
	Date string   `json:"date"` // 日期,如 "20260116"
	Data []string `json:"data"` // 分时数据数组
}

// StockQuoteData 实时行情数据 (完整版)
type StockQuoteData struct {
	RawData         map[string]json.RawMessage `json:"-"`
	Vff             []interface{}              `json:"-"` // 分时成交明细
	StockQuoteArray []string                   `json:"-"` // 主要行情数据数组
}

// UnmarshalJSON 自定义JSON解析
func (s *StockQuoteData) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.RawData = make(map[string]json.RawMessage)
	for key, value := range raw {
		s.RawData[key] = value
		if strings.HasPrefix(key, "v_ff_") {
			var vffData []interface{}
			json.Unmarshal(value, &vffData)
			s.Vff = vffData
		} else {
			var arr []string
			if err := json.Unmarshal(value, &arr); err == nil && len(arr) > 0 {
				s.StockQuoteArray = arr
			}
		}
	}
	return nil
}

// StockMarketInfo 市场状态信息
type StockMarketInfo struct {
	Markets []string `json:"market"`
}

// StockMxPrice 价格相关信息
type StockMxPrice struct {
	Price string `json:"price"`
	Mx    string `json:"mx"`
}

// ==================== 扩展的详细结构体 ====================

// StockQuoteItem 完整的实时行情详细项 (74个字段)
type StockQuoteItem struct {
	// 基础信息 [0-3]
	Index        string `json:"index"`         // 序号
	Name         string `json:"name"`          // 股票名称
	Code         string `json:"code"`          // 股票代码
	CurrentPrice string `json:"current_price"` // 当前价

	// 价格信息 [4-7]
	OpenPrice  string `json:"open_price"`  // 开盘价
	ClosePrice string `json:"close_price"` // 收盘价
	Volume     string `json:"volume"`      // 成交量(手)
	Turnover   string `json:"turnover"`    // 成交额

	// 昨日信息 [8-10]
	YesterdayClose string `json:"yesterday_close"` // 昨收
	Volume2        string `json:"volume_2"`        // 成交量2
	BuyPrice       string `json:"buy_price"`       // 买一价

	// 买卖盘信息 [11-13]
	BuyVolume  string `json:"buy_volume"`  // 买一量
	SellPrice  string `json:"sell_price"`  // 卖一价
	SellVolume string `json:"sell_volume"` // 卖一量

	// 成交量信息 [14-17]
	Volume3   string `json:"volume_3"`   // 成交量3
	HighPrice string `json:"high_price"` // 最高价
	LowPrice  string `json:"low_price"`  // 最低价
	Volume4   string `json:"volume_4"`   // 成交量4

	// 五档行情 [18-20]
	Price5  string `json:"price_5"`  // 价格5
	Volume5 string `json:"volume_5"` // 成交量5
	Volume6 string `json:"volume_6"` // 成交量6

	// 时间和涨跌 [21-25]
	Time          string `json:"time"`           // 时间
	Change        string `json:"change"`         // 涨跌额
	ChangePercent string `json:"change_percent"` // 涨跌幅
	High          string `json:"high"`           // 最高
	Low           string `json:"low"`            // 最低

	// 量价数据 [26-30]
	VolumePrice string `json:"volume_price"` // 量价数据
	Volume7     string `json:"volume_7"`     // 成交量7
	Turnover2   string `json:"turnover_2"`   // 成交额2
	Amplitude   string `json:"amplitude"`    // 振幅
	Volume8     string `json:"volume_8"`     // 成交量8

	// 更多成交量 [31-35]
	Volume9  string `json:"volume_9"`  // 成交量9
	Volume10 string `json:"volume_10"` // 成交量10
	Volume11 string `json:"volume_11"` // 成交量11
	High2    string `json:"high_2"`    // 最高2
	Low2     string `json:"low_2"`     // 最低2

	// 成交量 [36-40]
	Volume12 string `json:"volume_12"` // 成交量12
	Volume13 string `json:"volume_13"` // 成交量13
	Volume14 string `json:"volume_14"` // 成交量14
	Change2  string `json:"change_2"`  // 涨跌2
	Volume15 string `json:"volume_15"` // 成交量15

	// 更多成交量 [41-45]
	Volume16 string `json:"volume_16"` // 成交量16
	Volume17 string `json:"volume_17"` // 成交量17
	Volume18 string `json:"volume_18"` // 成交量18
	Volume19 string `json:"volume_19"` // 成交量19
	High3    string `json:"high_3"`    // 最高3

	// 更多数据 [46-50]
	Low3           string `json:"low_3"`            // 最低3
	ChangePercent2 string `json:"change_percent_2"` // 涨跌幅2
	Volume20       string `json:"volume_20"`        // 成交量20
	Volume21       string `json:"volume_21"`        // 成交量21
	Change2Value   string `json:"change_2_value"`   // 涨跌2值

	// 市盈率和成交量 [52-57]
	Volume22 string `json:"volume_22"` // 成交量22
	Volume23 string `json:"volume_23"` // 成交量23
	PERatio  string `json:"pe_ratio"`  // 市盈率
	Volume24 string `json:"volume_24"` // 成交量24
	Volume25 string `json:"volume_25"` // 成交量25
	Volume26 string `json:"volume_26"` // 成交量26

	// 更多成交量 [57-61]
	Volume27 string `json:"volume_27"` // 成交量27
	Volume28 string `json:"volume_28"` // 成交量28
	Volume29 string `json:"volume_29"` // 成交量29
	Volume30 string `json:"volume_30"` // 成交量30
	Volume31 string `json:"volume_31"` // 成交量31

	// 更多成交量 [62-66]
	Volume32 string `json:"volume_32"` // 成交量32
	Volume33 string `json:"volume_33"` // 成交量33
	Volume34 string `json:"volume_34"` // 成交量34
	Volume35 string `json:"volume_35"` // 成交量35
	Volume36 string `json:"volume_36"` // 成交量36

	// 更多数据 [67-70]
	Volume37  string `json:"volume_37"`  // 成交量37
	Volume38  string `json:"volume_38"`  // 成交量38
	Turnover3 string `json:"turnover_3"` // 成交额3
	Volume39  string `json:"volume_39"`  // 成交量39

	// 更多成交量 [70-74]
	Volume40 string `json:"volume_40"` // 成交量40
	Volume41 string `json:"volume_41"` // 成交量41
	Volume42 string `json:"volume_42"` // 成交量42
	Volume43 string `json:"volume_43"` // 成交量43

	// 更多成交量 [74-78]
	Volume44 string `json:"volume_44"` // 成交量44
	Volume45 string `json:"volume_45"` // 成交量45
	Volume46 string `json:"volume_46"` // 成交量46
	Volume47 string `json:"volume_47"` // 成交量47

	// 更多成交量 [78-82]
	Volume48 string `json:"volume_48"` // 成交量48
	Volume49 string `json:"volume_49"` // 成交量49
	Volume50 string `json:"volume_50"` // 成交量50
	Volume51 string `json:"volume_51"` // 成交量51

	// 更多成交量 [82-86]
	Volume52 string `json:"volume_52"` // 成交量52
	Volume53 string `json:"volume_53"` // 成交量53
	Volume54 string `json:"volume_54"` // 成交量54
	Volume55 string `json:"volume_55"` // 成交量55

	// 更多成交量 [86-91]
	Volume56 string `json:"volume_56"` // 成交量56
	Volume57 string `json:"volume_57"` // 成交量57
	Volume58 string `json:"volume_58"` // 成交量58
	Volume59 string `json:"volume_59"` // 成交量59
	Volume60 string `json:"volume_60"` // 成交量60

	// 市值信息 [92-94]
	TotalMarketCapitalization       string `json:"total_market_cap"`       // 总市值
	CirculatingMarketCapitalization string `json:"circulating_market_cap"` // 流通市值
	PE                              string `json:"pe"`                     // 市盈率(简化版)

	// 更多市值 [95-98]
	Volume61                         string `json:"volume_61"`                // 成交量61
	TotalMarketCapitalization2       string `json:"total_market_cap_2"`       // 总市值2
	CirculatingMarketCapitalization2 string `json:"circulating_market_cap_2"` // 流通市值2
	TurnoverRate                     string `json:"turnover_rate"`            // 换手率

	// 涨跌停和价格 [99-104]
	Volume62       string `json:"volume_62"`        // 成交量62
	LimitPriceHigh string `json:"limit_price_high"` // 涨停价
	LimitPriceLow  string `json:"limit_price_low"`  // 跌停价
	Volume63       string `json:"volume_63"`        // 成交量63
	Volume64       string `json:"volume_64"`        // 成交量64
	Open2          string `json:"open_2"`           // 开盘2

	// 更多价格信息 [105-111]
	Volume65   string `json:"volume_65"`    // 成交量65
	PrevClose2 string `json:"prev_close_2"` // 昨收2
	High4      string `json:"high_4"`       // 最高4
	Low4       string `json:"low_4"`        // 最低4
	Volume66   string `json:"volume_66"`    // 成交量66
	Volume67   string `json:"volume_67"`    // 成交量67
	Volume68   string `json:"volume_68"`    // 成交量68

	// 更多数据 [111-118]
	Volume69 string `json:"volume_69"` // 成交量69
	Volume70 string `json:"volume_70"` // 成交量70
	Currency string `json:"currency"`  // 货币
	Volume71 string `json:"volume_71"` // 成交量71
	Status   string `json:"status"`    // 状态
	Volume72 string `json:"volume_72"` // 成交量72
	Volume73 string `json:"volume_73"` // 成交量73
}

// ==================== 转换方法 ====================

// ToStockCurrentResp 转换为StockCurrentResp结构 (与Python实现对齐)
func (s *StockKLineDetail) ToStockCurrentResp() *StockCurrentResp {
	if s == nil {
		return nil
	}

	resp := &StockCurrentResp{}

	// 解析分时数据
	if s.Data != nil {
		resp.Data = s.Data.Data
		resp.Date = s.Data.Date
	}

	// 解析实时行情数据
	if s.Qt != nil && len(s.Qt.StockQuoteArray) > 0 {
		qtData := s.Qt.StockQuoteArray
		if len(qtData) > 1 {
			resp.Lot2Share = qtData[1]
		}
		if len(qtData) > 3 {
			resp.CurrentPrice = qtData[3]
		}
		if len(qtData) > 4 {
			resp.PrevClose = qtData[4]
		}
		if len(qtData) > 5 {
			resp.CurrentStartPrice = qtData[5]
		}
		if len(qtData) > 6 {
			resp.VolumeByLot = qtData[6]
		}
		if len(qtData) > 33 {
			resp.CurrentMaxPrice = qtData[33]
		}
		if len(qtData) > 34 {
			resp.CurrentMinPrice = qtData[34]
		}
		if len(qtData) > 39 {
			resp.PERatio = qtData[39]
		}
		if len(qtData) > 43 {
			resp.Amplitude = qtData[43]
		}
		if len(qtData) > 48 {
			resp.WeekHighIn52 = qtData[48]
		}
		if len(qtData) > 49 {
			resp.WeekLowIn52 = qtData[49]
		}
	}

	return resp
}

// ==================== 辅助方法 ====================

// GetStockCode 获取股票代码
func (s *StockKLineDataWrapper) GetStockCode() string {
	if s == nil || len(s.StockData) == 0 {
		return ""
	}
	for code := range s.StockData {
		return code
	}
	return ""
}

// GetKLineDetail 获取指定股票代码的K线数据
func (s *StockKLineDataWrapper) GetKLineDetail(code string) *StockKLineDetail {
	if s == nil {
		return nil
	}
	if code != "" {
		return s.StockData[code]
	}
	for _, detail := range s.StockData {
		return detail
	}
	return nil
}

// GetStockCurrentResp 获取股票当前行情数据 (与Python实现完全对齐)
func (s *StockKLineDataWrapper) GetStockCurrentResp(stockCode string) (*StockCurrentResp, error) {
	if s == nil {
		return nil, fmt.Errorf("StockKLineDataWrapper is nil")
	}
	if stockCode == "" {
		stockCode = s.GetStockCode()
		if stockCode == "" {
			return nil, fmt.Errorf("no stock code found")
		}
	}
	kLineDetail := s.StockData[stockCode]
	if kLineDetail == nil {
		return nil, fmt.Errorf("stock data not found for code: %s", stockCode)
	}
	return kLineDetail.ToStockCurrentResp(), nil
}
