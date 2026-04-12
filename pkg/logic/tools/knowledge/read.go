package knowledge

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"

	"msa/pkg/logic/message"
	"msa/pkg/logic/tools/safetool"
	"msa/pkg/model"
)

// ReadKnowledgeParam read_knowledge 工具的输入参数
type ReadKnowledgeParam struct {
	Type string `json:"type" jsonschema:"description=读取类型: errors(只读错误), summary(只读总结), all(读取两者)"`
}

// ReadKnowledgeData read_knowledge 返回数据
type ReadKnowledgeData struct {
	Errors      []ErrorEntry `json:"errors,omitempty"`
	PrevSummary *SummaryData `json:"prev_summary,omitempty"`
	HasErrors   bool         `json:"has_errors"`
	HasSummary  bool         `json:"has_summary"`
	ErrorsPath  string       `json:"errors_path,omitempty"`
	SummaryPath string       `json:"summary_path,omitempty"`
}

// ReadKnowledgeTool 读取知识库工具
type ReadKnowledgeTool struct{}

func (t *ReadKnowledgeTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), ReadKnowledge)
}

func (t *ReadKnowledgeTool) GetName() string {
	return "read_knowledge"
}

func (t *ReadKnowledgeTool) GetDescription() string {
	return "读取交易知识库（历史错误 + 昨日总结）| Read trading knowledge base (historical errors + yesterday's summary)"
}

func (t *ReadKnowledgeTool) GetToolGroup() model.ToolGroup {
	return model.KnowledgeToolGroup
}

// ReadKnowledge 读取知识库
func ReadKnowledge(ctx context.Context, param *ReadKnowledgeParam) (string, error) {
	return safetool.SafeExecute("read_knowledge", fmt.Sprintf("type: %s", param.Type), func() (string, error) {
		return doReadKnowledge(ctx, param)
	})
}

func doReadKnowledge(ctx context.Context, param *ReadKnowledgeParam) (string, error) {
	log.Infof("ReadKnowledge start, type: %s", param.Type)
	message.BroadcastToolStart("read_knowledge", fmt.Sprintf("type: %s", param.Type))

	// 默认读取全部
	readType := param.Type
	if readType == "" {
		readType = "all"
	}

	data := &ReadKnowledgeData{}

	// 根据类型读取
	switch readType {
	case "errors", "all":
		errorsPath, err := GetErrorsPath()
		if err != nil {
			log.Errorf("GetErrorsPath error: %v", err)
			message.BroadcastToolEnd("read_knowledge", "", err)
			return model.NewErrorResult(fmt.Sprintf("获取错误文件路径失败: %v", err)), nil
		}
		data.ErrorsPath = errorsPath

		entries, err := ParseErrorsFile(errorsPath)
		if err != nil {
			log.Errorf("ParseErrorsFile error: %v", err)
			message.BroadcastToolEnd("read_knowledge", "", err)
			return model.NewErrorResult(fmt.Sprintf("解析错误文件失败: %v", err)), nil
		}
		data.Errors = entries
		data.HasErrors = len(entries) > 0
	}

	switch readType {
	case "summary", "all":
		summaryPath, err := FindPrevSummaryFile()
		if err != nil {
			log.Errorf("FindPrevSummaryFile error: %v", err)
			message.BroadcastToolEnd("read_knowledge", "", err)
			return model.NewErrorResult(fmt.Sprintf("查找总结文件失败: %v", err)), nil
		}

		if summaryPath != "" {
			data.SummaryPath = summaryPath
			summary, err := ParseSummaryFile(summaryPath)
			if err != nil {
				log.Errorf("ParseSummaryFile error: %v", err)
				message.BroadcastToolEnd("read_knowledge", "", err)
				return model.NewErrorResult(fmt.Sprintf("解析总结文件失败: %v", err)), nil
			}
			data.PrevSummary = summary
			data.HasSummary = true
		}
	}

	// 构建结果消息
	resultMsg := fmt.Sprintf("读取知识库完成: has_errors=%v, has_summary=%v",
		data.HasErrors, data.HasSummary)
	message.BroadcastToolEnd("read_knowledge", resultMsg, nil)

	return model.NewSuccessResult(data, resultMsg), nil
}
