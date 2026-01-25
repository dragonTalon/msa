package stock

import (
	"context"
	"fmt"
	"msa/pkg/logic/message"
	"msa/pkg/model"
	mas_utils "msa/pkg/utils"

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
	return "get stock current quote"
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
	message.BroadcastToolStart("get stock current quote", fmt.Sprintf("stock_code: %s", param.StockCode))

	if param == nil {
		err := fmt.Errorf("param is nil")
		message.BroadcastToolEnd("get stock current quote", "", err)
		return "", err
	}

	// 调用公共函数获取股票数据
	stockCurrentResp, err := fetchStockData(param.StockCode)
	if err != nil {
		message.BroadcastToolEnd("get stock current quote", "", err)
		return "", err
	}

	// 输出详细的行情信息
	log.Infof("日期: %s", stockCurrentResp.Date)
	log.Infof("每手股数: %s", stockCurrentResp.Lot2Share)
	log.Infof("当前价: %s", stockCurrentResp.CurrentPrice)
	log.Infof("昨收: %s", stockCurrentResp.PrevClose)
	log.Infof("开盘价: %s", stockCurrentResp.CurrentStartPrice)
	log.Infof("成交量(手): %s", stockCurrentResp.VolumeByLot)
	log.Infof("最高价: %s", stockCurrentResp.CurrentMaxPrice)
	log.Infof("最低价: %s", stockCurrentResp.CurrentMinPrice)
	log.Infof("市盈率: %s", stockCurrentResp.PERatio)
	log.Infof("振幅: %s", stockCurrentResp.Amplitude)
	log.Infof("52周最高价: %s", stockCurrentResp.WeekHighIn52)
	log.Infof("52周最低价: %s", stockCurrentResp.WeekLowIn52)

	result := mas_utils.ToJSONString(stockCurrentResp)
	message.BroadcastToolEnd("get stock current quote", fmt.Sprintf("获取股票行情成功, 当前价: %s", stockCurrentResp.CurrentPrice), nil)
	return result, nil
}
