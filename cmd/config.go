package cmd

import (
	"github.com/spf13/cobra"
	"msa/pkg/tui/config"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "配置 MSA",
	Long:  `打开 TUI 配置界面，管理 API Key、Provider、日志等配置。`,
	RunE:  runConfig,
}

func runConfig(cmd *cobra.Command, args []string) error {
	return config.Run()
}
