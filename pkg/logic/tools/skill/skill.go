package skill

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"

	"msa/pkg/logic/skills"
	"msa/pkg/logic/tools/safetool"
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
	return safetool.SafeExecute("get_skill_content", fmt.Sprintf("skill_name: %s", param.SkillName), func() (string, error) {
		return doGetSkillContent(ctx, param)
	})
}

func doGetSkillContent(ctx context.Context, param *SkillContentParam) (string, error) {
	log.Infof("GetSkillContent start, skill_name: %s", param.SkillName)

	if param.SkillName == "" {
		err := fmt.Errorf("skill_name is required")
		return model.NewErrorResult(err.Error()), nil
	}

	manager := skills.GetManager()
	sk, err := manager.GetSkill(param.SkillName)
	if err != nil {
		log.Errorf("GetSkillContent: failed to get skill %s: %v", param.SkillName, err)
		return model.NewErrorResult(err.Error()), nil
	}
	if sk == nil {
		err = fmt.Errorf("skill '%s' not found", param.SkillName)
		log.Warnf("GetSkillContent: %v", err)
		return model.NewErrorResult(err.Error()), nil
	}

	content, err := sk.GetContent()
	if err != nil {
		log.Errorf("GetSkillContent: failed to load content for skill %s: %v", param.SkillName, err)
		return model.NewErrorResult(err.Error()), nil
	}

	data := &SkillContentData{
		SkillName:     param.SkillName,
		Content:       content,
		Length:        len(content),
		HasReferences: sk.HasReferences(),
		HasAssets:     sk.HasAssets(),
	}

	return model.NewSuccessResult(data, fmt.Sprintf("加载 skill: %s (%d 字符)", param.SkillName, len(content))), nil
}

// ========== Skill Reference Tool ==========

// SkillReferenceParam get_skill_reference 工具的输入参数
type SkillReferenceParam struct {
	SkillName string `json:"skill_name" jsonschema:"description=the name of the skill (can be omitted if file_name contains a cross-skill path like 'other-skill/references/file.md')"`
	FileName  string `json:"file_name"  jsonschema:"description=the reference file to load. Supports cross-skill paths: 'other-skill/references/file.md' auto-resolves the skill, 'references/file.md' or 'file.md' for same skill"`
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
	return "获取指定 skill 的 reference 文件内容（按需加载知识库）。支持跨 Skill 引用: file_name 可使用 \"other-skill/references/file.md\" 格式自动解析目标 Skill | Get the content of a reference file from a specified skill (on-demand knowledge loading). Cross-skill: file_name can use \"other-skill/references/file.md\" format for auto-resolution"
}

func (s *SkillReferenceTool) GetToolGroup() model.ToolGroup {
	return model.SkillToolGroup
}

// SkillReferenceData skill reference 数据
type SkillReferenceData struct {
	SkillName string `json:"skill_name"`
	FileName  string `json:"file_name"`
	Content   string `json:"content"`
	Length    int    `json:"length"`
}

// GetSkillReference 根据 skill name 和 file name 返回 reference 文件内容
func GetSkillReference(ctx context.Context, param *SkillReferenceParam) (string, error) {
	return safetool.SafeExecute("get_skill_reference", fmt.Sprintf("skill_name: %s, file_name: %s", param.SkillName, param.FileName), func() (string, error) {
		return doGetSkillReference(ctx, param)
	})
}

// parseReferencePath 解析跨 Skill 引用路径
// 支持三种格式:
//   - "skill-name/references/filename.md" → 提取 skill_name 和 filename
//   - "references/filename.md"             → 提取 filename (同 Skill)
//   - "filename.md"                        → 原样返回 (同 Skill)
//
// 返回: skillName (空字符串表示使用原始 skill_name), fileName
func parseReferencePath(fileName string) (string, string) {
	if !strings.Contains(fileName, "/") {
		return "", fileName
	}

	parts := strings.SplitN(fileName, "/", 3)
	if len(parts) == 2 {
		// "references/filename.md" → 同 Skill
		if parts[0] == "references" {
			return "", parts[1]
		}
		// "skill-name/filename.md" → 跨 Skill (无 references 前缀)
		return parts[0], parts[1]
	}

	if len(parts) >= 3 {
		// "skill-name/references/filename.md" → 跨 Skill
		if parts[1] == "references" {
			return parts[0], parts[2]
		}
		// "skill-name/other/filename.md" → 尝试提取第一段作为 skill_name
		return parts[0], parts[2]
	}

	return "", fileName
}

