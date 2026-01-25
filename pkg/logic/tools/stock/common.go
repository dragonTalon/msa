package stock

import (
	"encoding/json"
	"fmt"
	"msa/pkg/model"
	mas_utils "msa/pkg/utils"

	log "github.com/sirupsen/logrus"
)

// fetchStockData 获取股票数据的公共逻辑
// 返回StockCurrentResp和error
func fetchStockData(stockCode string) (*model.StockCurrentResp, error) {
	if stockCode == "" {
		return nil, fmt.Errorf("stock code is empty")
	}

	// 发起HTTP请求
	getResp, err := mas_utils.GetRestyClient().R().Get(model.FinanceSearchCurrentKLine + stockCode)
	if err != nil {
		return nil, err
	}
	log.Infof("API Response: %v", getResp.String())

	// 解析响应
	var rawResp map[string]interface{}
	if err := json.Unmarshal(getResp.Body(), &rawResp); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	// 检查返回码
	if code, ok := rawResp["code"].(float64); ok && int(code) != 0 {
		return nil, fmt.Errorf("API返回错误, code=%d, msg=%v", int(code), rawResp["msg"])
	}

	// 检查data字段
	dataRaw, ok := rawResp["data"]
	if !ok || dataRaw == nil {
		return nil, fmt.Errorf("请求失败，没有返回数据 data")
	}

	// 将data转换为map
	dataMap, ok := dataRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data字段格式错误,期望map[string]interface{}, 实际%T", dataRaw)
	}

	// 创建StockKLineDataWrapper
	wrapper := &model.StockKLineDataWrapper{
		StockData: make(map[string]*model.StockKLineDetail),
	}

	// 解析每个股票的数据
	for code, stockData := range dataMap {
		stockDataBytes, err := json.Marshal(stockData)
		if err != nil {
			return nil, fmt.Errorf("序列化股票数据失败: %w", err)
		}

		var detail model.StockKLineDetail
		if err := json.Unmarshal(stockDataBytes, &detail); err != nil {
			return nil, fmt.Errorf("解析股票详细数据失败: %w", err)
		}

		wrapper.StockData[code] = &detail
		log.Infof("成功解析股票代码: %s", code)
	}

	// 获取股票代码并转换为StockCurrentResp
	actualStockCode := wrapper.GetStockCode()
	log.Infof("股票代码: %s", actualStockCode)

	stockCurrentResp, err := wrapper.GetStockCurrentResp(actualStockCode)
	if err != nil {
		return nil, fmt.Errorf("获取股票行情数据失败: %w", err)
	}
	return stockCurrentResp, nil
}
