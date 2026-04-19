package tui

import (
	"context"
	"fmt"
	"strings"

	coreagent "msa/pkg/core/agent"
	"msa/pkg/core/event"
	"msa/pkg/core/runner"
	"msa/pkg/config"
	command "msa/pkg/logic/command"
	"msa/pkg/model"
	"msa/pkg/session"
	"msa/pkg/tui/style"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	listStyle "github.com/charmbracelet/lipgloss/list"
	log "github.com/sirupsen/logrus"
)

// contextKey 是 context 的键类型
type contextKey string

const (
	// UI 相关常量
	defaultInputWidth   = 50
	inputPaddingWidth   = 10
	minAPIKeyDisplayLen = 8
	apiKeyPrefixLen     = 4
	apiKeySuffixLen     = 4

	// 提示信息
	placeholderText     = "输入你的理财问题..."
	promptText          = "MSA > "
	welcomeMessage      = "欢迎使用 MSA！输入你的理财问题吧..."
	thinkingMessage     = "⏳ 正在思考..."
	clearSuccessMessage = "对话已清空，重新开始吧！"
	helpMessage         = "📋 可用命令:\n  • clear - 清空对话\n  • /skills - 列出所有可用的 Skills\n  • help/? - 显示帮助\n  • quit/exit - 退出程序"
	helpHint            = "ESC/Ctrl+C: 退出 | Ctrl+K: 清空 | Tab: 命令补全 | Enter: 发送"
)

// Chat TUI聊天模型
type Chat struct {
	textInput            textinput.Model             // 文本输入组件
	history              []model.Message             // 历史消息（只包含 User 和 Assistant）
	pendingMsgs          []model.Message             // 待 flush 的消息
	ctx                  context.Context             // 上下文
	width                int                         // 终端宽度
	height               int                         // 终端高度
	isCommandMode        bool                        // 是否处于命令模式
	commandList          []string                    // 命令列表
	commandSelectedIndex int                         // 命令建议选中索引
	commandSuggestions   []command.CommandSuggestion // 命令建议列表（带描述）
	streamingMsg         string                      // 流式输出的临时内容
	isStreaming          bool                        // 是否正在流式输出
	fullStreamContent    strings.Builder             // 当前 segment 的流式内容
	allStreamContent     strings.Builder             // 完整的 assistant 回复（跨 segment 累积）
	eventCh              chan event.Event             // 事件 channel
	currentSegment       strings.Builder             // 当前消息段内容
	currentSegmentType   model.StreamMsgType         // 当前消息段类型
	streamSegments       []model.Message             // 流式输出的消息段
	sessionMgr           *session.Manager            // 会话管理器
	resumeSession        *session.ParsedSession      // 恢复的会话（可选）
}

// Option Chat 配置选项
type Option func(*Chat)

// WithResumeSession 设置恢复的会话
func WithResumeSession(parsed *session.ParsedSession) Option {
	return func(c *Chat) {
		c.resumeSession = parsed
	}
}

// maskAPIKey 隐藏 APIKey，只显示前4个和后4个字符
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "未设置"
	}
	if len(apiKey) <= minAPIKeyDisplayLen {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:apiKeyPrefixLen] + "****" + apiKey[len(apiKey)-apiKeySuffixLen:]
}

