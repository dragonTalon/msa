package tui

import (
	"context"
	"fmt"
	"strings"

	"msa/pkg/config"
	"msa/pkg/logic/agent"
	command "msa/pkg/logic/command"
	"msa/pkg/logic/message"
	"msa/pkg/model"
	"msa/pkg/tui/style"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	listStyle "github.com/charmbracelet/lipgloss/list"
	log "github.com/sirupsen/logrus"
)

// contextKey 是 context 的键类型
type contextKey string

// manualSkillsKey 是手动指定 skills 的 context 键（与 cmd/root.go 中定义的相同）
const manualSkillsKey contextKey = "manualSkills"

const (
	// UI 相关常量
	defaultInputWidth   = 50
	inputPaddingWidth   = 10
	minAPIKeyDisplayLen = 8
	apiKeyPrefixLen     = 4
	apiKeySuffixLen     = 4

	// 流式输出相关常量
	streamBufferSize = 100

	// 提示信息
	placeholderText     = "输入你的理财问题..."
	promptText          = "MSA > "
	welcomeMessage      = "欢迎使用 MSA！输入你的理财问题吧..."
	thinkingMessage     = "⏳ 正在思考..."
	clearSuccessMessage = "对话已清空，重新开始吧！"
	helpMessage         = "📋 可用命令:\n  • clear - 清空对话\n  • /skills - 列出所有可用的 Skills\n  • /skills: <name1>,<name2> - 手动指定 skills\n  • help/? - 显示帮助\n  • quit/exit - 退出程序"
	helpHint            = "ESC/Ctrl+C: 退出 | Ctrl+K: 清空 | /skills 查看列表 | Enter: 发送"
)

