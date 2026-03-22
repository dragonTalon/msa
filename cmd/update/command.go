// Package cmd_update 提供 update 子命令
package cmd_update

import (
	"fmt"

	"github.com/spf13/cobra"

	"msa/pkg/updater"
	"msa/pkg/version"
)

var (
	// checkOnly 是否只检查不更新
	checkOnly bool
)

// NewCommand 创建 update 子命令
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "更新 MSA 到最新版本",
		Long: `检查并更新 MSA 到 GitHub Releases 上的最新版本。

示例:
  msa update          # 检查并更新到最新版本
  msa update --check  # 仅检查是否有新版本`,
		RunE: runUpdate,
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "仅检查是否有新版本，不执行更新")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	v, c, _ := version.Get()
	u := updater.New(v, c)

	result, err := u.Update(checkOnly)
	if err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}

	fmt.Println()
	fmt.Printf("当前版本: %s\n", result.CurrentVersion)
	fmt.Printf("最新版本: %s\n", result.LatestVersion)

	if result.Updated {
		fmt.Printf("\n✅ %s\n", result.Message)
		fmt.Println("请重启 MSA 以使用新版本")
	} else {
		fmt.Printf("\n%s\n", result.Message)
	}

	return nil
}