// NewChat 创建新的聊天模型
func NewChat(ctx context.Context, opts ...Option) *Chat {
	// 初始化 Markdown 渲染器
	if err := style.InitMarkdownRenderer(80); err != nil {
		log.Warnf("初始化 Markdown 渲染器失败: %v", err)
	}

	// 初始化文本输入组件
	ti := textinput.New()
	ti.Placeholder = placeholderText
	ti.Focus()
	ti.CharLimit = 0
	ti.Width = defaultInputWidth
	ti.PromptStyle = style.ChatInputPromptStyle
	ti.Prompt = promptText
	ti.TextStyle = style.ChatInputTextStyle

	cfg := config.GetLocalStoreConfig()
	modelName := cfg.Model
	if modelName == "" {
		modelName = "未设置"
	}

	// 获取会话管理器
	sessionMgr := session.GetManager()

	c := &Chat{
		textInput:            ti,
		ctx:                  ctx,
		history:              make([]model.Message, 0),
		commandSelectedIndex: 0,
		commandSuggestions:   make([]command.CommandSuggestion, 0),
		sessionMgr:           sessionMgr,
	}

	// 应用选项（可能在 opts 中传入 resumeSession）
	for _, opt := range opts {
		opt(c)
	}

	// 如果有恢复的会话，加载历史消息
	if c.resumeSession != nil {
		sessionMgr.SetCurrent(c.resumeSession.Session)
		c.history = c.resumeSession.Messages

		// 构建欢迎信息和历史消息
		c.pendingMsgs = []model.Message{
			{Role: model.RoleLogo, Content: style.GetStyledLogo()},
			{Role: model.RoleSystem, Content: fmt.Sprintf("已恢复会话: %s", c.resumeSession.Session.ShortID())},
		}

		// 添加历史消息到 pendingMsgs 以便显示
		for _, msg := range c.resumeSession.Messages {
			c.pendingMsgs = append(c.pendingMsgs, msg)
		}

		// 添加继续对话提示
		c.pendingMsgs = append(c.pendingMsgs, model.Message{
			Role:    model.RoleSystem,
			Content: welcomeMessage,
		})
	} else {
		// 创建新会话
		sess := sessionMgr.NewSession(session.ModeTUI)
		if err := sessionMgr.CreateSessionFile(sess); err != nil {
			log.Warnf("创建会话文件失败: %v", err)
		}
		sessionMgr.SetCurrent(sess)

		c.pendingMsgs = []model.Message{
			{Role: model.RoleLogo, Content: style.GetStyledLogo()},
			{Role: model.RoleSystem, Content: fmt.Sprintf("模型供应商: %s", cfg.Provider)},
			{Role: model.RoleSystem, Content: fmt.Sprintf("模型 : %s", modelName)},
			{Role: model.RoleSystem, Content: fmt.Sprintf("APIKey : %s", maskAPIKey(cfg.APIKey))},
			{Role: model.RoleSystem, Content: fmt.Sprintf("会话ID: %s", sess.ShortID())},
			{Role: model.RoleSystem, Content: welcomeMessage},
		}
	}

	return c
}

// Init 实现 tea.Model 接口
func (c *Chat) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, c.Flush())
}

// Flush 将待输出消息 flush 到终端
func (c *Chat) Flush() tea.Cmd {
	if len(c.pendingMsgs) == 0 {
		return nil
	}

	// 渲染所有待输出消息
	content := c.renderPendingMessages()
	// 清空待输出队列
	c.pendingMsgs = nil

	return tea.Println(content)
}

