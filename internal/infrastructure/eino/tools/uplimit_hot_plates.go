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

type UplimitHotPlatesTool struct{ client *resty.Client }

func NewUplimitHotPlatesTool() *UplimitHotPlatesTool {
	return &UplimitHotPlatesTool{client: resty.New().SetTimeout(10*time.Second).SetHeader("User-Agent", "Mozilla/5.0")}
}

func (t *UplimitHotPlatesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetUplimitHotPlates",
		Desc: "获取今日涨停热门板块数据。无需参数。",
		ParamsOneOf: schema.NewParamsOneOfByParams(nil),
	}, nil
}

func (t *UplimitHotPlatesTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?cb=&pn=1&pz=30&po=1&np=1&fields=f2,f3,f12,f14,f62,f184,f66,f69,f70,f71&fid=f3&fs=m:90+t:3+f:!50&_=%d", time.Now().UnixMilli())
	resp, err := t.client.R().SetContext(ctx).SetHeader("Referer", "https://data.eastmoney.com/").Get(url)
	if err != nil { return "", fmt.Errorf("eastmoney: %w", err) }

	var raw struct {
		Data *struct {
			Diff []struct {
				F3  float64 `json:"f3"`
				F12 string  `json:"f12"`
				F14 string  `json:"f14"`
				F62 float64 `json:"f62"`
				F70 float64 `json:"f70"`
			} `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &raw); err != nil { return string(resp.Body()), nil }
	if raw.Data == nil { return "暂无数据", nil }

	var b strings.Builder
	b.WriteString("## 涨停热门板块\n\n| 板块 | 涨跌幅 | 主力净流入 | 涨停数 |\n|---|---|---|---|\n")
	for _, d := range raw.Data.Diff {
		if d.F14 != "" {
			b.WriteString(fmt.Sprintf("| %s | %.2f%% | %.2f亿 | %.0f |\n", d.F14, d.F3, d.F62/1e8, d.F70))
		}
	}
	if b.Len() < 50 { return "今日暂无涨停板块数据", nil }
	return b.String(), nil
}
var _ tool.InvokableTool = (*UplimitHotPlatesTool)(nil)