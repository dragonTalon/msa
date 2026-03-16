package knowledge

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"msa/pkg/logic/message"
	"msa/pkg/model"
)

// WriteKnowledgeFileParam 写入知识文件参数
type WriteKnowledgeFileParam struct {
	Path     string                 `json:"path" jsonschema:"description=文件相对路径，如 summaries/2026-03-15.md，required"`
	Content  string                 `json:"content" jsonschema:"description=文件内容（Markdown 格式），required"`
	Metadata map[string]interface{} `json:"metadata" jsonschema:"description=可选的 YAML frontmatter 元数据"`
}

// WriteKnowledgeFileTool 写入知识文件工具
type WriteKnowledgeFileTool struct{}

func (t *WriteKnowledgeFileTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), WriteKnowledgeFile)
}

func (t *WriteKnowledgeFileTool) GetName() string {
	return "write_knowledge_file"
}

func (t *WriteKnowledgeFileTool) GetDescription() string {
	return "写入知识库文件 | Write a knowledge base file (summaries, errors, insights, etc.)"
}

func (t *WriteKnowledgeFileTool) GetToolGroup() model.ToolGroup {
	return model.KnowledgeToolGroup
}

// WriteKnowledgeFileData 写入知识文件数据
type WriteKnowledgeFileData struct {
	Path string `json:"path"`
	Size int    `json:"size"`
}

// WriteKnowledgeFile 写入知识文件
func WriteKnowledgeFile(ctx context.Context, param *WriteKnowledgeFileParam) (string, error) {
	message.BroadcastToolStart("write_knowledge_file", fmt.Sprintf("路径: %s", param.Path))

	// 路径安全校验
	fullPath, err := SafePath(param.Path)
	if err != nil {
		message.BroadcastToolEnd("write_knowledge_file", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 生成文件内容
	var content string
	if len(param.Metadata) > 0 {
		content, err = GenerateFileContent(param.Metadata, param.Content)
		if err != nil {
			message.BroadcastToolEnd("write_knowledge_file", "", err)
			return model.NewErrorResult(fmt.Sprintf("生成文件内容失败: %v", err)), nil
		}
	} else {
		content = param.Content
	}

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		message.BroadcastToolEnd("write_knowledge_file", "", err)
		return model.NewErrorResult(fmt.Sprintf("创建目录失败: %v", err)), nil
	}

	// 原子写入
	if err := atomicWrite(fullPath, []byte(content)); err != nil {
		message.BroadcastToolEnd("write_knowledge_file", "", err)
		return model.NewErrorResult(fmt.Sprintf("写入文件失败: %v", err)), nil
	}

	data := &WriteKnowledgeFileData{
		Path: param.Path,
		Size: len(content),
	}

	message.BroadcastToolEnd("write_knowledge_file", "文件写入成功", nil)
	return model.NewSuccessResult(data, "文件写入成功"), nil
}

// atomicWrite 原子写入文件
// 先写入临时文件，成功后重命名，避免部分写入导致数据损坏
func atomicWrite(path string, content []byte) error {
	// 创建临时文件
	tmpPath := path + ".tmp"

	// 写入临时文件
	if err := os.WriteFile(tmpPath, content, 0644); err != nil {
		return err
	}

	// 原子重命名
	if err := os.Rename(tmpPath, path); err != nil {
		// 重命名失败，尝试清理临时文件
		os.Remove(tmpPath)
		return err
	}

	return nil
}
