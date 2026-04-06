package knowledge

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ErrorEntry 错误记录条目
type ErrorEntry struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Stock   string `json:"stock,omitempty"`
	Error   string `json:"error"`
	Reason  string `json:"reason"`
	Lesson  string `json:"lesson"`
	RawText string `json:"raw_text,omitempty"` // 原始文本，供 Agent 参考
}

// SummaryData 总结数据
type SummaryData struct {
	Date      string  `json:"date"`
	PrevAsset float64 `json:"prev_asset"`
	CurrAsset float64 `json:"curr_asset"`
	PNL       float64 `json:"pnl"`
	PNLRate   float64 `json:"pnl_rate"`
	Raw       string  `json:"raw"` // 原始内容，供 Agent 参考
}

// ParseErrorsFile 解析错误记录文件
func ParseErrorsFile(path string) ([]ErrorEntry, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在时返回空列表，不报错
			return []ErrorEntry{}, nil
		}
		return nil, err
	}

	return parseErrorsContent(string(content)), nil
}

// parseErrorsContent 解析错误记录内容
func parseErrorsContent(content string) []ErrorEntry {
	var entries []ErrorEntry

	// 匹配错误条目格式：
	// ## YYYY-MM-DD 标题
	// 或
	// ### YYYY-MM-DD 标题
	entryRegex := regexp.MustCompile(`(?m)^##\s+(\d{4}-\d{2}-\d{2})\s+(.+)$`)

	// 查找所有条目起始位置
	matches := entryRegex.FindAllStringSubmatchIndex(content, -1)

	for i, match := range matches {
		entry := ErrorEntry{
			Date: content[match[2]:match[3]],
		}

		// 提取标题
		entry.Title = strings.TrimSpace(content[match[4]:match[5]])

		// 提取条目内容（到下一个条目或文件末尾）
		var entryContent string
		if i+1 < len(matches) {
			entryContent = content[match[0]:matches[i+1][0]]
		} else {
			entryContent = content[match[0]:]
		}

		entry.RawText = entryContent

		// 解析字段
		entry.Stock = extractField(entryContent, "股票")
		entry.Error = extractField(entryContent, "错误")
		entry.Reason = extractField(entryContent, "原因")
		entry.Lesson = extractField(entryContent, "教训")

		entries = append(entries, entry)
	}

	return entries
}

// extractField 从内容中提取指定字段
func extractField(content, fieldName string) string {
	// 匹配格式：**字段名**: 值 或 **字段名**：值
	pattern := fmt.Sprintf(`\*\*%s\*\*[：:]\s*(.+?)(?:\n|$)`, fieldName)
	regex := regexp.MustCompile(pattern)
	match := regex.FindStringSubmatch(content)
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

// FindPrevSummaryFile 查找最近一个交易日的总结文件
func FindPrevSummaryFile() (string, error) {
	summariesDir, err := GetSummariesDir()
	if err != nil {
		return "", err
	}

	// 检查目录是否存在
	if _, err := os.Stat(summariesDir); os.IsNotExist(err) {
		return "", nil
	}

	// 读取目录下的所有文件
	entries, err := os.ReadDir(summariesDir)
	if err != nil {
		return "", err
	}

	if len(entries) == 0 {
		return "", nil
	}

	// 查找今天之前的最近文件
	today := time.Now().Format("2006-01-02")
	var prevFiles []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// 检查文件名格式是否为 YYYY-MM-DD.md
		if !isValidDate(strings.TrimSuffix(name, ".md")) {
			continue
		}

		// 排除今天的文件
		if strings.TrimSuffix(name, ".md") == today {
			continue
		}

		prevFiles = append(prevFiles, name)
	}

	if len(prevFiles) == 0 {
		return "", nil
	}

	// 按日期排序，返回最近的
	sortFilesByDate(prevFiles)

	return filepath.Join(summariesDir, prevFiles[0]), nil
}

// sortFilesByDate 按日期降序排序文件名
func sortFilesByDate(files []string) {
	// 简单冒泡排序，按日期降序
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i] < files[j] {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
}

// ParseSummaryFile 解析总结文件
func ParseSummaryFile(path string) (*SummaryData, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data := &SummaryData{
		Raw: string(content),
	}

	// 从文件名提取日期
	filename := filepath.Base(path)
	data.Date = strings.TrimSuffix(filename, ".md")

	// 解析 YAML frontmatter 或 markdown 内容中的数值
	parseSummaryFields(string(content), data)

	return data, nil
}

// parseSummaryFields 解析总结文件中的字段
func parseSummaryFields(content string, data *SummaryData) {
	// 尝试解析 YAML frontmatter
	if strings.HasPrefix(content, "---") {
		endIndex := strings.Index(content[3:], "---")
		if endIndex > 0 {
			frontmatter := content[4 : endIndex+3]
			parseYAMLFrontmatter(frontmatter, data)
			return
		}
	}

	// 如果没有 frontmatter，尝试从内容中提取
	// 匹配格式：**昨日资产**: 12345.67
	if prevAsset := extractField(content, "昨日资产"); prevAsset != "" {
		fmt.Sscanf(prevAsset, "%f", &data.PrevAsset)
	}
	if currAsset := extractField(content, "今日资产"); currAsset != "" {
		fmt.Sscanf(currAsset, "%f", &data.CurrAsset)
	}
	if pnl := extractField(content, "今日盈亏"); pnl != "" {
		fmt.Sscanf(pnl, "%f", &data.PNL)
	}
	if pnlRate := extractField(content, "收益率"); pnlRate != "" {
		// 移除百分号
		rate := strings.TrimSuffix(pnlRate, "%")
		fmt.Sscanf(rate, "%f", &data.PNLRate)
	}
}

// parseYAMLFrontmatter 解析 YAML frontmatter
func parseYAMLFrontmatter(frontmatter string, data *SummaryData) {
	scanner := bufio.NewScanner(strings.NewReader(frontmatter))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "prev_asset":
			fmt.Sscanf(value, "%f", &data.PrevAsset)
		case "curr_asset":
			fmt.Sscanf(value, "%f", &data.CurrAsset)
		case "pnl":
			fmt.Sscanf(value, "%f", &data.PNL)
		case "pnl_rate":
			fmt.Sscanf(value, "%f", &data.PNLRate)
		}
	}
}

// isValidDate 验证日期格式是否为 YYYY-MM-DD
func isValidDate(date string) bool {
	if len(date) != 10 {
		return false
	}

	_, err := time.Parse("2006-01-02", date)
	return err == nil
}
