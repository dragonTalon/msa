package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"msa/pkg/app"
	"msa/pkg/config"
	"msa/pkg/tui"
)

var rootCmd = &cobra.Command{
	Use:   "msa",
	Short: "My Stock Agent CLI",
	Long:  tui.Logo,
	RunE:  runRoot,
}

// runRoot 根命令执行函数，仅做路由调用
func runRoot(cmd *cobra.Command, args []string) error {
	return app.Run(cmd.Context())
}

// Execute 程序执行入口
func Execute() {
	ExecuteWithSignal(rootCmd)
}

// ExecuteWithSignal 执行命令并处理信号
func ExecuteWithSignal(rootCmd *cobra.Command) {
	// 初始化配置
	if err := config.InitConfig(); err != nil {
		log.Warnf("初始化配置失败: %v", err)
	} else {
		log.Info("配置初始化成功")
	}

	ctx, cancel := NotifySignal(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Errorf("MSA execute failed: %v", err)
		log.Fatal(err)
	}
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
