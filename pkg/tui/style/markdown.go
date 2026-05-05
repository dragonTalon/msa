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
		heading := n.(*goldmarkAst.Heading)
		var style lipgloss.Style
		switch heading.Level {
		case 1:
			style = MDH1Style
		case 2:
			style = MDH2Style
		case 3:
			style = MDH3Style
		case 4:
			style = MDH4Style
		case 5:
			style = MDH5Style
		case 6:
			style = MDH6Style
		default:
			style = MDHeadingStyle
		}
		r.buf.WriteString(style.Render(text))
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
			link := c.(*goldmarkAst.Link)
			url := string(link.Destination)
			buf.WriteString(text)
			if url != "" {
				buf.WriteString(" " + MDLinkURLStyle.Render("("+url+")"))
			}
		case goldmarkAst.KindAutoLink:
			link := c.(*goldmarkAst.AutoLink)
			url := string(link.URL(r.source))
			buf.WriteString(MDLinkURLStyle.Render(url))
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

// findTaskCheckBox 递归查找 TaskCheckBox 节点
func findTaskCheckBox(n goldmarkAst.Node) *extast.TaskCheckBox {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == extast.KindTaskCheckBox {
			return c.(*extast.TaskCheckBox)
		}
		if found := findTaskCheckBox(c); found != nil {
			return found
		}
	}
	return nil
}

// renderList 渲染有序/无序列表
func (r *mdRenderer) renderList(n goldmarkAst.Node) {
	list := n.(*goldmarkAst.List)
	itemIndex := 0
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == goldmarkAst.KindListItem {
			// 检查是否有 TaskCheckBox 子节点（可能嵌套在 TextBlock 下）
			var checkbox string
			if tcb := findTaskCheckBox(c); tcb != nil {
				if tcb.IsChecked {
					checkbox = "[x] "
				} else {
					checkbox = "[ ] "
				}
			}
			text := strings.TrimSpace(r.collectText(c))
			if list.IsOrdered() {
				itemIndex++
				r.buf.WriteString(fmt.Sprintf("  %d. %s%s\n", itemIndex, checkbox, text))
			} else {
				r.buf.WriteString(fmt.Sprintf("  %s %s%s\n", MDListBullet, checkbox, text))
			}
		}
	}
}

// visibleWidth 计算去除 ANSI 转义序列后的可见字符宽度
func visibleWidth(s string) int {
	var buf strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		if s[i] == '\x1b' {
			inEscape = true
			continue
		}
		buf.WriteByte(s[i])
	}
	return len(buf.String())
}

// padToWidth 将字符串填充到指定宽度，支持左/中/右对齐
func padToWidth(s string, width int, align extast.Alignment) string {
	visible := visibleWidth(s)
	padding := width - visible
	if padding <= 0 {
		return s
	}
	switch align {
	case extast.AlignRight:
		return strings.Repeat(" ", padding) + s
	case extast.AlignCenter:
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
	default: // AlignLeft, AlignNone
		return s + strings.Repeat(" ", padding)
	}
}

// tableCellData 存储表格单元格的渲染文本和对齐方式
type tableCellData struct {
	text      string
	alignment extast.Alignment
}

// renderTable 渲染 GFM 表格（两遍扫描：收集数据 → 计算列宽 → 渲染）
func (r *mdRenderer) renderTable(n goldmarkAst.Node) {
	table := n.(*extast.Table)
	alignments := table.Alignments

	// 第一遍：收集所有行数据
	type rowData struct {
		cells    []tableCellData
		isHeader bool
	}
	var rows []rowData

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind() {
		case extast.KindTableHeader:
			var cells []tableCellData
			for cell := c.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if cell.Kind() == extast.KindTableCell {
					tc := cell.(*extast.TableCell)
					text := strings.TrimSpace(r.renderTableCells(cell))
					cells = append(cells, tableCellData{text: text, alignment: tc.Alignment})
				}
			}
			rows = append(rows, rowData{cells: cells, isHeader: true})
		case extast.KindTableRow:
			var cells []tableCellData
			for cell := c.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if cell.Kind() == extast.KindTableCell {
					tc := cell.(*extast.TableCell)
					text := strings.TrimSpace(r.renderTableCells(cell))
					cells = append(cells, tableCellData{text: text, alignment: tc.Alignment})
				}
			}
			rows = append(rows, rowData{cells: cells, isHeader: false})
		}
	}

	if len(rows) == 0 {
		return
	}

	// 计算列数（取最多列的行）
	numCols := len(rows[0].cells)
	for _, row := range rows {
		if len(row.cells) > numCols {
			numCols = len(row.cells)
		}
	}

	// 第二遍：计算每列最大可见宽度
	colWidths := make([]int, numCols)
	for _, row := range rows {
		for i, cell := range row.cells {
			w := visibleWidth(cell.text)
			if w > colWidths[i] {
				colWidths[i] = w
			}
		}
	}

	// 使用 table 级别的对齐信息映射到列
	colAlignments := make([]extast.Alignment, numCols)
	for i := range colAlignments {
		if i < len(alignments) {
			colAlignments[i] = alignments[i]
		} else {
			colAlignments[i] = extast.AlignNone
		}
	}

	// 第三遍：渲染各行（列对齐）
	for _, row := range rows {
		var cells []string
		for i, cell := range row.cells {
			align := cell.alignment
			if align == extast.AlignNone {
				align = colAlignments[i]
			}
			var styled string
			if row.isHeader {
				styled = MDTableHeaderStyle.Render(padToWidth(cell.text, colWidths[i], align))
			} else {
				styled = MDTableCellStyle.Render(padToWidth(cell.text, colWidths[i], align))
			}
			cells = append(cells, styled)
		}
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
		if r := recover(); r != nil {
			log.Warnf("chroma syntax highlight panic for lang=%s: %v", lang, r)
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
