package market

import (
	"context"
	"fmt"
)

// Stock 股票基础信息。
type Stock struct {
	TsCode string
	Symbol string
	Name   string
	Market string
}

// Quote 实时行情。
type Quote struct {
	Code           string
	Name           string
	Price          string
	Open           string
	PreClose       string
	High           string
	Low            string
	ChangePercent  float64
	ChangePrice    float64
	Date           string
	Time           string
	Volume         string
	Amount         string
}

// MarketRepo 行情仓储接口。
type MarketRepo interface {
	SearchStocks(ctx context.Context, keyword string, limit int) ([]*Stock, error)
}

// QuoteProvider 外部行情数据源。
type QuoteProvider interface {
	GetRealtimeQuotes(ctx context.Context, codes []string) ([]*Quote, error)
}

// MarketUsecase 行情用例。
type MarketUsecase struct {
	repo      MarketRepo
	quotes    QuoteProvider
	eastMoney KLineProvider
	sina      KLineProvider
}

// NewMarketUsecase 创建行情用例。
func NewMarketUsecase(repo MarketRepo, quotes QuoteProvider, eastMoney KLineProvider, sina KLineProvider) *MarketUsecase {
	return &MarketUsecase{repo: repo, quotes: quotes, eastMoney: eastMoney, sina: sina}
}

// SearchStocks 按关键词搜索股票。
func (uc *MarketUsecase) SearchStocks(ctx context.Context, keyword string, limit int) ([]*Stock, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return uc.repo.SearchStocks(ctx, keyword, limit)
}

// GetRealtimeQuotes 批量获取实时行情。
func (uc *MarketUsecase) GetRealtimeQuotes(ctx context.Context, codes []string) ([]*Quote, error) {
	normalized := NormalizeStockCodes(codes)
	if len(normalized) == 0 {
		return nil, nil
	}
	return uc.quotes.GetRealtimeQuotes(ctx, normalized)
}

// GetQuote 获取单只股票实时行情。
func (uc *MarketUsecase) GetQuote(ctx context.Context, code string) (*Quote, error) {
	quotes, err := uc.GetRealtimeQuotes(ctx, []string{code})
	if err != nil {
		return nil, err
	}
	if len(quotes) == 0 {
		return nil, nil
	}
	return quotes[0], nil
}

// GetKLines 获取K线数据，东方财富主源 + 新浪fallback。
func (uc *MarketUsecase) GetKLines(ctx context.Context, code string, ktype KLineType, limit int) ([]*KLine, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// 优先东方财富
	if uc.eastMoney != nil {
		kLines, err := uc.eastMoney.GetKLines(ctx, code, ktype, limit)
		if err == nil && len(kLines) > 0 {
			return kLines, nil
		}
		if err != nil {
			fmt.Printf("eastmoney kline fallback: %v\n", err)
		}
	}

	// fallback到新浪
	if uc.sina != nil {
		return uc.sina.GetKLines(ctx, code, ktype, limit)
	}
	return nil, fmt.Errorf("no kline provider available for %s", code)
}
