package extcli

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	coreagent "msa/pkg/core/agent"
	"msa/pkg/core/runner"
	"msa/pkg/config"
	"msa/pkg/model"
	"msa/pkg/renderer"
	"msa/pkg/session"
)

// Run executes a single-round CLI conversation.
// Returns exit code: 0 for success, 1 for failure.
func Run(ctx context.Context, question string, modelOverride string) int {
	if question == "" {
		log.Error("问题内容不能为空")
		return 1
	}

	cfg := config.GetLocalStoreConfig()
	if cfg == nil {
		log.Error("配置未初始化，请运行 'msa config'")
		return 1
	}
	if cfg.APIKey == "" {
		log.Error("请先配置 API Key，运行 'msa config'")
		return 1
	}
	if cfg.BaseURL == "" {
		log.Error("请先配置 Base URL，运行 'msa config'")
		return 1
	}

	// Apply model override
	if modelOverride != "" {
		cfg.Model = modelOverride
		log.Infof("使用命令行指定的模型: %s", modelOverride)
	}
	if cfg.Model == "" {
		log.Error("请指定模型 (-m) 或配置默认模型")
		return 1
	}

	// Create Agent (fresh instance each time, no global cache)
	ag, err := coreagent.New(ctx)
	if err != nil {
		log.Errorf("创建 Agent 失败: %v", err)
		return 1
	}

	// Create session and persist user message
	sessionMgr := session.GetManager()
	sess := sessionMgr.NewSession(session.ModeCLI)
	if err := sessionMgr.CreateSessionFile(sess); err != nil {
		log.Warnf("创建会话文件失败: %v", err)
	}
	sessionMgr.SetCurrent(sess)
	sessionMgr.AppendMessage(sess, "user", question)

	// Create Runner with CLIRenderer
	r := runner.New(ag, sessionMgr, renderer.NewCLI(os.Stdout, false))

	// Start conversation (no history for single-round mode)
	if err := r.Ask(ctx, question, []model.Message{}); err != nil {
		log.Errorf("对话失败: %v", err)
		return 1
	}

	// Print session ID for resuming
	fmt.Printf("\n---\n📌 会话ID: %s\n", sess.SessionID())
	fmt.Printf("   msa --resume %s\n", sess.SessionID())

	return 0
}
