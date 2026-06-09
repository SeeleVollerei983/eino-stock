package eastmoney

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	bizmarket "eino-stock/internal/biz/market"

	"github.com/go-resty/resty/v2"
)

type KLineClient struct {
	client *resty.Client
}

func NewKLineClient(timeout time.Duration) *KLineClient {
	client := resty.New()
	client.SetTransport(&http.Transport{
		DisableCompression: true,
	})
	client.SetTimeout(timeout)
	client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	client.SetHeader("Accept", "*/*")
	client.SetHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	client.SetHeader("Referer", "https://quote.eastmoney.com")
	client.SetHeader("Connection", "keep-alive")
	return &KLineClient{client: client}
}

func (c *KLineClient) GetKLines(ctx context.Context, code string, ktype bizmarket.KLineType, limit int) ([]*bizmarket.KLine, error) {
	secID := convertStockCode(code)
	if secID == "" {
		return nil, fmt.Errorf("eastmoney: invalid stock code %q", code)
	}
	if limit <= 0 {
		limit = 100
	}

	fields := "f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61,f116"
	baseURL := "https://push2his.eastmoney.com/api/qt/stock/kline/get"
	params := url.Values{}
	params.Set("secid", secID)
	params.Set("klt", string(ktype))
	params.Set("fqt", "0")
	params.Set("end", "20500101")
	params.Set("lmt", fmt.Sprintf("%d", limit))
	params.Set("fields1", "f1,f2,f3,f4,f5,f6")
	params.Set("fields2", fields)
	params.Set("wbp2u", "|0|0|0|web")
	params.Set("_", fmt.Sprintf("%d", time.Now().UnixMilli()))
	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := c.client.R().SetContext(ctx).Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("eastmoney HTTP: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("eastmoney HTTP %d", resp.StatusCode())
	}

	body := decompressBody(resp)
	var result struct {
		Rc      int    `json:"rc"`
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *struct {
			Klines []string `json:"klines"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("eastmoney JSON: %w", err)
	}
	if result.Rc != 0 {
		return nil, fmt.Errorf("eastmoney API error: rc=%d code=%d %s", result.Rc, result.Code, result.Message)
	}
	if result.Data == nil {
		return nil, fmt.Errorf("eastmoney: empty data")
	}

	kLines := make([]*bizmarket.KLine, 0, len(result.Data.Klines))
	for _, ks := range result.Data.Klines {
		parts := strings.Split(ks, ",")
		if len(parts) < 11 {
			continue
		}
		kLines = append(kLines, &bizmarket.KLine{
			Day:           parts[0],
			Open:          parts[1],
			Close:         parts[2],
			High:          parts[3],
			Low:           parts[4],
			Volume:        parts[5],
			Amount:        parts[6],
			ChangePercent: parts[8],
			ChangeValue:   parts[9],
			TurnoverRate:  parts[10],
		})
	}
	return kLines, nil
}

func convertStockCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if strings.Contains(code, ".") {
		parts := strings.Split(code, ".")
		if len(parts) == 2 {
			switch parts[1] {
			case "SH", "SS":
				return "1." + parts[0]
			case "SZ":
				return "0." + parts[0]
			case "BJ":
				return "0." + parts[0]
			case "HK":
				return "128." + parts[0]
			}
		}
	}
	if strings.HasPrefix(code, "SH") || strings.HasPrefix(code, "SZ") || strings.HasPrefix(code, "BJ") {
		market := code[:2]
		pure := code[2:]
		switch market {
		case "SH":
			return "1." + pure
		case "SZ", "BJ":
			return "0." + pure
		}
	}
	if len(code) >= 1 && code[0] >= '0' && code[0] <= '9' {
		switch code[0] {
		case '6':
			return "1." + code
		case '0', '3', '8', '9':
			return "0." + code
		}
	}
	if strings.HasPrefix(code, "HK") {
		return "128." + code[2:]
	}
	return ""
}

func decompressBody(resp *resty.Response) []byte {
	raw := resp.Body()
	enc := resp.Header().Get("Content-Encoding")
	if strings.ToLower(enc) != "gzip" {
		return raw
	}
	reader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return raw
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return raw
	}
	return data
}