func doGetSkillReference(ctx context.Context, param *SkillReferenceParam) (string, error) {
	log.Infof("GetSkillReference start, skill_name: %s, file_name: %s", param.SkillName, param.FileName)

	if param.SkillName == "" && !strings.Contains(param.FileName, "/") {
		err := fmt.Errorf("skill_name is required (file_name does not contain a path)")
		return model.NewErrorResult(err.Error()), nil
	}
	if param.FileName == "" {
		err := fmt.Errorf("file_name is required")
		return model.NewErrorResult(err.Error()), nil
	}

	// 智能解析跨 Skill 引用路径
	resolvedSkill, resolvedFile := parseReferencePath(param.FileName)
	skillName := param.SkillName
	fileName := param.FileName

	if resolvedSkill != "" {
		// 从路径中提取了 skill_name
		if param.SkillName != "" && param.SkillName != resolvedSkill {
			log.Infof("GetSkillReference: overriding skill_name from '%s' to '%s' based on path", param.SkillName, resolvedSkill)
		}
		skillName = resolvedSkill
		fileName = resolvedFile
		log.Infof("GetSkillReference: resolved cross-skill reference -> skill: %s, file: %s", skillName, fileName)
	} else if resolvedFile != fileName {
		// 同 Skill 的 "references/filename.md" 格式
		fileName = resolvedFile
		log.Infof("GetSkillReference: resolved same-skill reference -> file: %s", fileName)
	}

	if skillName == "" {
		err := fmt.Errorf("skill_name is required (could not resolve from path: %s)", param.FileName)
		return model.NewErrorResult(err.Error()), nil
	}

	manager := skills.GetManager()
	sk, err := manager.GetSkill(skillName)
	if err != nil {
		log.Errorf("GetSkillReference: failed to get skill %s: %v", skillName, err)
		return model.NewErrorResult(err.Error()), nil
	}
	if sk == nil {
		err = fmt.Errorf("skill '%s' not found", skillName)
		log.Warnf("GetSkillReference: %v", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 检查是否有 references 目录
	if !sk.HasReferences() {
		err = fmt.Errorf("skill '%s' has no references directory", skillName)
		log.Warnf("GetSkillReference: %v", err)
		return model.NewErrorResult(err.Error()), nil
	}

	content, err := sk.GetReference(fileName)
	if err != nil {
		log.Errorf("GetSkillReference: failed to load reference %s for skill %s: %v", fileName, skillName, err)
		return model.NewErrorResult(err.Error()), nil
	}

	data := &SkillReferenceData{
		SkillName: skillName,
		FileName:  fileName,
		Content:   content,
		Length:    len(content),
	}

	return model.NewSuccessResult(data, fmt.Sprintf("加载 reference: %s/%s (%d 字符)", skillName, fileName, len(content))), nil
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
	return safetool.SafeExecute("get_skill_asset", fmt.Sprintf("skill_name: %s, file_name: %s", param.SkillName, param.FileName), func() (string, error) {
		return doGetSkillAsset(ctx, param)
	})
}

func doGetSkillAsset(ctx context.Context, param *SkillAssetParam) (string, error) {
	log.Infof("GetSkillAsset start, skill_name: %s, file_name: %s", param.SkillName, param.FileName)

	if param.SkillName == "" {
		err := fmt.Errorf("skill_name is required")
		return model.NewErrorResult(err.Error()), nil
	}
	if param.FileName == "" {
		err := fmt.Errorf("file_name is required")
		return model.NewErrorResult(err.Error()), nil
	}

	manager := skills.GetManager()
	sk, err := manager.GetSkill(param.SkillName)
	if err != nil {
		log.Errorf("GetSkillAsset: failed to get skill %s: %v", param.SkillName, err)
		return model.NewErrorResult(err.Error()), nil
	}
	if sk == nil {
		err = fmt.Errorf("skill '%s' not found", param.SkillName)
		log.Warnf("GetSkillAsset: %v", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 检查是否有 assets 目录
	if !sk.HasAssets() {
		err = fmt.Errorf("skill '%s' has no assets directory", param.SkillName)
		log.Warnf("GetSkillAsset: %v", err)
		return model.NewErrorResult(err.Error()), nil
	}

	content, err := sk.GetAsset(param.FileName)
	if err != nil {
		log.Errorf("GetSkillAsset: failed to load asset %s for skill %s: %v", param.FileName, param.SkillName, err)
		return model.NewErrorResult(err.Error()), nil
	}

	data := &SkillAssetData{
		SkillName: param.SkillName,
		FileName:  param.FileName,
		Content:   content,
		Length:    len(content),
	}

	return model.NewSuccessResult(data, fmt.Sprintf("加载 asset: %s/%s (%d 字符)", param.SkillName, param.FileName, len(content))), nil
}
