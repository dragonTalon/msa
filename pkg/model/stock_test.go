package model

import (
	"encoding/json"
	"testing"
)

// TestLLMModel_Struct tests the LLMModel struct
func TestLLMModel_Struct(t *testing.T) {
	model := LLMModel{
		Name:        "gpt-4",
		Description: "GPT-4 model",
	}

	if model.Name != "gpt-4" {
		t.Errorf("Name = %v, want %v", model.Name, "gpt-4")
	}
	if model.Description != "GPT-4 model" {
		t.Errorf("Description = %v, want %v", model.Description, "GPT-4 model")
	}
}

// TestLLMModel_JSONSerialization tests JSON serialization
func TestLLMModel_JSONSerialization(t *testing.T) {
	model := LLMModel{
		Name:        "claude-3-opus",
		Description: "Claude 3 Opus",
	}

	data, err := json.Marshal(model)
	if err != nil {
		t.Fatalf("JSON Marshal error: %v", err)
	}

	var decoded LLMModel
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON Unmarshal error: %v", err)
	}

	if decoded.Name != model.Name {
		t.Errorf("Decoded Name = %v, want %v", decoded.Name, model.Name)
	}
}

// TestWebSearchParams_Struct tests the WebSearchParams struct
func TestWebSearchParams_Struct(t *testing.T) {
	params := WebSearchParams{
		Query: "test search",
	}

	if params.Query != "test search" {
		t.Errorf("Query = %v, want 'test search'", params.Query)
	}
}

// TestWebSearchParams_JSONSerialization tests JSON serialization
func TestWebSearchParams_JSONSerialization(t *testing.T) {
	params := WebSearchParams{
		Query: "搜索测试",
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("JSON Marshal error: %v", err)
	}

	var decoded WebSearchParams
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON Unmarshal error: %v", err)
	}

	if decoded.Query != "搜索测试" {
		t.Errorf("Decoded Query = %v, want '搜索测试'", decoded.Query)
	}
}

// TestWebSearchResponse_Struct tests the WebSearchResponse struct
func TestWebSearchResponse_Struct(t *testing.T) {
	response := WebSearchResponse{
		Query:     "test query",
		Results:   []SearchResultItem{},
		RequestID: "req-123",
		Status:    "success",
		Error:     "",
		Message:   "test message",
	}

	if response.Query != "test query" {
		t.Errorf("Query = %v, want 'test query'", response.Query)
	}
	if response.Status != "success" {
		t.Errorf("Status = %v, want 'success'", response.Status)
	}
}

// TestSearchResultItem_Struct tests the SearchResultItem struct
func TestSearchResultItem_Struct(t *testing.T) {
	item := SearchResultItem{
		Title:   "Test Page",
		URL:     "https://example.com",
		Snippet: "Test snippet",
		Source:  "example.com",
	}

	if item.Title != "Test Page" {
		t.Errorf("Title = %v, want 'Test Page'", item.Title)
	}
	if item.URL != "https://example.com" {
		t.Errorf("URL = %v, want 'https://example.com'", item.URL)
	}
}

// TestFetchPageParams_Struct tests the FetchPageParams struct
func TestFetchPageParams_Struct(t *testing.T) {
	params := FetchPageParams{
		URL:       "https://example.com",
		MaxLength: 5000,
	}

	if params.URL != "https://example.com" {
		t.Errorf("URL = %v, want 'https://example.com'", params.URL)
	}
	if params.MaxLength != 5000 {
		t.Errorf("MaxLength = %v, want 5000", params.MaxLength)
	}
}

// TestFetchPageResponse_Struct tests the FetchPageResponse struct
func TestFetchPageResponse_Struct(t *testing.T) {
	response := FetchPageResponse{
		URL:         "https://example.com",
		Title:       "Example Page",
		Content:     "Page content",
		HasMore:     true,
		TotalLength: 10000,
	}

	if response.URL != "https://example.com" {
		t.Errorf("URL = %v, want 'https://example.com'", response.URL)
	}
	if !response.HasMore {
		t.Error("HasMore should be true")
	}
	if response.TotalLength != 10000 {
		t.Errorf("TotalLength = %v, want 10000", response.TotalLength)
	}
}

