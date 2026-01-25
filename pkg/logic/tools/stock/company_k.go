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

type CompanyKParam struct {
	StockCode string `json:"stock_code" jsonschema:"description=stock_code of a stock code"`
}

type CompanyK struct {
}

func (ck *CompanyK) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(ck.GetName(), ck.GetDescription(), GetStockCompanyK)
}

func (ck *CompanyK) GetName() string {
	return "get stock company k"
}

func (ck *CompanyK) GetDescription() string {
	return "获取A股及港股上市公司对应的K线	|Get the current k line of a stock"
}

func (ck *CompanyK) GetToolGroup() model.ToolGroup {
	return model.StockToolGroup
}

func GetStockCompanyK(ctx context.Context, param *CompanyKParam) (string, error) {
	log.Infof("GetStockCompanyK start")

	// 使用公共函数记录工具调用开始
	message.BroadcastToolStart("get stock company k", fmt.Sprintf("stock_code: %s", param.StockCode))

	if param == nil {
		err := fmt.Errorf("param is nil")
		message.BroadcastToolEnd("get stock company k", "", err)
		return "", err
	}

	// 调用公共函数获取股票数据
	stockCurrentResp, err := fetchStockData(param.StockCode)
	if err != nil {
		message.BroadcastToolEnd("get stock company k", "", err)
		return "", err
	}

	// 输出K线相关信息
	log.Infof("日期: %s", stockCurrentResp.Date)
	// 显示分时数据数量
	if len(stockCurrentResp.Data) > 0 {
		log.Infof("分时数据点数: %d", len(stockCurrentResp.Data))
		// 显示前3个时间点
		for i := 0; i < len(stockCurrentResp.Data) && i < 3; i++ {
			log.Infof("  数据点%d: %s", i+1, stockCurrentResp.Data[i])
		}
	}

	result := mas_utils.ToJSONString(stockCurrentResp)
	message.BroadcastToolEnd("get stock company k", fmt.Sprintf("获取K线数据成功, 分时数据点数: %d", len(stockCurrentResp.Data)), nil)
	return result, nil
}
