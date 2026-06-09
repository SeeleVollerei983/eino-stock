package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/go-resty/resty/v2"
	"gopkg.in/yaml.v3"
)

type SelectAStockTool struct {
	client  *resty.Client
	qgqpBId string
}

func NewSelectAStockTool() *SelectAStockTool {
	qgqpBId := os.Getenv("QGQP_B_ID")
	if qgqpBId == "" {
		paths := []string{"configs/config.yaml", "../../configs/config.yaml"}
		for _, p := range paths {
			data, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			var raw struct {
				DataSource *struct {
					QgqpBId string `yaml:"qgqp_b_id"`
				} `yaml:"data_source"`
			}
			if err := yaml.Unmarshal(data, &raw); err == nil && raw.DataSource != nil && raw.DataSource.QgqpBId != "" {
				qgqpBId = raw.DataSource.QgqpBId
				break
			}
		}
	}
	return &SelectAStockTool{
		client: resty.New().
			SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:145.0) Gecko/20100101 Firefox/145.0").
			SetHeader("Origin", "https://xuangu.eastmoney.com").
			SetHeader("Referer", "https://xuangu.eastmoney.com/"),
		qgqpBId: qgqpBId,
	}
}

func (t *SelectAStockTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "SelectAStock",
		Desc: "A股智能选股。通过自然语言查询进行A股股票筛选，支持行情指标、技术形态、财务指标、行业概念等多条件组合筛选。输入选股条件即可。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"words": {
				Type:     schema.String,
				Desc:     "选股条件描述，如：非ST非退市股;换手率前200;市盈率小于20",
				Required: true,
			},
		}),
	}, nil
}

func (t *SelectAStockTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var params struct {
		Words string `json:"words"`
	}
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("parse params: %w", err)
	}
	if t.qgqpBId == "" {
		return "请先在 config.yaml 中配置 qgqp_b_id", nil
	}

	body := fmt.Sprintf(`{
		"keyWord": "%s", "pageSize": 20, "pageNo": 1,
		"fingerprint": "%s", "gids": [], "matchWord": "",
		"timestamp": %d, "shareToGuba": false, "requestId": "",
		"needCorrect": true, "removedConditionIdList": [],
		"xcId": "", "ownSelectAll": false, "dxInfo": [],
		"extraCondition": ""
	}`, params.Words, t.qgqpBId, time.Now().Unix())

	resp, err := t.client.R().SetContext(ctx).
		SetHeader("Host", "np-tjxg-g.eastmoney.com").
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("https://np-tjxg-g.eastmoney.com/api/smart-tag/stock/v3/pw/search-code")
	if err != nil {
		return "", fmt.Errorf("eastmoney: %w", err)
	}

	var raw struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data *struct {
			Result *struct {
				DataList []map[string]interface{} `json:"dataList"`
			} `json:"result"`
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &raw); err != nil {
		return fmt.Sprintf("解析失败: %s", string(resp.Body())), nil
	}
	if raw.Data == nil || raw.Data.Result == nil {
		return fmt.Sprintf("查询无结果: %s", raw.Msg), nil
	}

	var rows []string
	for _, item := range raw.Data.Result.DataList {
		code := getStr(item, "SECURITY_CODE", "code")
		name := getStr(item, "SECURITY_SHORT_NAME", "name")
		price := getStr(item, "NEWEST_PRICE")
		chg := getStr(item, "CHG")
		if code != "" && name != "" {
			if price != "" {
				rows = append(rows, fmt.Sprintf("%s(%s) 最新价:%s 涨跌幅:%s%%", name, code, price, chg))
			} else {
				rows = append(rows, fmt.Sprintf("%s(%s)", name, code))
			}
		}
	}
	if len(rows) == 0 {
		return "未找到符合条件的股票", nil
	}
	return fmt.Sprintf("找到 %d 只股票:\n%s", len(rows), strings.Join(rows, "\n")), nil
}

func getStr(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

var _ tool.InvokableTool = (*SelectAStockTool)(nil)