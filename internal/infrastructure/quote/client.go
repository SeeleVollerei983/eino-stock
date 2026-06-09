package quote

import (
	"context"
	"fmt"
	"strings"
	"time"

	bizmarket "eino-stock/internal/biz/market"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-resty/resty/v2"
)

const (
	sinaStockURL = "http://hq.sinajs.cn/rn=%d&list=%s"
	txStockURL   = "http://qt.gtimg.cn/?_=%d&q=%s"
)

// Client 新浪/腾讯实时行情客户端。
type Client struct {
	http *resty.Client
	log  *log.Helper
}

// NewClient 创建行情客户端。
func NewClient(timeout time.Duration, logger log.Logger) *Client {
	c := resty.New().
		SetTimeout(timeout).
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")
	return &Client{
		http: c,
		log:  log.NewHelper(logger),
	}
}

// GetRealtimeQuotes 批量获取实时行情。
func (c *Client) GetRealtimeQuotes(ctx context.Context, codes []string) ([]*bizmarket.Quote, error) {
	if len(codes) == 0 {
		return nil, nil
	}

	var quotes []*bizmarket.Quote

	txCodes := filterCodes(codes, func(code string) bool {
		p := strings.ToLower(code)
		return strings.HasPrefix(p, "hk") || strings.HasPrefix(p, "sh") || strings.HasPrefix(p, "sz")
	})
	if len(txCodes) > 0 {
		txList := joinCodes(txCodes, func(code string) string {
			p := strings.ToLower(code)
			if strings.HasPrefix(p, "hk") {
				return "r_" + p
			}
			return p
		})
		txQuotes, err := c.fetchTencent(ctx, txList)
		if err != nil {
			c.log.Warnf("tencent quote error: %v", err)
		} else {
			quotes = append(quotes, txQuotes...)
		}
	}

	sinaCodes := filterCodes(codes, func(code string) bool {
		p := strings.ToLower(code)
		return !strings.HasPrefix(p, "hk") && !strings.HasPrefix(p, "sh") && !strings.HasPrefix(p, "sz")
	})
	if len(sinaCodes) > 0 {
		sinaList := joinCodes(sinaCodes, func(code string) string {
			p := strings.ToLower(code)
			if strings.HasPrefix(p, "us") {
				return strings.Replace(p, "us", "gb_", 1)
			}
			return p
		})
		sinaQuotes, err := c.fetchSina(ctx, sinaList)
		if err != nil {
			c.log.Warnf("sina quote error: %v", err)
		} else {
			quotes = append(quotes, sinaQuotes...)
		}
	}

	return quotes, nil
}

func (c *Client) fetchSina(ctx context.Context, codes string) ([]*bizmarket.Quote, error) {
	url := fmt.Sprintf(sinaStockURL, time.Now().Unix(), codes)
	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Host", "hq.sinajs.cn").
		SetHeader("Referer", "https://finance.sina.com.cn/").
		Get(url)
	if err != nil {
		return nil, err
	}
	body := GB18030ToUTF8(resp.Body())
	lines := strings.Split(strings.TrimSpace(body), "\n")
	out := make([]*bizmarket.Quote, 0, len(lines))
	for _, line := range lines {
		q, err := parseSinaLine(line)
		if err != nil || q == nil {
			continue
		}
		out = append(out, q)
	}
	return out, nil
}

func (c *Client) fetchTencent(ctx context.Context, codes string) ([]*bizmarket.Quote, error) {
	url := fmt.Sprintf(txStockURL, time.Now().Unix(), codes)
	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Host", "qt.gtimg.cn").
		SetHeader("Referer", "https://gu.qq.com/").
		Get(url)
	if err != nil {
		return nil, err
	}
	body := GB18030ToUTF8(resp.Body())
	lines := strings.Split(strings.Trim(strings.TrimSpace(body), ";"), ";")
	out := make([]*bizmarket.Quote, 0, len(lines))
	for _, line := range lines {
		q, err := parseTencentLine(line)
		if err != nil || q == nil {
			continue
		}
		out = append(out, q)
	}
	return out, nil
}

func filterCodes(codes []string, keep func(string) bool) []string {
	out := make([]string, 0, len(codes))
	for _, code := range codes {
		if keep(code) {
			out = append(out, code)
		}
	}
	return out
}

func joinCodes(codes []string, mapFn func(string) string) string {
	parts := make([]string, 0, len(codes))
	for _, code := range codes {
		parts = append(parts, mapFn(code))
	}
	return strings.Join(parts, ",")
}
