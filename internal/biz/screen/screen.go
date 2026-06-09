package screen

import "context"

type BkItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ETFItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type HotStrategyItem struct {
	Question string  `json:"question"`
	Chg      float64 `json:"chg"`
	Heat     int     `json:"heat"`
}

type ScreenSource interface {
	SearchBk(ctx context.Context, keyword string, pageSize int) ([]*BkItem, error)
	SearchETF(ctx context.Context, keyword string, pageSize int) ([]*ETFItem, error)
	HotStrategy(ctx context.Context) ([]*HotStrategyItem, error)
}

type ScreenUsecase struct {
	source ScreenSource
}

func NewScreenUsecase(source ScreenSource) *ScreenUsecase {
	return &ScreenUsecase{source: source}
}

func (uc *ScreenUsecase) SearchBk(ctx context.Context, keyword string, pageSize int) ([]*BkItem, error) {
	return uc.source.SearchBk(ctx, keyword, pageSize)
}

func (uc *ScreenUsecase) SearchETF(ctx context.Context, keyword string, pageSize int) ([]*ETFItem, error) {
	return uc.source.SearchETF(ctx, keyword, pageSize)
}

func (uc *ScreenUsecase) HotStrategy(ctx context.Context) ([]*HotStrategyItem, error) {
	return uc.source.HotStrategy(ctx)
}