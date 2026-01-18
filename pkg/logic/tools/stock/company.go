package stock

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"
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
	return "get stock company code of a stock name"
}

func (c *CompanyCode) GetDescription() string {
	return "获取A股及港股上市公司对应的股票代码 | Get the stock code for A-share and Hong Kong listed companies"
}

func (c *CompanyCode) GetToolGroup() model.ToolGroup {
	return model.StockToolGroup
}

func GetStockCompanyCode(ctx context.Context, param *CompanyParam) (string, error) {
	log.Infof("GetStockCompanyK start")
	if param.StockName == "" {
		return "", fmt.Errorf("stock name is empty")
	}
	stockName := param.StockName
	if stockName == "" && len(stockName) < 2 {
		return "", fmt.Errorf("stock name is too short")
	}
	resp := &model.SearchResponse{}
	client := mas_utils.GetRestyClient()
	getResp, err := client.R().SetResult(resp).Get(model.FinanceSearchCode + stockName)
	if err != nil {
		return "", err
	}
	if len(resp.Stock) > 0 {
		log.Infof("get company info  : %s", mas_utils.ToJSONString(resp.Stock[0]))
		return mas_utils.ToJSONString(resp.Stock[0]), nil
	}
	log.Errorf("getResp fail : %v", getResp)
	return "", nil

}
