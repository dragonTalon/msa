package knowledge

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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

// summaryTableFieldMap 表格字段名到 SummaryData 字段的映射
var summaryTableFieldMap = map[string]string{
	"初始资金":  "prev_asset",
	"昨日资产":  "prev_asset",
	"当前总资产": "curr_asset",
	"今日资产":  "curr_asset",
	"总盈亏":   "pnl",
	"今日盈亏":  "pnl",
	"总收益率":  "pnl_rate",
	"收益率":   "pnl_rate",
}

// parseTableFields 从 Markdown 表格格式中提取字段值
// 在 YAML frontmatter 和 extractField 都失败后作为 fallback
func parseTableFields(content string, data *SummaryData) {
	// 查找表格分隔行（|---| 或 |------|）来确认存在表格
	separatorRe := regexp.MustCompile(`(?m)^\|[-:]+\|[-:|\s]+\|`)
	if !separatorRe.MatchString(content) {
		return
	}

	// 逐行匹配表格行：| 字段名 | 数值 |
	tableRowRe := regexp.MustCompile(`^\|\s*(.+?)\s*\|\s*(.+?)\s*\|`)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		matches := tableRowRe.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}

		fieldName := strings.TrimSpace(matches[1])
		fieldValue := strings.TrimSpace(matches[2])

		// 跳过表头行（"项目"、"------" 等）
		if fieldName == "项目" || strings.Contains(fieldName, "---") {
			continue
		}

		// 查找字段名映射
		dataField, exists := summaryTableFieldMap[fieldName]
		if !exists {
			continue
		}

		// 提取数值：去除 Markdown 格式标记（**bold**、单位文字等）
		numStr := extractNumberFromCellValue(fieldValue)

		switch dataField {
		case "prev_asset":
			fmt.Sscanf(numStr, "%f", &data.PrevAsset)
		case "curr_asset":
			fmt.Sscanf(numStr, "%f", &data.CurrAsset)
		case "pnl":
			fmt.Sscanf(numStr, "%f", &data.PNL)
		case "pnl_rate":
			rateStr := strings.TrimSuffix(numStr, "%")
			fmt.Sscanf(rateStr, "%f", &data.PNLRate)
		}
	}
}

// extractNumberFromCellValue 从表格单元格值中提取数值
func extractNumberFromCellValue(value string) string {
	// 去除所有 Markdown 标记符
	cleaned := strings.NewReplacer(
		"**", "",
		"*", "",
		"_", "",
		"`", "",
	).Replace(value)

	// 提取数字部分（含负号、小数点、逗号）
	re := regexp.MustCompile(`-?[\d,]+\.?\d*`)
	matches := re.FindStringSubmatch(cleaned)
	if len(matches) > 0 {
		return strings.ReplaceAll(matches[0], ",", "")
	}
	return ""
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
	hasBoldFieldData := false
	if prevAsset := extractField(content, "昨日资产"); prevAsset != "" {
		fmt.Sscanf(prevAsset, "%f", &data.PrevAsset)
		hasBoldFieldData = true
	}
	if currAsset := extractField(content, "今日资产"); currAsset != "" {
		fmt.Sscanf(currAsset, "%f", &data.CurrAsset)
		hasBoldFieldData = true
	}
	if pnl := extractField(content, "今日盈亏"); pnl != "" {
		fmt.Sscanf(pnl, "%f", &data.PNL)
		hasBoldFieldData = true
	}
	if pnlRate := extractField(content, "收益率"); pnlRate != "" {
		// 移除百分号
		rate := strings.TrimSuffix(pnlRate, "%")
		fmt.Sscanf(rate, "%f", &data.PNLRate)
		hasBoldFieldData = true
	}

	// 如果 **字段**: 值 格式未提取到数据，尝试表格格式
	if !hasBoldFieldData {
		log.Debugf("Bold field extraction found no data, trying table format fallback")
		parseTableFields(content, data)
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
