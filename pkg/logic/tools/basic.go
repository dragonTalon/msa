package tools

import (
	"github.com/cloudwego/eino/components/tool"
	log "github.com/sirupsen/logrus"
	"msa/pkg/model"
)

type MsaTool interface {
	// GetToolInfo 获取工具信息
	GetToolInfo() (tool.BaseTool, error)
	// GetName 获取工具名称
	GetName() string
	// GetDescription 获取工具描述
	GetDescription() string
	// GetToolGroup 获取工具组
	GetToolGroup() model.ToolGroup
}

var toolGroupMap = map[model.ToolGroup][]*model.Pair{}

var toolMap = map[string]MsaTool{}

// GetAllTools 获取所有工具
func GetAllTools() []tool.BaseTool {
	list := []tool.BaseTool{}
	for _, t := range toolMap {
		info, err := t.GetToolInfo()
		if err != nil {
			log.Errorf("GetToolInfo %s  fail : %v", t.GetName(), err)
			continue
		}
		list = append(list, info)
	}
	return list
}

func RegisterTool(tool MsaTool) {
	if toolGroupMap[tool.GetToolGroup()] == nil {
		toolGroupMap[tool.GetToolGroup()] = []*model.Pair{}
	}
	toolGroupMap[tool.GetToolGroup()] = append(toolGroupMap[tool.GetToolGroup()], &model.Pair{
		Key:   tool.GetName(),
		Value: tool.GetDescription(),
	})
	toolMap[tool.GetName()] = tool
}
