package stock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"
	"msa/pkg/logic/tools/safetool"
	"msa/pkg/model"
	mas_utils "msa/pkg/utils"
)

type HistoryKParam struct {
	StockCode string `json:"stock_code" jsonschema:"description=stock code (e.g. sh600519), required"`
	Period    string `json:"period" jsonschema:"description=K-line period: day/week/month, default: day"`
	Count     int    `json:"count" jsonschema:"description=number of bars (max 640 for day), default: 30"`
	Adjust    string `json:"adjust" jsonschema:"description=adjust type: qfq (前复权)/hfq (后复权)/empty (不复权), default: qfq"`
}

type HistoryK struct{}

func (h *HistoryK) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(h.GetName(), h.GetDescription(), GetStockHistoryK)
}

func (h *HistoryK) GetName() string { return "get_stock_history_k" }

func (h *HistoryK) GetDescription() string {
	return "获取A股及港股历史K线数据，支持日/周/月K线，返回日期、开盘、收盘、最高、最低、成交量 | Get historical K-line data for A-share and HK stocks, returns date, open, close, high, low, volume"
}

func (h *HistoryK) GetToolGroup() model.ToolGroup { return model.StockToolGroup }

func GetStockHistoryK(ctx context.Context, param *HistoryKParam) (string, error) {
	return safetool.SafeExecute("get_stock_history_k", fmt.Sprintf("stock_code: %s", param.StockCode), func() (string, error) {
		return doGetStockHistoryK(ctx, param)
	})
}

func doGetStockHistoryK(ctx context.Context, param *HistoryKParam) (string, error) {
	if param == nil || param.StockCode == "" {
		return model.NewErrorResult("stock_code is required"), nil
	}
	if param.Period == "" {
		param.Period = "day"
	}
	if param.Count <= 0 || param.Count > 640 {
		param.Count = 30
	}
	if param.Adjust == "" {
		param.Adjust = "qfq"
	}

	paramStr := fmt.Sprintf("%s,%s,,,%d,%s", param.StockCode, param.Period, param.Count, param.Adjust)
	apiURL := model.FinanceKLineAPI + url.QueryEscape(paramStr)

	resp, err := mas_utils.GetRestyClient().R().Get(apiURL)
	if err != nil {
		return model.NewErrorResult(fmt.Sprintf("HTTP request failed: %v", err)), nil
	}

	var klineResp model.StockKLineResp
	if err := json.Unmarshal(resp.Body(), &klineResp); err != nil {
		return model.NewErrorResult(fmt.Sprintf("JSON解析失败: %v", err)), nil
	}
	if klineResp.Code != 0 {
		return model.NewErrorResult(fmt.Sprintf("API error: code=%d, msg=%s", klineResp.Code, klineResp.Msg)), nil
	}

	for _, stockData := range klineResp.Data {
		var rawData [][]string
		switch param.Period {
		case "day":
			if len(stockData.QfqDay) > 0 {
				rawData = stockData.QfqDay
			} else {
				rawData = stockData.Day
			}
		case "week":
			if len(stockData.QfqWeek) > 0 {
				rawData = stockData.QfqWeek
			} else {
				rawData = stockData.Week
			}
		case "month":
			if len(stockData.QfqMonth) > 0 {
				rawData = stockData.QfqMonth
			} else {
				rawData = stockData.Month
			}
		}
		bars := model.ToKLineBars(rawData)
		if len(bars) > 0 {
			log.Infof("get_stock_history_k: period=%s count=%d first=%s last=%s",
				param.Period, len(bars), bars[0].Date, bars[len(bars)-1].Date)
			return model.NewSuccessResult(bars, fmt.Sprintf("获取%s %s K线数据成功, %d条", param.StockCode, param.Period, len(bars))), nil
		}
	}
	return model.NewErrorResult("no K-line data found"), nil
}
