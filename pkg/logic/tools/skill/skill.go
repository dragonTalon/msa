package skill

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"

	"msa/pkg/logic/message"
	"msa/pkg/logic/skills"
	"msa/pkg/model"
)

// SkillContentParam get_skill_content 工具的输入参数
type SkillContentParam struct {
	SkillName string `json:"skill_name" jsonschema:"description=the name of the skill to retrieve content for"`
}

// SkillContentTool 查询 skill 完整内容的工具
type SkillContentTool struct{}

func (s *SkillContentTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(s.GetName(), s.GetDescription(), GetSkillContent)
}

func (s *SkillContentTool) GetName() string {
	return "get_skill_content"
}

func (s *SkillContentTool) GetDescription() string {
	return "获取指定 skill 的完整内容（处理流程和规则）| Get the full content (workflow and rules) of a specified skill"
}

func (s *SkillContentTool) GetToolGroup() model.ToolGroup {
	return model.SkillToolGroup
}

// GetSkillContent 根据 skill name 返回对应 SKILL.md 的完整 body 内容
func GetSkillContent(ctx context.Context, param *SkillContentParam) (string, error) {
	log.Infof("GetSkillContent start, skill_name: %s", param.SkillName)
	message.BroadcastToolStart("get_skill_content", fmt.Sprintf("skill_name: %s", param.SkillName))

	if param.SkillName == "" {
		err := fmt.Errorf("skill_name is required")
		message.BroadcastToolEnd("get_skill_content", "", err)
		return "", err
	}

	manager := skills.GetManager()
	sk, err := manager.GetSkill(param.SkillName)
	if err != nil {
		log.Errorf("GetSkillContent: failed to get skill %s: %v", param.SkillName, err)
		message.BroadcastToolEnd("get_skill_content", "", err)
		return "", err
	}
	if sk == nil {
		err = fmt.Errorf("skill '%s' not found", param.SkillName)
		log.Warnf("GetSkillContent: %v", err)
		message.BroadcastToolEnd("get_skill_content", "", err)
		return "", err
	}

	content, err := sk.GetContent()
	if err != nil {
		log.Errorf("GetSkillContent: failed to load content for skill %s: %v", param.SkillName, err)
		message.BroadcastToolEnd("get_skill_content", "", err)
		return "", err
	}

	message.BroadcastToolEnd("get_skill_content", fmt.Sprintf("loaded skill: %s (%d chars)", param.SkillName, len(content)), nil)
	return content, nil
}
