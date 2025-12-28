# msa (My Stock Agent)

## Project Structure 
```
msa/ (Project Root | 项目根目录)
├── go.mod (Go Module Dependency Configuration | Go模块依赖配置)
├── go.sum (Go Module Dependency Verification | Go模块依赖校验文件)
├── main.go (Project Entry File, Initializes Cobra CLI | 项目入口文件，初始化Cobra CLI)
├── cmd/ (Cobra CLI Command Definitions Directory | Cobra CLI 命令定义目录，每个子命令对应一个文件)
├── pkg/ (Public Utility Package, Exposed for Reuse | 公共工具包，对外可暴露，供其他项目复用)
│   ├── config/ (Configuration Management Module | 配置管理模块)
├── api/ (HTTP API Interface Module | HTTP API 接口模块)
├── data/ (Local Data Storage Directory, Auto-generated | 本地数据存储目录，自动生成)
└── docs/ (Project Documentation Directory | 项目文档目录)
```