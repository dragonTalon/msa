package command

import (
	"context"
	"fmt"

	"msa/pkg/model"
)

// RememberCommand 记忆命令
type RememberCommand struct{}

func init() {
	RegisterCommand(&RememberCommand{})
}

// Name 返回命令名称
func (c *RememberCommand) Name() string {
	return "remember"
}

// Description 返回命令描述
func (c *RememberCommand) Description() string {
	return "打开记忆浏览器 - 查看历史会话、知识库和搜索记忆"
}

// Run 执行命令
func (c *RememberCommand) Run(ctx context.Context, args []string) (*model.CmdResult, error) {
	return &model.CmdResult{
		Code: 0,
		Msg:  "正在打开记忆浏览器...",
		Type: "memory-browser",
		Data: nil,
	}, nil
}

// ToSelect 返回选择器配置（已弃用，使用 memory-browser 类型）
func (c *RememberCommand) ToSelect(items []*model.SelectorItem) (*model.BaseSelector, error) {
	return nil, fmt.Errorf("remember 命令使用 memory-browser 类型，不支持 ToSelect")
}