// renderPendingMessages 渲染待输出的消息
func (c *Chat) renderPendingMessages() string {
	var sb strings.Builder
	var lastMsgType model.StreamMsgType

	for i, msg := range c.pendingMsgs {
		// 检查是否需要在消息类型切换时添加间隔
		if i > 0 && msg.Role == model.RoleAssistant && lastMsgType != "" && lastMsgType != msg.MsgType {
			sb.WriteString("\n")
		}

		switch msg.Role {
		case model.RoleLogo:
			sb.WriteString(msg.Content)
		case model.RoleUser:
			sb.WriteString(style.ChatUserMsgStyle.Render("👤 你: "))
			sb.WriteString(style.ChatNormalMsgStyle.Render(msg.Content))
		case model.RoleSystem:
			sb.WriteString(style.ChatSystemMsgStyle.Render("🔧 系统: "))
			sb.WriteString(style.ChatNormalMsgStyle.Render(msg.Content))
		case model.RoleAssistant:
			sb.WriteString(style.ChatSystemMsgStyle.Render("🤖 MSA: "))
			// 根据消息类型选择样式和渲染方式
			switch msg.MsgType {
			case model.StreamMsgTypeTool:
				// 工具消息 - 黄色，不渲染 Markdown
				sb.WriteString(style.ChatToolMsgStyle.Render(msg.Content))
			case model.StreamMsgTypeReason:
				// 思考消息 - 灰色，不渲染 Markdown
				sb.WriteString(style.ChatReasonMsgStyle.Render(msg.Content))
			case model.StreamMsgTypeText:
				fallthrough
			default:
				// 正文消息 - 渲染 Markdown
				sb.WriteString(style.RenderMarkdown(msg.Content))
			}
			// 记录当前消息类型
			lastMsgType = msg.MsgType
		}
		if i < len(c.pendingMsgs)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// addMessage 添加消息到待 flush 队列
func (c *Chat) addMessage(role model.MessageRole, content string, msgType model.StreamMsgType) {
	c.pendingMsgs = append(c.pendingMsgs, model.Message{
		Role:    role,
		Content: content,
		MsgType: msgType,
	})
}

// Update 实现 tea.Model 接口，处理消息更新
func (c *Chat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var tiCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.textInput.Width = msg.Width - inputPaddingWidth
		// 更新 Markdown 渲染器宽度
		if err := style.InitMarkdownRenderer(msg.Width - 10); err != nil {
			log.Warnf("更新 Markdown 渲染器宽度失败: %v", err)
		}

	case event.Event:
		return c.handleEvent(msg)
	case tea.KeyMsg:
		log.Debugf("捕获按键: %s, Type: %v", msg.String(), msg.Type)
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return c, tea.Quit

		case tea.KeyEnter:
			// 如果正在流式输出，忽略回车键
			if c.isStreaming {
				log.Debug("流式输出中，忽略回车键")
				return c, nil
			}
			return c.handleEnterKey()

		case tea.KeyCtrlK:
			return c.handleClearHistory()

		case tea.KeyTab:
			// Tab 补全命令
			if c.isCommandMode && len(c.commandSuggestions) > 0 {
				selected := c.commandSuggestions[c.commandSelectedIndex]
				c.textInput.SetValue("/" + selected.Name + " ")
				c.textInput.CursorEnd()
				c.isCommandMode = false
				c.commandSuggestions = nil
				return c, nil
			}

		case tea.KeyUp:
			// 在命令模式下向上选择
			if c.isCommandMode && len(c.commandSuggestions) > 0 && c.commandSelectedIndex > 0 {
				c.commandSelectedIndex--
			}

		case tea.KeyDown:
			// 在命令模式下向下选择
			if c.isCommandMode && len(c.commandSuggestions) > 0 && c.commandSelectedIndex < len(c.commandSuggestions)-1 {
				c.commandSelectedIndex++
			}

		default:
			// 如果正在流式输出，不更新输入框
			if !c.isStreaming {
				c.textInput, tiCmd = c.textInput.Update(msg)
				// 检测命令模式
				if strings.HasPrefix(c.textInput.Value(), "/") {
					c.isCommandMode = true
					c.commandSuggestions = command.GetCommandSuggestions(c.textInput.Value())
					c.commandSelectedIndex = 0
					// 保持 commandList 兼容性
					c.commandList = make([]string, len(c.commandSuggestions))
					for i, s := range c.commandSuggestions {
						c.commandList[i] = s.Name
					}
				} else {
					c.isCommandMode = false
					c.commandSuggestions = nil
				}
			}
		}
	}
	return c, tiCmd
}

// View 实现 tea.Model 接口，渲染界面（只渲染输入框和帮助信息）
func (c *Chat) View() string {
	var sb strings.Builder

	// 如果正在流式输出，显示临时内容
	if c.isStreaming {
		sb.WriteString(c.streamingMsg)
		sb.WriteString("\n")
	}

	// 输入区域
	inputBox := lipgloss.NewStyle().
		Padding(0, 1).
		Render(c.textInput.View())
	sb.WriteString(inputBox)

	// 显示命令提示
	if c.isCommandMode && len(c.commandSuggestions) > 0 {
		sb.WriteString("\n")
		sb.WriteString(style.ChatSystemMsgStyle.Render("匹配的命令:"))
		sb.WriteString("\n")
		for i, suggestion := range c.commandSuggestions {
			var line string
			if i == c.commandSelectedIndex {
				// 选中项：高亮显示
				nameStyle := style.CommandSuggestionSelectedStyle
				descStyle := style.CommandSuggestionDescStyle
				line = fmt.Sprintf("  > %s  %s",
					nameStyle.Render("/"+suggestion.Name),
					descStyle.Render(suggestion.Description))
			} else {
				// 普通项
				nameStyle := style.CommandSuggestionNormalStyle
				descStyle := style.CommandSuggestionDescStyle
				line = fmt.Sprintf("    %s  %s",
					nameStyle.Render("/"+suggestion.Name),
					descStyle.Render(suggestion.Description))
			}
			sb.WriteString(line + "\n")
		}
		// 快捷键提示
		sb.WriteString(style.CommandSuggestionHintStyle.Render("Tab: 补全 │ ↑↓: 选择 │ Enter: 执行 │ Esc: 取消"))
	}

	// 帮助提示
	sb.WriteString("\n" + style.ChatHelpStyle.Render(helpHint))

	return sb.String()
}

