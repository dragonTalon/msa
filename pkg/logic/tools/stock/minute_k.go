package stock

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"
	"msa/pkg/logic/tools/safetool"
	"msa/pkg/model"
)

// FetchStockMinuteData 获取并解析分钟K线数据
// 复用 FetchStockData() 返回的 StockCurrentResp.Data 字段
func FetchStockMinuteData(stockCode string) (*model.StockMinuteKResp, error) {
	resp, err := FetchStockData(stockCode)
	if err != nil {
		return nil, err
	}

	result := &model.StockMinuteKResp{
		StockCode: stockCode,
		Date:      resp.Date,
	}

	if len(resp.Data) == 0 {
		log.Infof("get_stock_minute_k: no minute data for %s", stockCode)
		return result, nil
	}

	bars := make([]model.StockMinuteBar, 0, len(resp.Data))
	var prevVolume float64
	var prevTurnover float64

	for _, raw := range resp.Data {
		parts := strings.Split(strings.TrimSpace(raw), " ")
		if len(parts) < 4 {
			log.Warnf("get_stock_minute_k: unexpected data format: %s", raw)
			continue
		}

		timeStr := parts[0]
		price, err1 := strconv.ParseFloat(parts[1], 64)
		cumVolume, err2 := strconv.ParseFloat(parts[2], 64)
		cumTurnover, err3 := strconv.ParseFloat(parts[3], 64)
		if err1 != nil || err2 != nil || err3 != nil {
			log.Warnf("get_stock_minute_k: parse error for %s: %v %v %v", raw, err1, err2, err3)
			continue
		}

		// 累计值差分得到每分钟实际值
		volume := cumVolume
		turnover := cumTurnover
		if len(bars) > 0 {
			volume = cumVolume - prevVolume
			turnover = cumTurnover - prevTurnover
		}

		bars = append(bars, model.StockMinuteBar{
			Time:     timeStr,
			Price:    price,
			Volume:   int64(volume),
			Turnover: turnover,
		})

		prevVolume = cumVolume
		prevTurnover = cumTurnover
	}

	result.Bars = bars
	result.Count = len(bars)
	log.Infof("get_stock_minute_k: stock=%s date=%s bars=%d", stockCode, resp.Date, result.Count)
	return result, nil
}

// MinuteKParam 分钟K线参数
type MinuteKParam struct {
	StockCode string `json:"stock_code" jsonschema:"description=股票代码（如 sh600519）, required"`
}

// MinuteK 分钟K线工具
type MinuteK struct{}

func (m *MinuteK) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(m.GetName(), m.GetDescription(), GetStockMinuteK)
}

func (m *MinuteK) GetName() string {
	return "get_stock_minute_k"
}

func (m *MinuteK) GetDescription() string {
	return "获取股票的分钟级K线数据，返回时间、价格、成交量、成交额（按分钟拆分）| Get minute-level K-line data for a stock, returns time, price, volume, turnover per minute"
}

func (m *MinuteK) GetToolGroup() model.ToolGroup {
	return model.StockToolGroup
}

// GetStockMinuteK 获取分钟K线数据
func GetStockMinuteK(ctx context.Context, param *MinuteKParam) (string, error) {
	return safetool.SafeExecute("get_stock_minute_k", param.StockCode, func() (string, error) {
		return doGetStockMinuteK(ctx, param)
	})
}

func doGetStockMinuteK(ctx context.Context, param *MinuteKParam) (string, error) {
	log.Infof("GetStockMinuteK start: %s", param.StockCode)
	if param == nil || param.StockCode == "" {
		return model.NewErrorResult("stock_code is required"), nil
	}

	data, err := FetchStockMinuteData(param.StockCode)
	if err != nil {
		return model.NewErrorResult(err.Error()), nil
	}

	if data.Count == 0 {
		return model.NewSuccessResult(data, "非交易日或无分钟数据"), nil
	}

	return model.NewSuccessResult(data, fmt.Sprintf("获取 %s 分钟K线成功, %d条", param.StockCode, data.Count)), nil
}
