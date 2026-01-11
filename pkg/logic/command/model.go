package command

import (
	"context"
	"fmt"
	"msa/pkg/config"
	"msa/pkg/logic/provider"
	"msa/pkg/model"
	"sort"

	log "github.com/sirupsen/logrus"
)

func init() {
	RegisterCommand(&ListModel{})
	// SetModel 命令已被交互式选择器替代，使用 /models 或 /model 命令
	// RegisterCommand(&SetModel{})
}

type ListModel struct {
}

func (l *ListModel) Name() string {
	return "models"
}

func (l *ListModel) Description() string {
	return "List all models"
}

func (l *ListModel) Run(ctx context.Context, args []string) (*model.CmdResult, error) {
	p := provider.GetProvider()
	if p == nil {
		return nil, fmt.Errorf("provider not found")
	}
	models, err := p.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为 SelectorItem
	var items []*model.SelectorItem
	for _, m := range models {
		items = append(items, &model.SelectorItem{
			Name:        m.Name,
			Description: m.Description,
		})
	}

	return &model.CmdResult{
		Code: 0,
		Msg:  "success",
		Type: "selector", // 返回 selector 类型，启动交互式选择器
		Data: items,
	}, nil
}

func (l *ListModel) ToSelect(items []*model.SelectorItem) (*model.BaseSelector, error) {
	// 按名称排序，保证顺序一致
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return &model.BaseSelector{
		Items:         items,
		FilteredItems: items, // 初始时显示所有项
		Cursor:        0,
		ViewportTop:   0,
		ViewportSize:  15, // 默认显示15行
		SearchQuery:   "", // 初始搜索为空
		OnConfirm: func(selected string) error {
			// 保存选中的模型到配置
			config.GetLocalStoreConfig().Model = selected
			err := config.SaveConfig()
			if err != nil {
				log.Errorf("保存配置失败: %v", err)
				return err
			}
			log.Infof("已选择模型: %s", selected)
			return nil
		},
	}, nil
}
