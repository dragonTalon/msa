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
	return "获取指定 skill 的完整内容（处理流程和规则）。返回内容包括是否有 references/assets 可按需加载 | Get the full content (workflow and rules) of a specified skill. Returns whether references/assets are available for on-demand loading"
}

func (s *SkillContentTool) GetToolGroup() model.ToolGroup {
	return model.SkillToolGroup
}

// SkillContentData skill 内容数据
type SkillContentData struct {
	SkillName     string `json:"skill_name"`
	Content       string `json:"content"`
	Length        int    `json:"length"`
	HasReferences bool   `json:"has_references"`
	HasAssets     bool   `json:"has_assets"`
}

// GetSkillContent 根据 skill name 返回对应 SKILL.md 的完整 body 内容
func GetSkillContent(ctx context.Context, param *SkillContentParam) (string, error) {
	log.Infof("GetSkillContent start, skill_name: %s", param.SkillName)
	message.BroadcastToolStart("get_skill_content", fmt.Sprintf("skill_name: %s", param.SkillName))

	if param.SkillName == "" {
		err := fmt.Errorf("skill_name is required")
		message.BroadcastToolEnd("get_skill_content", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	manager := skills.GetManager()
	sk, err := manager.GetSkill(param.SkillName)
	if err != nil {
		log.Errorf("GetSkillContent: failed to get skill %s: %v", param.SkillName, err)
		message.BroadcastToolEnd("get_skill_content", "", err)
		return model.NewErrorResult(err.Error()), nil
	}
	if sk == nil {
		err = fmt.Errorf("skill '%s' not found", param.SkillName)
		log.Warnf("GetSkillContent: %v", err)
		message.BroadcastToolEnd("get_skill_content", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	content, err := sk.GetContent()
	if err != nil {
		log.Errorf("GetSkillContent: failed to load content for skill %s: %v", param.SkillName, err)
		message.BroadcastToolEnd("get_skill_content", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	data := &SkillContentData{
		SkillName:     param.SkillName,
		Content:       content,
		Length:        len(content),
		HasReferences: sk.HasReferences(),
		HasAssets:     sk.HasAssets(),
	}

	message.BroadcastToolEnd("get_skill_content", fmt.Sprintf("loaded skill: %s (%d chars)", param.SkillName, len(content)), nil)
	return model.NewSuccessResult(data, fmt.Sprintf("加载 skill: %s (%d 字符)", param.SkillName, len(content))), nil
}

// ========== Skill Reference Tool ==========

// SkillReferenceParam get_skill_reference 工具的输入参数
type SkillReferenceParam struct {
	SkillName string `json:"skill_name" jsonschema:"description=the name of the skill"`
	FileName  string `json:"file_name"  jsonschema:"description=the name of the reference file to load"`
}

// SkillReferenceTool 查询 skill reference 文件的工具
type SkillReferenceTool struct{}

func (s *SkillReferenceTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(s.GetName(), s.GetDescription(), GetSkillReference)
}

func (s *SkillReferenceTool) GetName() string {
	return "get_skill_reference"
}

func (s *SkillReferenceTool) GetDescription() string {
	return "获取指定 skill 的 reference 文件内容（按需加载知识库）| Get the content of a reference file from a specified skill (on-demand knowledge loading)"
}

func (s *SkillReferenceTool) GetToolGroup() model.ToolGroup {
	return model.SkillToolGroup
}

// SkillReferenceData skill reference 数据
type SkillReferenceData struct {
	SkillName  string `json:"skill_name"`
	FileName   string `json:"file_name"`
	Content    string `json:"content"`
	Length     int    `json:"length"`
}

// GetSkillReference 根据 skill name 和 file name 返回 reference 文件内容
func GetSkillReference(ctx context.Context, param *SkillReferenceParam) (string, error) {
	log.Infof("GetSkillReference start, skill_name: %s, file_name: %s", param.SkillName, param.FileName)
	message.BroadcastToolStart("get_skill_reference", fmt.Sprintf("skill_name: %s, file_name: %s", param.SkillName, param.FileName))

	if param.SkillName == "" {
		err := fmt.Errorf("skill_name is required")
		message.BroadcastToolEnd("get_skill_reference", "", err)
		return model.NewErrorResult(err.Error()), nil
	}
	if param.FileName == "" {
		err := fmt.Errorf("file_name is required")
		message.BroadcastToolEnd("get_skill_reference", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	manager := skills.GetManager()
	sk, err := manager.GetSkill(param.SkillName)
	if err != nil {
		log.Errorf("GetSkillReference: failed to get skill %s: %v", param.SkillName, err)
		message.BroadcastToolEnd("get_skill_reference", "", err)
		return model.NewErrorResult(err.Error()), nil
	}
	if sk == nil {
		err = fmt.Errorf("skill '%s' not found", param.SkillName)
		log.Warnf("GetSkillReference: %v", err)
		message.BroadcastToolEnd("get_skill_reference", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 检查是否有 references 目录
	if !sk.HasReferences() {
		err = fmt.Errorf("skill '%s' has no references directory", param.SkillName)
		log.Warnf("GetSkillReference: %v", err)
		message.BroadcastToolEnd("get_skill_reference", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	content, err := sk.GetReference(param.FileName)
	if err != nil {
		log.Errorf("GetSkillReference: failed to load reference %s for skill %s: %v", param.FileName, param.SkillName, err)
		message.BroadcastToolEnd("get_skill_reference", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	data := &SkillReferenceData{
		SkillName: param.SkillName,
		FileName:  param.FileName,
		Content:   content,
		Length:    len(content),
	}

	message.BroadcastToolEnd("get_skill_reference", fmt.Sprintf("loaded reference: %s/%s (%d chars)", param.SkillName, param.FileName, len(content)), nil)
	return model.NewSuccessResult(data, fmt.Sprintf("加载 reference: %s/%s (%d 字符)", param.SkillName, param.FileName, len(content))), nil
}

// ========== Skill Asset Tool ==========

// SkillAssetParam get_skill_asset 工具的输入参数
type SkillAssetParam struct {
	SkillName string `json:"skill_name" jsonschema:"description=the name of the skill"`
	FileName  string `json:"file_name"  jsonschema:"description=the name of the asset file to load"`
}

// SkillAssetTool 查询 skill asset 文件的工具
type SkillAssetTool struct{}

func (s *SkillAssetTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(s.GetName(), s.GetDescription(), GetSkillAsset)
}

func (s *SkillAssetTool) GetName() string {
	return "get_skill_asset"
}

func (s *SkillAssetTool) GetDescription() string {
	return "获取指定 skill 的 asset 文件内容（按需加载模板）| Get the content of an asset file from a specified skill (on-demand template loading)"
}

func (s *SkillAssetTool) GetToolGroup() model.ToolGroup {
	return model.SkillToolGroup
}

// SkillAssetData skill asset 数据
type SkillAssetData struct {
	SkillName string `json:"skill_name"`
	FileName  string `json:"file_name"`
	Content   string `json:"content"`
	Length    int    `json:"length"`
}

// GetSkillAsset 根据 skill name 和 file name 返回 asset 文件内容
func GetSkillAsset(ctx context.Context, param *SkillAssetParam) (string, error) {
	log.Infof("GetSkillAsset start, skill_name: %s, file_name: %s", param.SkillName, param.FileName)
	message.BroadcastToolStart("get_skill_asset", fmt.Sprintf("skill_name: %s, file_name: %s", param.SkillName, param.FileName))

	if param.SkillName == "" {
		err := fmt.Errorf("skill_name is required")
		message.BroadcastToolEnd("get_skill_asset", "", err)
		return model.NewErrorResult(err.Error()), nil
	}
	if param.FileName == "" {
		err := fmt.Errorf("file_name is required")
		message.BroadcastToolEnd("get_skill_asset", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	manager := skills.GetManager()
	sk, err := manager.GetSkill(param.SkillName)
	if err != nil {
		log.Errorf("GetSkillAsset: failed to get skill %s: %v", param.SkillName, err)
		message.BroadcastToolEnd("get_skill_asset", "", err)
		return model.NewErrorResult(err.Error()), nil
	}
	if sk == nil {
		err = fmt.Errorf("skill '%s' not found", param.SkillName)
		log.Warnf("GetSkillAsset: %v", err)
		message.BroadcastToolEnd("get_skill_asset", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 检查是否有 assets 目录
	if !sk.HasAssets() {
		err = fmt.Errorf("skill '%s' has no assets directory", param.SkillName)
		log.Warnf("GetSkillAsset: %v", err)
		message.BroadcastToolEnd("get_skill_asset", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	content, err := sk.GetAsset(param.FileName)
	if err != nil {
		log.Errorf("GetSkillAsset: failed to load asset %s for skill %s: %v", param.FileName, param.SkillName, err)
		message.BroadcastToolEnd("get_skill_asset", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	data := &SkillAssetData{
		SkillName: param.SkillName,
		FileName:  param.FileName,
		Content:   content,
		Length:    len(content),
	}

	message.BroadcastToolEnd("get_skill_asset", fmt.Sprintf("loaded asset: %s/%s (%d chars)", param.SkillName, param.FileName, len(content)), nil)
	return model.NewSuccessResult(data, fmt.Sprintf("加载 asset: %s/%s (%d 字符)", param.SkillName, param.FileName, len(content))), nil
}
