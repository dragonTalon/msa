package model

// KLineBar 单根K线数据
type KLineBar struct {
	Date   string `json:"date"`
	Open   string `json:"open"`
	Close  string `json:"close"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Volume string `json:"volume"`
}

// StockKLineResp fqkline API 响应
type StockKLineResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data map[string]struct {
		QfqDay   [][]string `json:"qfqday"`
		QfqWeek  [][]string `json:"qfqweek"`
		QfqMonth [][]string `json:"qfqmonth"`
		Day      [][]string `json:"day"`
		Week     [][]string `json:"week"`
		Month    [][]string `json:"month"`
	} `json:"data"`
}

// ToKLineBars 将原始数组转换为 KLineBar 切片
func ToKLineBars(raw [][]string) []KLineBar {
	bars := make([]KLineBar, 0, len(raw))
	for _, r := range raw {
		if len(r) >= 6 {
			bars = append(bars, KLineBar{
				Date: r[0], Open: r[1], Close: r[2],
				High: r[3], Low: r[4], Volume: r[5],
			})
		}
	}
	return bars
}

// StockIndustryResp jiankuang API 响应
type StockIndustryResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Zyzb struct {
			Date   string            `json:"date"`
			Detail map[string]string `json:"detail"`
		} `json:"zyzb"`
		Gsjj struct {
			Gsmz  string          `json:"gsmz"`
			Yw    string          `json:"yw"`
			Dy    string          `json:"dy"`
			Jg    string          `json:"jg"`
			Riqi  string          `json:"riqi"`
			Plate []IndustryPlate `json:"plate"`
		} `json:"gsjj"`
	} `json:"data"`
}

// IndustryPlate 行业分类板块
type IndustryPlate struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Level string `json:"level"`
}

// BoardRankResp mktHs/rank API 响应
type BoardRankResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data []BoardItem `json:"data"`
}

// BoardItem 板块排行项
type BoardItem struct {
	BdName  string `json:"bd_name"`
	BdCode  string `json:"bd_code"`
	BdZxj   string `json:"bd_zxj"`
	BdZd    string `json:"bd_zd"`
	BdZdf   string `json:"bd_zdf"`
	BdZs    string `json:"bd_zs"`
	NzgCode string `json:"nzg_code"`
	NzgName string `json:"nzg_name"`
	NzgZxj  string `json:"nzg_zxj"`
	NzgZd   string `json:"nzg_zd"`
	NzgZdf  string `json:"nzg_zdf"`
	BdZdf5  string `json:"bd_zdf5"`
	BdZdf20 string `json:"bd_zdf20"`
}
