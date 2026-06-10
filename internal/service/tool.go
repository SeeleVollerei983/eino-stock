package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"eino-stock/internal/infrastructure/eino/tools"

	"github.com/go-resty/resty/v2"
	"gopkg.in/yaml.v3"
)

func ToolScreen(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q"); if q == "" { writeError(w, 400, fmt.Errorf("missing q")); return }
	t := tools.NewSelectAStockTool()
	r2, e := t.InvokableRun(r.Context(), fmt.Sprintf(`{"words":"%s"}`, q))
	writeJSON(w, map[string]any{"result": r2}, e)
}

func ToolMinute(w http.ResponseWriter, r *http.Request) {
	c := r.URL.Query().Get("code"); if c == "" { writeError(w, 400, fmt.Errorf("missing code")); return }
	t := tools.NewMinuteDataTool()
	r2, e := t.InvokableRun(r.Context(), fmt.Sprintf(`{"code":"%s"}`, c))
	writeJSON(w, map[string]any{"result": r2}, e)
}

func ToolDetail(w http.ResponseWriter, r *http.Request) {
	c := r.URL.Query().Get("code"); if c == "" { writeError(w, 400, fmt.Errorf("missing code")); return }
	t := tools.NewStockDetailTool()
	r2, e := t.InvokableRun(r.Context(), fmt.Sprintf(`{"code":"%s"}`, c))
	writeJSON(w, map[string]any{"result": r2}, e)
}

func ToolNotice(w http.ResponseWriter, r *http.Request) {
	c := r.URL.Query().Get("code"); if c == "" { writeError(w, 400, fmt.Errorf("missing code")); return }
	t := tools.NewStockNoticeTool()
	r2, e := t.InvokableRun(r.Context(), fmt.Sprintf(`{"code":"%s"}`, c))
	writeJSON(w, map[string]any{"result": r2}, e)
}

func ToolReport(w http.ResponseWriter, r *http.Request) {
	c := r.URL.Query().Get("code"); if c == "" { writeError(w, 400, fmt.Errorf("missing code")); return }
	d := r.URL.Query().Get("days"); if d == "" { d = "30" }
	t := tools.NewResearchReportTool()
	r2, e := t.InvokableRun(r.Context(), fmt.Sprintf(`{"code":"%s","days":"%s"}`, c, d))
	writeJSON(w, map[string]any{"result": r2}, e)
}

func ToolGlobalIndexes(w http.ResponseWriter, r *http.Request) {
	t := tools.NewGlobalStockIndexesTool()
	r2, e := t.InvokableRun(r.Context(), "{}")
	writeJSON(w, map[string]any{"result": r2}, e)
}

func ToolHotPlates(w http.ResponseWriter, r *http.Request) {
	t := tools.NewUplimitHotPlatesTool()
	r2, e := t.InvokableRun(r.Context(), "{}")
	writeJSON(w, map[string]any{"result": r2}, e)
}

func ToolLongTiger(w http.ResponseWriter, r *http.Request) { t := tools.NewLongTigerListTool()
	r2, e := t.InvokableRun(r.Context(), "{}")
	writeJSON(w, map[string]any{"result": r2}, e)
}

func ToolIndustryValuation(w http.ResponseWriter, r *http.Request) {
	t := tools.NewIndustryValuationTool()
	r2, e := t.InvokableRun(r.Context(), "{}")
	writeJSON(w, map[string]any{"result": r2}, e)
}

func ToolIndustryMoneyRank(w http.ResponseWriter, r *http.Request) {
	t := tools.NewIndustryMoneyRankTool()
	r2, e := t.InvokableRun(r.Context(), "{}")
	writeJSON(w, map[string]any{"result": r2}, e)
}


