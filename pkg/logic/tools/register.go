package tools

import (
	"msa/pkg/logic/tools/finance"
	"msa/pkg/logic/tools/search"
	skilltools "msa/pkg/logic/tools/skill"
	"msa/pkg/logic/tools/stock"
	"msa/pkg/logic/tools/todo"
)

var _ MsaTool = (*skilltools.SkillContentTool)(nil)
var _ MsaTool = (*skilltools.SkillReferenceTool)(nil)
var _ MsaTool = (*skilltools.SkillAssetTool)(nil)

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

var _ MsaTool = (*todo.CheckTodoTool)(nil)
var _ MsaTool = (*todo.CreateTodoTool)(nil)
var _ MsaTool = (*todo.UpdateTodoTool)(nil)
var _ MsaTool = (*todo.VerifyTodoTool)(nil)
var _ MsaTool = (*todo.FillSummaryTool)(nil)

func init() {
	registerStock()
	registerSearch()
	registerFinance()
	registerSkill()
	registerTodo()
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

func registerSkill() {
	RegisterTool(&skilltools.SkillContentTool{})
	RegisterTool(&skilltools.SkillReferenceTool{})
	RegisterTool(&skilltools.SkillAssetTool{})
}

func registerTodo() {
	RegisterTool(&todo.CheckTodoTool{})
	RegisterTool(&todo.CreateTodoTool{})
	RegisterTool(&todo.UpdateTodoTool{})
	RegisterTool(&todo.VerifyTodoTool{})
	RegisterTool(&todo.FillSummaryTool{})
}
