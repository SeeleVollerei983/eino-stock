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

type IndustryValuationTool struct{ client *resty.Client }

func NewIndustryValuationTool() *IndustryValuationTool {
	return &IndustryValuationTool{client: resty.New().SetTimeout(10*time.Second).SetHeader("User-Agent", "Mozilla/5.0")}
}

func (t *IndustryValuationTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetIndustryValuation",
		Desc: "获取行业/板块平均估值和中值（PE,PEG等）。无需参数。",
		ParamsOneOf: schema.NewParamsOneOfByParams(nil),
	}, nil
}

func (t *IndustryValuationTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?cb=&pn=1&pz=50&po=1&np=1&fields=f2,f3,f7,f8,f9,f10,f12,f14,f20,f21,f23&fid=f3&fs=m:90+t:2&_=%d", time.Now().UnixMilli())
	resp, err := t.client.R().SetContext(ctx).SetHeader("Referer", "https://quote.eastmoney.com/").Get(url)
	if err != nil { return "", fmt.Errorf("eastmoney: %w", err) }

	var raw struct {
		Data *struct {
			Diff []struct {
				F12 string  `json:"f12"`
				F14 string  `json:"f14"`
				F7  float64 `json:"f7"`
				F8  float64 `json:"f8"`
				F9  float64 `json:"f9"`
				F23 float64 `json:"f23"`
			} `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &raw); err != nil { return string(resp.Body()), nil }
	if raw.Data == nil { return "暂无数据", nil }

	var b strings.Builder
	b.WriteString("## 行业估值排名\n\n| 行业 | 涨跌幅 | 市盈率 | 市净率 | 市销率 |\n|---|---|---|---|---|\n")
	for _, d := range raw.Data.Diff {
		if d.F14 != "" {
			b.WriteString(fmt.Sprintf("| %s | %.2f%% | %.2f | %.2f | %.2f |\n", d.F14, d.F7, d.F9, d.F23, d.F8))
		}
	}
	return b.String(), nil
}
var _ tool.InvokableTool = (*IndustryValuationTool)(nil)