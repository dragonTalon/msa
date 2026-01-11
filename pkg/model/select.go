package model

import "context"

// SelectorItem 选择器项
type SelectorItem struct {
	Name        string
	Description string
}

// BaseSelector 通用选择器
type BaseSelector struct {
	Items         []*SelectorItem             // 选择项列表
	FilteredItems []*SelectorItem             // 过滤后的选择项列表
	Cursor        int                         // 当前光标位置
	Selected      string                      // 选中的项
	Ctx           context.Context             // 上下文
	Width         int                         // 终端宽度
	Height        int                         // 终端高度
	Confirmed     bool                        // 是否已确认
	Err           error                       // 错误信息
	ViewportTop   int                         // 视口顶部位置
	ViewportSize  int                         // 视口大小（可见行数）
	SearchQuery   string                      // 搜索关键字
	OnConfirm     func(selected string) error // 确认回调函数
}
