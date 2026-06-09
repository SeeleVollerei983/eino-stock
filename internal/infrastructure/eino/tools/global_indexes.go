package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/go-resty/resty/v2"
)

type GlobalStockIndexesTool struct{ client *resty.Client }

func NewGlobalStockIndexesTool() *GlobalStockIndexesTool {
	return &GlobalStockIndexesTool{client: resty.New().SetTimeout(10*time.Second).SetHeader("User-Agent", "Mozilla/5.0")}
}

func (t *GlobalStockIndexesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GlobalStockIndexes",
		Desc: "获取全球主要指数行情，包括美股(道指/纳指/标普)、A股(上证/深证/创业板)、港股(恒指)、欧股、日股等。无需参数。",
		ParamsOneOf: schema.NewParamsOneOfByParams(nil),
	}, nil
}

func (t *GlobalStockIndexesTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	resp, err := t.client.R().SetContext(ctx).
		SetHeader("Referer", "https://stockapp.finance.qq.com/mstats").
		Get("https://proxy.finance.qq.com/ifzqgtimg/appstock/app/rank/indexRankDetail2")
	if err != nil { return "", fmt.Errorf("qq: %w", err) }

	var raw struct { Data map[string]any `json:"data"` }
	if err := json.Unmarshal(resp.Body(), &raw); err != nil { return string(resp.Body()), nil }

	var b strings.Builder
	for _, g := range []struct{ Key, Title string }{
		{"common", "主要指数"}, {"america", "美洲"}, {"asia", "亚太"}, {"europe", "欧洲"}, {"other", "其他"},
	} {
		items, ok := raw.Data[g.Key].([]any)
		if !ok || len(items) == 0 { continue }
		b.WriteString(fmt.Sprintf("\n## %s\n\n", g.Title))
		b.WriteString("| 名称 | 最新价 | 涨跌幅 |\n|---|---|---|\n")
		for _, item := range items {
			if m, ok := item.(map[string]any); ok {
				price := m["zxj"]; if price == nil { price = m["price"] }
				b.WriteString(fmt.Sprintf("| %s | %v | %v |\n", m["name"], price, m["zdf"]))
			}
		}
	}
	return b.String(), nil
}
var _ tool.InvokableTool = (*GlobalStockIndexesTool)(nil)
