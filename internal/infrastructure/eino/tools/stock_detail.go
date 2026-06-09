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

type StockDetailTool struct{ client *resty.Client }

func NewStockDetailTool() *StockDetailTool {
	return &StockDetailTool{client: resty.New().SetTimeout(10 * time.Second).SetHeader("User-Agent", "Mozilla/5.0")}
}

func (t *StockDetailTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetStockDetail",
		Desc: "获取股票实时行情和五档盘口数据（买五卖五、开高低收、成交量）。输入股票代码即可。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"code": {Type: schema.String, Desc: "股票代码", Required: true},
		}),
	}, nil
}

func (t *StockDetailTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var p struct{ Code string `json:"code"` }
	if err := json.Unmarshal([]byte(input), &p); err != nil { return "", err }
	code := strings.ToLower(strings.TrimSpace(p.Code))
	if code == "" { return "请提供股票代码", nil }
	code = strings.ReplaceAll(code, ".", "")
	if !strings.HasPrefix(code, "sh") && !strings.HasPrefix(code, "sz") {
		if len(code) >= 6 && code[0] == '6' { code = "sh" + code } else { code = "sz" + code }
	}
	resp, err := t.client.R().SetContext(ctx).SetHeader("Host", "hq.sinajs.cn").SetHeader("Referer", "https://finance.sina.com.cn").Get("https://hq.sinajs.cn/list=" + code)
	if err != nil { return "", fmt.Errorf("sina: %w", err) }
	raw := string(resp.Body())
	s := strings.Index(raw, "\""); e := strings.LastIndex(raw, "\"")
	if s < 0 || e <= s+1 { return "未获取到数据", nil }
	f := strings.Split(raw[s+1:e], ",")
	if len(f) < 32 { return "数据不完整", nil }
	return fmt.Sprintf(`名称:%s  代码:%s
现价:%s  涨跌幅:%.2f%%
开盘:%s  最高:%s  最低:%s  昨收:%s
成交量:%s手  成交额:%s
买一:%s(%s)  卖一:%s(%s)
买二:%s(%s)  卖二:%s(%s)
买三:%s(%s)  卖三:%s(%s)
买四:%s(%s)  卖四:%s(%s)
买五:%s(%s)  卖五:%s(%s)`, f[0], code, f[3], (pf(f[3])-pf(f[2]))/pf(f[2])*100, f[1], f[4], f[5], f[2], f[8], f[9], f[11], f[12], f[21], f[22], f[13], f[14], f[23], f[24], f[15], f[16], f[25], f[26], f[17], f[18], f[27], f[28], f[19], f[20], f[29], f[30]), nil
}
func pf(s string) float64 { var v float64; fmt.Sscanf(s, "%f", &v); return v }
var _ tool.InvokableTool = (*StockDetailTool)(nil)