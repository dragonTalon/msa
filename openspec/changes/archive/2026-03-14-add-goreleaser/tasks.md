# 实现任务：添加 GoReleaser 自动发布

## 1. 版本变量声明

- [x] 1.1 在 `main.go` 中添加版本变量声明
  - 声明 `Version`、`Commit`、`Date` 变量
  - 默认值分别为 `"dev"`、`"none"`、`"unknown"`

## 2. version 子命令

- [x] 2.1 创建 `cmd/version/` 目录
- [x] 2.2 创建 `cmd/version/command.go`
  - 包名为 `cmd_version`
  - 实现 `NewCommand() *cobra.Command`
  - 输出版本号、commit、构建时间
- [x] 2.3 在 `cmd/root.go` 中注册 version 命令
  - 导入 `msa/cmd/version`
  - 调用 `AddCommand(cmd_version.NewCommand())`

## 3. GoReleaser 配置

- [x] 3.1 创建 `.goreleaser.yml` 配置文件
  - 配置项目名称为 `msa`
  - 配置目标平台（linux/darwin/windows，amd64/arm64）
  - 配置 ldflags 注入版本信息
  - 配置打包格式（tar.gz/zip）
  - 配置校验和生成
  - 配置 Changelog 生成规则
- [x] 3.2 配置版本信息注入的 ldflags
  - `-X main.Version={{.Version}}`
  - `-X main.Commit={{.Commit}}`
  - `-X main.Date={{.Date}}`

## 4. GitHub Actions 工作流

- [x] 4.1 创建 `.github/workflows/release.yml`
  - 触发条件：`push: tags: ['v*']`
  - 使用 `goreleaser/goreleaser-action@v6`
  - 配置 `GITHUB_TOKEN` 权限

## 5. 验证

- [x] 5.1 运行 `go vet ./...` 验证编译
- [x] 5.2 本地测试 version 命令
  - 执行 `go run . version` 显示开发版本信息
- [ ] 5.3 本地测试 GoReleaser（可选）
  - 安装 goreleaser
  - 运行 `goreleaser release --snapshot --clean`
  - 验证生成的产物