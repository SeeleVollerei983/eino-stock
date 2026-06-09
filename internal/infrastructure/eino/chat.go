package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"eino-stock/internal/infrastructure/eino/tools"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/go-resty/resty/v2"
)

type chatChoice struct {
	FinishReason string      `json:"finish_reason"`
	Message      chatMessage `json:"message"`
}

type chatMessage struct {
	Content   *string        `json:"content"`
	ToolCalls []chatToolCall `json:"tool_calls"`
}

type chatToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function chatFunction `json:"function"`
}

type chatFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ChatAgent struct{}

func NewChatAgent(ctx context.Context, cfg *AIConfig) (*ChatAgent, error) {
	log.Printf("[Agent] created (direct API mode)")
	return &ChatAgent{}, nil
}

func (a *ChatAgent) Stream(ctx context.Context, userQuery string, toolCb func(name, args string)) (*schema.StreamReader[*schema.Message], error) {
	start := time.Now()
	log.Printf("[Agent] === CHAT START === query=%q", userQuery)
	cfg := ReadAIConfig()
	if cfg.LLMApiKey == "" {
		return nil, fmt.Errorf("[Agent] LLM_API_KEY not configured")
	}
	client := resty.New().SetTimeout(60*time.Second).SetHeader("Authorization", "Bearer "+cfg.LLMApiKey).SetBaseURL(cfg.LLMBaseURL)
	toolMap := map[string]tool.InvokableTool{}
	apiTools := buildToolDefs(toolMap)
	log.Printf("[Agent] built %d tool definitions", len(apiTools))

	sr, sw := schema.Pipe[*schema.Message](10)
	go func() {
		defer sw.Close()
		messages := buildMessages(userQuery)

		for step := 0; step < 10; step++ {
			raw, err := callLLM(ctx, client, cfg.LLMModel, messages, apiTools)
			if err != nil {
				log.Printf("[Agent] llm error: %v", err)
				sw.Send(&schema.Message{Content: fmt.Sprintf("LLM失败: %v", err)}, nil)
				return
			}

			var resp struct {
				Choices []chatChoice `json:"choices"`
			}
			if err := json.Unmarshal(raw, &resp); err != nil {
				log.Printf("[Agent] parse error: %v", err)
				sw.Send(&schema.Message{Content: fmt.Sprintf("解析错误: %v", err)}, nil)
				return
			}
			if len(resp.Choices) == 0 {
				sw.Send(&schema.Message{Content: "LLM无响应"}, nil)
				return
			}

			c := resp.Choices[0]
			asm := c.Message

			asMap := map[string]any{"role": "assistant", "content": nil}
			if len(asm.ToolCalls) > 0 {
				var tcs []map[string]any
				for _, tc := range asm.ToolCalls {
					tcs = append(tcs, map[string]any{
						"id": tc.ID, "type": tc.Type,
						"function": map[string]any{"name": tc.Function.Name, "arguments": tc.Function.Arguments},
					})
				}
				asMap["tool_calls"] = tcs
			}
			if asm.Content != nil {
				asMap["content"] = *asm.Content
			}
			messages = append(messages, asMap)

			if len(asm.ToolCalls) > 0 {
				for _, tc := range asm.ToolCalls {
					log.Printf("[Agent] tool call: %s(args=%s)", tc.Function.Name, tc.Function.Arguments)
					if toolCb != nil {
						toolCb(tc.Function.Name, tc.Function.Arguments)
					}
					t, ok := toolMap[tc.Function.Name]
					if !ok {
						log.Printf("[Agent] unknown tool: %s", tc.Function.Name)
						continue
					}
					result, err := t.InvokableRun(ctx, tc.Function.Arguments)
					if err != nil {
						result = fmt.Sprintf("工具执行失败: %v", err)
					}
					tr := result
					if len(tr) > 100 {
						tr = tr[:100]
					}
					log.Printf("[Agent] tool result: %s...", tr)
					messages = append(messages, map[string]any{
						"role": "tool", "content": result, "tool_call_id": tc.ID,
					})
				}
				apiTools = nil
				continue
			}
			if asm.Content != nil {
				log.Printf("[Agent] final response (%d chars), elapsed=%.2fs", len(*asm.Content), time.Since(start).Seconds())
				sw.Send(&schema.Message{Content: *asm.Content}, nil)
			}
			return
		}
		log.Printf("[Agent] max steps")
		sw.Send(&schema.Message{Content: "处理超时"}, nil)
	}()
	return sr, nil
}

func buildToolDefs(toolMap map[string]tool.InvokableTool) []map[string]any {
	var defs []map[string]any
	for _, t := range tools.GetAllTools() {
		info, _ := t.Info(context.Background())
		if info == nil {
			continue
		}
		props, req := map[string]any{}, []string{}
		switch info.Name {
		case "SelectAStock":
			props = map[string]any{"words": map[string]any{"type": "string", "description": "选股条件"}}
			req = []string{"words"}
		case "GetStockMinuteData", "GetStockDetail", "GetStockNotice", "GetStockResearchReport":
			props = map[string]any{"code": map[string]any{"type": "string", "description": "股票代码"}}
			req = []string{"code"}
		}
		defs = append(defs, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name": info.Name, "description": info.Desc,
				"parameters": map[string]any{"type": "object", "properties": props, "required": req},
			},
		})
		if inv, ok := t.(tool.InvokableTool); ok {
			toolMap[info.Name] = inv
		}
	}
	return defs
}

func buildMessages(query string) []map[string]any {
	return []map[string]any{
		{"role": "system", "content": "你是股票AI专家，擅长通过技术面分析筛选优质股票。你的分析框架分三步：\n\n第一步：K线趋势分析\n- 查看K线图，检查是否多头排列(MA5>MA10>MA20>MA60)\n- 检查均线是否粘合向上\n- 股价回踩到均线粘合处是最佳参与位置\n\n第二步：反转形态确认\n- 横盘趋势的股票强势突破时关注\n- 下跌趋势的股票突破前期高点时关注\n- 回调不破突破临界点是最佳参与位置\n\n第三步：量价关系验证\n- 价涨量升、价跌量缩说明趋势稳定\n- 回踩不破前一个高点就是机会\n- 放量突破+缩量回调是最佳信号\n\n分析步骤：\n1. 先调用SelectAStock筛选符合条件的股票\n2. 对候选股票逐一调用ExpertStockAnalysis进行技术面分析\n3. 综合所有分析结果，给出推荐排序\n4. 重点推荐同时满足：多头排列+均线粘合+回踩确认+量价配合的股票\n\n严禁不调用工具直接回答。必须使用真实数据。"},
		{"role": "user", "content": query},
	}
}

func callLLM(ctx context.Context, client *resty.Client, model string, messages []map[string]any, tools []map[string]any) ([]byte, error) {
	body := map[string]any{"model": model, "messages": messages}
	if len(tools) > 0 {
		body["tools"] = tools
		body["tool_choice"] = "required"
	}

	b, _ := json.Marshal(body)
	log.Printf("[Agent] API REQ: %s", string(b))
	resp, err := client.R().SetContext(ctx).SetHeader("Content-Type", "application/json").SetBody(body).Post("/v1/chat/completions")
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}
