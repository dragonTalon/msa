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
	log.Debugf("GetStockCompanyCode start, stock_name: %s", param.StockName)

	// 使用公共函数记录工具调用开始
	message.BroadcastToolStart("get stock company code", fmt.Sprintf("stock_name: %s", param.StockName))

	// 参数校验
	if param.StockName == "" {
		err := fmt.Errorf("stock name is empty")
		message.BroadcastToolEnd("get stock company code", "", err)
		return "", err
	}
	if len(param.StockName) < 2 {
		err := fmt.Errorf("stock name is too short, minimum length is 2")
		message.BroadcastToolEnd("get stock company code", "", err)
		return "", err
	}

	// 发起请求
	resp := &model.SearchResponse{}
	client := mas_utils.GetRestyClient()
	_, err := client.R().SetResult(resp).Get(model.FinanceSearchCode + param.StockName)
	if err != nil {
		log.Errorf("failed to get stock company code: %v", err)
		message.BroadcastToolEnd("get stock company code", "", err)
		return "", err
	}

	// 处理响应
	if len(resp.Stock) > 0 {
		msg := mas_utils.ToJSONString(resp.Stock[0])
		message.BroadcastToolEnd("get stock company code", fmt.Sprintf("找到股票: %s", msg), nil)
		log.Debugf("found company info: %s", msg)
		return msg, nil
	}

	err = fmt.Errorf("no stock found for name: %s", param.StockName)
	message.BroadcastToolEnd("get stock company code", "", err)
	log.Warnf("no stock found for name: %s", param.StockName)
	return "", err
}