func ToolScreenV2(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q"); if q == "" { writeError(w, 400, fmt.Errorf("missing q")); return }
	qid := readQgqpBIdFromFile(); if qid == "" { qid = os.Getenv("QGQP_B_ID") }
	if qid == "" { writeJSON(w, map[string]any{"error":"no qgqp_b_id"}, nil); return }
	bd := fmt.Sprintf(`{"keyWord":"%s","pageSize":50,"pageNo":1,"fingerprint":"%s","gids":[],"matchWord":"","timestamp":%d,"shareToGuba":false,"requestId":"","needCorrect":true,"removedConditionIdList":[],"xcId":"","ownSelectAll":false,"dxInfo":[],"extraCondition":""}`, q, qid, time.Now().Unix())
	cl := resty.New().SetTimeout(15*time.Second).SetHeader("User-Agent","Mozilla/5.0").SetHeader("Origin","https://xuangu.eastmoney.com").SetHeader("Referer","https://xuangu.eastmoney.com/")
	resp, e := cl.R().SetContext(r.Context()).SetHeader("Host","np-tjxg-g.eastmoney.com").SetHeader("Content-Type","application/json").SetBody(bd).Post("https://np-tjxg-g.eastmoney.com/api/smart-tag/stock/v3/pw/search-code")
	if e != nil { writeError(w, 500, e); return }
	var raw map[string]any; json.Unmarshal(resp.Body(), &raw)
	d, _ := raw["data"].(map[string]any)
	res, _ := d["result"].(map[string]any)
	dl, _ := res["dataList"].([]any)
	var out []map[string]any
	for _, item := range dl {
		it, _ := item.(map[string]any)
		out = append(out, map[string]any{
			"code": getS(it,"SECURITY_CODE"), "name": getS(it,"SECURITY_SHORT_NAME"),
			"price": getF(it,"NEWEST_PRICE"), "chg": getF(it,"CHG"),
			"turnover": getF(it,"TURNOVER_RATE"), "qrr": getF(it,"QRR"),
			"volume": getS(it,"VOLUME"), "amount": getS(it,"TRADING_VOLUMES"),
			"mvval": getS(it,"MVVAL"), "high": getF(it,"PEAK_PRICE"), "low": getF(it,"BOTTOM_PRICE"),
		})
	}
	writeJSON(w, out, nil)
}
func readQgqpBIdFromFile() string {
	d, e := os.ReadFile("configs/config.yaml")
	if e != nil { return "" }
	var m map[string]any
	yaml.Unmarshal(d, &m)
	if ds, ok := m["data_source"].(map[string]any); ok {
		if v, ok := ds["qgqp_b_id"].(string); ok { return v }
	}
	return ""
}
func getS(m map[string]any, k string) string { v, _ := m[k]; return fmt.Sprintf("%v", v) }
func getF(m map[string]any, k string) float64 { if v, ok := m[k]; ok { switch t := v.(type) { case float64: return t; case string: var f float64; fmt.Sscanf(t, "%f", &f); return f } }; return 0 }
func ToolNewsList(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	t := tools.NewNewsListTool()
	input := "{}"
	if keyword != "" {
		input = fmt.Sprintf(`{"keyword":"%s"}`, keyword)
	}
	r2, e := t.InvokableRun(r.Context(), input)
	writeJSON(w, map[string]any{"result": r2}, e)
}

func ToolEconomicData(w http.ResponseWriter, r *http.Request) {
	flag := r.URL.Query().Get("flag")
	t := tools.NewQueryEconomicDataTool()
	input := "{}"
	if flag != "" {
		input = fmt.Sprintf(`{"flag":"%s"}`, flag)
	}
	r2, e := t.InvokableRun(r.Context(), input)
	writeJSON(w, map[string]any{"result": r2}, e)
}

func ToolMutualTop10(w http.ResponseWriter, r *http.Request) {
	mutualType := r.URL.Query().Get("mutualType")
	tradeDate := r.URL.Query().Get("tradeDate")
	t := tools.NewMutualTop10Tool()
	input := `{"mutualType":"001"}`
	if mutualType != "" || tradeDate != "" {
		if mutualType == "" { mutualType = "001" }
		input = fmt.Sprintf(`{"mutualType":"%s","tradeDate":"%s"}`, mutualType, tradeDate)
	}
	r2, e := t.InvokableRun(r.Context(), input)
	writeJSON(w, map[string]any{"result": r2}, e)
}

// GlobalIndexesJSON 结构化全球指数
func GlobalIndexesJSON(w http.ResponseWriter, r *http.Request) {
	client := resty.New().SetTimeout(10*time.Second).SetHeader("User-Agent", "Mozilla/5.0")
	resp, err := client.R().SetContext(r.Context()).
		SetHeader("Referer", "https://stockapp.finance.qq.com/mstats").
		Get("https://proxy.finance.qq.com/ifzqgtimg/appstock/app/rank/indexRankDetail2")
	if err != nil { writeError(w, 500, err); return }
	var raw struct { Data map[string]any `json:"data"` }
	if err := json.Unmarshal(resp.Body(), &raw); err != nil { writeError(w, 500, fmt.Errorf("parse: %w", err)); return }
	groups := []string{"common", "america", "asia", "europe", "other"}
	result := make(map[string][]map[string]any)
	for _, g := range groups {
		items, ok := raw.Data[g].([]any)
		if !ok { continue }
		for _, item := range items {
			if m, ok := item.(map[string]any); ok {
				result[g] = append(result[g], map[string]any{"name": m["name"], "zxj": m["zxj"], "zdf": m["zdf"], "code": m["code"], "img": m["img"], "state": m["state"], "location": m["location"]})
			}
		}
	}
	writeJSON(w, result, nil)
}

