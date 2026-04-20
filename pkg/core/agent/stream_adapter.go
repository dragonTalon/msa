package agent

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"

	"msa/pkg/core/event"
	corelogger "msa/pkg/core/logger"
)

// ProcessResult is the return value of StreamAdapter.Process.
// Contains tool calls collected during this LLM round and the assembled assistant message.
type ProcessResult struct {
	ToolCalls    []schema.ToolCall
	AssistantMsg *schema.Message // for appending to messages to continue the conversation
	HasContent   bool            // whether any text content was output
}

// StreamAdapter consumes a raw Eino stream and emits clean Event values.
// Its sole responsibility: handle Eino chunk format instability so callers always get clean Events.
type StreamAdapter struct{}

// Process handles one LLM stream, sending Events to the out channel.
// Also collects tool calls and text content, returning ProcessResult for the ReAct loop.
func (a *StreamAdapter) Process(
	ctx context.Context,
	reader *schema.StreamReader[*schema.Message],
	out chan<- event.Event,
) (ProcessResult, error) {
	logger := corelogger.FromCtx(ctx)

	var (
		textBuf         strings.Builder
		toolCalls       []schema.ToolCall
		chunkCount      int
		specialDetected bool
		toolCallBuf     strings.Builder
		contentOnlyLen  int
		allChunks       []*schema.Message
	)

	// Thinking progress ticker goroutine
	thinkingDone := make(chan struct{})
	thinkingReset := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		elapsed := 0
		for {
			select {
			case <-thinkingDone:
				return
			case <-thinkingReset:
				ticker.Reset(15 * time.Second)
				elapsed = 0
			case <-ticker.C:
				elapsed += 15
				logger.Infof("[StreamAdapter] 正在思考中...(%ds)", elapsed)
			}
		}
	}()
	defer close(thinkingDone)

	for {
		msg, err := reader.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Debugf("[StreamAdapter] stream 结束, chunks=%d, toolCalls=%d", chunkCount, len(toolCalls))
				break
			}
			if errors.Is(err, context.Canceled) {
				logger.Warnf("[StreamAdapter] context canceled，正常退出")
				reader.Close()
				return ProcessResult{}, context.Canceled
			}
			reader.Close() // drain to prevent goroutine leak
			return ProcessResult{}, err
		}
		chunkCount++

		// Reset thinking ticker on each chunk received
		select {
		case thinkingReset <- struct{}{}:
		default:
		}

		// Standard tool call: collect directly, don't broadcast
		// (executeTools will emit EventToolStart/Result later)
		if len(msg.ToolCalls) > 0 {
			logger.Infof("[StreamAdapter] chunk#%d 标准 toolCall: %v", chunkCount, toolCallNames(msg.ToolCalls))
			toolCalls = append(toolCalls, msg.ToolCalls...)
			allChunks = append(allChunks, msg)
			continue
		}

		// Handle ReasoningContent (thinking)
		if msg.ReasoningContent != "" {
			if !specialDetected {
				if tagIdx := findToolCallTag(msg.ReasoningContent); tagIdx >= 0 {
					if tagIdx > 0 {
						sendEvent(ctx, out, event.Event{Type: event.EventThinking, Text: msg.ReasoningContent[:tagIdx]})
					}
					toolCallBuf.WriteString(msg.ReasoningContent[tagIdx:])
					specialDetected = true
					logger.Infof("[StreamAdapter] chunk#%d 在 ReasoningContent 中检测到非标准工具调用标记", chunkCount)
				} else {
					logger.Debugf("[StreamAdapter] chunk#%d thinking len=%d", chunkCount, len(msg.ReasoningContent))
					sendEvent(ctx, out, event.Event{Type: event.EventThinking, Text: msg.ReasoningContent})
				}
			} else {
				toolCallBuf.WriteString(msg.ReasoningContent)
			}
			msg.ReasoningContent = ""
		}

		// Handle Content (main text)
		if msg.Content != "" {
			if !specialDetected {
				if tagIdx := findToolCallTagWithOffset(msg.Content, contentOnlyLen); tagIdx >= 0 {
					relIdx := tagIdx - contentOnlyLen
					if relIdx > 0 {
						sendEvent(ctx, out, event.Event{Type: event.EventTextChunk, Text: msg.Content[:relIdx]})
						textBuf.WriteString(msg.Content[:relIdx])
					}
					toolCallBuf.WriteString(msg.Content[relIdx:])
					contentOnlyLen += len(msg.Content)
					specialDetected = true
					logger.Infof("[StreamAdapter] chunk#%d 检测到非标准工具调用标记（Content）", chunkCount)
					// Clear previous chunks' Content to avoid ConcatMessages merging dirty data
					for _, prev := range allChunks {
						prev.Content = ""
					}
					msg.Content = ""
				} else {
					logger.Debugf("[StreamAdapter] chunk#%d text len=%d", chunkCount, len(msg.Content))
					sendEvent(ctx, out, event.Event{Type: event.EventTextChunk, Text: msg.Content})
					textBuf.WriteString(msg.Content)
					contentOnlyLen += len(msg.Content)
				}
			} else {
				toolCallBuf.WriteString(msg.Content)
				msg.Content = ""
			}
		}

		allChunks = append(allChunks, msg)
	}

	// Handle non-standard tool call format
	if specialDetected {
		fullContent := toolCallBuf.String()
		parsedCalls := parseToolCalls(fullContent)
		if len(parsedCalls) > 0 {
			logger.Infof("[StreamAdapter] 解析到非标准工具调用: %v", toolCallNames(parsedCalls))
			// Inject into last chunk (eino internal processing needs this)
			if len(allChunks) > 0 {
				lastChunk := allChunks[len(allChunks)-1]
				lastChunk.ToolCalls = parsedCalls
				lastChunk.Content = ""
			}
			toolCalls = append(toolCalls, parsedCalls...)
		} else {
			logger.Warnf("[StreamAdapter] specialDetected=true 但解析失败，内容: %q",
				truncate(fullContent, 200))
		}
	}

	// Signal end of text output
	if textBuf.Len() > 0 {
		sendEvent(ctx, out, event.Event{Type: event.EventTextDone})
	}

	// Assemble assistant message for appending to messages
	assistantMsg := buildAssistantMsg(allChunks)

	// For standard tool calls, use the fully merged ToolCalls from assistantMsg
	// (ConcatMessages properly merges streaming chunks into complete ToolCalls).
	// For non-standard tool calls (specialDetected), toolCalls was already set from parseToolCalls.
	if !specialDetected && len(assistantMsg.ToolCalls) > 0 {
		toolCalls = assistantMsg.ToolCalls
	}

	return ProcessResult{
		ToolCalls:    toolCalls,
		AssistantMsg: assistantMsg,
		HasContent:   textBuf.Len() > 0,
	}, nil
}

