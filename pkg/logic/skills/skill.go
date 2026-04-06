package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// SkillSource 表示 Skill 的来源
type SkillSource int

const (
	// SkillSourceBuiltin 表示内置 Skill
	SkillSourceBuiltin SkillSource = iota
	// SkillSourceUser 表示用户自定义 Skill
	SkillSourceUser
)

// String 返回 SkillSource 的字符串表示
func (s SkillSource) String() string {
	switch s {
	case SkillSourceBuiltin:
		return "builtin"
	case SkillSourceUser:
		return "user"
	default:
		return "unknown"
	}
}

// SkillPattern 定义 Skill 的设计模式类型
type SkillPattern string

const (
	PatternToolWrapper SkillPattern = "tool-wrapper" // 工具包装器：让 agent 成为领域专家
	PatternGenerator   SkillPattern = "generator"    // 生成器：从模板生成结构化输出
	PatternReviewer    SkillPattern = "reviewer"     // 审查器：按检查清单评分
	PatternInversion   SkillPattern = "inversion"    // 反转：Agent 先采访用户再行动
	PatternPipeline    SkillPattern = "pipeline"     // 流水线：严格的顺序工作流
)

// SkillTrigger 定义 Skill 的触发条件
type SkillTrigger struct {
	Time     string   `yaml:"time,omitempty"`     // 触发时间段，如 "9:30-11:30"
	Session  string   `yaml:"session,omitempty"`  // Session 标签，如 "morning-session"
	Keywords []string `yaml:"keywords,omitempty"` // 触发关键词
}

// SkillMetadata 扩展的 Skill 元数据
type SkillMetadata struct {
	Pattern      SkillPattern   `yaml:"pattern,omitempty"`       // 设计模式类型
	Triggers     []SkillTrigger `yaml:"triggers,omitempty"`      // 触发条件
	Tools        []string       `yaml:"tools,omitempty"`         // 依赖的工具
	Dependencies []string       `yaml:"dependencies,omitempty"`  // 依赖的其他 Skill
	Steps        int            `yaml:"steps,omitempty"`         // Pipeline 模式的步骤数量
	OutputFormat string         `yaml:"output-format,omitempty"` // Generator 模式的输出格式
	RequiresTodo bool           `yaml:"requires_todo,omitempty"` // 是否需要创建 TODO 列表
}

// Skill 表示一个技能单元
type Skill struct {
	// 公共字段
	Name        string        // Skill 名称（唯一标识）
	Description string        // Skill 描述
	Version     string        // Skill 版本
	Priority    int           // Skill 优先级（0-10，默认 5）
	Source      SkillSource   // Skill 来源
	Metadata    SkillMetadata // 扩展元数据

	// 私有字段
	dirPath    string            // Skill 目录路径
	content    string            // Skill 主内容（懒加载）
	references map[string]string // references/ 目录内容（懒加载）
	assets     map[string]string // assets/ 目录内容（懒加载）
	loaded     bool              // 主内容是否已加载
	mu         sync.RWMutex      // 并发保护
}

// GetContent 懒加载 Skill 主内容
// 使用双重检查锁定模式确保线程安全
func (s *Skill) GetContent() (string, error) {
	// 快速路径：已加载，直接返回
	s.mu.RLock()
	if s.loaded {
		content := s.content
		s.mu.RUnlock()
		return content, nil
	}
	s.mu.RUnlock()

	// 慢速路径：需要加载
	s.mu.Lock()
	defer s.mu.Unlock()

	// 双重检查
	if s.loaded {
		return s.content, nil
	}

	// 从文件加载内容
	skillPath := filepath.Join(s.dirPath, "SKILL.md")
	content, err := loadSkillContent(skillPath)
	if err != nil {
		return "", err
	}

	s.content = content
	s.loaded = true
	return s.content, nil
}