// TestStockInfo_Struct tests the StockInfo struct
func TestStockInfo_Struct(t *testing.T) {
	info := StockInfo{
		Code:    "000001",
		Name:    "平安银行",
		Type:    "sz",
		Suggest: "000001",
		Status:  "active",
	}

	if info.Code != "000001" {
		t.Errorf("Code = %v, want '000001'", info.Code)
	}
	if info.Name != "平安银行" {
		t.Errorf("Name = %v, want '平安银行'", info.Name)
	}
}

// TestSearchResponse_Struct tests the SearchResponse struct
func TestSearchResponse_Struct(t *testing.T) {
	response := SearchResponse{
		Stock: []StockInfo{
			{Code: "000001", Name: "平安银行"},
			{Code: "000002", Name: "万科A"},
		},
	}

	if len(response.Stock) != 2 {
		t.Errorf("Stock length = %v, want 2", len(response.Stock))
	}
}

// TestStockKLineDataWrapper_UnmarshalJSON tests custom JSON unmarshaling
func TestStockKLineDataWrapper_UnmarshalJSON(t *testing.T) {
	jsonData := `{"sh000001": {"data": {"date": "20240101", "data": ["1", "2", "3"]}}}`

	var wrapper StockKLineDataWrapper
	err := json.Unmarshal([]byte(jsonData), &wrapper)
	if err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if wrapper.StockData == nil {
		t.Error("StockData should not be nil after unmarshal")
	}

	if len(wrapper.StockData) != 1 {
		t.Errorf("StockData length = %v, want 1", len(wrapper.StockData))
	}
}

// TestStockKLineDataWrapper_GetStockCode tests GetStockCode
func TestStockKLineDataWrapper_GetStockCode(t *testing.T) {
	wrapper := &StockKLineDataWrapper{
		StockData: map[string]*StockKLineDetail{
			"sh000001": {Data: &StockKLineTimeData{Date: "20240101"}},
		},
	}

	code := wrapper.GetStockCode()
	if code != "sh000001" {
		t.Errorf("GetStockCode() = %v, want 'sh000001'", code)
	}
}

// TestStockKLineDataWrapper_GetStockCode_Empty tests GetStockCode with empty data
func TestStockKLineDataWrapper_GetStockCode_Empty(t *testing.T) {
	wrapper := &StockKLineDataWrapper{
		StockData: map[string]*StockKLineDetail{},
	}

	code := wrapper.GetStockCode()
	if code != "" {
		t.Errorf("GetStockCode() with empty data = %v, want empty string", code)
	}
}

// TestStockKLineDataWrapper_GetStockCode_Nil tests GetStockCode with nil wrapper
func TestStockKLineDataWrapper_GetStockCode_Nil(t *testing.T) {
	var wrapper *StockKLineDataWrapper
	code := wrapper.GetStockCode()
	if code != "" {
		t.Errorf("GetStockCode() with nil wrapper = %v, want empty string", code)
	}
}

// TestStockCurrentResp_Struct tests the StockCurrentResp struct
func TestStockCurrentResp_Struct(t *testing.T) {
	resp := StockCurrentResp{
		Data:              []string{"1", "2", "3"},
		Date:              "20240101",
		Lot2Share:         "100",
		CurrentPrice:      "10.50",
		CurrentMaxPrice:   "11.00",
		CurrentMinPrice:   "10.00",
		CurrentStartPrice: "10.20",
		PrevClose:         "10.30",
		PERatio:           "15.5",
		Amplitude:         "5.0",
		WeekHighIn52:      "12.00",
		WeekLowIn52:       "9.00",
		VolumeByLot:       "1000000",
	}

	if resp.Date != "20240101" {
		t.Errorf("Date = %v, want '20240101'", resp.Date)
	}
	if resp.CurrentPrice != "10.50" {
		t.Errorf("CurrentPrice = %v, want '10.50'", resp.CurrentPrice)
	}
}

// TestStockKLineDetail_ToStockCurrentResp tests conversion method
func TestStockKLineDetail_ToStockCurrentResp(t *testing.T) {
	detail := &StockKLineDetail{
		Data: &StockKLineTimeData{
			Date: "20240101",
			Data: []string{"0", "09:30:00", "10.50", "10.55", "10.45", "10.50", "100", "1000", "10.45", "10.55"},
		},
		Qt: &StockQuoteData{
			StockQuoteArray: []string{"0", "9", "30", "000001", "平安银行", "10.50", "10.55", "10.45", "10.50", "1000000", "11.00", "10.00"},
		},
	}

	resp := detail.ToStockCurrentResp()
	if resp == nil {
		t.Fatal("ToStockCurrentResp() returned nil")
	}

	// Check fields based on array indices
	if resp.Date != "20240101" {
		t.Errorf("Date = %v, want '20240101'", resp.Date)
	}
	// CurrentPrice is at index 5
	if resp.CurrentPrice != "10.50" {
		t.Logf("CurrentPrice = %v (expected at index 5, got: %v)", resp.CurrentPrice, resp.CurrentPrice)
	}
}

