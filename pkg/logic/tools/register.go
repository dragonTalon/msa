package tools

import (
	"msa/pkg/logic/tools/search"
	"msa/pkg/logic/tools/stock"
)

var _ MsaTool = (*stock.CompanyCode)(nil)
var _ MsaTool = (*stock.CompanyK)(nil)
var _ MsaTool = (*stock.CompanyInfo)(nil)
var _ MsaTool = (*search.SearchTool)(nil)
var _ MsaTool = (*search.FetcherTool)(nil)

func init() {
	registerStock()
	registerSearch()
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
