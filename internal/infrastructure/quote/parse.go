package quote

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	bizmarket "eino-stock/internal/biz/market"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// GB18030ToUTF8 新浪/腾讯行情响应转 UTF-8。
func GB18030ToUTF8(bs []byte) string {
	reader := transform.NewReader(bytes.NewReader(bs), simplifiedchinese.GB18030.NewDecoder())
	d, err := io.ReadAll(reader)
	if err != nil {
		return string(bs)
	}
	return string(d)
}

func parseSinaLine(data string) (*bizmarket.Quote, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return nil, fmt.Errorf("empty line")
	}
	parts := strings.SplitN(data, "=", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid format")
	}
	prefix := parts[0]
	raw := strings.Trim(parts[1], "\";")

	var fields map[string]string
	var err error
	switch {
	case strings.Contains(prefix, "hq_str_sz"), strings.Contains(prefix, "hq_str_sh"),
		strings.Contains(prefix, "hq_str_bj"), strings.Contains(prefix, "hq_str_sb"):
		fields, err = parseSHSZ(raw, strings.TrimPrefix(prefix, "hq_str_"))
	case strings.Contains(prefix, "hq_str_hk"):
		fields, err = parseHK(raw, strings.TrimPrefix(prefix, "hq_str_"))
	case strings.Contains(prefix, "hq_str_gb"):
		fields, err = parseUS(raw, strings.TrimPrefix(prefix, "hq_str_"))
	default:
		return nil, fmt.Errorf("unknown prefix: %s", prefix)
	}
	if err != nil {
		return nil, err
	}
	return mapToQuote(fields), nil
}

func parseTencentLine(data string) (*bizmarket.Quote, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return nil, fmt.Errorf("empty line")
	}
	parts := strings.SplitN(data, "=", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid format")
	}
	code := strings.TrimPrefix(strings.TrimPrefix(parts[0], "v_r_"), "v_")
	raw := strings.Trim(parts[1], "\"")
	items := strings.Split(raw, "~")
	if len(items) < 35 {
		return nil, fmt.Errorf("invalid tencent data")
	}
	fields := map[string]string{
		"code":      code,
		"name":      items[1],
		"price":     items[3],
		"pre_close": items[4],
		"open":      items[5],
		"high":      items[33],
		"low":       items[34],
	}
	if strings.Contains(items[30], "/") {
		t := strings.ReplaceAll(items[30], "/", "-")
		ts := strings.Fields(t)
		if len(ts) >= 2 {
			fields["date"] = ts[0]
			fields["time"] = ts[1]
		}
	} else if len(items[29]) >= 14 {
		fields["date"] = items[29][0:4] + "-" + items[29][4:6] + "-" + items[29][6:8]
		fields["time"] = items[29][8:10] + ":" + items[29][10:12] + ":" + items[29][12:14]
		fields["high"] = items[32]
		fields["low"] = items[33]
	}
	return mapToQuote(fields), nil
}

func parseSHSZ(raw, code string) (map[string]string, error) {
	items := strings.Split(raw, ",")
	if len(items) < 32 {
		return nil, fmt.Errorf("invalid shsz data")
	}
	return map[string]string{
		"code":      code,
		"name":      items[0],
		"open":      items[1],
		"pre_close": items[2],
		"price":     items[3],
		"high":      items[4],
		"low":       items[5],
		"volume":    items[8],
		"amount":    items[9],
		"date":      items[30],
		"time":      strings.Trim(items[31], "\";"),
	}, nil
}

func parseHK(raw, code string) (map[string]string, error) {
	items := strings.Split(raw, ",")
	if len(items) < 19 {
		return nil, fmt.Errorf("invalid hk data")
	}
	return map[string]string{
		"code":      code,
		"name":      items[1],
		"open":      items[2],
		"pre_close": items[3],
		"high":      items[4],
		"low":       items[5],
		"price":     items[6],
		"date":      strings.ReplaceAll(items[17], "/", "-"),
		"time":      strings.Trim(items[18], "\";"),
	}, nil
}

func parseUS(raw, code string) (map[string]string, error) {
	items := strings.Split(raw, ",")
	if len(items) < 4 {
		return nil, fmt.Errorf("invalid us data")
	}
	preClose := items[len(items)-1]
	if len(items) >= 36 {
		preClose = strings.Trim(items[26], "\";")
	}
	ts := strings.Fields(items[3])
	date, tm := "", ""
	if len(ts) >= 2 {
		date, tm = ts[0], ts[1]
	}
	return map[string]string{
		"code":      code,
		"name":      items[0],
		"price":     items[1],
		"open":      items[5],
		"pre_close": preClose,
		"high":      items[6],
		"low":       items[7],
		"date":      date,
		"time":      tm,
	}, nil
}

func mapToQuote(fields map[string]string) *bizmarket.Quote {
	price, _ := strconv.ParseFloat(fields["price"], 64)
	preClose, _ := strconv.ParseFloat(fields["pre_close"], 64)
	changePrice := price - preClose
	changePercent := 0.0
	if preClose > 0 {
		changePercent = changePrice / preClose * 100
	}
	return &bizmarket.Quote{
		Code:          fields["code"],
		Name:          fields["name"],
		Price:         fields["price"],
		Open:          fields["open"],
		PreClose:      fields["pre_close"],
		High:          fields["high"],
		Low:           fields["low"],
		ChangePrice:   changePrice,
		ChangePercent: changePercent,
		Date:          fields["date"],
		Time:          fields["time"],
		Volume:        fields["volume"],
		Amount:        fields["amount"],
	}
}
