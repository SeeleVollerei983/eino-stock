package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/go-resty/resty/v2"
)

type ResearchReportTool struct{ client *resty.Client }

func NewResearchReportTool() *ResearchReportTool {
	return &ResearchReportTool{client: resty.New().SetTimeout(10 * time.Second).SetHeader("User-Agent", "Mozilla/5.0")}
}

func (t *ResearchReportTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetStockResearchReport",
		Desc: "获取券商研究报告。输入股票代码，可选参数days(近N天，默认30)。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"code": {Type: schema.String, Desc: "股票代码", Required: true},
			"days": {Type: schema.String, Desc: "近N天(可选，默认30)", Required: false},
		}),
	}, nil
}

func (t *ResearchReportTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var p struct{ Code, Days string }
	if err := json.Unmarshal([]byte(input), &p); err != nil { return "", err }
	if p.Code == "" { return "请提供股票代码", nil }
	days := 30
	if p.Days != "" { if d, e := strconv.Atoi(p.Days); e == nil && d > 0 { days = d } }
	code := strings.ReplaceAll(p.Code, ".", "")
	secu := code
	if strings.HasPrefix(code, "6") || strings.HasPrefix(code, "5") { secu = code + ".SH" } else { secu = code + ".SZ" }
	bd := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	ed := time.Now().Format("2006-01-02")
	url := fmt.Sprintf("https://reportapi.eastmoney.com/report/list?pageSize=15&beginTime=%s&endTime=%s&pageNo=1&code=%s", bd, ed, secu)
	resp, err := t.client.R().SetContext(ctx).SetHeader("Host", "reportapi.eastmoney.com").SetHeader("Referer", "https://data.eastmoney.com/report/").Get(url)
	if err != nil { return "", fmt.Errorf("report: %w", err) }
	var raw struct{ Data []struct{ Title, OrgName, PublishDate, Predict string } }
	if err := json.Unmarshal(resp.Body(), &raw); err != nil { return string(resp.Body()), nil }
	if len(raw.Data) == 0 { return fmt.Sprintf("近%d天暂无研报", days), nil }
	var rows []string
	for _, r := range raw.Data { rows = append(rows, fmt.Sprintf("[%s] %s - %s(%s)", r.PublishDate[:10], r.Title, r.OrgName, r.Predict)) }
	return fmt.Sprintf("近%d天研报 (%d篇):\n%s", days, len(rows), strings.Join(rows, "\n")), nil
}
var _ tool.InvokableTool = (*ResearchReportTool)(nil)