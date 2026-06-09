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

type MarketDataTool struct{ client *resty.Client }

func NewMarketDataTool() *MarketDataTool {
	return &MarketDataTool{client: resty.New().SetTimeout(10*time.Second).SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")}
}

func (t *MarketDataTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetMarketData",
		Desc: "获取A股市场行情概况，包括主要指数行情、涨跌家数分布、今日申购信息等。无需参数。",
		ParamsOneOf: schema.NewParamsOneOfByParams(nil),
	}, nil
}

func (t *MarketDataTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	resp, err := t.client.R().SetContext(ctx).
		SetHeader("Referer", "https://www.cls.cn/").
		Get("https://x-quote.cls.cn/quote/index/home?app=CailianpressWeb&os=web&sv=8.4.6")
	if err != nil { return "", fmt.Errorf("cls: %w", err) }

	var raw struct {
		Data struct {
			IndexQuote []struct {
				SecuCode  string  `json:"secu_code"`
				SecuName  string  `json:"secu_name"`
				LastPrice float64 `json:"last_price"`
				ChgPct    float64 `json:"chg_pct"`
			} `json:"index_quote"`
			UpDownDis struct {
				UpNum   int `json:"up_num"`
				DownNum int `json:"down_num"`
			} `json:"up_down_dis"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &raw); err != nil { return string(resp.Body()), nil }

	var b strings.Builder
	b.WriteString("## 主要指数\n\n| 名称 | 最新价 | 涨跌幅 |\n|---|---|---|\n")
	for _, idx := range raw.Data.IndexQuote {
		b.WriteString(fmt.Sprintf("| %s | %.2f | %.2f%% |\n", idx.SecuName, idx.LastPrice, idx.ChgPct*100))
	}
	b.WriteString(fmt.Sprintf("\n## 涨跌分布\n\n上涨:%d 下跌:%d", raw.Data.UpDownDis.UpNum, raw.Data.UpDownDis.DownNum))
	return b.String(), nil
}
var _ tool.InvokableTool = (*MarketDataTool)(nil)