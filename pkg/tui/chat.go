package tui

import (
	"context"
	"fmt"
	"msa/pkg/config"
	"msa/pkg/logic/agent"
	command "msa/pkg/logic/command"
	"msa/pkg/model"
	"msa/pkg/tui/style"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	listStyle "github.com/charmbracelet/lipgloss/list"
	log "github.com/sirupsen/logrus"
)

// Chat TUI聊天模型
type Chat struct {
	textInput         textinput.Model           // 文本输入组件
	history           []model.Message           // 历史消息
	pendingMsgs       []model.Message           // 待 flush 的消息
	ctx               context.Context           // 上下文
	width             int                       // 终端宽度
	height            int                       // 终端高度
	cmdFlag           bool                      // 是否处于命令模式
	cmdList           []string                  // 命令列表
	streamingMsg      string                    // 流式输出的临时内容
	isStreaming       bool                      // 是否正在流式输出
	fullStreamContent strings.Builder           // 完整的流式内容
	streamOutputCh    <-chan *agent.StreamChunk // 流式输出 channel
	streamUnregister  func()                    // 取消订阅函数
}

// maskAPIKey 隐藏 APIKey，只显示前4个和后4个字符
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "未设置"
	}
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// NewChat 创建新的聊天模型
func NewChat(ctx context.Context) *Chat {
	// 初始化文本输入组件
	ti := textinput.New()
	ti.Placeholder = "输入你的理财问题..."
	ti.Focus()
	ti.CharLimit = 0
	ti.Width = 50
	ti.PromptStyle = style.ChatInputPromptStyle
	ti.Prompt = "MSA > "
	ti.TextStyle = style.ChatInputTextStyle
	cfg := config.GetLocalStoreConfig()
	m := cfg.Model
	if m == "" {
		m = "未设置"
	}
	return &Chat{
		textInput: ti,
		pendingMsgs: []model.Message{
			{Role: model.RoleLogo, Content: style.GetStyledLogo()},
			{Role: model.RoleSystem, Content: fmt.Sprintf("模型供应商: %s", cfg.Provider)},
			{Role: model.RoleSystem, Content: fmt.Sprintf("模型 : %s", m)},
			{Role: model.RoleSystem, Content: fmt.Sprintf("APIKey : %s", maskAPIKey(cfg.APIKey))},
			{Role: model.RoleSystem, Content: "欢迎使用 MSA！输入你的理财问题吧..."},
		},
		ctx: ctx,
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

	for i, msg := range c.pendingMsgs {
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
			sb.WriteString(style.ChatNormalMsgStyle.Render(msg.Content))
		}
		if i < len(c.pendingMsgs)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// addMessage 添加消息到待 flush 队列
func (c *Chat) addMessage(role model.MessageRole, content string) {
	c.pendingMsgs = append(c.pendingMsgs, model.Message{
		Role:    role,
		Content: content,
	})
}

// Update 实现 tea.Model 接口，处理消息更新
func (c *Chat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var tiCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.textInput.Width = msg.Width - 10

	case *agent.StreamChunk:
		if msg.Err != nil {
			c.clearStreamState()
			c.addMessage(model.RoleSystem, fmt.Sprintf("接收消息失败: %v", msg.Err))
			return c, c.Flush()
		}

		if msg.IsDone {
			fullContent := c.fullStreamContent.String()
			log.Infof("stream end: %s", fullContent)
			c.clearStreamState()

			if fullContent != "" {
				c.history = append(c.history, model.Message{
					Role:    model.RoleAssistant,
					Content: fullContent,
				})
				c.addMessage(model.RoleAssistant, fullContent)
			}
			return c, c.Flush()
		}

		// 跳过空消息（继续接收下一个）
		if msg.Content == "" {
			return c, c.receiveNextChunk()
		}

		// 正常流式内容
		c.fullStreamContent.WriteString(msg.Content)

		isFirst := c.fullStreamContent.Len() == len(msg.Content)
		if isFirst {
			c.streamingMsg = style.ChatSystemMsgStyle.Render("🤖 MSA: ") +
				style.ChatNormalMsgStyle.Render(msg.Content)
		} else {
			c.streamingMsg += style.ChatNormalMsgStyle.Render(msg.Content)
		}

		return c, c.receiveNextChunk()

	case tea.KeyMsg:
		log.Debugf("捕获按键: %s, Type: %v", msg.String(), msg.Type)
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return c, tea.Quit

		case tea.KeyEnter:
			c.cmdFlag = false
			input := strings.TrimSpace(c.textInput.Value())
			if input == "" {
				return c, nil
			}

			c.addMessage(model.RoleUser, input)
			c.textInput.Reset()

			// 处理命令
			if strings.HasPrefix(input, "/") {
				return c.commandHandler(input)
			}

			// 处理特殊命令
			switch strings.ToLower(input) {
			case "clear":
				c.history = []model.Message{}
				c.addMessage(model.RoleSystem, "对话已清空，重新开始吧！")
				return c, c.Flush()
			case "help", "?":
				c.addMessage(model.RoleSystem, "📋 可用命令:\n  • clear - 清空对话\n  • help/? - 显示帮助\n  • quit/exit - 退出程序")
				return c, c.Flush()
			case "quit", "exit":
				return c, tea.Quit
			}

			// 先注册订阅，确保不丢失消息
			c.streamOutputCh, c.streamUnregister = agent.RegisterStreamOutput(100)

			// 发起聊天请求（Ask 会在后台异步处理并通过 StreamOutputManager 广播流式数据）
			err := agent.Ask(c.ctx, input, c.history)
			if err != nil {
				log.Errorf("chat error: %v", err)
				// 清理订阅
				if c.streamUnregister != nil {
					c.streamUnregister()
					c.streamUnregister = nil
				}
				c.streamOutputCh = nil
				c.addMessage(model.RoleSystem, "聊天出错: "+err.Error())
				return c, c.Flush()
			}

			return c, tea.Batch(c.Flush(), c.startStreaming())

		case tea.KeyCtrlK:
			c.textInput.Reset()
			c.history = []model.Message{}
			c.addMessage(model.RoleSystem, "对话已清空，重新开始吧！")
			return c, c.Flush()

		default:
			c.textInput, tiCmd = c.textInput.Update(msg)
			if strings.HasPrefix(c.textInput.Value(), "/") {
				c.cmdFlag = true
				c.cmdList = command.GetLikeCommand(c.textInput.Value())
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
	if c.cmdFlag {
		styles := listStyle.New()
		for _, cmdStr := range c.cmdList {
			styles.Item("/" + cmdStr)
		}
		log.Infof("view styles %s", styles)
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("%s", styles))
	}
	// 帮助提示
	help := style.ChatHelpStyle.Render(
		"ESC/Ctrl+C: 退出 | Ctrl+K: 清空 | Enter: 发送",
	)
	sb.WriteString("\n")
	sb.WriteString(help)

	return sb.String()
}

// analyzeResult 分析结果
func analyzeResult(result *model.CmdResult) string {
	if result == nil {
		return "结果为空"
	}
	sb := strings.Builder{}
	sb.WriteString("\n")
	switch result.Type {
	case "list":
		list, ok := result.Data.([]string)
		if !ok {
			return "结果类型错误"
		}
		styles := listStyle.New()
		for _, v := range list {
			styles = styles.Item(v)
		}
		log.Infof("list styles %s", styles)
		sb.WriteString(fmt.Sprintf("%s", styles))

	case "table":
		table, ok := result.Data.(map[string]string)
		if !ok {
			return "结果类型错误"
		}
		// 渲染表格
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
	sb.WriteString(style.TableHeaderStyle.Render("描述"))
	sb.WriteString("\n")

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
		sb.WriteString(rowStyle.Render(value))
		sb.WriteString("\n")
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
		c.addMessage(model.RoleSystem, "未找到命令: "+input)
		c.addMessage(model.RoleSystem, fmt.Sprintf("可用命令: %v", command.GetLikeCommand("/")))
		return c, c.Flush()
	}

	var args []string
	if len(split) > 1 {
		args = split[1:]
	}

	// 执行命令
	runResult, err := msaCmd.Run(c.ctx, args)
	if err != nil {
		c.addMessage(model.RoleSystem, "执行命令失败: "+err.Error())
		log.Errorf("执行命令失败: %v", err)
		return c, c.Flush()
	}

	log.Infof("执行命令成功: %v", runResult)

	// 检查是否需要启动交互式选择器
	// 如果命令返回的是 selector 类型，则启动选择器
	if runResult.Type == "selector" {
		items, ok := runResult.Data.([]*model.SelectorItem)
		if !ok {
			c.addMessage(model.RoleSystem, "选择器数据类型错误")
			log.Errorf("选择器数据类型错误")
			return c, c.Flush()
		}

		// 调用命令的 ToSelect 方法创建选择器
		selector, err := msaCmd.ToSelect(items)
		if err != nil {
			c.addMessage(model.RoleSystem, "创建选择器失败: "+err.Error())
			log.Errorf("创建选择器失败: %v", err)
			return c, c.Flush()
		}

		// 设置上下文
		selector.Ctx = c.ctx
		c.textInput.Reset()

		// 使用 SelectorView 包装 BaseSelector，并传入当前聊天模型
		selectorView := NewSelectorView(selector, c)

		// 启动交互式选择器
		return selectorView, nil
	}

	// 普通命令结果，直接显示
	c.addMessage(model.RoleSystem, analyzeResult(runResult))
	c.textInput.Reset()
	return c, c.Flush()
}

// startStreaming 启动流式输出（订阅已在调用 Ask 之前完成）
func (c *Chat) startStreaming() tea.Cmd {
	c.isStreaming = true
	c.streamingMsg = style.ChatNormalMsgStyle.Render("⏳ 正在思考...")
	c.fullStreamContent.Reset()

	// 注意：订阅已在调用 Ask 之前完成，Ask 会通过 toolCallChecker 广播数据
	// 这里直接开始从 streamOutputCh 接收
	return c.receiveNextChunk()
}

// receiveNextChunk 接收下一个流式消息块（从 StreamOutputManager 的 channel 接收）
func (c *Chat) receiveNextChunk() tea.Cmd {
	return func() tea.Msg {
		if c.streamOutputCh == nil {
			return &agent.StreamChunk{Err: fmt.Errorf("stream output channel is nil")}
		}

		// 从 channel 接收数据
		chunk, ok := <-c.streamOutputCh
		if !ok {
			// channel 已关闭
			return &agent.StreamChunk{IsDone: true}
		}

		log.Infof("recv chunk: Content=%s, IsDone=%v, Err=%v", chunk.Content, chunk.IsDone, chunk.Err)
		// 直接返回接收到的 chunk
		return chunk
	}
}

// clearStreamState 清除流式输出状态
func (c *Chat) clearStreamState() {
	c.isStreaming = false
	c.streamingMsg = ""

	// 取消订阅并清理 channel
	if c.streamUnregister != nil {
		c.streamUnregister()
		c.streamUnregister = nil
	}
	c.streamOutputCh = nil
}
