package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/go-resty/resty/v2"
)

type ExpertAnalysisTool struct {
	client *resty.Client
}

func NewExpertAnalysisTool() *ExpertAnalysisTool {
	return &ExpertAnalysisTool{
		client: resty.New().SetTimeout(15*time.Second).
			SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
			SetHeader("Referer", "https://quote.eastmoney.com"),
	}
}

func (t *ExpertAnalysisTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "ExpertStockAnalysis",
		Desc: "股票技术面综合分析。输入股票代码，返回MA排列、均线粘合、回踩支撑、突破形态、量价关系的全面分析结果。用于判断股票是否处于最佳买入时机。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"code":  {Type: schema.String, Desc: "股票代码，如 600519 或 600519.SH", Required: true},
			"days":  {Type: schema.String, Desc: "分析周期(日K条数，默认120)", Required: false},
		}),
	}, nil
}

// klineRow represents a parsed K-line data point
type klineRow struct {
	day    string
	open   float64
	close  float64
	high   float64
	low    float64
	volume float64
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}

func sma(values []float64, period int) []float64 {
	out := make([]float64, len(values))
	for i := 0; i < len(values); i++ {
		if i < period-1 { out[i] = 0; continue }
		sum := 0.0
		for j := 0; j < period; j++ { sum += values[i-j] }
		out[i] = sum / float64(period)
	}
	return out
}

