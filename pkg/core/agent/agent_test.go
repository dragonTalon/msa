package agent

import (
	"context"
	"github.com/cloudwego/eino/schema"
	log "github.com/sirupsen/logrus"
	"msa/pkg/core/event"
	"msa/pkg/logic/skills"
	"msa/pkg/utils"
	"testing"
	"time"
)

func TestAgent_Run(t *testing.T) {
	t.Skip("Skipping agent test - requires API key and live LLM connection")
	ctx := context.Background()
	agent, err := New(ctx)
	if err != nil {
		t.Fatalf("new agent error: %v", err)
		return
	}
	skillManager := skills.GetManager()
	if err := skillManager.Initialize(); err != nil {
		log.Warnf("[Runner] skills initialize warning: %v", err)
	}

	// Build query messages (system prompt + history + user input)
	now := time.Now()
	messages, err := BuildQueryMessages(ctx, "查看一下我的账户", []*schema.Message{}, map[string]any{
		"role":    "专业股票分析助手",
		"style":   "理性、专业、客观且严谨",
		"time":    now.Format("2006-01-02 15:04:05"),
		"weekday": now.Format("2006年01月02日 星期Monday"),
	})
	if err != nil {
		t.Fatalf("build query messages error: %v", err)
		return
	}
	sr, err := agent.einoAgent.Stream(context.Background(), messages)
	if err != nil {
		t.Fatalf("stream error: %v", err)
		return
	}
	c := make(chan event.Event, 1)
	go func() {
		for {
			select {
			case e := <-c:
				log.Infof("resulst event: %v", utils.ToJSONString(e))
			}
		}
	}()
	process, err := agent.adapter.Process(ctx, sr, c)
	if err != nil {
		return
	}

	log.Infof("process: %v", utils.ToJSONString(process))
}
