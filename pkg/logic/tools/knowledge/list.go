package knowledge

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"msa/pkg/logic/message"
	"msa/pkg/model"
)

// ListKnowledgeFilesParam 列出知识文件参数
type ListKnowledgeFilesParam struct {
	Directory string `json:"directory" jsonschema:"description=目录名称（如 errors, notes），为空则列出根目录"`
}

// ListKnowledgeFilesTool 列出知识文件工具
type ListKnowledgeFilesTool struct{}

func (t *ListKnowledgeFilesTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), ListKnowledgeFiles)
}

func (t *ListKnowledgeFilesTool) GetName() string {
	return "list_knowledge_files"
}

func (t *ListKnowledgeFilesTool) GetDescription() string {
	return "列出知识库目录下的文件 | List files in knowledge base directory"
}

func (t *ListKnowledgeFilesTool) GetToolGroup() model.ToolGroup {
	return model.KnowledgeToolGroup
}

// ListKnowledgeFilesData 列出知识文件数据
type ListKnowledgeFilesData struct {
	Directory string     `json:"directory"`
	Total     int        `json:"total"`
	Items     []FileInfo `json:"items"`
}

// ListKnowledgeFiles 列出知识文件
func ListKnowledgeFiles(ctx context.Context, param *ListKnowledgeFilesParam) (string, error) {
	message.BroadcastToolStart("list_knowledge_files", fmt.Sprintf("目录: %s", param.Directory))

	basePath := GetKnowledgeBasePath()
	var targetPath string

	if param.Directory == "" {
		// 列出根目录下的子目录
		targetPath = basePath
	} else {
		// 列出指定目录下的文件
		safeDir, err := SafePath(param.Directory)
		if err != nil {
			message.BroadcastToolEnd("list_knowledge_files", "", err)
			return model.NewErrorResult(err.Error()), nil
		}
		targetPath = safeDir
	}

	// 读取目录
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 目录不存在，返回空列表
			message.BroadcastToolEnd("list_knowledge_files", "目录不存在或为空", nil)
			return model.NewSuccessResult(&ListKnowledgeFilesData{Directory: param.Directory, Total: 0, Items: []FileInfo{}}, "目录不存在或为空"), nil
		}
		message.BroadcastToolEnd("list_knowledge_files", "", err)
		return model.NewErrorResult(fmt.Sprintf("读取目录失败: %v", err)), nil
	}

	// 收集文件信息
	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		relativePath := entry.Name()
		if param.Directory != "" {
			relativePath = param.Directory + "/" + entry.Name()
		}

		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    relativePath,
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
			IsDir:   entry.IsDir(),
		})
	}

	// 按修改时间排序（最新的在前）
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime > files[j].ModTime
	})

	data := &ListKnowledgeFilesData{
		Directory: param.Directory,
		Total:     len(files),
		Items:     files,
	}

	message.BroadcastToolEnd("list_knowledge_files", fmt.Sprintf("列出 %d 项", len(files)), nil)
	return model.NewSuccessResult(data, fmt.Sprintf("列出 %d 项", len(files))), nil
}
