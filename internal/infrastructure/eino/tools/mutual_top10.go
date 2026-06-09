package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/go-resty/resty/v2"
)

// MutualTop10Tool 互联互通十大成交股工具。
type MutualTop10Tool struct {
	client *resty.Client
}

func NewMutualTop10Tool() *MutualTop10Tool {
	return &MutualTop10Tool{
		client: resty.New().SetTimeout(15 * time.Second).SetHeader("User-Agent", "Mozilla/5.0"),
	}
}

func (t *MutualTop10Tool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetMutualTop10",
		Desc: "获取北向资金（沪股通、深股通）和南向资金（港股通）十大成交股数据。mutualType=001沪股通,002港股通(沪),003深股通,004港股通(深)。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"mutualType": {Type: schema.String, Desc: "001=沪股通,002=港股通(沪),003=深股通,004=港股通(深)", Required: true},
			"tradeDate":  {Type: schema.String, Desc: "交易日期 YYYY-MM-DD，默认最新", Required: false},
		}),
	}, nil
}

func (t *MutualTop10Tool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var p struct {
		MutualType string `json:"mutualType"`
		TradeDate  string `json:"tradeDate"`
	}
	if err := json.Unmarshal([]byte(input), &p); err != nil {
		return "", fmt.Errorf("parse: %w", err)
	}
	if p.MutualType == "" {
		p.MutualType = "001"
	}
	if p.TradeDate == "" {
		p.TradeDate = time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	}

	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/kamt.kamtkline/get?fields1=f1,f2,f3,f4,f5,f6&fields2=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f13&kamtType=%s&begintDate=%s&endDate=%s&pageNo=1&pageSize=10&secids=_&_=%d",
		p.MutualType, p.TradeDate, p.TradeDate, time.Now().UnixMilli())

	resp, err := t.client.R().SetContext(ctx).Get(url)
	if err != nil {
		return "", fmt.Errorf("http: %w", err)
	}

	var raw struct {
		Data *struct {
			KlineData []struct {
				Code  string  `json:"code"`
				Name  string  `json:"name"`
				Price float64 `json:"price"`
				Chg   float64 `json:"change"`
			} `json:"klineData"`
		} `json:"data"`
	}

	typeName := map[string]string{"001": "沪股通", "002": "港股通(沪)", "003": "深股通", "004": "港股通(深)"}

	result := fmt.Sprintf("## %s 十大成交股 (%s)\n\n", typeName[p.MutualType], p.TradeDate)
	if err := json.Unmarshal(resp.Body(), &raw); err != nil || raw.Data == nil {
		result += "暂未获取到数据（数据17:00-18:00左右更新）"
	} else {
		result += "| 代码 | 名称 | 成交价 | 涨跌幅 |\n|---|---|---|---|\n"
		for _, k := range raw.Data.KlineData {
			result += fmt.Sprintf("| %s | %s | %.2f | %.2f%% |\n", k.Code, k.Name, k.Price, k.Chg)
		}
	}
	return result, nil
}

var _ tool.InvokableTool = (*MutualTop10Tool)(nil)
