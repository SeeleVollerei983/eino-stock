package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"eino-stock/internal/infrastructure/search"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type WebSearchTool struct {
	client *search.Client
}

func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{client: search.NewClient()}
}

func (t *WebSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "WebSearch",
		Desc: "搜索互联网获取最新信息。输入搜索关键词，返回相关网页标题、链接和摘要。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type: "string", Desc: "搜索关键词", Required: true,
			},
		}),
	}, nil
}

func (t *WebSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	if args.Query == "" {
		return "", fmt.Errorf("query is required")
	}
	results, err := t.client.Search(ctx, args.Query)
	if err != nil {
		return fmt.Sprintf("搜索失败: %v", err), nil
	}
	if len(results) == 0 {
		return "未找到相关搜索结果", nil
	}
	var out string
	for i, r := range results {
		if i >= 10 { break }
		out += fmt.Sprintf("%d. [%s](%s)\n   %s\n\n", i+1, r.Title, r.URL, r.Snippet)
	}
	return out, nil
}
