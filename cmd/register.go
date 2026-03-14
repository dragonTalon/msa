package cmd

import "github.com/spf13/cobra"

// AddCommand 注册子命令到 rootCmd
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}
