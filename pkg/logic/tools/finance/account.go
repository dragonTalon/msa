package finance

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"msa/pkg/db"
	msadb "msa/pkg/db"
	"msa/pkg/logic/tools/safetool"
	"msa/pkg/model"
)

// CreateAccountParam 创建账户参数
type CreateAccountParam struct {
	InitialAmount float64 `json:"initial_amount" jsonschema:"description=初始金额，单位：元"`
}

// CreateAccountTool 创建账户工具
type CreateAccountTool struct{}

func (t *CreateAccountTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), CreateAccount)
}

func (t *CreateAccountTool) GetName() string {
	return "create_account"
}

func (t *CreateAccountTool) GetDescription() string {
	return "创建新的交易账户（全局只能有一个活跃账户）| Create a new trading account (only one active account globally)"
}

func (t *CreateAccountTool) GetToolGroup() model.ToolGroup {
	return model.FinanceToolGroup
}

// AccountData 账户数据
type AccountData struct {
	ID            int64  `json:"id"`
	InitialAmount string `json:"initial_amount"`
	AvailableAmt  string `json:"available_amt"`
	LockedAmt     string `json:"locked_amt,omitempty"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at,omitempty"`
}

// CreateAccount 创建账户
func CreateAccount(ctx context.Context, param *CreateAccountParam) (string, error) {
	return safetool.SafeExecute("create_account", fmt.Sprintf("初始金额: %.2f 元", param.InitialAmount), func() (string, error) {
		return doCreateAccount(ctx, param)
	})
}

func doCreateAccount(ctx context.Context, param *CreateAccountParam) (string, error) {
	database := msadb.GetDB()
	if database == nil {
		err := fmt.Errorf("数据库未初始化")
		return model.NewErrorResult(err.Error()), nil
	}

	// 检查是否已存在活跃账户
	var count int64
	database.Model(&model.Account{}).
		Where("status = ?", model.AccountStatusActive).
		Count(&count)
	if count > 0 {
		err := fmt.Errorf("已存在活跃账户，不支持多账户")
		return model.NewErrorResult(err.Error()), nil
	}

	// 创建账户
	accountID, err := db.CreateAccount(database, "default", model.YuanToHao(param.InitialAmount))
	if err != nil {
		return model.NewErrorResult(err.Error()), nil
	}

	// 查询创建的账户
	account, err := db.GetAccountByID(database, accountID)
	if err != nil {
		return model.NewErrorResult(err.Error()), nil
	}

	data := &AccountData{
		ID:            int64(account.ID),
		InitialAmount: formatHaoToYuan(account.InitialAmount),
		AvailableAmt:  formatHaoToYuan(account.AvailableAmt),
		Status:        string(account.Status),
	}

	return model.NewSuccessResult(data, "账户创建成功"), nil
}

// GetAccountParam 查询账户参数
type GetAccountParam struct{}

// GetAccountTool 查询账户工具
type GetAccountTool struct{}

func (t *GetAccountTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), GetAccount,
		utils.WithUnmarshalArguments(unmarshalEmptyParam[GetAccountParam]))
}

func (t *GetAccountTool) GetName() string {
	return "get_account"
}

func (t *GetAccountTool) GetDescription() string {
	return "查询当前活跃账户信息 | Get current active account information"
}

func (t *GetAccountTool) GetToolGroup() model.ToolGroup {
	return model.FinanceToolGroup
}

// GetAccount 查询账户
func GetAccount(ctx context.Context, param *GetAccountParam) (string, error) {
	return safetool.SafeExecute("get_account", "", func() (string, error) {
		return doGetAccount(ctx, param)
	})
}

func doGetAccount(ctx context.Context, param *GetAccountParam) (string, error) {
	database := msadb.GetDB()
	if database == nil {
		err := fmt.Errorf("数据库未初始化")
		return model.NewErrorResult(err.Error()), nil
	}

	account, err := getActiveAccount(database)
	if err != nil {
		return model.NewErrorResult(err.Error()), nil
	}

	data := &AccountData{
		ID:            int64(account.ID),
		InitialAmount: formatHaoToYuan(account.InitialAmount),
		AvailableAmt:  formatHaoToYuan(account.AvailableAmt),
		LockedAmt:     formatHaoToYuan(account.LockedAmt),
		Status:        string(account.Status),
		CreatedAt:     account.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return model.NewSuccessResult(data, "查询账户成功"), nil
}

// UpdateAccountStatusParam 修改账户状态参数
type UpdateAccountStatusParam struct {
	Action string `json:"action" jsonschema:"description=操作类型: freeze(冻结)/unfreeze(解冻)/close(关闭)"`
}

// UpdateAccountStatusTool 修改账户状态工具
type UpdateAccountStatusTool struct{}

func (t *UpdateAccountStatusTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), UpdateAccountStatus)
}

func (t *UpdateAccountStatusTool) GetName() string {
	return "update_account_status"
}

func (t *UpdateAccountStatusTool) GetDescription() string {
	return "修改账户状态（冻结/解冻/关闭）| Update account status (freeze/unfreeze/close)"
}

func (t *UpdateAccountStatusTool) GetToolGroup() model.ToolGroup {
	return model.FinanceToolGroup
}

// UpdateAccountStatus 修改账户状态
func UpdateAccountStatus(ctx context.Context, param *UpdateAccountStatusParam) (string, error) {
	return safetool.SafeExecute("update_account_status", fmt.Sprintf("操作: %s", param.Action), func() (string, error) {
		return doUpdateAccountStatus(ctx, param)
	})
}

func doUpdateAccountStatus(ctx context.Context, param *UpdateAccountStatusParam) (string, error) {
	database := msadb.GetDB()
	if database == nil {
		err := fmt.Errorf("数据库未初始化")
		return model.NewErrorResult(err.Error()), nil
	}

	account, err := getActiveAccount(database)
	if err != nil {
		return model.NewErrorResult(err.Error()), nil
	}

	var result string
	switch param.Action {
	case "freeze":
		err = db.UpdateAccountStatus(database, account.ID, model.AccountStatusFrozen)
		if err != nil {
			return model.NewErrorResult(err.Error()), nil
		}
		result = "账户已冻结"
	case "unfreeze":
		err = db.UpdateAccountStatus(database, account.ID, model.AccountStatusActive)
		if err != nil {
			return model.NewErrorResult(err.Error()), nil
		}
		result = "账户已解冻"
	case "close":
		// 验证余额为 0
		if account.AvailableAmt != 0 || account.LockedAmt != 0 {
			err := fmt.Errorf("无法关闭账户：余额不为零（可用: %s 元，锁定: %s 元）",
				formatHaoToYuan(account.AvailableAmt),
				formatHaoToYuan(account.LockedAmt))
			return model.NewErrorResult(err.Error()), nil
		}
		err = db.UpdateAccountStatus(database, account.ID, model.AccountStatusClosed)
		if err != nil {
			return model.NewErrorResult(err.Error()), nil
		}
		result = "账户已关闭"
	default:
		err := fmt.Errorf("无效的操作类型: %s，支持的类型: freeze/unfreeze/close", param.Action)
		return model.NewErrorResult(err.Error()), nil
	}

	return model.NewSuccessResult(nil, result), nil
}
