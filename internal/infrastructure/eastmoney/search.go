package eastmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"eino-stock/internal/biz/market"

	"github.com/go-resty/resty/v2"
)

// SearchClient 东财搜索仓储 — 实现 market.MarketRepo。
type SearchClient struct {
	client  *resty.Client
	qgqpBId string
}

func NewSearchClient(qgqpBId string) *SearchClient {
	return &SearchClient{
		client: resty.New().
			SetTimeout(15 * time.Second).
			SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:145.0) Gecko/20100101 Firefox/145.0").
			SetHeader("Origin", "https://xuangu.eastmoney.com").
			SetHeader("Referer", "https://xuangu.eastmoney.com/"),
		qgqpBId: qgqpBId,
	}
}

func (s *SearchClient) SearchStocks(ctx context.Context, keyword string, limit int) ([]*market.Stock, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	if s.qgqpBId == "" {
		return nil, fmt.Errorf("eastmoney: QGQP_B_ID not configured")
	}

	url := "https://np-tjxg-g.eastmoney.com/api/smart-tag/stock/v3/pw/search-code"
	body := fmt.Sprintf(`{
		"keyWord": "%s", "pageSize": %d, "pageNo": 1,
		"fingerprint": "%s", "gids": [], "matchWord": "",
		"timestamp": %d, "shareToGuba": false, "requestId": "",
		"needCorrect": true, "removedConditionIdList": [],
		"xcId": "", "ownSelectAll": false, "dxInfo": [],
		"extraCondition": ""
	}`, keyword, limit, s.qgqpBId, time.Now().UnixMilli())

	resp, err := s.client.R().SetContext(ctx).
		SetHeader("Host", "np-tjxg-g.eastmoney.com").
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(url)
	if err != nil {
		return nil, fmt.Errorf("eastmoney search: %w", err)
	}

	var raw struct {
		Data *struct {
			Result *struct {
				DataList []struct {
					Code   string `json:"SECURITY_CODE"`
					Name   string `json:"SECURITY_SHORT_NAME"`
					Market string `json:"TRADING_MARKET"`
				} `json:"dataList"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &raw); err != nil {
		return nil, fmt.Errorf("parse search: %w", err)
	}
	if raw.Data == nil || raw.Data.Result == nil {
		return nil, nil
	}

	stocks := make([]*market.Stock, 0, len(raw.Data.Result.DataList))
	for _, item := range raw.Data.Result.DataList {
		symbol := item.Code
		if len(symbol) >= 6 {
			symbol = symbol[:6]
		}
		stocks = append(stocks, &market.Stock{
			Symbol: symbol,
			TsCode: item.Code,
			Name:   item.Name,
			Market: item.Market,
		})
	}
	return stocks, nil
}