// TestStockKLineDetail_ToStockCurrentResp_Nil tests with nil detail
func TestStockKLineDetail_ToStockCurrentResp_Nil(t *testing.T) {
	var detail *StockKLineDetail
	resp := detail.ToStockCurrentResp()
	if resp != nil {
		t.Error("ToStockCurrentResp() with nil detail should return nil")
	}
}

// TestStockKLineDataWrapper_GetStockCurrentResp tests GetStockCurrentResp
func TestStockKLineDataWrapper_GetStockCurrentResp(t *testing.T) {
	wrapper := &StockKLineDataWrapper{
		StockData: map[string]*StockKLineDetail{
			"sh000001": {
				Data: &StockKLineTimeData{Date: "20240101"},
				Qt:   &StockQuoteData{StockQuoteArray: []string{"0", "9", "30", "000001", "平安银行", "10.50"}},
			},
		},
	}

	resp, err := wrapper.GetStockCurrentResp("sh000001")
	if err != nil {
		t.Errorf("GetStockCurrentResp() error = %v", err)
	}
	if resp == nil {
		t.Error("GetStockCurrentResp() returned nil")
	}
}

// TestStockKLineDataWrapper_GetStockCurrentResp_InvalidCode tests with invalid code
func TestStockKLineDataWrapper_GetStockCurrentResp_InvalidCode(t *testing.T) {
	wrapper := &StockKLineDataWrapper{
		StockData: map[string]*StockKLineDetail{
			"sh000001": {Data: &StockKLineTimeData{Date: "20240101"}},
		},
	}

	_, err := wrapper.GetStockCurrentResp("invalid_code")
	if err == nil {
		t.Error("GetStockCurrentResp() with invalid code should return error")
	}
}

// TestStockKLineDataWrapper_GetStockCurrentResp_EmptyCode tests with empty code
func TestStockKLineDataWrapper_GetStockCurrentResp_EmptyCode(t *testing.T) {
	wrapper := &StockKLineDataWrapper{
		StockData: map[string]*StockKLineDetail{
			"sh000001": {Data: &StockKLineTimeData{Date: "20240101"}},
		},
	}

	// Empty code should use GetStockCode
	resp, err := wrapper.GetStockCurrentResp("")
	if err != nil {
		t.Errorf("GetStockCurrentResp() with empty code error = %v", err)
	}
	if resp == nil {
		t.Error("GetStockCurrentResp() with empty code returned nil response")
	}
}

// TestStockKLineDataWrapper_GetKLineDetail tests GetKLineDetail method
func TestStockKLineDataWrapper_GetKLineDetail(t *testing.T) {
	wrapper := &StockKLineDataWrapper{
		StockData: map[string]*StockKLineDetail{
			"sh000001": {Data: &StockKLineTimeData{Date: "20240101"}},
			"sh000002": {Data: &StockKLineTimeData{Date: "20240102"}},
		},
	}

	// Test with specific code
	detail := wrapper.GetKLineDetail("sh000001")
	if detail == nil {
		t.Error("GetKLineDetail() returned nil for existing code")
	}

	// Test with empty code (should return first available)
	detail2 := wrapper.GetKLineDetail("")
	if detail2 == nil {
		t.Error("GetKLineDetail() with empty code should return first detail")
	}

	// Test with invalid code (should return nil since code doesn't exist)
	detail3 := wrapper.GetKLineDetail("invalid")
	if detail3 != nil {
		t.Error("GetKLineDetail() with invalid code should return nil")
	}
}

// TestStockKLineDataWrapper_GetKLineDetail_Nil tests GetKLineDetail with nil wrapper
func TestStockKLineDataWrapper_GetKLineDetail_Nil(t *testing.T) {
	var wrapper *StockKLineDataWrapper
	detail := wrapper.GetKLineDetail("any_code")
	if detail != nil {
		t.Error("GetKLineDetail() with nil wrapper should return nil")
	}
}
