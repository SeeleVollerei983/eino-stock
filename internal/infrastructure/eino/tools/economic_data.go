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

// QueryEconomicDataTool 宏观经济数据查询工具。
type QueryEconomicDataTool struct {
	client *resty.Client
}

func NewQueryEconomicDataTool() *QueryEconomicDataTool {
	return &QueryEconomicDataTool{
		client: resty.New().SetTimeout(15 * time.Second).SetHeader("User-Agent", "Mozilla/5.0"),
	}
}

func (t *QueryEconomicDataTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "QueryEconomicData",
		Desc: "查询宏观经济数据(GDP,CPI,PPI,PMI)。参数flag=all返回全部，或指定GDP/CPI/PPI/PMI返回对应数据。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"flag": {
				Type:    schema.String,
				Desc:    "all=全部;GDP=国内生产总值;CPI=居民消费价格指数;PPI=工业品出厂价格指数;PMI=采购经理人指数",
				Required: false,
			},
		}),
	}, nil
}

func (t *QueryEconomicDataTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var p struct{ Flag string `json:"flag"` }
	json.Unmarshal([]byte(input), &p)
	flag := strings.ToUpper(strings.TrimSpace(p.Flag))

	var result strings.Builder
	result.WriteString("## 宏观经济数据\n\n")

	all := flag == "" || flag == "ALL"

	switch flag {
	case "GDP":
		result.WriteString(fetchMacroTable(ctx, t.client, "RPT_ECONOMY_GDP", "GDP 国内生产总值"))
	case "CPI":
		result.WriteString(fetchMacroTable(ctx, t.client, "RPT_ECONOMY_CPI", "CPI 居民消费价格指数"))
	case "PPI":
		result.WriteString(fetchMacroTable(ctx, t.client, "RPT_ECONOMY_PPI", "PPI 工业品出厂价格指数"))
	case "PMI":
		result.WriteString(fetchMacroTable(ctx, t.client, "RPT_ECONOMY_PMI", "PMI 采购经理人指数"))
	default:
		if !all {
			result.WriteString("未知参数，支持: GDP, CPI, PPI, PMI, ALL")
		}
	}

	if all {
		result.WriteString(fetchMacroTable(ctx, t.client, "RPT_ECONOMY_GDP", "GDP 国内生产总值"))
		result.WriteString(fetchMacroTable(ctx, t.client, "RPT_ECONOMY_CPI", "CPI 居民消费价格指数"))
		result.WriteString(fetchMacroTable(ctx, t.client, "RPT_ECONOMY_PPI", "PPI 工业品出厂价格指数"))
		result.WriteString(fetchMacroTable(ctx, t.client, "RPT_ECONOMY_PMI", "PMI 采购经理人指数"))
	}

	return result.String(), nil
}

func fetchMacroTable(ctx context.Context, client *resty.Client, reportName, title string) string {
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=%s&columns=REPORT_DATE,INDICATOR_ID,INDICATOR_VALUE&pageNumber=1&pageSize=10&sortTypes=-1&sortColumns=REPORT_DATE&source=WEB&client=WEB&filter=&_=%d",
		reportName, time.Now().UnixMilli())
	resp, err := client.R().SetContext(ctx).
		SetHeader("Referer", "https://data.eastmoney.com/").
		Get(url)
	if err != nil {
		return ""
	}
	var raw struct {
		Success bool `json:"success"`
		Data    *struct {
			Items []struct {
				REPORT_DATE     string  `json:"REPORT_DATE"`
				INDICATOR_VALUE float64 `json:"INDICATOR_VALUE"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &raw); err != nil || !raw.Success || raw.Data == nil || len(raw.Data.Items) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("### %s\n\n", title))
	b.WriteString("| 日期 | 数值 |\n|---|---|\n")
	for _, item := range raw.Data.Items {
		b.WriteString(fmt.Sprintf("| %s | %.2f |\n", item.REPORT_DATE, item.INDICATOR_VALUE))
	}
	b.WriteString("\n")
	return b.String()
}

var _ tool.InvokableTool = (*QueryEconomicDataTool)(nil)
