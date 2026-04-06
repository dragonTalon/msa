package skills

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Loader 扫描和加载 Skills
type Loader struct {
	builtinDir string
	userDir    string
	registry   *Registry
}

// NewLoader 创建一个新的 Loader
func NewLoader(builtinDir, userDir string, registry *Registry) *Loader {
	return &Loader{
		builtinDir: builtinDir,
		userDir:    userDir,
		registry:   registry,
	}
}

// skillMetadataYAML 表示 SKILL.md 的 YAML frontmatter（用于解析）
type skillMetadataYAML struct {
	Name         string         `yaml:"name"`
	Description  string         `yaml:"description"`
	Version      string         `yaml:"version"`
	Priority     int            `yaml:"priority"`
	Pattern      string         `yaml:"pattern"`
	Triggers     []SkillTrigger `yaml:"triggers"`
	Tools        []string       `yaml:"tools"`
	Dependencies []string       `yaml:"dependencies"`
	Steps        int            `yaml:"steps"`
	OutputFormat string         `yaml:"output-format"`
	RequiresTodo bool           `yaml:"requires_todo"`
}

// LoadAll 扫描并加载所有 Skills
func (l *Loader) LoadAll() error {
	// 扫描内置 Skills
	if err := l.scanBuiltinSkills(); err != nil {
		log.Warnf("Failed to scan builtin skills: %v", err)
		// 不返回错误，继续扫描用户 Skills
	}

	// 扫描用户自定义 Skills
	if err := l.scanUserSkills(); err != nil {
		log.Warnf("Failed to scan user skills: %v", err)
		// 不返回错误
	}

	return nil
}

// scanBuiltinSkills 扫描内置 Skills 目录
func (l *Loader) scanBuiltinSkills() error {
	log.Infof("Scanning builtin skills from %s", l.builtinDir)

	entries, err := os.ReadDir(l.builtinDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf("Builtin skills directory does not exist: %s", l.builtinDir)
			return nil
		}
		return fmt.Errorf("failed to read builtin skills directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillDir := filepath.Join(l.builtinDir, entry.Name())
		skillPath := filepath.Join(skillDir, "SKILL.md")
		if err := l.loadSkill(skillDir, skillPath, SkillSourceBuiltin); err != nil {
			log.Warnf("Failed to load builtin skill %s: %v", skillPath, err)
			// 继续处理其他 Skills
		}
	}

	return nil
}

// scanUserSkills 扫描用户自定义 Skills 目录
func (l *Loader) scanUserSkills() error {
	log.Infof("Scanning user skills from %s", l.userDir)

	// 目录不存在时忽略（不报错）
	if _, err := os.Stat(l.userDir); os.IsNotExist(err) {
		log.Infof("User skills directory does not exist: %s", l.userDir)
		return nil
	}

	entries, err := os.ReadDir(l.userDir)
	if err != nil {
		return fmt.Errorf("failed to read user skills directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillDir := filepath.Join(l.userDir, entry.Name())
		skillPath := filepath.Join(skillDir, "SKILL.md")
		if err := l.loadSkill(skillDir, skillPath, SkillSourceUser); err != nil {
			log.Warnf("Failed to load user skill %s: %v", skillPath, err)
			// 继续处理其他 Skills
		}
	}

	return nil
}

// loadSkill 加载单个 Skill
func (l *Loader) loadSkill(skillDir, skillPath string, source SkillSource) error {
	// 解析 Skill 元数据
	metadata, err := l.parseSkillMetadata(skillPath)
	if err != nil {
		return err
	}

	// 创建 Skill
	skill := &Skill{
		Name:        metadata.Name,
		Description: metadata.Description,
		Version:     metadata.Version,
		Priority:    metadata.Priority,
		Source:      source,
		Metadata: SkillMetadata{
			Pattern:      SkillPattern(metadata.Pattern),
			Triggers:     metadata.Triggers,
			Tools:        metadata.Tools,
			Dependencies: metadata.Dependencies,
			Steps:        metadata.Steps,
			OutputFormat: metadata.OutputFormat,
			RequiresTodo: metadata.RequiresTodo,
		},
		dirPath: skillDir,
		loaded:  false,
	}

	// 注册到 Registry
	l.registry.Register(skill)

	// 记录详细信息
	log.Debugf("Loaded skill: %s (pattern: %s, priority: %d, has_refs: %v, has_assets: %v)",
		skill.Name, skill.Metadata.Pattern, skill.Priority, skill.HasReferences(), skill.HasAssets())

	return nil
}

// parseSkillMetadata 解析 SKILL.md 文件的 YAML frontmatter
func (l *Loader) parseSkillMetadata(path string) (*skillMetadataYAML, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open skill file: %w", err)
	}
	defer file.Close()

	// 读取文件
	scanner := bufio.NewScanner(file)
	var frontmatterLines []string
	inFrontmatter := false

	for scanner.Scan() {
		line := scanner.Text()

		// 检测 frontmatter 开始
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				// frontmatter 结束
				break
			}
		}

		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		} else {
			// 没有 frontmatter，直接使用默认值
			break
		}
	}

	if len(frontmatterLines) == 0 {
		return nil, fmt.Errorf("no frontmatter found in skill file")
	}

	// 解析 YAML
	frontmatter := strings.Join(frontmatterLines, "\n")
	var metadata skillMetadataYAML
	if err := yaml.Unmarshal([]byte(frontmatter), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// 验证必需字段
	if metadata.Name == "" {
		return nil, fmt.Errorf("skill name is required")
	}
	if metadata.Description == "" {
		return nil, fmt.Errorf("skill description is required")
	}

	// 使用默认值
	if metadata.Priority == 0 {
		metadata.Priority = 5 // 默认优先级
	}

	log.Debugf("Parsed skill metadata: %s (priority: %d, pattern: %s)",
		metadata.Name, metadata.Priority, metadata.Pattern)

	return &metadata, nil
}
