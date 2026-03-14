package cmd_version

import (
	"fmt"

	"github.com/spf13/cobra"

	"msa/pkg/version"
)

// NewCommand 创建 version 子命令
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Long:  `显示 MSA 的版本号、commit hash 和构建时间。`,
		Run:   runVersion,
	}

	return cmd
}

func runVersion(cmd *cobra.Command, args []string) {
	v, c, d := version.Get()
	fmt.Println("MSA 版本:", v)
	fmt.Println("Commit:", c)
	fmt.Println("构建时间:", d)
}