// handleEvent handles pipeline events from the Runner.
func (c *Chat) handleEvent(e event.Event) (tea.Model, tea.Cmd) {
	switch e.Type {
	case event.EventError:
		log.Errorf("事件错误: %v", e.Err)
		c.clearStreamState()
		c.addMessage(model.RoleSystem, fmt.Sprintf("接收消息失败: %v", e.Err), model.StreamMsgTypeText)
		return c, c.Flush()

	case event.EventRoundDone:
		// Accumulate full reply for history
		if c.fullStreamContent.Len() > 0 && c.currentSegmentType != model.StreamMsgTypeTool {
			if c.allStreamContent.Len() > 0 {
				c.allStreamContent.WriteString("\n")
			}
			c.allStreamContent.WriteString(c.fullStreamContent.String())
		}

		allContent := c.allStreamContent.String()
		latestSegmentContent := c.fullStreamContent.String()
		latestSegmentType := c.currentSegmentType

		c.clearStreamState()

		if allContent != "" {
			c.history = append(c.history, model.Message{
				Role:    model.RoleAssistant,
				Content: allContent,
				MsgType: model.StreamMsgTypeText,
			})
			sess := c.sessionMgr.Current()
			if sess != nil {
				c.sessionMgr.AppendMessage(sess, "assistant", allContent)
			}
		}

		if latestSegmentContent != "" {
			c.addMessage(model.RoleAssistant, latestSegmentContent, latestSegmentType)
		}
		return c, c.Flush()

	case event.EventTextChunk:
		if e.Text == "" {
			return c, c.receiveNextChunk()
		}
		return c.handleStreamContent(e.Text, model.StreamMsgTypeText)

	case event.EventThinking:
		if e.Text == "" {
			return c, c.receiveNextChunk()
		}
		return c.handleStreamContent(e.Text, model.StreamMsgTypeReason)

	case event.EventToolStart:
		toolMsg := fmt.Sprintf("\n正在调用工具: %s, 参数: %s", e.Tool.Name, e.Tool.Input)
		return c.handleStreamContent(toolMsg, model.StreamMsgTypeTool)

	case event.EventToolResult:
		toolMsg := fmt.Sprintf("工具 %s 执行完成\n", e.Result.Name)
		return c.handleStreamContent(toolMsg, model.StreamMsgTypeTool)

	case event.EventToolError:
		toolMsg := fmt.Sprintf("工具 %s 执行失败: %v\n", e.Result.Name, e.Err)
		return c.handleStreamContent(toolMsg, model.StreamMsgTypeTool)

	case event.EventTextDone:
		// Text output ended for this segment — keep streaming state
		return c, c.receiveNextChunk()
	}

	return c, c.receiveNextChunk()
}

// handleStreamContent handles streaming content of a specific type.
// Extracted to reduce duplication in handleEvent.
func (c *Chat) handleStreamContent(content string, msgType model.StreamMsgType) (tea.Model, tea.Cmd) {
	isFirst := c.fullStreamContent.Len() == 0

	// If segment type changed, flush previous segment
	if !isFirst && c.currentSegmentType != "" && c.currentSegmentType != msgType {
		prevContent := c.fullStreamContent.String()
		prevMsgType := c.currentSegmentType

		if prevContent != "" && prevMsgType != model.StreamMsgTypeTool {
			if c.allStreamContent.Len() > 0 {
				c.allStreamContent.WriteString("\n")
			}
			c.allStreamContent.WriteString(prevContent)
		}

		c.fullStreamContent.Reset()
		c.streamingMsg = ""

		if prevContent != "" {
			c.addMessage(model.RoleAssistant, prevContent, prevMsgType)
		}

		isFirst = true
	}

	c.fullStreamContent.WriteString(content)
	c.currentSegmentType = msgType

	var contentStyle lipgloss.Style
	switch msgType {
	case model.StreamMsgTypeTool:
		contentStyle = style.ChatToolMsgStyle
	case model.StreamMsgTypeReason:
		contentStyle = style.ChatReasonMsgStyle
	default:
		contentStyle = style.ChatNormalMsgStyle
	}

	if isFirst {
		c.streamingMsg = style.ChatSystemMsgStyle.Render("🤖 MSA: ") +
			contentStyle.Render(content) + "\n"
	} else {
		c.streamingMsg += contentStyle.Render(content)
	}
	return c, c.receiveNextChunk()
}

