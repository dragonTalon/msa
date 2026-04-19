package stock

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"
	"msa/pkg/logic/tools/safetool"
	"msa/pkg/model"
	mas_utils "msa/pkg/utils"
)

type CompanyParam struct {
	StockName string `json:"stock_name" jsonschema:"description=stock_name of a stock name"`
}

type CompanyCode struct {
}

func (c *CompanyCode) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(c.GetName(), c.GetDescription(), GetStockCompanyCode)
}

func (c *CompanyCode) GetName() string {
	return "get_stock_company_code"
}

func (c *CompanyCode) GetDescription() string {
	return "获取A股及港股上市公司对应的股票代码 | Get the stock code for A-share and Hong Kong listed companies"
}

func (c *CompanyCode) GetToolGroup() model.ToolGroup {
	return model.StockToolGroup
}

func GetStockCompanyCode(ctx context.Context, param *CompanyParam) (string, error) {
	return safetool.SafeExecute("get_stock_company_code", fmt.Sprintf("stock_name: %s", param.StockName), func() (string, error) {
		return doGetStockCompanyCode(ctx, param)
	})
}

func doGetStockCompanyCode(ctx context.Context, param *CompanyParam) (string, error) {
	log.Debugf("GetStockCompanyCode start, stock_name: %s", param.StockName)

	// 参数校验
	if param.StockName == "" {
		err := fmt.Errorf("stock name is empty")
		return model.NewErrorResult(err.Error()), nil
	}
	if len(param.StockName) < 2 {
		err := fmt.Errorf("stock name is too short, minimum length is 2")
		return model.NewErrorResult(err.Error()), nil
	}

	// 发起请求
	resp := &model.SearchResponse{}
	client := mas_utils.GetRestyClient()
	log.Infof("get_stock_company_code url: %s", model.FinanceSearchCode+param.StockName)
	_, err := client.R().SetResult(resp).Get(model.FinanceSearchCode + param.StockName)
	if err != nil {
		log.Errorf("failed to get stock company code: %v", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 处理响应
	if len(resp.Stock) > 0 {
		log.Debugf("found company info: %v", resp.Stock)
		return model.NewSuccessResult(resp.Stock, fmt.Sprintf("找到股票: %d 只", len(resp.Stock))), nil
	}
	if len(resp.Fund) > 0 {
		log.Debugf("found fund info: %v", resp.Fund)
		return model.NewSuccessResult(resp.Fund, fmt.Sprintf("找到基金: %d 只", len(resp.Fund))), nil
	}

	err = fmt.Errorf("no stock found for name: %s", param.StockName)
	log.Warnf("no stock found for name: %s", param.StockName)
	return model.NewErrorResult(err.Error()), nil
}
