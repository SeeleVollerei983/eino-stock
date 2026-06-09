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

type StockNoticeTool struct{ client *resty.Client }

func NewStockNoticeTool() *StockNoticeTool {
	return &StockNoticeTool{client: resty.New().SetTimeout(10 * time.Second).SetHeader("User-Agent", "Mozilla/5.0")}
}

func (t *StockNoticeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetStockNotice",
		Desc: "获取上市公司公告列表。输入股票代码即可。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"code": {Type: schema.String, Desc: "股票代码", Required: true},
		}),
	}, nil
}

func (t *StockNoticeTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var p struct{ Code string `json:"code"` }
	if err := json.Unmarshal([]byte(input), &p); err != nil { return "", err }
	code := strings.ReplaceAll(p.Code, ".", "")
	if code == "" { return "请提供股票代码", nil }
	secu := code
	if strings.HasPrefix(code, "6") { secu = code + ".SH" } else { secu = code + ".SZ" }
	url := fmt.Sprintf("https://np-anotice-stock.eastmoney.com/api/security/announcement/page?sr=-1&page_size=15&page_index=1&stock_list=%s&f_node=0&s_node=0", secu)
	resp, err := t.client.R().SetContext(ctx).SetHeader("Host", "np-anotice-stock.eastmoney.com").SetHeader("Referer", "https://data.eastmoney.com/notices").Get(url)
	if err != nil { return "", fmt.Errorf("notice: %w", err) }
	var raw struct{ Data struct{ Total int; List []struct{ Title, NoticeDate string } } }
	if err := json.Unmarshal(resp.Body(), &raw); err != nil { return string(resp.Body()), nil }
	if len(raw.Data.List) == 0 { return "暂无公告", nil }
	var rows []string
	for _, n := range raw.Data.List { rows = append(rows, fmt.Sprintf("[%s] %s", n.NoticeDate[:10], n.Title)) }
	return fmt.Sprintf("最新公告 (%d条,共%d条):\n%s", len(rows), raw.Data.Total, strings.Join(rows, "\n")), nil
}
var _ tool.InvokableTool = (*StockNoticeTool)(nil)