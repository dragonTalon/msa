package skills

import (
	"os"
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

// Skill 表示一个技能单元
type Skill struct {
	// 公共字段
	Name        string      // Skill 名称（唯一标识）
	Description string      // Skill 描述
	Version     string      // Skill 版本
	Priority    int         // Skill 优先级（0-10，默认 5）
	Source      SkillSource // Skill 来源

	// 私有字段
	path    string       // SKILL.md 文件路径
	content string       // Skill 内容（懒加载）
	loaded  bool         // 内容是否已加载
	mu      sync.RWMutex // 并发保护
}

// GetContent 懒加载 Skill 内容
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
	content, err := loadSkillContent(s.path)
	if err != nil {
		return "", err
	}

	s.content = content
	s.loaded = true
	return s.content, nil
}

// GetPath 返回 Skill 文件路径
func (s *Skill) GetPath() string {
	return s.path
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
