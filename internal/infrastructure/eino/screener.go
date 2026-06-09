package eino

import (
	"context"
	"encoding/json"
	"fmt"

	"eino-stock/internal/infrastructure/eino/tools"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

type ScreenResult struct {
	Query      string   `json:"query"`
	Candidates []string `json:"candidates,omitempty"`
	Raw        string   `json:"raw,omitempty"`
}

func RunScreener(ctx context.Context, userQuery string, aiCfg *AIConfig) (*ScreenResult, error) {
	chatModel, llmErr := NewChatModel(ctx, aiCfg)
	parsedQuery, err := ParseConditions(ctx, userQuery, chatModel)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	result := &ScreenResult{Query: parsedQuery}
	iwencaiTool := tools.NewSelectAStockTool()

	if llmErr == nil && chatModel != nil {
		agent, err := react.NewAgent(ctx, &react.AgentConfig{
			ToolCallingModel: chatModel,
			ToolsConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{iwencaiTool},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("create agent: %w", err)
		}
		msg, err := agent.Generate(ctx, []*schema.Message{
			{Role: schema.System, Content: `你是一个股票选股助手。使用 SelectAStock 工具筛选股票。用表格返回结果。`},
			{Role: schema.User, Content: parsedQuery},
		})
		if err != nil {
			return nil, fmt.Errorf("screener: %w", err)
		}
		result.Raw = msg.Content
	} else {
		toolInput, _ := json.Marshal(map[string]string{"words": parsedQuery})
		output, err := iwencaiTool.InvokableRun(ctx, string(toolInput))
		if err != nil {
			return nil, fmt.Errorf("tool: %w", err)
		}
		result.Raw = output
	}
	return result, nil
}