// analyzeResult 分析结果
func analyzeResult(result *model.CmdResult) string {
	if result == nil {
		return "结果为空"
	}

	var sb strings.Builder
	sb.WriteString("\n")

	switch result.Type {
	case "message":
		msg, ok := result.Data.(string)
		if !ok {
			return "结果类型错误"
		}
		sb.WriteString(msg)

	case "list":
		list, ok := result.Data.([]string)
		if !ok {
			return "结果类型错误"
		}
		styles := listStyle.New()
		for _, v := range list {
			styles = styles.Item(v)
		}
		sb.WriteString(fmt.Sprintf("%s", styles))

	case "table":
		table, ok := result.Data.(map[string]string)
		if !ok {
			return "结果类型错误"
		}
		sb.WriteString(renderTable(table))

	case "boolean":
		b, ok := result.Data.(bool)
		if !ok {
			return "结果类型错误"
		}
		if b {
			sb.WriteString(style.ChatNormalMsgStyle.Render(result.Msg))
		} else {
			sb.WriteString(style.ChatNormalMsgStyle.Render(result.Error.Error()))
		}
	}

	return sb.String()
}

// renderTable 渲染表格，展示 key-value 数据
func renderTable(data map[string]string) string {
	if len(data) == 0 {
		return "无数据"
	}

	var sb strings.Builder

	// 表头
	sb.WriteString(style.TableHeaderStyle.Render("模型名称"))
	sb.WriteString(style.TableHeaderStyle.Render("描述") + "\n")

	// 表格内容
	rowIndex := 0
	for key, value := range data {
		var rowStyle lipgloss.Style
		if rowIndex%2 == 0 {
			rowStyle = style.TableEvenRowStyle
		} else {
			rowStyle = style.TableOddRowStyle
		}

		sb.WriteString(rowStyle.Render(key))
		sb.WriteString(rowStyle.Render(value) + "\n")
		rowIndex++
	}

	return sb.String()
}

// commandHandler 命令处理器
func (c *Chat) commandHandler(input string) (tea.Model, tea.Cmd) {
	input = strings.TrimPrefix(input, "/")
	split := strings.Split(input, " ")
	cmdName := split[0]

	msaCmd := command.GetCommand(cmdName)
	if msaCmd == nil {
		c.addMessage(model.RoleSystem, "未找到命令: "+input, model.StreamMsgTypeText)
		c.addMessage(model.RoleSystem, fmt.Sprintf("可用命令: %v", command.GetLikeCommand("/")), model.StreamMsgTypeText)
		return c, c.Flush()
	}

	var args []string
	if len(split) > 1 {
		args = split[1:]
	}

	runResult, err := msaCmd.Run(c.ctx, args)
	if err != nil {
		c.addMessage(model.RoleSystem, "执行命令失败: "+err.Error(), model.StreamMsgTypeText)
		log.Errorf("执行命令失败: %v", err)
		return c, c.Flush()
	}

	log.Debugf("执行命令成功: %v", runResult)

	// 如果命令返回的是 selector 类型，则启动选择器
	if runResult.Type == "selector" {
		items, ok := runResult.Data.([]*model.SelectorItem)
		if !ok {
			c.addMessage(model.RoleSystem, "选择器数据类型错误", model.StreamMsgTypeText)
			log.Errorf("选择器数据类型错误")
			return c, c.Flush()
		}

		selector, err := msaCmd.ToSelect(items)
		if err != nil {
			c.addMessage(model.RoleSystem, "创建选择器失败: "+err.Error(), model.StreamMsgTypeText)
			log.Errorf("创建选择器失败: %v", err)
			return c, c.Flush()
		}

		selector.Ctx = c.ctx
		c.textInput.Reset()
		selectorView := NewSelectorView(selector, c)
		return selectorView, nil
	}

	c.addMessage(model.RoleSystem, analyzeResult(runResult), model.StreamMsgTypeText)
	c.textInput.Reset()
	return c, c.Flush()
}

// startStreaming 启动流式输出
func (c *Chat) startStreaming() tea.Cmd {
	c.isStreaming = true
	c.streamingMsg = style.ChatNormalMsgStyle.Render(thinkingMessage)
	c.fullStreamContent.Reset()
	c.allStreamContent.Reset()
	c.currentSegment.Reset()
	c.currentSegmentType = ""
	c.streamSegments = make([]model.Message, 0)
	// 禁用输入框
	c.textInput.Blur()
	log.Debug("流式输出开始，禁用输入框")
	return c.receiveNextChunk()
}

