package cmd_config

import (
	"github.com/spf13/cobra"
	"msa/pkg/tui/config"
)

// NewCommand 创建 config 子命令
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "配置 MSA",
		Long:  `打开 TUI 配置界面，管理 API Key、Provider、日志等配置。`,
		RunE:  runConfig,
	}
	return cmd
}

func runConfig(cmd *cobra.Command, args []string) error {
	return config.Run()
}