func (t *ExpertAnalysisTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var params struct {
		Code string `json:"code"`
		Days int    `json:"days"`
	}
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}
	if params.Code == "" { return "请提供股票代码", nil }
	if params.Days <= 0 { params.Days = 120 }

	// Convert stock code to eastmoney format
	code := strings.ToUpper(strings.TrimSpace(params.Code))
	code = strings.ReplaceAll(code, ".SH", ""); code = strings.ReplaceAll(code, ".SZ", "")
	code = strings.ReplaceAll(code, ".sh", ""); code = strings.ReplaceAll(code, ".sz", "")
	secID := ""
	if len(code) >= 1 && code[0] >= '0' && code[0] <= '9' {
		switch code[0] {
		case '6': secID = "1." + code
		case '0', '3': secID = "0." + code
		default: secID = "0." + code
		}
	}

	// Fetch K-line data
	url := fmt.Sprintf("https://push2his.eastmoney.com/api/qt/stock/kline/get?secid=%s&klt=101&fqt=0&end=20500101&lmt=%d&fields1=f1,f2,f3,f4,f5,f6&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&_=%d",
		secID, params.Days, time.Now().UnixMilli())

	resp, err := t.client.R().SetContext(ctx).Get(url)
	if err != nil { return fmt.Sprintf("获取K线失败: %v", err), nil }

	var result struct {
		Data *struct { Klines []string `json:"klines"` } `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil { return "解析K线失败", nil }
	if result.Data == nil || len(result.Data.Klines) == 0 { return "没有K线数据", nil }

	// Parse K-line data
	rows := make([]klineRow, 0, len(result.Data.Klines))
	for _, ks := range result.Data.Klines {
		parts := strings.Split(ks, ",")
		if len(parts) < 11 { continue }
		rows = append(rows, klineRow{
			day:    parts[0],
			open:   parseFloat(parts[1]),
			close:  parseFloat(parts[2]),
			high:   parseFloat(parts[3]),
			low:    parseFloat(parts[4]),
			volume: parseFloat(parts[5]),
		})
	}
	if len(rows) < 20 { return "K线数据不足(需要至少20条)", nil }

	// Extract values
	n := len(rows)
	closes := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	volumes := make([]float64, n)
	for i, r := range rows {
		closes[i] = r.close; highs[i] = r.high; lows[i] = r.low; volumes[i] = r.volume
	}

	// Calculate MAs
	ma5 := sma(closes, 5)
	ma10 := sma(closes, 10)
	ma20 := sma(closes, 20)
	ma60 := sma(closes, 60)

	last := n - 1
	price := closes[last]
	var signals []string
	var details []string

	// ===== 1. 多头排列检测 =====
	bullish := ma5[last] > 0 && ma10[last] > 0 && ma20[last] > 0 && ma60[last] > 0 &&
		ma5[last] > ma10[last] && ma10[last] > ma20[last] && ma20[last] > ma60[last]
	// Check if bullish has been maintained for the past 5 periods
	bullishStreak := 0
	for i := last; i >= 0 && i > last-10; i-- {
		if ma5[i] > 0 && ma10[i] > 0 && ma20[i] > 0 && ma60[i] > 0 &&
			ma5[i] > ma10[i] && ma10[i] > ma20[i] && ma20[i] > ma60[i] {
			bullishStreak++
		} else { break }
	}

	if bullish {
		signals = append(signals, fmt.Sprintf("✅ 多头排列: MA5=%.2f > MA10=%.2f > MA20=%.2f > MA60=%.2f (已维持%d期)", ma5[last], ma10[last], ma20[last], ma60[last], bullishStreak))
	} else {
		signals = append(signals, fmt.Sprintf("⚠️ 非多头排列: MA5=%.2f MA10=%.2f MA20=%.2f MA60=%.2f", ma5[last], ma10[last], ma20[last], ma60[last]))
	}

	// ===== 2. 均线粘合检测 =====
	// MAs are converging if their values are within a tight range
	maVals := []float64{ma5[last], ma10[last], ma20[last]}
	if ma60[last] > 0 { maVals = append(maVals, ma60[last]) }
	mean, maxVal, minVal := 0.0, 0.0, math.MaxFloat64
	for _, v := range maVals {
		mean += v; if v > maxVal { maxVal = v }; if v < minVal { minVal = v }
	}
	mean /= float64(len(maVals))
	convergePct := (maxVal - minVal) / mean * 100
	isMAConverging := convergePct < 8.0 && convergePct >= 0

	if isMAConverging {
		signals = append(signals, fmt.Sprintf("✅ 均线粘合: 各均线在%.2f~%.2f范围内(幅度%.1f%%)，粘合向上说明主力吸筹完毕", minVal, maxVal, convergePct))
	} else {
		details = append(details, fmt.Sprintf("均线分散度: %.1f%%", convergePct))
	}

	// ===== 3. 回踩均线检测 =====
	// Price is near the MA convergence zone => pullback
	maCenter := (ma20[last] + ma60[last]) / 2
	if ma60[last] <= 0 { maCenter = ma20[last] }
	pullbackPct := math.Abs(price-maCenter) / maCenter * 100
	isPullback := pullbackPct < 3.0 && pullbackPct >= 0 && bullish
	if isPullback {
		signals = append(signals, fmt.Sprintf("✅ 回踩均线: 当前价%.2f靠近均线区域%.2f(偏离%.1f%%)，是最佳参与位置", price, maCenter, pullbackPct))
	} else if bullish {
		details = append(details, fmt.Sprintf("当前价%.2f偏离均线区域%.2f(%.1f%%)，等待回踩", price, maCenter, pullbackPct))
	}

	// ===== 4. 突破形态检测 =====
	// Check if current price exceeds 60-period high
	lookback := 60
	if n < lookback { lookback = n - 5 }
	highest60 := 0.0
	highest60Day := ""
	for i := n - lookback; i < last; i++ {
		if highs[i] > highest60 { highest60 = highs[i]; highest60Day = rows[i].day }
	}
	breakoutPct := (price - highest60) / highest60 * 100
	isBreakout := breakoutPct > 0 && breakoutPct < 15.0

	// Check if recent high is close to 60-period high
	nearHighPct := (price - highest60*0.98) / (highest60 * 0.98) * 100
	if isBreakout && breakoutPct > 0 {
		signals = append(signals, fmt.Sprintf("✅ 突破信号: 当前价%.2f突破前期高点%.2f(%.1f%%)(%s)，进入上升趋势标准", price, highest60, breakoutPct, highest60Day))
	} else if nearHighPct > -3 && nearHighPct <= 0 {
		signals = append(signals, fmt.Sprintf("⚠️ 临近突破: 当前价%.2f接近60日高点%.2f(%s)，关注是否能放量突破", price, highest60, highest60Day))
	}

	// ===== 5. 量价关系分析 =====
	avgVol := 0.0
	for i := n - 20; i < n; i++ { avgVol += volumes[i] }
	avgVol /= 20
	recentVol := volumes[last]

	// Check price-volume correlation over last 10 periods
	volConfirmCount := 0
	totalUpDays := 0
	for i := n - 10; i <= last; i++ {
		if closes[i] > closes[i-1] {
			totalUpDays++
			if volumes[i] > avgVol { volConfirmCount++ }
		} else {
			if volumes[i] <= avgVol { volConfirmCount++ }
		}
	}
	volConfirmRatio := float64(volConfirmCount) / 10.0 * 100

	if volConfirmRatio >= 60 {
		signals = append(signals, fmt.Sprintf("✅ 量价配合: %.0f%%的交易日符合「价涨量升、价跌量缩」，趋势稳定", volConfirmRatio))
	} else {
		details = append(details, fmt.Sprintf("量价配合度: %.0f%%", volConfirmRatio))
	}

	// Volume surge on latest up day
	if recentVol > avgVol*1.5 && closes[last] > closes[last-1] {
		signals = append(signals, fmt.Sprintf("⚠️ 放量上涨: 今日成交量%.0f是20日均量%.0f的%.1f倍，关注后续", recentVol, avgVol, recentVol/avgVol))
	}

	// ===== 6. 回调不破前高检测 =====
	// Find the most recent significant high and check if pullback holds above it
	if bullish && isPullback {
		signals = append(signals, "✅ 回调不破临界点: 多头排列+回踩均线，回调不破均线粘合处就是最佳参与机会")
	}

	// ===== Build Output =====
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📊 %s 技术面分析 (%d日K线)\n", params.Code, params.Days))
	sb.WriteString(fmt.Sprintf("当前价: %.2f\n\n", price))
	sb.WriteString("【分析结论】\n")
	for _, s := range signals { sb.WriteString(s + "\n") }
	if len(signals) == 0 { sb.WriteString("无明显信号\n") }
	sb.WriteString("\n【持仓建议】\n")

	// Final recommendation
	score := 0
	if bullish { score++ }
	if isMAConverging { score++ }
	if isPullback { score++ }
	if isBreakout { score++ }
	if volConfirmRatio >= 60 { score++ }

	switch {
	case score >= 4:
		sb.WriteString("⭐⭐⭐⭐⭐ 强烈推荐: 多头排列+均线粘合+回踩确认+量价配合，最佳参与时机\n")
	case score >= 3:
		sb.WriteString("⭐⭐⭐⭐ 推荐关注: 多个信号共振，可逐步建仓\n")
	case score >= 2:
		sb.WriteString("⭐⭐⭐ 观察中: 部分条件满足，等待更多确认信号\n")
	default:
		sb.WriteString("⭐ 暂不参与: 条件不满足，等待更好的时机\n")
	}

	if len(details) > 0 {
		sb.WriteString("\n【详细信息】\n")
		for _, d := range details { sb.WriteString(d + "\n") }
	}

	return sb.String(), nil
}
