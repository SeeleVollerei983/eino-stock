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

type LongTigerListTool struct{ client *resty.Client }

func NewLongTigerListTool() *LongTigerListTool {
	return &LongTigerListTool{client: resty.New().SetTimeout(10*time.Second).SetHeader("User-Agent", "Mozilla/5.0")}
}

func (t *LongTigerListTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetLongTigerList",
		Desc: "获取龙虎榜数据，包括营业部买入排行。可选参数date(日期YYYYMMDD，默认今日)。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"date": {Type: schema.String, Desc: "日期(YYYYMMDD,可选)", Required: false},
		}),
	}, nil
}

func (t *LongTigerListTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var p struct{ Date string `json:"date"` }
	json.Unmarshal([]byte(input), &p)
	date := p.Date
	if date == "" { date = time.Now().AddDate(0,0,-1).Format("20060102") }

	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_DAILYBILLBOARD_DEPLIST&columns=ALL&pageNumber=1&pageSize=20&sortTypes=-1&sortColumns=BUY_AMOUNT&source=WEB&client=WEB&filter=(TRADE_DATE%%3D%%27%s%%27)", date)
	resp, err := t.client.R().SetContext(ctx).SetHeader("Referer", "https://data.eastmoney.com/").Get(url)
	if err != nil { return "", fmt.Errorf("eastmoney: %w", err) }

	var raw struct {
		Result *struct {
			Data []struct {
				OperateDeptName string  `json:"OPERATE_DEPT_NAME"`
				StockName       string  `json:"SECURITY_NAME"`
				StockCode       string  `json:"SECURITY_CODE"`
				BuyAmount       float64 `json:"BUY_AMOUNT"`
				SellAmount      float64 `json:"SELL_AMOUNT"`
				NetAmount       float64 `json:"NET_AMOUNT"`
			} `json:"data"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp.Body(), &raw); err != nil { return string(resp.Body()), nil }
	if raw.Result == nil || len(raw.Result.Data) == 0 { return fmt.Sprintf("%s暂无龙虎榜数据", date), nil }

	var b strings.Builder
	b.WriteString(fmt.Sprintf("## 龙虎榜营业部排名 (%s)\n\n", date))
	b.WriteString("| 营业部 | 股票 | 买入 | 卖出 | 净额 |\n|---|---|---|---|---|\n")
	for _, d := range raw.Result.Data {
		b.WriteString(fmt.Sprintf("| %s | %s(%s) | %.2f万 | %.2f万 | %.2f万 |\n", d.OperateDeptName, d.StockName, d.StockCode, d.BuyAmount/1e4, d.SellAmount/1e4, d.NetAmount/1e4))
	}
	return b.String(), nil
}
var _ tool.InvokableTool = (*LongTigerListTool)(nil)