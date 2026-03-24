package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"msa/cmd/config"
	"msa/cmd/skill"
	"msa/cmd/update"
	"msa/cmd/version"
	"msa/pkg/app"
	"msa/pkg/config"
	"msa/pkg/db"
	"msa/pkg/extcli"
	"msa/pkg/session"
	"msa/pkg/tui"
	"msa/pkg/tui/style"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	// configArgs --config 参数的值
	configArgs []string

	// question -q/--question 参数的值（单轮对话）
	question string

	// modelOverride -m/--model 参数的值（模型覆盖）
	modelOverride string

	// resumeSessionID --resume 参数的值（恢复会话）
	resumeSessionID string
)

var rootCmd = &cobra.Command{
	Use:   "msa",
	Short: "My Stock Agent CLI",
	Long:  style.Logo,
	RunE:  runRoot,
}

func init() {
	rootCmd.PersistentFlags().StringArrayVar(&configArgs, "config", nil, "配置参数（格式：key=value 或文件路径）")
	rootCmd.PersistentFlags().StringVarP(&question, "question", "q", "", "单轮对话问题（不进入TUI）")
	rootCmd.PersistentFlags().StringVarP(&modelOverride, "model", "m", "", "指定模型（覆盖配置文件）")
	rootCmd.PersistentFlags().StringVar(&resumeSessionID, "resume", "", "恢复会话（格式：YYYY-MM-DD_uuid）")

	// 注册子命令
	AddCommand(cmd_config.NewCommand())
	AddCommand(cmd_skill.NewCommand())
	AddCommand(cmd_version.NewCommand())
	AddCommand(cmd_update.NewCommand())
}

// runRoot 根命令执行函数，仅做路由调用
func runRoot(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// 如果指定了 -q 参数（即使为空），执行 CLI 单轮对话
	if cmd.Flags().Changed("question") {
		exitCode := extcli.Run(ctx, question, modelOverride)
		os.Exit(exitCode)
		return nil
	}

	// 处理 --resume 参数
	if resumeSessionID != "" {
		return runResume(ctx, resumeSessionID)
	}

	return app.Run(ctx)
}

// runResume 处理 --resume 参数
func runResume(ctx context.Context, sessionID string) error {
	// 加载会话
	sessionMgr := session.GetManager()
	parsed, err := sessionMgr.LoadSession(sessionID)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Printf("正确格式：msa --resume YYYY-MM-DD_uuid\n")
		os.Exit(1)
		return nil
	}

	log.Infof("恢复会话: %s, 历史消息: %d 条", parsed.Session.SessionID(), len(parsed.Messages))

	// 启动 TUI 并注入恢复的会话
	p := tea.NewProgram(tui.NewChat(ctx, tui.WithResumeSession(parsed)))
	if _, err := p.Run(); err != nil {
		log.Errorf("运行 TUI 失败: %v", err)
		return err
	}

	// TUI 退出后输出会话 ID
	sess := sessionMgr.Current()
	if sess != nil {
		fmt.Printf("\n📌 会话ID: %s\n", sess.SessionID())
		fmt.Printf("   msa --resume %s\n", sess.SessionID())
	}

	return nil
}

// Execute 程序执行入口
func Execute() {
	ExecuteWithSignal(rootCmd)
}

// ExecuteWithSignal 执行命令并处理信号
func ExecuteWithSignal(rootCmd *cobra.Command) {
	// 解析 --config 参数
	cliCfg := parseConfigArgs(configArgs)
	if cliCfg != nil {
		config.SetCLIConfig(cliCfg)
	}

	// 初始化配置
	if err := config.InitConfig(); err != nil {
		log.Warnf("初始化配置失败: %v", err)
	} else {
		log.Info("配置初始化成功")
	}

	// 注册数据库清理函数
	defer func() {
		if err := db.CloseGlobalDB(); err != nil {
			log.Warnf("关闭数据库时出错: %v", err)
		}
	}()

	ctx, cancel := NotifySignal(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Errorf("MSA execute failed: %v", err)
		log.Fatal(err)
	}
}

// parseConfigArgs 解析 --config 参数
func parseConfigArgs(args []string) *config.LocalStoreConfig {
	if len(args) == 0 {
		return nil
	}

	result := &config.LocalStoreConfig{}

	for _, arg := range args {
		cfg, err := config.ParseConfigArg(arg)
		if err != nil {
			log.Warnf("解析配置参数失败: %s, 错误: %v", arg, err)
			fmt.Printf("使用说明:\n")
			fmt.Printf("  --config key=value    设置配置项\n")
			fmt.Printf("  --config /path/to/file 加载配置文件\n")
			fmt.Printf("  支持的配置项: provider, apikey, baseurl, loglevel, logfile\n")
			continue
		}

		// 合并配置
		if cfg.Provider != "" {
			result.Provider = cfg.Provider
		}
		if cfg.APIKey != "" {
			result.APIKey = cfg.APIKey
		}
		if cfg.BaseURL != "" {
			result.BaseURL = cfg.BaseURL
		}
		if cfg.Model != "" {
			result.Model = cfg.Model
		}
		if cfg.LogConfig != nil {
			if result.LogConfig == nil {
				result.LogConfig = &config.LogConfig{}
			}
			if cfg.LogConfig.Level != "" {
				result.LogConfig.Level = cfg.LogConfig.Level
			}
			if cfg.LogConfig.Format != "" {
				result.LogConfig.Format = cfg.LogConfig.Format
			}
			if cfg.LogConfig.Output != "" {
				result.LogConfig.Output = cfg.LogConfig.Output
			}
			if cfg.LogConfig.File != "" {
				result.LogConfig.File = cfg.LogConfig.File
			}
			if cfg.LogConfig.TimeFormat != "" {
				result.LogConfig.TimeFormat = cfg.LogConfig.TimeFormat
			}
			if cfg.LogConfig.ShowColor {
				result.LogConfig.ShowColor = cfg.LogConfig.ShowColor
			}
		}
	}

	return result
}

// NotifySignal 创建带信号监听的上下文
func NotifySignal(parent context.Context, signals ...os.Signal) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	// 绑定信号通知
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)

	if ctx.Err() == nil {
		// 监听信号
		go func() {
			// 第一次收到信号取消上下文
			select {
			case <-ctx.Done():
				return
			case <-ch:
				cancel()
			}
			// 第二次直接退出
			select {
			case s, ok := <-ch:
				if !ok || s == nil {
					os.Exit(1)
				}
				if syscallSignal, isSyscallSignal := s.(syscall.Signal); isSyscallSignal {
					os.Exit(128 + int(syscallSignal)) // 128+n 被信号终止的退出码
				}
				os.Exit(1)
			}
		}()
	}

	return ctx, cancel
}
