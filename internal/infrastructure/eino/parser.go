package eino

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ParseConditions extracts structured stock screening conditions from a query.
// For structured queries (containing ";"), it passes them through.
// For natural language queries, it uses LLM to extract conditions.
func ParseConditions(ctx context.Context, query string, chatModel model.ToolCallingChatModel) (string, error) {
	// If already structured (contains ";" or is short/direct), pass through
	if strings.Contains(query, ";") {
		return query, nil
	}

	if chatModel == nil {
		return query, nil
	}

	sysPrompt := `你是一个股票选股条件提取器。将用户的自然语言查询转换为i问财格式的选股条件。
要求：
1. 提取所有筛选条件
2. 用中文分号(;)连接多个条件
3. 保留原始数值和单位
4. 不添加原查询中没有的条件

输出格式：仅输出条件字符串，不要有其他文字。

示例：
用户：帮我找找市盈率低于20的医药股，市值要大于100亿
输出：市盈率小于20;所属行业包含医药;总市值大于100亿

用户：筛选换手率前200名的非ST股票，股价在5到50之间
输出：非ST;换手率前200;股价大于5小于50`

	messages := []*schema.Message{
		{Role: schema.System, Content: sysPrompt},
		{Role: schema.User, Content: query},
	}

	result, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("parser: %w", err)
	}
	return strings.TrimSpace(result.Content), nil
}

func NewChatModel(ctx context.Context, cfg *AIConfig) (model.ToolCallingChatModel, error) {
	if cfg == nil || cfg.LLMApiKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY not configured")
	}
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.LLMBaseURL,
		APIKey:  cfg.LLMApiKey,
		Model:   cfg.LLMModel,
	})
}