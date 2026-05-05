package style

import (
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromaFormatter "github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
	log "github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
	goldmarkAst "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

var (
	goldmarkParser goldmark.Markdown
)

// InitMarkdownRenderer 初始化 Markdown 渲染器
func InitMarkdownRenderer(width int) error {
	goldmarkParser = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
	)
	_ = width // 宽度由 lipgloss 样式控制，不再通过 glamour wordwrap
	return nil
}

// mdRenderer 遍历 goldmark AST 并用 lipgloss 渲染
type mdRenderer struct {
	buf    strings.Builder
	source []byte
}

// renderBlock 渲染块级节点
func (r *mdRenderer) renderBlock(n goldmarkAst.Node) {
	switch n.Kind() {
	case goldmarkAst.KindHeading:
		text := r.collectText(n)
		r.buf.WriteString(MDHeadingStyle.Render(text))
		r.buf.WriteString("\n")

	case goldmarkAst.KindParagraph:
		text := r.renderInlines(n)
		if strings.TrimSpace(text) != "" {
			r.buf.WriteString(ChatNormalMsgStyle.Render(text))
			r.buf.WriteString("\n")
		}

	case goldmarkAst.KindFencedCodeBlock:
		node := n.(*goldmarkAst.FencedCodeBlock)
		lines := node.Lines()
		var code strings.Builder
		for i := 0; i < lines.Len(); i++ {
			seg := lines.At(i)
			code.Write(seg.Value(r.source))
		}
		var lang string
		if node.Info != nil {
			lang = string(node.Info.Text(r.source))
		}
		r.buf.WriteString(r.highlightCode(code.String(), lang))
		r.buf.WriteString("\n")

	case goldmarkAst.KindCodeBlock:
		lines := n.Lines()
		var code strings.Builder
		for i := 0; i < lines.Len(); i++ {
			seg := lines.At(i)
			code.Write(seg.Value(r.source))
		}
		for _, line := range strings.Split(strings.TrimRight(code.String(), "\n"), "\n") {
			r.buf.WriteString("  ")
			r.buf.WriteString(MDCodeStyle.Render(line))
			r.buf.WriteString("\n")
		}

	case goldmarkAst.KindList:
		r.renderList(n)

	case goldmarkAst.KindBlockquote:
		text := r.collectText(n)
		text = strings.TrimSpace(text)
		r.buf.WriteString(MDBlockquoteStyle.Render(text))
		r.buf.WriteString("\n")

	case extast.KindTable:
		r.renderTable(n)

	case goldmarkAst.KindThematicBreak:
		r.buf.WriteString(ChatDividerStyle.Render(DividerLine))
		r.buf.WriteString("\n")

	default:
		// 递归处理未知块元素的子节点
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			r.renderBlock(c)
		}
	}
}

// renderInlines 渲染段落内的内联元素
func (r *mdRenderer) renderInlines(n goldmarkAst.Node) string {
	var buf strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind() {
		case goldmarkAst.KindText, goldmarkAst.KindString:
			buf.WriteString(string(c.Text(r.source)))
		case goldmarkAst.KindEmphasis:
			e := c.(*goldmarkAst.Emphasis)
			text := r.collectText(c)
			if e.Level == 2 {
				buf.WriteString(MDBoldStyle.Render(text))
			} else {
				buf.WriteString(MDItalicStyle.Render(text))
			}
		case goldmarkAst.KindCodeSpan:
			text := r.collectText(c)
			buf.WriteString(MDCodeStyle.Render(strings.TrimSpace(text)))
		case goldmarkAst.KindLink:
			text := r.collectText(c)
			buf.WriteString(text)
		case extast.KindStrikethrough:
			text := r.collectText(c)
			buf.WriteString(MDStrikethroughStyle.Render(text))
		default:
			buf.WriteString(r.collectText(c))
		}
	}
	return buf.String()
}

// collectText 递归收集节点内所有文本内容
func (r *mdRenderer) collectText(n goldmarkAst.Node) string {
	var buf strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind() {
		case goldmarkAst.KindText, goldmarkAst.KindString:
			buf.WriteString(string(c.Text(r.source)))
		default:
			buf.WriteString(r.collectText(c))
		}
	}
	return buf.String()
}

// renderList 渲染有序/无序列表
func (r *mdRenderer) renderList(n goldmarkAst.Node) {
	list := n.(*goldmarkAst.List)
	itemIndex := 0
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == goldmarkAst.KindListItem {
			text := strings.TrimSpace(r.collectText(c))
			if list.IsOrdered() {
				itemIndex++
				r.buf.WriteString(fmt.Sprintf("  %d. %s\n", itemIndex, text))
			} else {
				r.buf.WriteString(fmt.Sprintf("  %s %s\n", MDListBullet, text))
			}
		}
	}
}

