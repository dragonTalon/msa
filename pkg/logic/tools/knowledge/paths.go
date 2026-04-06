package knowledge

import (
	"os"
	"path/filepath"
)

// 知识库目录和文件常量
const (
	KnowledgeDir = ".msa"
	ErrorsFile   = "errors.md"
	SummariesDir = "summaries"
)

// GetKnowledgeDir 获取知识库根目录路径
func GetKnowledgeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, KnowledgeDir), nil
}

// GetErrorsPath 获取错误记录文件路径
func GetErrorsPath() (string, error) {
	knowledgeDir, err := GetKnowledgeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(knowledgeDir, ErrorsFile), nil
}

// GetSummariesDir 获取总结目录路径
func GetSummariesDir() (string, error) {
	knowledgeDir, err := GetKnowledgeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(knowledgeDir, SummariesDir), nil
}

// GetSummaryPath 获取指定日期的总结文件路径
func GetSummaryPath(date string) (string, error) {
	summariesDir, err := GetSummariesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(summariesDir, date+".md"), nil
}

// EnsureDir 确保目录存在，不存在则创建
func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}
