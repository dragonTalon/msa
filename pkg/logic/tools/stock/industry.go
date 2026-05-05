package stock

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"
	"msa/pkg/logic/tools/safetool"
	"msa/pkg/model"
	mas_utils "msa/pkg/utils"
)

type IndustryParam struct {
	StockCode string `json:"stock_code" jsonschema:"description=stock code (e.g. sh600519), required"`
}

type Industry struct{}

func (ind *Industry) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(ind.GetName(), ind.GetDescription(), GetStockIndustry)
}

func (ind *Industry) GetName() string { return "get_stock_industry" }

func (ind *Industry) GetDescription() string {
	return "获取个股的行业分类（申万一级和二级行业）、主要财务指标和公司基本信息 | Get industry classification (Shenwan level 1 & 2), key financials, and company info"
}

func (ind *Industry) GetToolGroup() model.ToolGroup { return model.StockToolGroup }

func GetStockIndustry(ctx context.Context, param *IndustryParam) (string, error) {
	return safetool.SafeExecute("get_stock_industry", fmt.Sprintf("stock_code: %s", param.StockCode), func() (string, error) {
		return doGetStockIndustry(ctx, param)
	})
}

func doGetStockIndustry(ctx context.Context, param *IndustryParam) (string, error) {
	if param == nil || param.StockCode == "" {
		return model.NewErrorResult("stock_code is required"), nil
	}

	apiURL := model.FinanceStockIndustry + param.StockCode
	resp, err := mas_utils.GetRestyClient().R().Get(apiURL)
	if err != nil {
		return model.NewErrorResult(fmt.Sprintf("HTTP request failed: %v", err)), nil
	}

	var indResp model.StockIndustryResp
	if err := json.Unmarshal(resp.Body(), &indResp); err != nil {
		return model.NewErrorResult(fmt.Sprintf("JSON解析失败: %v", err)), nil
	}
	if indResp.Code != 0 {
		return model.NewErrorResult(fmt.Sprintf("API error: code=%d, msg=%s", indResp.Code, indResp.Msg)), nil
	}

	result := map[string]interface{}{
		"company_name": indResp.Data.Gsjj.Gsmz,
		"business":     indResp.Data.Gsjj.Yw,
		"region":       indResp.Data.Gsjj.Dy,
		"ipo_date":     indResp.Data.Gsjj.Riqi,
		"industry":     indResp.Data.Gsjj.Plate,
	}
	if indResp.Data.Zyzb.Detail != nil {
		result["financial"] = indResp.Data.Zyzb.Detail
	}

	log.Infof("get_stock_industry: %s industries=%d", param.StockCode, len(indResp.Data.Gsjj.Plate))
	return model.NewSuccessResult(result, fmt.Sprintf("获取%s行业分类成功", param.StockCode)), nil
}
