package knowledge

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// 知识库根目录
const (
	KnowledgeBaseDir = ".msa/knowledge"

	// 知识库子目录
	DirUserProfile = "user-profile"
	DirStrategies  = "strategies"
	DirInsights    = "insights"
	DirNotes       = "notes"
	DirSummaries   = "summaries"
	DirErrors      = "errors"
)

// FileInfo 文件信息
type FileInfo struct {
	Name    string `json:"name"`    // 文件名
	Path    string `json:"path"`    // 相对路径
	Size    int64  `json:"size"`    // 文件大小（字节）
	ModTime string `json:"modTime"` // 修改时间
	IsDir   bool   `json:"isDir"`   // 是否为目录
}

// KnowledgeFileContent 知识文件内容
type KnowledgeFileContent struct {
	Metadata map[string]interface{} `json:"metadata"` // YAML frontmatter 元数据
	Content  string                 `json:"content"`  // Markdown 正文
}

// 单例模式：知识库路径
var (
	knowledgeBasePath string
	knowledgeBaseOnce sync.Once
)

// GetKnowledgeBasePath 获取知识库根目录路径（单例模式）
func GetKnowledgeBasePath() string {
	knowledgeBaseOnce.Do(func() {
		// 优先从环境变量获取
		if envPath := os.Getenv("MSA_KNOWLEDGE_PATH"); envPath != "" {
			knowledgeBasePath = envPath
			return
		}
		// 默认使用当前目录下的 .msa/knowledge
		wd, err := os.Getwd()
		if err != nil {
			log.Warnf("获取工作目录失败: %v，使用默认路径", err)
			knowledgeBasePath = KnowledgeBaseDir
			return
		}
		knowledgeBasePath = filepath.Join(wd, KnowledgeBaseDir)
	})
	return knowledgeBasePath
}

// SafePath 安全路径校验，防止路径遍历攻击
// 返回安全后的完整路径，如果路径不安全则返回错误
func SafePath(userPath string) (string, error) {
	if userPath == "" {
		return "", ErrEmptyPath
	}

	// 清理路径
	cleanPath := filepath.Clean(userPath)

	// 检查路径遍历
	if strings.Contains(cleanPath, "..") {
		return "", ErrPathTraversal
	}

	// 检查绝对路径（Unix 风格）
	if filepath.IsAbs(cleanPath) {
		return "", ErrAbsolutePath
	}

	// 检查绝对路径（Windows 风格，如 C:\ 或 C:/）
	// 格式：字母 + 冒号 + 分隔符
	if len(cleanPath) >= 3 {
		c := cleanPath[0]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			if cleanPath[1] == ':' && (cleanPath[2] == '\\' || cleanPath[2] == '/') {
				return "", ErrAbsolutePath
			}
		}
	}

	// 构建完整路径
	basePath := GetKnowledgeBasePath()
	fullPath := filepath.Join(basePath, cleanPath)

	// 验证最终路径在允许范围内
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", ErrInvalidPath
	}
	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", ErrInvalidPath
	}

	if !strings.HasPrefix(absFull, absBase) {
		return "", ErrPathTraversal
	}

	return fullPath, nil
}

// 错误定义
var (
	ErrEmptyPath     = &KnowledgeError{Code: "EMPTY_PATH", Message: "路径不能为空"}
	ErrPathTraversal = &KnowledgeError{Code: "PATH_TRAVERSAL", Message: "路径遍历不允许 | Path traversal not allowed"}
	ErrAbsolutePath  = &KnowledgeError{Code: "ABSOLUTE_PATH", Message: "不允许使用绝对路径 | Absolute path not allowed"}
	ErrInvalidPath   = &KnowledgeError{Code: "INVALID_PATH", Message: "无效的路径 | Invalid path"}
	ErrFileNotFound  = &KnowledgeError{Code: "FILE_NOT_FOUND", Message: "文件不存在 | File not found"}
	ErrDirNotFound   = &KnowledgeError{Code: "DIR_NOT_FOUND", Message: "目录不存在 | Directory not found"}
	ErrReadFailed    = &KnowledgeError{Code: "READ_FAILED", Message: "读取文件失败 | Failed to read file"}
	ErrWriteFailed   = &KnowledgeError{Code: "WRITE_FAILED", Message: "写入文件失败 | Failed to write file"}
	ErrInvalidTag    = &KnowledgeError{Code: "INVALID_TAG", Message: "无效的标签格式 | Invalid tag format"}
)

// KnowledgeError 知识库错误
type KnowledgeError struct {
	Code    string
	Message string
}

func (e *KnowledgeError) Error() string {
	return e.Message
}
