// Package agent provides the core agent implementation for MSA.
package agent

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/schema"
	log "github.com/sirupsen/logrus"
)

// dsmlStartPattern 用于检测 DSML 开始标记（大小写不敏感）
var dsmlStartPattern = regexp.MustCompile(`(?i)<[｜|]DSML[｜|]function_calls>`)

// pipeToolCallPattern 用于检测 <|tool_calls_section_begin|> 格式的工具调用
// 同时兼容半角竖线 | (U+007C) 和全角竖线 ｜ (U+FF5C)
var pipeToolCallPattern = regexp.MustCompile(`<[|｜]tool_calls_section_begin[|｜]>`)

// parseToolCalls 解析非标准格式的工具调用，支持 DSML 和 pipe 两种格式
func parseToolCalls(content string) []schema.ToolCall {
	if pipeToolCallPattern.MatchString(content) {
		if calls := parsePipeToolCalls(content); len(calls) > 0 {
			return calls
		}
	}
	return parseDSMLToolCalls(content)
}

func parseDSMLToolCalls(content string) []schema.ToolCall {
	invokePattern := regexp.MustCompile(`(?i)<[｜|]DSML[｜|]invoke\s+name="([^"]+)"[^>]*>([\s\S]*?)</[｜|]DSML[｜|]invoke>`)
	paramPattern := regexp.MustCompile(`(?i)<[｜|]DSML[｜|]parameter\s+name="([^"]+)"[^>]*>([\s\S]*?)</[｜|]DSML[｜|]parameter>`)

	var toolCalls []schema.ToolCall
	invokeMatches := invokePattern.FindAllStringSubmatch(content, -1)

	for i, invokeMatch := range invokeMatches {
		if len(invokeMatch) < 3 {
			continue
		}
		toolName := strings.TrimSpace(invokeMatch[1])
		invokeBody := invokeMatch[2]

		params := make(map[string]interface{})
		paramMatches := paramPattern.FindAllStringSubmatch(invokeBody, -1)
		for _, paramMatch := range paramMatches {
			if len(paramMatch) < 3 {
				continue
			}
			params[strings.TrimSpace(paramMatch[1])] = strings.TrimSpace(paramMatch[2])
		}

		argsJSON, err := json.Marshal(params)
		if err != nil {
			log.Warnf("parseDSMLToolCalls: 序列化参数失败: %v", err)
			continue
		}

		idx := i
		toolCalls = append(toolCalls, schema.ToolCall{
			Index: &idx,
			ID:    fmt.Sprintf("dsml_call_%d", i),
			Type:  "function",
			Function: schema.FunctionCall{
				Name:      toolName,
				Arguments: string(argsJSON),
			},
		})
	}

	return toolCalls
}

func parsePipeToolCalls(content string) []schema.ToolCall {
	pipeExtractPattern := regexp.MustCompile(`<[|｜]tool_calls_section_begin[|｜]>(\[.*?\])`)
	matches := pipeExtractPattern.FindStringSubmatch(content)
	if len(matches) < 2 {
		pipeExtractPatternMultiline := regexp.MustCompile(`(?s)<[|｜]tool_calls_section_begin[|｜]>(\[.*?\])`)
		matches = pipeExtractPatternMultiline.FindStringSubmatch(content)
		if len(matches) < 2 {
			log.Warnf("parsePipeToolCalls: 未找到 JSON 数组")
			return nil
		}
	}

	jsonStr := strings.TrimSpace(matches[1])

	type pipeToolCall struct {
		Name       string                 `json:"name"`
		Parameters map[string]interface{} `json:"parameters"`
	}
	var rawCalls []pipeToolCall
	if err := json.Unmarshal([]byte(jsonStr), &rawCalls); err != nil {
		log.Warnf("parsePipeToolCalls: JSON 解析失败: %v, json: %s", err, jsonStr)
		return nil
	}

	var toolCalls []schema.ToolCall
	for i, raw := range rawCalls {
		argsJSON, err := json.Marshal(raw.Parameters)
		if err != nil {
			log.Warnf("parsePipeToolCalls: 序列化参数失败: %v", err)
			continue
		}
		idx := i
		toolCalls = append(toolCalls, schema.ToolCall{
			Index: &idx,
			ID:    fmt.Sprintf("pipe_call_%d", i),
			Type:  "function",
			Function: schema.FunctionCall{
				Name:      strings.TrimSpace(raw.Name),
				Arguments: string(argsJSON),
			},
		})
	}
	return toolCalls
}
