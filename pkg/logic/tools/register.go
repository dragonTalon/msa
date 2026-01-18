package tools

import (
	"msa/pkg/logic/tools/stock"
)

var _ MsaTool = (*stock.CompanyCode)(nil)
var _ MsaTool = (*stock.CompanyK)(nil)

func init() {
	RegisterTool(&stock.CompanyCode{})
	RegisterTool(&stock.CompanyK{})
}