// TelegraphJSON 结构化电报新闻
func TelegraphJSON(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	client := resty.New().SetTimeout(15*time.Second).SetHeader("User-Agent", "Mozilla/5.0")

	// Cailianpress
	if keyword == "" || strings.Contains(keyword, "财联社") || strings.Contains(keyword, "电报") {
		resp, err := client.R().SetContext(r.Context()).SetHeader("Referer", "https://www.cls.cn/").Get("https://www.cls.cn/api/cache?app=CailianpressWeb&name=telegraph&os=web&sv=8.7.9")
		if err == nil {
			var m map[string]any
			if json.Unmarshal(resp.Body(), &m) == nil {
				if errno, _ := m["errno"].(float64); errno == 0 {
					if data, _ := m["data"].(map[string]any); data != nil {
						if rd, _ := data["roll_data"].([]any); len(rd) > 0 {
							var items []map[string]any
							for i, item := range rd {
								if i >= 30 { break }
								im, ok := item.(map[string]any)
								if !ok { continue }
								ctime, _ := im["ctime"].(float64)
								tm := time.Unix(int64(ctime), 0)
								dt := tm.Format("2006-01-02 15:04")
								to := tm.Format("15:04")
								title, _ := im["title"].(string)
								content, _ := im["content"].(string)
								level, _ := im["level"].(string)
								id := fmt.Sprintf("%v", im["id"])
								items = append(items, map[string]any{
									"title": title, "content": content, "time": to, "dataTime": dt,
									"isRed": level != "" && level != "C",
									"source": "财联社电报", "url": "https://www.cls.cn/telegraph/" + id,
								})
							}
							if len(items) > 0 { writeJSON(w, items, nil); return }
						}
					}
				}
			}
		}
	}

	// Sina
	if strings.Contains(keyword, "新浪") || strings.Contains(keyword, "财经") {
		resp, err := client.R().SetContext(r.Context()).SetHeader("Referer", "https://finance.sina.com.cn/").Get("https://feed.mix.sina.com.cn/api/roll/get?pageid=153&lid=2516&k=&num=15&page=1")
		if err == nil {
			var sina struct { Result struct { Data []map[string]any } }
			if json.Unmarshal(resp.Body(), &sina) == nil && len(sina.Result.Data) > 0 {
				items := make([]map[string]any, 0, len(sina.Result.Data))
				for _, item := range sina.Result.Data {
					title, _ := item["title"].(string)
					ct := fmt.Sprintf("%v", item["ctime"])
					if len(ct) >= 16 { ct = ct[11:16] }
					items = append(items, map[string]any{"title": title, "time": ct, "dataTime": ct, "isRed": false, "source": "新浪财经", "url": item["url"]})
				}
				writeJSON(w, items, nil); return
			}
		}
	}

	writeJSON(w, []any{}, nil)
}

