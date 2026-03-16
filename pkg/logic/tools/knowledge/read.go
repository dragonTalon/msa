package knowledge

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"msa/pkg/logic/message"
	"msa/pkg/model"
)

// ReadKnowledgeFileParam 读取知识文件参数
type ReadKnowledgeFileParam struct {
	Path string `json:"path" jsonschema:"description=文件相对路径，如 notes/2026-03-15.md，required"`
}

// ReadKnowledgeFileTool 读取知识文件工具
type ReadKnowledgeFileTool struct{}

func (t *ReadKnowledgeFileTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), ReadKnowledgeFile)
}

func (t *ReadKnowledgeFileTool) GetName() string {
	return "read_knowledge_file"
}

func (t *ReadKnowledgeFileTool) GetDescription() string {
	return "读取知识库文件 | Read a knowledge base file (notes, strategies, errors, etc.)"
}

func (t *ReadKnowledgeFileTool) GetToolGroup() model.ToolGroup {
	return model.KnowledgeToolGroup
}

// KnowledgeFileData 知识文件数据
type KnowledgeFileData struct {
	Path     string                 `json:"path"`
	Size     int64                  `json:"size"`
	ModTime  string                 `json:"mod_time"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Content  string                 `json:"content"`
}

// ReadKnowledgeFile 读取知识文件
func ReadKnowledgeFile(ctx context.Context, param *ReadKnowledgeFileParam) (string, error) {
	message.BroadcastToolStart("read_knowledge_file", fmt.Sprintf("路径: %s", param.Path))

	// 路径安全校验
	fullPath, err := SafePath(param.Path)
	if err != nil {
		message.BroadcastToolEnd("read_knowledge_file", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 检查文件是否存在
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			message.BroadcastToolEnd("read_knowledge_file", "", ErrFileNotFound)
			return model.NewErrorResult(fmt.Sprintf("%s: %s", ErrFileNotFound.Message, param.Path)), nil
		}
		message.BroadcastToolEnd("read_knowledge_file", "", err)
		return model.NewErrorResult(fmt.Sprintf("读取文件失败: %v", err)), nil
	}

	// 不允许读取目录
	if info.IsDir() {
		message.BroadcastToolEnd("read_knowledge_file", "", ErrReadFailed)
		return model.NewErrorResult(fmt.Sprintf("路径是目录，不是文件: %s", param.Path)), nil
	}

	// 读取文件内容
	content, err := os.ReadFile(fullPath)
	if err != nil {
		message.BroadcastToolEnd("read_knowledge_file", "", err)
		return model.NewErrorResult(fmt.Sprintf("读取文件失败: %v", err)), nil
	}

	// 解析 frontmatter
	metadata, body := ParseFileWithFrontmatter(content)

	// 构建返回数据
	data := &KnowledgeFileData{
		Path:     param.Path,
		Size:     info.Size(),
		ModTime:  info.ModTime().Format("2006-01-02 15:04:05"),
		Metadata: metadata,
		Content:  body,
	}

	message.BroadcastToolEnd("read_knowledge_file", "读取文件成功", nil)
	return model.NewSuccessResult(data, "读取文件成功"), nil
}

// formatReadResult 格式化读取结果（保留用于日志）
func formatReadResult(path string, metadata map[string]interface{}, body string, info os.FileInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("文件: %s\n", path))
	sb.WriteString(fmt.Sprintf("大小: %d 字节\n", info.Size()))
	sb.WriteString(fmt.Sprintf("修改时间: %s\n", info.ModTime().Format("2006-01-02 15:04:05")))

	if len(metadata) > 0 {
		sb.WriteString("\n--- 元数据 ---\n")
		for key, value := range metadata {
			sb.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
	}

	sb.WriteString("\n--- 内容 ---\n")
	sb.WriteString(body)

	return sb.String()
}
