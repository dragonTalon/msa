package tools

import (
	"msa/pkg/logic/tools/finance"
	"msa/pkg/logic/tools/search"
	"msa/pkg/logic/tools/stock"
)

var _ MsaTool = (*stock.CompanyCode)(nil)
var _ MsaTool = (*stock.CompanyK)(nil)
var _ MsaTool = (*stock.CompanyInfo)(nil)
var _ MsaTool = (*search.SearchTool)(nil)
var _ MsaTool = (*search.FetcherTool)(nil)

var _ MsaTool = (*finance.CreateAccountTool)(nil)
var _ MsaTool = (*finance.GetAccountTool)(nil)
var _ MsaTool = (*finance.UpdateAccountStatusTool)(nil)
var _ MsaTool = (*finance.GetPositionsTool)(nil)
var _ MsaTool = (*finance.GetAccountSummaryTool)(nil)
var _ MsaTool = (*finance.SubmitBuyOrderTool)(nil)
var _ MsaTool = (*finance.SubmitSellOrderTool)(nil)
var _ MsaTool = (*finance.GetTransactionsTool)(nil)

func init() {
	registerStock()
	registerSearch()
	registerFinance()
}

func registerStock() {
	RegisterTool(&stock.CompanyCode{})
	RegisterTool(&stock.CompanyK{})
	RegisterTool(&stock.CompanyInfo{})
}

func registerSearch() {
	RegisterTool(&search.SearchTool{})
	RegisterTool(&search.FetcherTool{})
}

func registerFinance() {
	RegisterTool(&finance.CreateAccountTool{})
	RegisterTool(&finance.GetAccountTool{})
	RegisterTool(&finance.UpdateAccountStatusTool{})
	RegisterTool(&finance.GetPositionsTool{})
	RegisterTool(&finance.GetAccountSummaryTool{})
	RegisterTool(&finance.SubmitBuyOrderTool{})
	RegisterTool(&finance.SubmitSellOrderTool{})
	RegisterTool(&finance.GetTransactionsTool{})
}
