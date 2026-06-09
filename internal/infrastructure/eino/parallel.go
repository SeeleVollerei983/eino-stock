package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"eino-stock/internal/infrastructure/eino/tools"

	"golang.org/x/sync/errgroup"
)

type ParallelResult struct {
	Query            string                `json:"query"`
	DimensionResults map[string]*DimResult `json:"dimensionResults"`
	Merged           []string              `json:"merged"`
	Summary          string                `json:"summary"`
}

type DimResult struct {
	Dimension string   `json:"dimension"`
	Query     string   `json:"query"`
	Stocks    []string `json:"stocks"`
	Count     int      `json:"count"`
	Raw       string   `json:"raw,omitempty"`
}

func RunParallelScreener(ctx context.Context, userQuery string, aiCfg *AIConfig) (*ParallelResult, error) {
	chatModel, _ := NewChatModel(ctx, aiCfg)
	parsedQuery, err := ParseConditions(ctx, userQuery, chatModel)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	result := &ParallelResult{Query: parsedQuery, DimensionResults: make(map[string]*DimResult)}
	dims := splitDimensions(parsedQuery)

	mu := sync.Mutex{}
	g, gctx := errgroup.WithContext(ctx)
	for name, dim := range dims {
		n, d := name, dim
		g.Go(func() error {
			dr := runDim(gctx, n, d)
			mu.Lock()
			result.DimensionResults[n] = dr
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("parallel: %w", err)
	}
	result.Merged = mergeResults(result.DimensionResults)
	result.Summary = formatSummary(result)
	return result, nil
}

func runDim(ctx context.Context, name, query string) *DimResult {
	dr := &DimResult{Dimension: name, Query: query}
	t := tools.NewSelectAStockTool()
	in, _ := json.Marshal(map[string]string{"words": query})
	out, err := t.InvokableRun(ctx, string(in))
	if err != nil {
		dr.Raw = fmt.Sprintf("错误: %v", err)
		return dr
	}
	dr.Raw = out
	dr.Stocks = extractCodes(out)
	dr.Count = len(dr.Stocks)
	return dr
}

func extractCodes(raw string) []string {
	seen := map[string]bool{}
	var codes []string
	for _, line := range strings.Split(raw, "\n") {
		// Line format: 股票名称(代码) 最新价:... 涨跌幅:...
		start := strings.Index(line, "(")
		if start < 0 {
			continue
		}
		end := strings.Index(line[start:], ")")
		if end < 1 {
			continue
		}
		code := line[start+1 : start+end]
		if len(code) >= 6 && isAllDigit(code) && !seen[code] {
			seen[code] = true
			codes = append(codes, code)
		}
	}
	return codes
}

func isAllDigit(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func mergeResults(dims map[string]*DimResult) []string {
	if len(dims) == 0 {
		return nil
	}
	// Count how many dimensions each stock appears in
	counts := map[string]int{}
	for _, d := range dims {
		for _, c := range d.Stocks {
			counts[c]++
		}
	}
	// Collect stocks that appear in ALL dimensions
	want := len(dims)
	var merged []string
	for _, c := range dims[firstKey(dims)].Stocks {
		if counts[c] == want {
			merged = append(merged, c)
		}
	}
	return merged
}

func firstKey(dims map[string]*DimResult) string {
	for k := range dims {
		return k
	}
	return ""
}

var dimKeywords = map[string][]string{
	"估值": {"PE", "PB", "市盈率", "市净率", "估值", "股息率", "PEG"},
	"财务": {"ROE", "毛利率", "净利率", "营收", "净利润", "EPS", "每股收益", "ROA", "资产负债率"},
	"市场": {"换手率", "市值", "流通市值", "量比", "成交额", "成交", "股价", "价格", "涨跌幅", "涨幅", "连板", "筹码", "集中度"},
}

func splitDimensions(query string) map[string]string {
	conds := strings.Split(query, ";")
	var common, val, fin, mkt []string

	for _, c := range conds {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		matched := false
		for dim, kws := range dimKeywords {
			for _, kw := range kws {
				if strings.Contains(c, kw) {
					switch dim {
					case "估值":
						val = append(val, c)
					case "财务":
						fin = append(fin, c)
					case "市场":
						mkt = append(mkt, c)
					}
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			common = append(common, c)
		}
	}

	build := func(conds []string) string {
		all := append([]string{}, common...)
		all = append(all, conds...)
		return strings.Join(all, ";")
	}

	dims := map[string]string{}
	if len(val) > 0 {
		dims["估值"] = build(val)
	}
	if len(fin) > 0 {
		dims["财务"] = build(fin)
	}
	if len(mkt) > 0 {
		dims["市场"] = build(mkt)
	}
	if len(dims) == 0 && len(common) > 0 {
		dims["综合筛选"] = strings.Join(common, ";")
	}
	return dims
}

func formatSummary(r *ParallelResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("多维并行选股完成！\n查询: %s\n\n", r.Query))
	for name, dr := range r.DimensionResults {
		b.WriteString(fmt.Sprintf("■ %s维度: 找到%d只股票\n", name, dr.Count))
	}
	b.WriteString(fmt.Sprintf("\n同时满足以上%d个维度的股票: %d只\n", len(r.DimensionResults), len(r.Merged)))
	if len(r.Merged) > 0 {
		b.WriteString(fmt.Sprintf("代码: %s\n", strings.Join(r.Merged, ", ")))
	}
	return b.String()
}