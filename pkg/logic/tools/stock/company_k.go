package stock

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"
	"msa/pkg/logic/message"
	"msa/pkg/logic/tools/safetool"
	"msa/pkg/model"
)

type CompanyKParam struct {
	StockCode string `json:"stock_code" jsonschema:"description=stock_code of a stock code"`
}

type CompanyK struct {
}

func (ck *CompanyK) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(ck.GetName(), ck.GetDescription(), GetStockCompanyK)
}

func (ck *CompanyK) GetName() string {
	return "get_stock_company_k"
}

func (ck *CompanyK) GetDescription() string {
	return "获取A股及港股上市公司对应的K线	|Get the current k line of a stock"
}

func (ck *CompanyK) GetToolGroup() model.ToolGroup {
	return model.StockToolGroup
}

func GetStockCompanyK(ctx context.Context, param *CompanyKParam) (string, error) {
	return safetool.SafeExecute("get_stock_company_k", fmt.Sprintf("stock_code: %s", param.StockCode), func() (string, error) {
		return doGetStockCompanyK(ctx, param)
	})
}

func doGetStockCompanyK(ctx context.Context, param *CompanyKParam) (string, error) {
	log.Infof("GetStockCompanyK start")

	// 使用公共函数记录工具调用开始
	message.BroadcastToolStart("get stock company k", fmt.Sprintf("stock_code: %s", param.StockCode))

	if param == nil {
		err := fmt.Errorf("param is nil")
		message.BroadcastToolEnd("get stock company k", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 调用公共函数获取股票数据
	stockCurrentResp, err := FetchStockData(param.StockCode)
	if err != nil {
		message.BroadcastToolEnd("get stock company k", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 输出K线相关信息
	log.Infof("日期: %s", stockCurrentResp.Date)
	log.Infof("分时数据点数: %d", len(stockCurrentResp.Data))

	message.BroadcastToolEnd("get stock company k", fmt.Sprintf("获取K线数据成功, 分时数据点数: %d", len(stockCurrentResp.Data)), nil)
	return model.NewSuccessResult(stockCurrentResp, fmt.Sprintf("获取K线数据成功, 分时数据点数: %d", len(stockCurrentResp.Data))), nil
}
