package eastmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	bizscreen "eino-stock/internal/biz/screen"

	"github.com/go-resty/resty/v2"
)

type ScreenClient struct {
	client  *resty.Client
	qgqpBId string
}

func NewScreenClient(qgqpBId string, timeout time.Duration) *ScreenClient {
	if qgqpBId == "" {
		qgqpBId = os.Getenv("QGQP_B_ID")
	}
	return &ScreenClient{
		client: resty.New().
			SetTimeout(timeout).
			SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:145.0) Gecko/20100101 Firefox/145.0").
			SetHeader("Origin", "https://xuangu.eastmoney.com").
			SetHeader("Referer", "https://xuangu.eastmoney.com/"),
		qgqpBId: qgqpBId,
	}
}

func (c *ScreenClient) search(ctx context.Context, url, keyword string, pageSize int) ([]map[string]interface{}, error) {
	if c.qgqpBId == "" {
		return nil, fmt.Errorf("请先配置 QGQP_B_ID 环境变量")
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	body := fmt.Sprintf(`{
		"keyWord": "%s", "pageSize": %d, "pageNo": 1,
		"fingerprint": "%s", "gids": [], "matchWord": "",
		"timestamp": %d, "shareToGuba": false, "requestId": "",
		"needCorrect": true, "removedConditionIdList": [],
		"xcId": "", "ownSelectAll": false, "dxInfo": [],
		"extraCondition": ""
	}`, keyword, pageSize, c.qgqpBId, time.Now().Unix())

	resp, err := c.client.R().SetContext(ctx).
		SetHeader("Host", "np-tjxg-g.eastmoney.com").
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP: %w", err)
	}

	var raw struct {
		Code string `json:"code"`
		Data *struct {
			Result *struct {
				DataList []map[string]interface{} `json:"dataList"`
			} `json:"result"`
		} `json:"data"`
	}
	fmt.Printf("DEBUG ETF resp=%s\\n", string(resp.Body()))
	if err := json.Unmarshal(resp.Body(), &raw); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	if raw.Data == nil || raw.Data.Result == nil {
		return nil, fmt.Errorf("no data in response: code=%s", raw.Code)
	}
	return raw.Data.Result.DataList, nil
}

func (c *ScreenClient) SearchBk(ctx context.Context, keyword string, pageSize int) ([]*bizscreen.BkItem, error) {
	items, err := c.search(ctx, "https://np-tjxg-b.eastmoney.com/api/smart-tag/bkc/v3/pw/search-code", keyword, pageSize)
	if err != nil {
		return nil, err
	}
	out := make([]*bizscreen.BkItem, 0, len(items))
	for _, item := range items {
		out = append(out, &bizscreen.BkItem{
			Code: getString(item, "SECURITY_CODE"),
			Name: getString(item, "SECURITY_SHORT_NAME"),
		})
	}
	return out, nil
}

func (c *ScreenClient) SearchStock(ctx context.Context, keyword string, pageSize int) ([]map[string]interface{}, error) {
	return c.search(ctx, "https://np-tjxg-g.eastmoney.com/api/smart-tag/stock/v3/pw/search-code", keyword, pageSize)
}

func (c *ScreenClient) SearchETF(ctx context.Context, keyword string, pageSize int) ([]*bizscreen.ETFItem, error) {
	items, err := c.search(ctx, "https://np-tjxg-b.eastmoney.com/api/smart-tag/etf/v3/pw/search-code", keyword, pageSize)
	if err != nil {
		return nil, err
	}
	out := make([]*bizscreen.ETFItem, 0, len(items))
	for _, item := range items {
		out = append(out, &bizscreen.ETFItem{
			Code: getString(item, "SECURITY_CODE"),
			Name: getString(item, "SECURITY_SHORT_NAME"),
		})
	}
	return out, nil
}

func (c *ScreenClient) HotStrategy(ctx context.Context) ([]*bizscreen.HotStrategyItem, error) {
	url := fmt.Sprintf("https://np-ipick.eastmoney.com/recommend/stock/heat/ranking?count=20&trace=%d&client=web&biz=web_smart_tag", time.Now().Unix())
	resp, err := c.client.R().SetContext(ctx).
		SetHeader("Host", "np-ipick.eastmoney.com").
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("HotStrategy: %w", err)
	}

	var result struct {
		Data []struct {
			Chg       float64 `json:"chg"`
			HeatValue int     `json:"heatValue"`
			Question  string  `json:"question"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("HotStrategy JSON: %w", err)
	}

	items := make([]*bizscreen.HotStrategyItem, 0, len(result.Data))
	for _, d := range result.Data {
		items = append(items, &bizscreen.HotStrategyItem{
			Question: d.Question,
			Chg:      d.Chg * 100,
			Heat:     d.HeatValue,
		})
	}
	return items, nil
}

func getString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}