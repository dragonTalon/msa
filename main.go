/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"msa/cmd"
	"msa/pkg/version"
)

// 版本信息，由 GoReleaser 在编译时注入
// ldflags: -X main.Version -X main.Commit -X main.Date
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func main() {
	version.Set(Version, Commit, Date)
	cmd.Execute()
}
