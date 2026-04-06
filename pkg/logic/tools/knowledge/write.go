package knowledge

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"

	"msa/pkg/logic/message"
	"msa/pkg/model"
)

// 写入相关常量
const (
	MaxRetries   = 3
	RetryDelay   = 100 * time.Millisecond
	ErrorsHeader = `# 错误记录

本文件记录交易过程中的错误操作和教训，供后续决策参考。

`
)

// WriteKnowledgeParam write_knowledge 工具的输入参数
type WriteKnowledgeParam struct {
	Type    string `json:"type" jsonschema:"description=写入类型: error(错误记录) 或 summary(每日总结)"`
	Content string `json:"content" jsonschema:"description=写入内容"`
	Date    string `json:"date,omitempty" jsonschema:"description=日期(YYYY-MM-DD格式，summary类型必填)"`
}

// WriteKnowledgeData write_knowledge 返回数据
type WriteKnowledgeData struct {
	Type       string `json:"type"`
	Path       string `json:"path"`
	Verified   bool   `json:"verified"`
	ContentLen int    `json:"content_len"`
	RetryCount int    `json:"retry_count"`
}

// WriteKnowledgeTool 写入知识库工具
type WriteKnowledgeTool struct{}

func (t *WriteKnowledgeTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), WriteKnowledge)
}

func (t *WriteKnowledgeTool) GetName() string {
	return "write_knowledge"
}

func (t *WriteKnowledgeTool) GetDescription() string {
	return "写入交易知识库（错误记录 / 每日总结），带验证和重试 | Write to trading knowledge base with verification and retry"
}

func (t *WriteKnowledgeTool) GetToolGroup() model.ToolGroup {
	return model.KnowledgeToolGroup
}