// MarketStatisticJSON 今日市场统计数据（涨跌家数比 + 涨跌停家数比）
func MarketStatisticJSON(w http.ResponseWriter, r *http.Request) {
	client := resty.New().SetTimeout(10*time.Second).SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	resp, err := client.R().SetContext(r.Context()).
		SetHeader("Referer", "https://www.cls.cn/").
		Get("https://x-quote.cls.cn/quote/index/home?app=CailianpressWeb&os=web&sv=8.4.6")
	if err != nil { writeError(w, 500, fmt.Errorf("cls: %w", err)); return }

	var raw struct {
		Data struct {
			UpDownDis struct {
				RiseNum int `json:"rise_num"`
				FallNum int `json:"fall_num"`
				UpNum   int `json:"up_num"`
				DownNum int `json:"down_num"`
			} `json:"up_down_dis"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &raw); err != nil { writeError(w, 500, fmt.Errorf("parse: %w", err)); return }

	d := raw.Data.UpDownDis
	now := time.Now()
	totalUp, totalDown := d.RiseNum, d.FallNum
	limitUp, limitDown := d.UpNum, d.DownNum

	var upRatio, upDownRatio, limitRatio float64
	total := totalUp + totalDown
	if total > 0 { upRatio = float64(totalUp) / float64(total) * 100 }
	if totalDown > 0 { upDownRatio = float64(totalUp) / float64(totalDown) } else if totalUp > 0 { upDownRatio = float64(totalUp) }
	if limitDown > 0 { limitRatio = float64(limitUp) / float64(limitDown) } else if limitUp > 0 { limitRatio = float64(limitUp) }

	sentimentDesc := "中性"
	switch {
	case upDownRatio >= 2: sentimentDesc = "普涨(极强)"
	case upDownRatio >= 1.5: sentimentDesc = "偏强"
	case upDownRatio > 1: sentimentDesc = "稍强"
	case upDownRatio == 1: sentimentDesc = "中性"
	case upDownRatio > 0.5: sentimentDesc = "稍弱"
	case upDownRatio > 0: sentimentDesc = "偏弱"
	default: sentimentDesc = "普跌(冰点)"
	}

	result := []map[string]any{{
		"dataDate": now.Format("2006-01-02"),
		"dataTime": now.Format("15:04"),
		"upCount": totalUp, "downCount": totalDown,
		"upRatio": upRatio, "upDownRatio": upDownRatio,
		"sentimentDesc": sentimentDesc,
		"limitUp": limitUp, "limitDown": limitDown, "limitRatio": limitRatio,
	}}
	writeJSON(w, result, nil)
}

// RecentMarketStatisticJSON 近N日市场统计（返回单日聚合数据）
func RecentMarketStatisticJSON(w http.ResponseWriter, r *http.Request) {
	MarketStatisticJSON(w, r)
}

// HotTopicJSON 东方财富股吧热门话题
func HotTopicJSON(w http.ResponseWriter, r *http.Request) {
	client := resty.New().SetTimeout(15*time.Second).SetHeader("User-Agent", "Mozilla/5.0")
	resp, err := client.R().SetContext(r.Context()).
		SetHeader("Host", "gubatopic.eastmoney.com").
		SetHeader("Origin", "https://gubatopic.eastmoney.com").
		SetHeader("Referer", "https://gubatopic.eastmoney.com/").
		Get("https://gubatopic.eastmoney.com/interface/GetData.aspx?path=newtopic/api/Topic/HomePageListRead")
	if err != nil { writeJSON(w, []any{}, nil); return }
	var m map[string]any
	if err := json.Unmarshal(resp.Body(), &m); err != nil { writeJSON(w, []any{}, nil); return }
	items, _ := m["re"].([]any)
	if items == nil { items = []any{} }
	writeJSON(w, items, nil)
}

// HotEventJSON 雪球热门事件（简化实现）
func HotEventJSON(w http.ResponseWriter, r *http.Request) {
	client := resty.New().SetTimeout(15*time.Second).SetHeader("User-Agent", "Mozilla/5.0")
	resp, err := client.R().SetContext(r.Context()).
		SetHeader("Host", "xueqiu.com").
		SetHeader("Referer", "https://xueqiu.com/").
		Get("https://xueqiu.com/hot_event/list.json?count=50")
	if err != nil { writeJSON(w, []any{}, nil); return }
	var m map[string]any
	if err := json.Unmarshal(resp.Body(), &m); err != nil { writeJSON(w, []any{}, nil); return }
	list, _ := m["list"].([]any)
	if list == nil { list = []any{} }
	writeJSON(w, list, nil)
}

// HotStockJSON 东方财富热门股票排行
func HotStockJSON(w http.ResponseWriter, r *http.Request) {
	marketType := r.URL.Query().Get("type")
	if marketType == "" { marketType = "10" }
	client := resty.New().SetTimeout(15*time.Second).SetHeader("User-Agent", "Mozilla/5.0")
	resp, err := client.R().SetContext(r.Context()).
		SetHeader("Referer", "https://quote.eastmoney.com/").
		Get(fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?cb=&pn=1&pz=50&po=1&np=1&fields=f12,f14,f3,f2,f4,f6,f9,f8,f18,f20&fid=f3&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23,m:0+t:81+s:2048&ut=bd1d9ddb04089700cf9c27f6f7426281"))
	if err != nil { writeJSON(w, []any{}, nil); return }
	var m map[string]any
	if err := json.Unmarshal(resp.Body(), &m); err != nil { writeJSON(w, []any{}, nil); return }
	data, _ := m["data"].(map[string]any)
	dl, _ := data["diff"].([]any)
	out := make([]map[string]any, 0)
	for _, item := range dl {
		if it, ok := item.(map[string]any); ok {
			code, _ := it["f12"].(string)
			name, _ := it["f14"].(string)
			price, _ := it["f2"].(float64)
			pct, _ := it["f3"].(float64)
			exchange := "SH"
			if len(code) == 6 && code[0] != '6' && code[0] != '9' { exchange = "SZ" }
			out = append(out, map[string]any{
				"code": code, "name": name, "exchange": exchange,
				"current": price, "percent": pct,
				"value": it["f20"], "increment": it["f4"],
				"rank_change": it["f9"],
			})
		}
	}
	if out == nil { out = []map[string]any{} }
	writeJSON(w, out, nil)
}
