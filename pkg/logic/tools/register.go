package tools

import (
	"msa/pkg/logic/tools/stock"
)

var _ MsaTool = (*stock.CompanyCode)(nil)
var _ MsaTool = (*stock.CompanyK)(nil)
var _ MsaTool = (*stock.CompanyInfo)(nil)

func init() {
	registerStock()
}

func registerStock() {
	RegisterTool(&stock.CompanyCode{})
	RegisterTool(&stock.CompanyK{})
	RegisterTool(&stock.CompanyInfo{})
}
