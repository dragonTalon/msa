package internal

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Extractor 内容提取器接口
type Extractor interface {
	Extract(html string) (title, content string, err error)
}

// BasicExtractor 基础内容提取器
type BasicExtractor struct {
	maxLength int
}

// NewBasicExtractor 创建基础提取器
func NewBasicExtractor() *BasicExtractor {
	return &BasicExtractor{
		maxLength: 5000, // 默认最大长度
	}
}

// Extract 提取 HTML 的标题和内容
func (e *BasicExtractor) Extract(htmlStr string) (title, content string, err error) {
	// 解析 HTML
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return "", "", fmt.Errorf("解析 HTML 失败: %w", err)
	}

	// 提取标题
	title = e.extractTitle(doc)

	// 清理 HTML 并提取内容
	cleanedDoc := e.cleanHTML(doc)
	content = e.extractContent(cleanedDoc)

	// 清理空白字符
	content = e.cleanWhitespace(content)

	// 应用长度限制
	if len(content) > e.maxLength {
		content = content[:e.maxLength]
	}

	return title, content, nil
}

// extractTitle 提取页面标题
func (e *BasicExtractor) extractTitle(doc *html.Node) string {
	var title string
	var findTitle func(*html.Node)
	findTitle = func(n *html.Node) {
		if n.Type == html.ElementNode && n.DataAtom == atom.Title {
			if n.FirstChild != nil {
				title = n.FirstChild.Data
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTitle(c)
			if title != "" {
				return
			}
		}
	}
	findTitle(doc)
	return title
}

// cleanHTML 清理 HTML，移除不需要的元素
func (e *BasicExtractor) cleanHTML(doc *html.Node) *html.Node {
	// 定义需要移除的标签
	removeTags := map[atom.Atom]bool{
		atom.Script:   true,
		atom.Style:    true,
		atom.Meta:     true,
		atom.Link:     true,
		atom.Noscript: true,
	}

	// 定义需要移除的元素（保留内容但移除标签本身）
	removeElements := map[string]bool{
		"nav":    true,
		"header": true,
		"footer": true,
		"aside":  true,
		"iframe": true,
	}

	var cleanFunc func(*html.Node) *html.Node
	cleanFunc = func(n *html.Node) *html.Node {
		if n == nil {
			return nil
		}

		// 移除脚本和样式标签（包括内容）
		if removeTags[n.DataAtom] {
			return nil
		}

		// 移除导航、页眉、页脚等元素
		if n.Type == html.ElementNode && removeElements[strings.ToLower(n.Data)] {
			return nil
		}

		// 克隆节点
		newNode := &html.Node{
			Type:      n.Type,
			Data:      n.Data,
			DataAtom:  n.DataAtom,
			Namespace: n.Namespace,
		}

		// 递归处理子节点
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			cleanedChild := cleanFunc(c)
			if cleanedChild != nil {
				newNode.AppendChild(cleanedChild)
			}
		}

		return newNode
	}

	return cleanFunc(doc)
}

// extractContent 提取正文内容
func (e *BasicExtractor) extractContent(doc *html.Node) string {
	var content strings.Builder
	var extractFunc func(*html.Node)

	extractFunc = func(n *html.Node) {
		if n == nil {
			return
		}

		switch n.Type {
		case html.TextNode:
			text := strings.TrimSpace(n.Data)
			if text != "" {
				content.WriteString(text)
				content.WriteString(" ")
			}

		case html.ElementNode:
			// 块级元素后添加换行
			isBlock := n.DataAtom == atom.P ||
				n.DataAtom == atom.Div ||
				n.DataAtom == atom.Br ||
				n.DataAtom == atom.H1 ||
				n.DataAtom == atom.H2 ||
				n.DataAtom == atom.H3 ||
				n.DataAtom == atom.H4 ||
				n.DataAtom == atom.H5 ||
				n.DataAtom == atom.H6 ||
				n.DataAtom == atom.Li ||
				n.DataAtom == atom.Tr ||
				n.DataAtom == atom.Th ||
				n.DataAtom == atom.Td

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				extractFunc(c)
			}

			if isBlock {
				content.WriteString("\n\n")
			}
		}
	}

	extractFunc(doc)
	return content.String()
}

// cleanWhitespace 清理多余的空白字符
func (e *BasicExtractor) cleanWhitespace(s string) string {
	// 替换多个连续空格为单个空格
	s = strings.Join(strings.Fields(s), " ")
	// 替换多个连续换行为双换行
	lines := strings.Split(s, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, "\n\n")
}

// extractPlainText 从 HTML 节点提取纯文本（辅助方法）
func (e *BasicExtractor) extractPlainText(n *html.Node) string {
	var buf strings.Builder
	var f func(*html.Node)

	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(n)
	return buf.String()
}

// renderNode 将 HTML 节点渲染为字符串（辅助方法）
func (e *BasicExtractor) renderNode(n *html.Node) string {
	var buf strings.Builder
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}
