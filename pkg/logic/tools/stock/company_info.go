package stock

import (
	"context"
	"fmt"
	"msa/pkg/logic/message"
	"msa/pkg/model"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"
)

type CompanyInfoParam struct {
	StockCode string `json:"stock_code" jsonschema:"description=stock_code of a stock code"`
}

type CompanyInfo struct {
}

func (ck *CompanyInfo) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(ck.GetName(), ck.GetDescription(), GetStockCompanyInfo)
}

func (ck *CompanyInfo) GetName() string {
	return "get_stock_quote"
}

func (ck *CompanyInfo) GetDescription() string {
	return "获取A股及港股股票的实时行情数据，包括当前价、昨收、开盘价、成交量、最高价、最低价、市盈率、振幅、52周高低价等 | Get the real-time quote data of A-share and Hong Kong stocks, including current price, previous close, open price, volume, high/low prices, P/E ratio, amplitude, 52-week high/low, etc."
}

func (ck *CompanyInfo) GetToolGroup() model.ToolGroup {
	return model.StockToolGroup
}

func GetStockCompanyInfo(ctx context.Context, param *CompanyInfoParam) (string, error) {
	log.Infof("GetStockCompanyInfo start")

	// 使用公共函数记录工具调用开始
	message.BroadcastToolStart("get_stock_quote", fmt.Sprintf("stock_code: %s", param.StockCode))

	if param == nil {
		err := fmt.Errorf("param is nil")
		message.BroadcastToolEnd("get stock current quote", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 调用公共函数获取股票数据
	stockCurrentResp, err := FetchStockData(param.StockCode)
	if err != nil {
		message.BroadcastToolEnd("get stock current quote", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 输出详细的行情信息
	log.Infof("日期: %s", stockCurrentResp.Date)
	log.Infof("当前价: %s", stockCurrentResp.CurrentPrice)

	message.BroadcastToolEnd("get stock current quote", fmt.Sprintf("获取股票行情成功, 当前价: %s", stockCurrentResp.CurrentPrice), nil)
	return model.NewSuccessResult(stockCurrentResp, fmt.Sprintf("获取股票行情成功, 当前价: %s", stockCurrentResp.CurrentPrice)), nil
}
