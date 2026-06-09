package sina

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	bizmarket "eino-stock/internal/biz/market"

	"github.com/go-resty/resty/v2"
)

type KLineClient struct {
	client *resty.Client
}

func NewKLineClient(timeout time.Duration) *KLineClient {
	return &KLineClient{
		client: resty.New().
			SetTimeout(timeout).
			SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
			SetHeader("Accept", "*/*").
			SetHeader("Referer", "https://finance.sina.com.cn").
			SetHeader("Accept-Language", "zh-CN,zh;q=0.9"),
	}
}

func (c *KLineClient) GetKLines(ctx context.Context, code string, ktype bizmarket.KLineType, limit int) ([]*bizmarket.KLine, error) {
	symbol := sinaSymbol(code)
	if symbol == "" {
		return nil, fmt.Errorf("sina: invalid stock code %q", code)
	}
	scale := klineScale(ktype)
	if scale == "" {
		return nil, fmt.Errorf("sina: unsupported kline type %q", ktype)
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 1023 {
		limit = 1023
	}

	ts := time.Now().UnixMilli()
	callback := fmt.Sprintf("callback_%d", ts)
	baseURL := "https://quotes.sina.cn/cn/api/jsonp_v2.php/" + callback + "/CN_MarketDataService.getKLineData"
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("scale", scale)
	params.Set("ma", "no")
	params.Set("datalen", fmt.Sprintf("%d", limit))
	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := c.client.R().SetContext(ctx).Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("sina HTTP: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("sina HTTP %d", resp.StatusCode())
	}

	items := parseJSONP(resp.Body())
	if items == nil {
		return nil, fmt.Errorf("sina: failed to parse JSONP response")
	}

	kLines := make([]*bizmarket.KLine, 0, len(items))
	for _, item := range items {
		kLines = append(kLines, &bizmarket.KLine{
			Day:           item.Day,
			Open:          item.Open,
			Close:         item.Close,
			High:          item.High,
			Low:           item.Low,
			Volume:        item.Volume,
			Amount:        item.Amount,
			ChangePercent: item.ChangePercent,
		})
	}
	return kLines, nil
}

type sinaKLineItem struct {
	Day           string `json:"day"`
	Open          string `json:"open"`
	High          string `json:"high"`
	Low           string `json:"low"`
	Close         string `json:"close"`
	Volume        string `json:"volume"`
	Amount        string `json:"amount"`
	ChangePercent string `json:"changePercent"`
}

func sinaSymbol(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if strings.HasPrefix(code, "SH") || strings.HasPrefix(code, "SZ") || strings.HasPrefix(code, "BJ") {
		return strings.ToLower(code[:2]) + code[2:]
	}
	if strings.Contains(code, ".") {
		parts := strings.Split(code, ".")
		if len(parts) == 2 {
			switch parts[1] {
			case "SH", "SS":
				return "sh" + parts[0]
			case "SZ":
				return "sz" + parts[0]
			case "BJ":
				return "bj" + parts[0]
			}
		}
	}
	if len(code) >= 1 && code[0] >= '0' && code[0] <= '9' {
		switch code[0] {
		case '6':
			return "sh" + code
		case '0', '3':
			return "sz" + code
		case '8', '9':
			return "bj" + code
		}
	}
	return strings.ToLower(code)
}

func klineScale(ktype bizmarket.KLineType) string {
	switch ktype {
	case bizmarket.KLine1Min:
		return "1"
	case bizmarket.KLine5Min:
		return "5"
	case bizmarket.KLine15Min:
		return "15"
	case bizmarket.KLine30Min:
		return "30"
	case bizmarket.KLine60Min:
		return "60"
	case bizmarket.KLineDay:
		return "240"
	case bizmarket.KLineWeek:
		return "1200"
	default:
		return ""
	}
}

var jsonpArrayStart = regexp.MustCompile(`\[\s*\{`)

func parseJSONP(body []byte) []sinaKLineItem {
	trimmed := strings.TrimSpace(string(body))
	loc := jsonpArrayStart.FindStringIndex(trimmed)
	if len(loc) != 2 {
		return nil
	}
	end := strings.LastIndex(trimmed, "]")
	if end < loc[0] {
		return nil
	}
	jsonStr := trimmed[loc[0] : end+1]
	var items []sinaKLineItem
	if err := json.Unmarshal([]byte(jsonStr), &items); err != nil {
		return nil
	}
	return items
}