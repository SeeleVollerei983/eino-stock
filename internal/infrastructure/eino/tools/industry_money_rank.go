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

type IndustryMoneyRankTool struct{ client *resty.Client }

func NewIndustryMoneyRankTool() *IndustryMoneyRankTool {
	return &IndustryMoneyRankTool{client: resty.New().SetTimeout(10*time.Second).SetHeader("User-Agent", "Mozilla/5.0")}
}

func (t *IndustryMoneyRankTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetIndustryMoneyRank",
		Desc: "获取行业资金流向排名（按主力净流入排序）。无需参数。",
		ParamsOneOf: schema.NewParamsOneOfByParams(nil),
	}, nil
}

func (t *IndustryMoneyRankTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?cb=&pn=1&pz=50&po=1&np=1&fields=f2,f3,f12,f14,f62,f184,f66,f69&fid=f62&fs=m:90+t:2&_=%d", time.Now().UnixMilli())
	resp, err := t.client.R().SetContext(ctx).SetHeader("Referer", "https://data.eastmoney.com/bkzj/").Get(url)
	if err != nil { return "", fmt.Errorf("eastmoney: %w", err) }

	var raw struct {
		Data *struct {
			Diff []struct {
				F14 string  `json:"f14"`
				F3  float64 `json:"f3"`
				F62 float64 `json:"f62"`
				F184 float64 `json:"f184"`
				F66 float64 `json:"f66"`
			} `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &raw); err != nil { return string(resp.Body()), nil }
	if raw.Data == nil { return "暂无数据", nil }

	var b strings.Builder
	b.WriteString("## 行业资金流向排名\n\n| 行业 | 涨跌幅 | 主力净流入 | 超大单 | 大单 |\n|---|---|---|---|---|\n")
	for _, d := range raw.Data.Diff {
		if d.F14 != "" {
			b.WriteString(fmt.Sprintf("| %s | %.2f%% | %.2f亿 | %.2f亿 | %.2f亿 |\n", d.F14, d.F3, d.F62/1e8, d.F184/1e8, d.F66/1e8))
		}
	}
	return b.String(), nil
}
var _ tool.InvokableTool = (*IndustryMoneyRankTool)(nil)