// renderTable 渲染 GFM 表格
func (r *mdRenderer) renderTable(n goldmarkAst.Node) {
	table := n.(*extast.Table)
	alignments := table.Alignments

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind() {
		case extast.KindTableHeader:
			r.renderTableHeader(c, alignments)
		case extast.KindTableRow:
			r.renderTableRow(c, alignments)
		}
	}
}

// renderTableHeader 渲染表格表头行
// TableHeader 直接包含 TableCell 子节点（见 goldmark extension/ast/table.go NewTableHeader）
func (r *mdRenderer) renderTableHeader(n goldmarkAst.Node, alignments []extast.Alignment) {
	var cells []string
	for cell := n.FirstChild(); cell != nil; cell = cell.NextSibling() {
		if cell.Kind() == extast.KindTableCell {
			text := strings.TrimSpace(r.collectText(cell))
			cells = append(cells, MDTableHeaderStyle.Render(text))
		}
	}
	if len(cells) > 0 {
		r.buf.WriteString(" " + strings.Join(cells, " "+MDTableSeparator+" "))
		r.buf.WriteString("\n")
	}
}

// renderTableRow 渲染表格数据行
func (r *mdRenderer) renderTableRow(n goldmarkAst.Node, alignments []extast.Alignment) {
	var cells []string
	for cell := n.FirstChild(); cell != nil; cell = cell.NextSibling() {
		if cell.Kind() == extast.KindTableCell {
			text := r.renderTableCells(cell)
			cells = append(cells, MDTableCellStyle.Render(text))
		}
	}
	if len(cells) > 0 {
		r.buf.WriteString(" " + strings.Join(cells, " "+MDTableSeparator+" "))
		r.buf.WriteString("\n")
	}
}

// renderTableCells 递归渲染表格单元格内的内联元素
func (r *mdRenderer) renderTableCells(n goldmarkAst.Node) string {
	var buf strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind() {
		case goldmarkAst.KindText, goldmarkAst.KindString:
			buf.WriteString(string(c.Text(r.source)))
		case goldmarkAst.KindEmphasis:
			e := c.(*goldmarkAst.Emphasis)
			text := r.collectText(c)
			if e.Level == 2 {
				buf.WriteString(MDBoldStyle.Render(text))
			} else {
				buf.WriteString(MDItalicStyle.Render(text))
			}
		case goldmarkAst.KindCodeSpan:
			text := r.collectText(c)
			buf.WriteString(MDCodeStyle.Render(strings.TrimSpace(text)))
		case extast.KindStrikethrough:
			text := r.collectText(c)
			buf.WriteString(MDStrikethroughStyle.Render(strings.TrimSpace(text)))
		default:
			buf.WriteString(r.collectText(c))
		}
	}
	return strings.TrimSpace(buf.String())
}

// highlightCode 使用 chroma 高亮代码
func (r *mdRenderer) highlightCode(code, lang string) string {
	if lang == "" {
		lang = "text"
	}

	// 恢复可能的 panic
	defer func() {
		if recover() != nil {
			// chroma 失败时已在调用方回退，此处静默
		}
	}()

	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		// 回退：纯文本代码块
		return r.plainCodeBlock(code)
	}

	formatter := chromaFormatter.Get("terminal256")
	if formatter == nil {
		formatter = chromaFormatter.Fallback
	}

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	var buf strings.Builder
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return r.plainCodeBlock(code)
	}
	return strings.TrimRight(buf.String(), "\n")
}

// plainCodeBlock 纯文本代码块（无语法高亮回退）
func (r *mdRenderer) plainCodeBlock(code string) string {
	var buf strings.Builder
	for _, line := range strings.Split(code, "\n") {
		buf.WriteString("  ")
		buf.WriteString(MDCodeStyle.Render(line))
		buf.WriteString("\n")
	}
	return strings.TrimRight(buf.String(), "\n")
}

// RenderMarkdown 渲染 Markdown 文本
func RenderMarkdown(content string) string {
	if goldmarkParser == nil {
		if err := InitMarkdownRenderer(80); err != nil {
			log.Warnf("goldmark parser 初始化失败，降级为原始文本: %v", err)
			return content
		}
	}

	if strings.TrimSpace(content) == "" {
		return ""
	}

	doc := goldmarkParser.Parser().Parse(text.NewReader([]byte(content)))

	r := &mdRenderer{source: []byte(content)}

	// 遍历文档的子节点（块级元素）
	for c := doc.FirstChild(); c != nil; c = c.NextSibling() {
		r.renderBlock(c)
	}

	return strings.TrimRight(r.buf.String(), "\n")
}

// RenderMarkdownWithStyle 渲染 Markdown 并应用外层样式
func RenderMarkdownWithStyle(content string, style lipgloss.Style) string {
	rendered := RenderMarkdown(content)
	return style.Render(rendered)
}
