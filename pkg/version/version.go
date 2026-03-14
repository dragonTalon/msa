// Package version 提供版本信息管理
package version

var (
	// Version 版本号
	Version = "dev"
	// Commit git commit hash
	Commit = "none"
	// Date 构建时间
	Date = "unknown"
)

// Set 设置版本信息
func Set(v, c, d string) {
	Version = v
	Commit = c
	Date = d
}

// Get 获取版本信息
func Get() (version, commit, date string) {
	return Version, Commit, Date
}