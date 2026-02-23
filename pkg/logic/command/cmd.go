package command

import (
	"context"
	"msa/pkg/model"
	"strings"

	log "github.com/sirupsen/logrus"
)

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