// receiveNextChunk 接收下一个事件
func (c *Chat) receiveNextChunk() tea.Cmd {
	return func() tea.Msg {
		if c.eventCh == nil {
			log.Error("event channel 为空")
			return event.Event{Type: event.EventError, Err: fmt.Errorf("event channel is nil")}
		}
		e, ok := <-c.eventCh
		if !ok {
			log.Debug("event channel 已关闭")
			return event.Event{Type: event.EventRoundDone}
		}
		log.Debugf("接收 event: type=%d", e.Type)
		return e
	}
}

// clearStreamState 清除流式输出状态
func (c *Chat) clearStreamState() {
	c.isStreaming = false
	c.streamingMsg = ""
	c.fullStreamContent.Reset()
	c.allStreamContent.Reset()
	c.currentSegment.Reset()
	c.currentSegmentType = ""
	c.streamSegments = nil

	c.eventCh = nil

	// 重新启用输入框
	c.textInput.Focus()
	log.Debug("流式输出结束，启用输入框")
}

// handleEnterKey 处理回车键事件
func (c *Chat) handleEnterKey() (tea.Model, tea.Cmd) {
	c.isCommandMode = false
	input := strings.TrimSpace(c.textInput.Value())
	if input == "" {
		return c, nil
	}

	// 先将用户消息存入 history（解耦 UI 渲染）
	c.history = append(c.history, model.Message{
		Role:    model.RoleUser,
		Content: input,
		MsgType: model.StreamMsgTypeText,
	})
	c.addMessage(model.RoleUser, input, model.StreamMsgTypeText)

	// 持久化用户消息
	sess := c.sessionMgr.Current()
	if sess != nil {
		c.sessionMgr.AppendMessage(sess, "user", input)
	}

	c.textInput.Reset()

	// 处理命令
	if strings.HasPrefix(input, "/") {
		return c.commandHandler(input)
	}

	// 处理特殊命令
	switch strings.ToLower(input) {
	case "clear":
		return c.handleClearHistory()
	case "help", "?":
		c.addMessage(model.RoleSystem, helpMessage, model.StreamMsgTypeText)
		return c, c.Flush()
	case "quit", "exit":
		return c, tea.Quit
	}

	// 发起聊天请求
	return c.startChatRequest(input)
}

// handleClearHistory 清空对话历史
func (c *Chat) handleClearHistory() (tea.Model, tea.Cmd) {
	c.textInput.Reset()
	c.history = make([]model.Message, 0)

	// 重置会话，开始新会话
	sess := c.sessionMgr.NewSession(session.ModeTUI)
	if err := c.sessionMgr.CreateSessionFile(sess); err != nil {
		log.Warnf("创建新会话文件失败: %v", err)
	}
	c.sessionMgr.SetCurrent(sess)

	c.addMessage(model.RoleSystem, clearSuccessMessage, model.StreamMsgTypeText)
	c.addMessage(model.RoleSystem, fmt.Sprintf("新会话ID: %s", sess.ShortID()), model.StreamMsgTypeText)
	return c, c.Flush()
}

// startChatRequest 发起聊天请求
func (c *Chat) startChatRequest(input string) (tea.Model, tea.Cmd) {
	ag, err := coreagent.New(c.ctx)
	if err != nil {
		log.Errorf("创建 Agent 失败: %v", err)
		c.addMessage(model.RoleSystem, "创建 Agent 失败: "+err.Error(), model.StreamMsgTypeText)
		return c, c.Flush()
	}

	c.eventCh = make(chan event.Event, 64)

	// Create a channel-based renderer that writes directly to eventCh
	chanRenderer := &channelRenderer{ch: c.eventCh}
	r := runner.New(ag, c.sessionMgr, chanRenderer)

	go func() {
		defer close(c.eventCh)
		if err := r.Ask(c.ctx, input, c.history); err != nil {
			log.Errorf("[TUI] Runner.Ask 失败: %v", err)
		}
	}()

	return c, tea.Batch(c.Flush(), c.startStreaming())
}

// channelRenderer is a Renderer that writes events directly to a channel.
// Used by TUI to receive events from Runner without going through program.Send.
type channelRenderer struct {
	ch chan<- event.Event
}

func (r *channelRenderer) Handle(ctx context.Context, e event.Event) error {
	select {
	case r.ch <- e:
	case <-ctx.Done():
	}
	return nil
}
