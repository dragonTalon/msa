/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "msa-cli",
	Short: "My Stock Agent CLI ",
	Long:  "                                              \n          ____                                \n        ,'  , `.   .--.--.       ,---,        \n     ,-+-,.' _ |  /  /    '.    '  .' \\       \n  ,-+-. ;   , || |  :  /`. /   /  ;    '.     \n ,--.'|'   |  ;| ;  |  |--`   :  :       \\    \n|   |  ,', |  ': |  :  ;_     :  |   /\\   \\   \n|   | /  | |  ||  \\  \\    `.  |  :  ' ;.   :  \n'   | :  | :  |,   `----.   \\ |  |  ;/  \\   \\ \n;   . |  ; |--'    __ \\  \\  | '  :  | \\  \\ ,' \n|   : |  | ,      /  /`--'  / |  |  '  '--'   \n|   : '  |/      '--'.     /  |  :  :         \n;   | |`-'         `--'---'   |  | ,'         \n|   ;/                        `--''           \n'---'                                         \n                                              (My Stock Agent CLI)",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(ConfigCmd)
	rootCmd.AddCommand(InitCmd)
	rootCmd.AddCommand(LoggerCmd)
}