// Chat TUI聊天模型
type Chat struct {
	textInput          textinput.Model           // 文本输入组件
	history            []model.Message           // 历史消息
	pendingMsgs        []model.Message           // 待 flush 的消息
	ctx                context.Context           // 上下文
	width              int                       // 终端宽度
	height             int                       // 终端高度
	isCommandMode      bool                      // 是否处于命令模式
	commandList        []string                  // 命令列表
	streamingMsg       string                    // 流式输出的临时内容
	isStreaming        bool                      // 是否正在流式输出
	fullStreamContent  strings.Builder           // 完整的流式内容
	streamOutputCh     <-chan *model.StreamChunk // 流式输出 channel
	streamUnregister   func()                    // 取消订阅函数
	currentSegment     strings.Builder           // 当前消息段内容
	currentSegmentType model.StreamMsgType       // 当前消息段类型
	streamSegments     []model.Message           // 流式输出的消息段
	sessionSkills      []string                  // 会话级别的手动 skills（来自 CLI --skills）
	oneTimeSkills      []string                  // 一次性 skills（来自 /skills: 命令，用后即焚）
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
func NewChat(ctx context.Context) *Chat {
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

	// 从 context 中提取手动指定的 skills
	var manualSkills []string
	if skills, ok := ctx.Value(manualSkillsKey).([]string); ok && len(skills) > 0 {
		manualSkills = skills
		log.Infof("从 CLI 参数获取手动 skills: %v", manualSkills)
	}

	pendingMsgs := []model.Message{
		{Role: model.RoleLogo, Content: style.GetStyledLogo()},
		{Role: model.RoleSystem, Content: fmt.Sprintf("模型供应商: %s", cfg.Provider)},
		{Role: model.RoleSystem, Content: fmt.Sprintf("模型 : %s", modelName)},
		{Role: model.RoleSystem, Content: fmt.Sprintf("APIKey : %s", maskAPIKey(cfg.APIKey))},
		{Role: model.RoleSystem, Content: welcomeMessage},
	}

	// 如果有手动指定的 skills，显示提示
	if len(manualSkills) > 0 {
		pendingMsgs = append(pendingMsgs, model.Message{
			Role:    model.RoleSystem,
			Content: fmt.Sprintf("✓ 已手动指定 skills: %v", manualSkills),
			MsgType: model.StreamMsgTypeText,
		})
		pendingMsgs = append(pendingMsgs, model.Message{
			Role:    model.RoleSystem,
			Content: "提示: 在此会话中，所有对话都将使用指定的 skills",
			MsgType: model.StreamMsgTypeText,
		})
	}

	return &Chat{
		textInput:     ti,
		pendingMsgs:   pendingMsgs,
		ctx:           ctx,
		history:       make([]model.Message, 0),
		sessionSkills: manualSkills,
		oneTimeSkills: nil,
	}
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
			c.history = append(c.history, msg)
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

	case *model.StreamChunk:
		return c.chatStreamMsg(msg)
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

		default:
			// 如果正在流式输出，不更新输入框
			if !c.isStreaming {
				c.textInput, tiCmd = c.textInput.Update(msg)
				// 检测命令模式
				if strings.HasPrefix(c.textInput.Value(), "/") {
					c.isCommandMode = true
					c.commandList = command.GetLikeCommand(c.textInput.Value())
				} else {
					c.isCommandMode = false
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
	if c.isCommandMode && len(c.commandList) > 0 {
		styles := listStyle.New()
		for _, cmdStr := range c.commandList {
			styles.Item("/" + cmdStr)
		}
		sb.WriteString("\n" + fmt.Sprintf("%s", styles))
	}

	// 帮助提示
	sb.WriteString("\n" + style.ChatHelpStyle.Render(helpHint))

	return sb.String()
}

// chatStreamMsg 处理流式消息
func (c *Chat) chatStreamMsg(msg *model.StreamChunk) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		log.Errorf("流式消息错误: %v", msg.Err)
		c.clearStreamState()
		c.addMessage(model.RoleSystem, fmt.Sprintf("接收消息失败: %v", msg.Err), model.StreamMsgTypeText)
		return c, c.Flush()
	}

	if msg.IsDone {
		fullContent := c.fullStreamContent.String()
		log.Debugf("流式输出结束，总长度: %d", len(fullContent))
		c.clearStreamState()

		if fullContent != "" {
			c.history = append(c.history, model.Message{
				Role:    model.RoleAssistant,
				Content: fullContent,
				MsgType: model.StreamMsgTypeText,
			})
			c.addMessage(model.RoleAssistant, fullContent, model.StreamMsgTypeText)
		}
		return c, c.Flush()
	}

	// 跳过空消息（继续接收下一个）
	if msg.Content == "" {
		return c, c.receiveNextChunk()
	}

	// 判断是否是第一个消息块
	isFirst := c.fullStreamContent.Len() == 0

	// 如果当前有流式内容，且新消息类型与当前不同，先flush之前的内容
	if !isFirst && c.currentSegmentType != "" && c.currentSegmentType != msg.MsgType {
		// 保存之前的内容
		prevContent := c.fullStreamContent.String()
		prevMsgType := c.currentSegmentType

		// 清空流式状态
		c.fullStreamContent.Reset()
		c.streamingMsg = ""

		// 将之前的内容添加到消息列表
		if prevContent != "" {
			c.addMessage(model.RoleAssistant, prevContent, prevMsgType)
		}

		// 重置为第一个消息块
		isFirst = true
	}

	// 正常流式内容
	c.fullStreamContent.WriteString(msg.Content)
	c.currentSegmentType = msg.MsgType

	// 根据消息类型选择样式
	var contentStyle lipgloss.Style
	switch msg.MsgType {
	case model.StreamMsgTypeTool:
		// 工具消息 - 黄色
		contentStyle = style.ChatToolMsgStyle
	case model.StreamMsgTypeReason:
		// 思考消息 - 灰色
		contentStyle = style.ChatReasonMsgStyle
	case model.StreamMsgTypeText:
		fallthrough
	default:
		// 正文消息 - 白色
		contentStyle = style.ChatNormalMsgStyle
	}

	if isFirst {
		c.streamingMsg = style.ChatSystemMsgStyle.Render("🤖 MSA: ") +
			contentStyle.Render(msg.Content) + "\n"
	} else {
		c.streamingMsg += contentStyle.Render(msg.Content)
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
	c.currentSegment.Reset()
	c.currentSegmentType = ""
	c.streamSegments = make([]model.Message, 0)
	// 禁用输入框
	c.textInput.Blur()
	log.Debug("流式输出开始，禁用输入框")
	return c.receiveNextChunk()
}

// receiveNextChunk 接收下一个流式消息块
func (c *Chat) receiveNextChunk() tea.Cmd {
	return func() tea.Msg {
		if c.streamOutputCh == nil {
			log.Error("流式输出 channel 为空")
			return &model.StreamChunk{Err: fmt.Errorf("stream output channel is nil")}
		}

		chunk, ok := <-c.streamOutputCh
		if !ok {
			log.Debug("流式输出 channel 已关闭")
			return &model.StreamChunk{IsDone: true}
		}

		log.Debugf("接收流式块: Content长度=%d, IsDone=%v, Err=%v", len(chunk.Content), chunk.IsDone, chunk.Err)
		return chunk
	}
}

// clearStreamState 清除流式输出状态
func (c *Chat) clearStreamState() {
	c.isStreaming = false
	c.streamingMsg = ""
	c.fullStreamContent.Reset()
	c.currentSegment.Reset()
	c.currentSegmentType = ""
	c.streamSegments = nil

	if c.streamUnregister != nil {
		log.Debug("取消流式输出订阅")
		c.streamUnregister()
		c.streamUnregister = nil
	}
	c.streamOutputCh = nil

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

	c.addMessage(model.RoleUser, input, model.StreamMsgTypeText)
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

	// 处理 /skills: 命令（手动指定 skills）
	if strings.HasPrefix(input, "/skills:") {
		return c.handleSkillsCommand(input)
	}

	// 发起聊天请求
	return c.startChatRequest(input)
}

// handleClearHistory 清空对话历史
func (c *Chat) handleClearHistory() (tea.Model, tea.Cmd) {
	c.textInput.Reset()
	c.history = make([]model.Message, 0)
	c.addMessage(model.RoleSystem, clearSuccessMessage, model.StreamMsgTypeText)
	return c, c.Flush()
}

// handleSkillsCommand 处理 /skills: 命令
// 格式: /skills: base,stock-analysis 或 /skills: base stock-analysis
func (c *Chat) handleSkillsCommand(input string) (tea.Model, tea.Cmd) {
	// 提取 skills 列表
	skillsStr := strings.TrimPrefix(input, "/skills:")
	skillsStr = strings.TrimSpace(skillsStr)

	if skillsStr == "" {
		c.addMessage(model.RoleSystem, "📋 使用方法: /skills: <skill1>,<skill2>...", model.StreamMsgTypeText)
		c.addMessage(model.RoleSystem, "示例: /skills: base,stock-analysis,output-formats", model.StreamMsgTypeText)
		c.addMessage(model.RoleSystem, "提示: 可用命令: /skills list - 查看所有可用的 skills", model.StreamMsgTypeText)
		c.textInput.Reset()
		return c, c.Flush()
	}

	// 解析 skills 列表（支持逗号或空格分隔）
	var skillList []string
	if strings.Contains(skillsStr, ",") {
		skillList = strings.Split(skillsStr, ",")
	} else {
		skillList = strings.Fields(skillsStr)
	}

	// 去除空格并验证
	c.oneTimeSkills = make([]string, 0, len(skillList))
	for _, skill := range skillList {
		skill = strings.TrimSpace(skill)
		if skill != "" {
			c.oneTimeSkills = append(c.oneTimeSkills, skill)
		}
	}

	if len(c.oneTimeSkills) == 0 {
		c.addMessage(model.RoleSystem, "❌ 未指定任何 skills", model.StreamMsgTypeText)
		c.textInput.Reset()
		return c, c.Flush()
	}

	// 显示确认信息
	c.addMessage(model.RoleSystem, fmt.Sprintf("✓ 已手动指定 skills: %v", c.oneTimeSkills), model.StreamMsgTypeText)
	c.addMessage(model.RoleSystem, "提示: 在此消息后直接输入问题，将使用指定的 skills（一次性使用）", model.StreamMsgTypeText)
	c.textInput.Reset()
	return c, c.Flush()
}

// startChatRequest 发起聊天请求
func (c *Chat) startChatRequest(input string) (tea.Model, tea.Cmd) {
	c.streamOutputCh, c.streamUnregister = message.RegisterStreamOutput(streamBufferSize)

	// 合并会话级别和一次性 skills（优先使用一次性 skills）
	var skillsForRequest []string
	if len(c.oneTimeSkills) > 0 {
		skillsForRequest = c.oneTimeSkills
		c.oneTimeSkills = nil // 一次性 skills 用后即焚
	} else {
		skillsForRequest = c.sessionSkills
	}

	err := agent.Ask(c.ctx, input, c.history, skillsForRequest)
	if err != nil {
		log.Errorf("聊天请求失败: %v", err)
		c.clearStreamState()
		c.addMessage(model.RoleSystem, "聊天出错: "+err.Error(), model.StreamMsgTypeText)
		return c, c.Flush()
	}

	return c, tea.Batch(c.Flush(), c.startStreaming())
}
