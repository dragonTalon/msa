package command

import (
	"context"
	"msa/pkg/model"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

// CommandSuggestion 命令建议结构体，包含命令名称和描述
type CommandSuggestion struct {
	Name        string
	Description string
}

type MsaCommand interface {
	Name() string
	Description() string
	Run(ctx context.Context, args []string) (*model.CmdResult, error)
	// 新tea模型入口
	ToSelect(item []*model.SelectorItem) (*model.BaseSelector, error)
}

var commandMap = map[string]MsaCommand{}

var listCommands = []string{}

func RegisterCommand(cmd MsaCommand) {
	commandMap[cmd.Name()] = cmd
	listCommands = append(listCommands, cmd.Name())
}

// GetLikeCommand 获取相似的命令
func GetLikeCommand(cmd string) []string {
	if cmd == "" || len(commandMap) == 0 {
		return nil
	}
	log.Infof("GetLikeCommand: %s", cmd)
	cmd = strings.ToLower(cmd)
	cmd = strings.TrimPrefix(cmd, "/")

	if cmd == "" {
		return listCommands
	}
	list := []string{}
	for _, v := range listCommands {
		if strings.HasPrefix(v, cmd) {
			list = append(list, v)
		}
	}
	return list
}

// GetCommand 获取命令
func GetCommand(cmd string) MsaCommand {
	if cmd == "" {
		return nil
	}
	cmd = strings.ToLower(cmd)
	cmd = strings.TrimPrefix(cmd, "/")
	command, ok := commandMap[cmd]
	if ok {
		return command
	}
	return nil
}

// GetCommandSuggestions 获取带描述的命令建议列表
// prefix: 命令前缀（可带或不带 "/"）
// 返回按名称排序的命令建议切片
func GetCommandSuggestions(prefix string) []CommandSuggestion {
	if len(commandMap) == 0 {
		return nil
	}

	// 处理前缀
	prefix = strings.ToLower(prefix)
	prefix = strings.TrimPrefix(prefix, "/")

	// 收集匹配的命令
	var suggestions []CommandSuggestion
	for _, name := range listCommands {
		if prefix == "" || strings.HasPrefix(name, prefix) {
			if cmd, ok := commandMap[name]; ok {
				suggestions = append(suggestions, CommandSuggestion{
					Name:        name,
					Description: cmd.Description(),
				})
			}
		}
	}

	// 按名称排序
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Name < suggestions[j].Name
	})

	return suggestions
}
