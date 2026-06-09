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

type MinuteDataTool struct {
	client *resty.Client
}

func NewMinuteDataTool() *MinuteDataTool {
	return &MinuteDataTool{client: resty.New().SetTimeout(10 * time.Second).SetHeader("User-Agent", "Mozilla/5.0")}
}

func (t *MinuteDataTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetStockMinuteData",
		Desc: "获取股票当日分时成交数据，包括每分钟的价格和成交量。输入股票代码即可。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"code": {Type: schema.String, Desc: "股票代码，如 600519.SH 或 000001", Required: true},
		}),
	}, nil
}

func (t *MinuteDataTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var p struct{ Code string `json:"code"` }
	if err := json.Unmarshal([]byte(input), &p); err != nil { return "", err }
	if p.Code == "" { return "请提供股票代码", nil }

	code := strings.ToLower(strings.ReplaceAll(p.Code, ".", ""))
	if !strings.HasPrefix(code, "sh") && !strings.HasPrefix(code, "sz") {
		if code[0] == '6' { code = "sh" + code } else { code = "sz" + code }
	}

	url := fmt.Sprintf("https://web.ifzq.gtimg.cn/appstock/app/minute/query?code=%s", code)
	resp, err := t.client.R().SetContext(ctx).SetHeader("Host", "web.ifzq.gtimg.cn").Get(url)
	if err != nil { return "", fmt.Errorf("tencent: %w", err) }

	var raw map[string]any
	if err := json.Unmarshal(resp.Body(), &raw); err != nil { return string(resp.Body()), nil }

	code = strings.Replace(code, "sh", "", 1)
	code = strings.Replace(code, "sz", "", 1)
	data, _ := raw[code].(map[string]any)
	minutes, _ := data["data"].(map[string]any)
	mins, _ := minutes["mins"].([]any)
	prices, _ := minutes["price"].([]any)
	volumes, _ := minutes["volume"].([]any)

	if len(mins) == 0 { return "暂未获取到分时数据（非交易时段或无数据）", nil }

	var rows []string
	for i := 0; i < len(mins) && i < 240; i++ {
		tm := fmt.Sprintf("%v", mins[i])
		pr := fmt.Sprintf("%v", prices[i])
		vo := fmt.Sprintf("%v", volumes[i])
		rows = append(rows, fmt.Sprintf("%s\t价格:%s\t成交量:%s", tm, pr, vo))
	}
	return fmt.Sprintf("当日分时数据 (%d条):\n%s", len(rows), strings.Join(rows, "\n")), nil
}

var _ tool.InvokableTool = (*MinuteDataTool)(nil)