// sendEvent sends an event to the channel, respecting ctx cancellation.
func sendEvent(ctx context.Context, ch chan<- event.Event, e event.Event) bool {
	select {
	case ch <- e:
		return true
	case <-ctx.Done():
		return false
	}
}

// findToolCallTag finds the start position of a non-standard tool call marker in content.
// Returns -1 if not found.
func findToolCallTag(content string) int {
	if loc := pipeToolCallPattern.FindStringIndex(content); loc != nil {
		return loc[0]
	}
	if loc := dsmlStartPattern.FindStringIndex(content); loc != nil {
		return loc[0]
	}
	return -1
}

// findToolCallTagWithOffset is like findToolCallTag but adds offset to get absolute position.
func findToolCallTagWithOffset(content string, offset int) int {
	idx := findToolCallTag(content)
	if idx < 0 {
		return -1
	}
	return idx + offset
}

// buildAssistantMsg assembles all stream chunks into an assistant message.
func buildAssistantMsg(chunks []*schema.Message) *schema.Message {
	if len(chunks) == 0 {
		return &schema.Message{Role: schema.Assistant}
	}
	merged, err := schema.ConcatMessages(chunks)
	if err != nil {
		// fallback: manual text concatenation
		var sb strings.Builder
		for _, c := range chunks {
			sb.WriteString(c.Content)
		}
		return &schema.Message{Role: schema.Assistant, Content: sb.String()}
	}
	return merged
}

// toolCallNames extracts tool call names for logging.
func toolCallNames(calls []schema.ToolCall) []string {
	names := make([]string, 0, len(calls))
	for _, c := range calls {
		names = append(names, c.Function.Name)
	}
	return names
}

// truncate truncates a string for logging.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