// GetReference 按需加载 references/ 目录下的文件
func (s *Skill) GetReference(name string) (string, error) {
	s.mu.RLock()
	if s.references != nil {
		if content, ok := s.references[name]; ok {
			s.mu.RUnlock()
			return content, nil
		}
	}
	s.mu.RUnlock()

	// 慢速路径：需要加载
	s.mu.Lock()
	defer s.mu.Unlock()

	// 初始化 references map
	if s.references == nil {
		s.references = make(map[string]string)
	}

	// 检查是否已加载
	if content, ok := s.references[name]; ok {
		return content, nil
	}

	// 从文件加载
	refPath := filepath.Join(s.dirPath, "references", name)
	data, err := os.ReadFile(refPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("reference file not found: %s", name)
		}
		return "", fmt.Errorf("failed to read reference file: %w", err)
	}

	content := string(data)
	s.references[name] = content
	log.Debugf("Loaded reference %s for skill %s", name, s.Name)
	return content, nil
}

// GetAsset 按需加载 assets/ 目录下的文件
func (s *Skill) GetAsset(name string) (string, error) {
	s.mu.RLock()
	if s.assets != nil {
		if content, ok := s.assets[name]; ok {
			s.mu.RUnlock()
			return content, nil
		}
	}
	s.mu.RUnlock()

	// 慢速路径：需要加载
	s.mu.Lock()
	defer s.mu.Unlock()

	// 初始化 assets map
	if s.assets == nil {
		s.assets = make(map[string]string)
	}

	// 检查是否已加载
	if content, ok := s.assets[name]; ok {
		return content, nil
	}

	// 从文件加载
	assetPath := filepath.Join(s.dirPath, "assets", name)
	data, err := os.ReadFile(assetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("asset file not found: %s", name)
		}
		return "", fmt.Errorf("failed to read asset file: %w", err)
	}

	content := string(data)
	s.assets[name] = content
	log.Debugf("Loaded asset %s for skill %s", name, s.Name)
	return content, nil
}

// GetDirPath 返回 Skill 目录路径
func (s *Skill) GetDirPath() string {
	return s.dirPath
}

// HasReferences 检查是否有 references 目录
func (s *Skill) HasReferences() bool {
	refDir := filepath.Join(s.dirPath, "references")
	_, err := os.Stat(refDir)
	return err == nil
}

// HasAssets 检查是否有 assets 目录
func (s *Skill) HasAssets() bool {
	assetDir := filepath.Join(s.dirPath, "assets")
	_, err := os.Stat(assetDir)
	return err == nil
}

// HasTodoTemplate 检查是否有 TODO 模板文件
func (s *Skill) HasTodoTemplate() bool {
	todoPath := filepath.Join(s.dirPath, "references", "todo-template.md")
	_, err := os.Stat(todoPath)
	return err == nil
}

// GetTodoTemplate 获取 TODO 模板内容
func (s *Skill) GetTodoTemplate() (string, error) {
	return s.GetReference("todo-template.md")
}

// loadSkillContent 从文件加载 Skill 内容，跳过 YAML frontmatter
func loadSkillContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	content := extractBody(data)
	log.Debugf("Loaded skill content from %s", path)
	return content, nil
}

// extractBody 提取 Markdown body 内容，跳过 YAML frontmatter
// 格式：---\nYAML\n---\nBody
func extractBody(data []byte) string {
	content := string(data)

	// 检查是否以 YAML frontmatter 开头
	if !strings.HasPrefix(content, "---") {
		return content
	}

	// 查找第二个 "---" 标记
	// 从第 4 个字符开始查找（跳过第一个 "---\n"）
	idx := strings.Index(content[4:], "---")
	if idx == -1 {
		// 没有找到结束标记，返回整个内容
		return content
	}

	// 提取第二个 "---" 之后的内容
	// idx + 4 是因为 content[4:] 跳过了前 4 个字符
	// idx + 4 + 3 是因为 "---" 是 3 个字符
	bodyStart := 4 + idx + 3
	if bodyStart >= len(content) {
		return ""
	}

	// 跳过换行符
	body := content[bodyStart:]
	body = strings.TrimLeft(body, "\n\r")

	return body
}
