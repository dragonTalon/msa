package knowledge

import (
	"bufio"
	"bytes"
	"strings"

	"gopkg.in/yaml.v3"
)

const frontmatterDelimiter = "---"

// ParseFrontmatter 解析 Markdown 文件中的 YAML frontmatter
// 返回元数据和正文内容
func ParseFrontmatter(content string) (map[string]interface{}, string) {
	metadata := make(map[string]interface{})

	// 检查是否以 --- 开头
	if !strings.HasPrefix(content, frontmatterDelimiter+"\n") &&
		!strings.HasPrefix(content, frontmatterDelimiter+"\r\n") {
		// 无 frontmatter，返回空元数据和原始内容
		return metadata, content
	}

	// 查找结束的 ---
	lines := strings.SplitN(content, "\n", -1)
	if len(lines) < 2 {
		return metadata, content
	}

	// 找到结束的分隔符
	endIndex := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == frontmatterDelimiter {
			endIndex = i
			break
		}
	}

	if endIndex == -1 {
		// 未找到结束分隔符，返回原始内容
		return metadata, content
	}

	// 提取 YAML 内容
	yamlContent := strings.Join(lines[1:endIndex], "\n")

	// 解析 YAML
	if err := yaml.Unmarshal([]byte(yamlContent), &metadata); err != nil {
		// 解析失败，返回空元数据和原始内容（容错处理）
		return metadata, content
	}

	// 提取正文（跳过 frontmatter）
	bodyStart := endIndex + 1
	if bodyStart >= len(lines) {
		return metadata, ""
	}

	// 跳过开头的空行
	for bodyStart < len(lines) && lines[bodyStart] == "" {
		bodyStart++
	}

	body := strings.Join(lines[bodyStart:], "\n")
	return metadata, body
}

// GenerateFrontmatter 从元数据生成 YAML frontmatter
func GenerateFrontmatter(metadata map[string]interface{}) (string, error) {
	if len(metadata) == 0 {
		return "", nil
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(metadata); err != nil {
		return "", err
	}

	if err := encoder.Close(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GenerateFileContent 生成带 frontmatter 的完整文件内容
func GenerateFileContent(metadata map[string]interface{}, body string) (string, error) {
	if len(metadata) == 0 {
		return body, nil
	}

	frontmatter, err := GenerateFrontmatter(metadata)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(frontmatterDelimiter)
	sb.WriteString("\n")
	sb.WriteString(frontmatter)
	if !strings.HasSuffix(frontmatter, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString(frontmatterDelimiter)
	sb.WriteString("\n")
	if body != "" {
		sb.WriteString(body)
	}

	return sb.String(), nil
}

// ParseFileWithFrontmatter 从字节数组解析文件内容
func ParseFileWithFrontmatter(data []byte) (map[string]interface{}, string) {
	return ParseFrontmatter(string(data))
}

// IsEmptyYAML 检查 YAML 是否为空
func IsEmptyYAML(metadata map[string]interface{}) bool {
	return len(metadata) == 0
}

// ReadFrontmatterLine 逐行读取 frontmatter（用于大文件）
func ReadFrontmatterLine(scanner *bufio.Scanner) (map[string]interface{}, string, error) {
	metadata := make(map[string]interface{})
	var yamlLines []string
	var bodyLines []string

	state := "start" // start, yaml, body

	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case "start":
			if line == frontmatterDelimiter {
				state = "yaml"
			} else {
				// 无 frontmatter
				bodyLines = append(bodyLines, line)
				state = "body"
			}
		case "yaml":
			if line == frontmatterDelimiter {
				state = "body"
			} else {
				yamlLines = append(yamlLines, line)
			}
		case "body":
			bodyLines = append(bodyLines, line)
		}
	}

	// 解析 YAML
	if len(yamlLines) > 0 {
		yamlContent := strings.Join(yamlLines, "\n")
		if err := yaml.Unmarshal([]byte(yamlContent), &metadata); err != nil {
			return metadata, strings.Join(bodyLines, "\n"), err
		}
	}

	return metadata, strings.Join(bodyLines, "\n"), scanner.Err()
}