// WriteKnowledge 写入知识库
func WriteKnowledge(ctx context.Context, param *WriteKnowledgeParam) (string, error) {
	log.Infof("WriteKnowledge start, type: %s", param.Type)
	message.BroadcastToolStart("write_knowledge", fmt.Sprintf("type: %s, content_len: %d", param.Type, len(param.Content)))

	// 参数验证
	if err := validateWriteParam(param); err != nil {
		log.Errorf("validateWriteParam error: %v", err)
		message.BroadcastToolEnd("write_knowledge", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	var path string
	var verifyErr error
	retryCount := 0

	// 重试循环
	for retryCount = 0; retryCount < MaxRetries; retryCount++ {
		if retryCount > 0 {
			time.Sleep(RetryDelay)
			log.Infof("WriteKnowledge retry %d/%d", retryCount, MaxRetries)
		}

		// 执行写入
		path, verifyErr = executeWrite(param)
		if verifyErr != nil {
			continue
		}

		// 验证写入
		if verifyWrite(path, param) {
			data := &WriteKnowledgeData{
				Type:       param.Type,
				Path:       path,
				Verified:   true,
				ContentLen: len(param.Content),
				RetryCount: retryCount,
			}

			resultMsg := fmt.Sprintf("写入成功: type=%s, path=%s, verified=true", param.Type, path)
			message.BroadcastToolEnd("write_knowledge", resultMsg, nil)
			return model.NewSuccessResult(data, resultMsg), nil
		}

		verifyErr = fmt.Errorf("验证失败")
	}

	// 重试耗尽
	data := &WriteKnowledgeData{
		Type:       param.Type,
		Path:       path,
		Verified:   false,
		ContentLen: len(param.Content),
		RetryCount: retryCount,
	}

	errMsg := fmt.Sprintf("写入失败: 重试%d次后仍验证失败", MaxRetries)
	log.Errorf("WriteKnowledge failed after %d retries", MaxRetries)
	message.BroadcastToolEnd("write_knowledge", "", fmt.Errorf("%s", errMsg))

	return model.NewSuccessResult(data, errMsg), nil
}

// validateWriteParam 验证写入参数
func validateWriteParam(param *WriteKnowledgeParam) error {
	if param.Type == "" {
		return fmt.Errorf("type 是必填项")
	}

	if param.Type != "error" && param.Type != "summary" {
		return fmt.Errorf("无效的 type，必须是 error 或 summary")
	}

	if param.Content == "" {
		return fmt.Errorf("content 是必填项")
	}

	// summary 类型需要日期
	if param.Type == "summary" {
		if param.Date == "" {
			return fmt.Errorf("summary 类型需要 date 参数")
		}
		if !isValidDate(param.Date) {
			return fmt.Errorf("date 格式无效，需要 YYYY-MM-DD 格式")
		}
	}

	return nil
}

// executeWrite 执行写入
func executeWrite(param *WriteKnowledgeParam) (string, error) {
	switch param.Type {
	case "error":
		return appendError(param.Content)
	case "summary":
		return writeSummary(param.Date, param.Content)
	default:
		return "", fmt.Errorf("未知的写入类型: %s", param.Type)
	}
}

// appendError 追加错误记录
func appendError(content string) (string, error) {
	// 确保目录存在
	knowledgeDir, err := GetKnowledgeDir()
	if err != nil {
		return "", err
	}
	if err := EnsureDir(knowledgeDir); err != nil {
		return "", err
	}

	// 获取错误文件路径
	errorsPath, err := GetErrorsPath()
	if err != nil {
		return "", err
	}

	// 检查文件是否存在
	var fileContent string
	if _, err := os.Stat(errorsPath); os.IsNotExist(err) {
		// 文件不存在，创建并写入标题
		fileContent = ErrorsHeader
	} else {
		// 读取现有内容
		data, err := os.ReadFile(errorsPath)
		if err != nil {
			return "", err
		}
		fileContent = string(data)
	}

	// 追加新内容（添加分隔符）
	newContent := fileContent
	if !strings.HasSuffix(newContent, "\n\n") {
		if strings.HasSuffix(newContent, "\n") {
			newContent += "\n"
		} else {
			newContent += "\n\n"
		}
	}
	newContent += content + "\n"

	// 写入文件
	if err := os.WriteFile(errorsPath, []byte(newContent), 0644); err != nil {
		return "", err
	}

	return errorsPath, nil
}

// writeSummary 写入总结
func writeSummary(date, content string) (string, error) {
	// 确保目录存在
	summariesDir, err := GetSummariesDir()
	if err != nil {
		return "", err
	}
	if err := EnsureDir(summariesDir); err != nil {
		return "", err
	}

	// 获取总结文件路径
	summaryPath, err := GetSummaryPath(date)
	if err != nil {
		return "", err
	}

	// 写入文件（覆盖）
	if err := os.WriteFile(summaryPath, []byte(content), 0644); err != nil {
		return "", err
	}

	return summaryPath, nil
}

// verifyWrite 验证写入
func verifyWrite(path string, param *WriteKnowledgeParam) bool {
	// 读取文件
	data, err := os.ReadFile(path)
	if err != nil {
		log.Errorf("verifyWrite: read file error: %v", err)
		return false
	}

	content := string(data)

	// 验证内容是否包含写入的内容
	switch param.Type {
	case "error":
		return verifyErrorContent(content, param.Content)
	case "summary":
		return verifySummaryContent(content, param.Content)
	default:
		return false
	}
}

// verifyErrorContent 验证错误记录内容
func verifyErrorContent(fileContent, writtenContent string) bool {
	// 检查文件是否包含写入的内容
	return strings.Contains(fileContent, writtenContent)
}

// verifySummaryContent 验证总结内容
func verifySummaryContent(fileContent, writtenContent string) bool {
	// 检查文件内容是否完全匹配（总结文件是覆盖写入）
	return strings.TrimSpace(fileContent) == strings.TrimSpace(writtenContent) ||
		strings.Contains(fileContent, writtenContent)
}
