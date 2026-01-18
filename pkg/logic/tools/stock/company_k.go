package stock

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"
	"msa/pkg/model"
	mas_utils "msa/pkg/utils"
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
	if param == nil {
		return "", fmt.Errorf("param is nil")
	}
	if param.StockCode == "" {
		return "", fmt.Errorf("stock code is empty")
	}
	getResp, err := mas_utils.GetRestyClient().R().Get(model.FinanceSearchCurrentKLine + param.StockCode)
	if err != nil {
		return "", err
	}
	log.Infof("getResp : %v", getResp.String())
	var rawResp map[string]interface{}
	if err := json.Unmarshal(getResp.Body(), &rawResp); err != nil {
		return "", fmt.Errorf("JSON解析失败: %w", err)
	}
	if code, ok := rawResp["code"].(float64); ok && int(code) != 0 {
		return "", fmt.Errorf("API返回错误, code=%d, msg=%v", int(code), rawResp["msg"])
	}
	// 检查data字段
	dataRaw, ok := rawResp["data"]
	if !ok || dataRaw == nil {
		return "", fmt.Errorf("请求失败，没有返回数据 data")
	}
	// 将data转换为map,因为它包含动态的股票代码key
	dataMap, ok := dataRaw.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("data字段格式错误,期望map[string]interface{}, 实际%T", dataRaw)
	}
	// 创建StockKLineDataWrapper
	wrapper := &model.StockKLineDataWrapper{
		StockData: make(map[string]*model.StockKLineDetail),
	}
	for stockCode, stockData := range dataMap {
		// 将stockData序列化为JSON
		stockDataBytes, err := json.Marshal(stockData)
		if err != nil {
			return "", fmt.Errorf("序列化股票数据失败: %w", err)
		}

		// 解析为StockKLineDetail
		var detail model.StockKLineDetail
		if err := json.Unmarshal(stockDataBytes, &detail); err != nil {
			return "", fmt.Errorf("解析股票详细数据失败: %w", err)
		}

		wrapper.StockData[stockCode] = &detail
		log.Infof("成功解析股票代码: %s\n", stockCode)
	}
	if wrapper != nil {
		stockCode := wrapper.GetStockCode()
		log.Infof("股票代码: %s\n", stockCode)

		// 获取Python对齐的StockCurrentResp
		stockCurrentResp, err := wrapper.GetStockCurrentResp(stockCode)
		if err != nil {
			return "", fmt.Errorf("获取股票行情数据失败: %w", err)
		}

		log.Infof("日期: %s\n", stockCurrentResp.Date)
		log.Infof("每手股数: %s\n", stockCurrentResp.Lot2Share)
		log.Infof("当前价: %s\n", stockCurrentResp.CurrentPrice)
		log.Infof("昨收: %s\n", stockCurrentResp.PrevClose)
		log.Infof("开盘价: %s\n", stockCurrentResp.CurrentStartPrice)
		log.Infof("成交量(手): %s\n", stockCurrentResp.VolumeByLot)
		log.Infof("最高价: %s\n", stockCurrentResp.CurrentMaxPrice)
		log.Infof("最低价: %s\n", stockCurrentResp.CurrentMinPrice)
		log.Infof("市盈率: %s\n", stockCurrentResp.PERatio)
		log.Infof("振幅: %s\n", stockCurrentResp.Amplitude)
		log.Infof("52周最高价: %s\n", stockCurrentResp.WeekHighIn52)
		log.Infof("52周最低价: %s\n", stockCurrentResp.WeekLowIn52)

		// 显示分时数据数量
		if len(stockCurrentResp.Data) > 0 {
			log.Infof("分时数据点数: %d\n", len(stockCurrentResp.Data))
			// 显示前3个时间点
			for i := 0; i < len(stockCurrentResp.Data) && i < 3; i++ {
				log.Infof("  数据点%d: %s\n", i+1, stockCurrentResp.Data[i])
			}
		}
		return mas_utils.ToJSONString(stockCurrentResp), nil
	}

	return mas_utils.ToJSONString(wrapper), nil
}
