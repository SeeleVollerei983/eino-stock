package ai

import (
	"context"

	"eino-stock/internal/infrastructure/eino"
)

type ScreenUsecase struct{}

func NewScreenUsecase() *ScreenUsecase {
	return &ScreenUsecase{}
}

func (uc *ScreenUsecase) ParallelScreen(ctx context.Context, query string) (*eino.ParallelResult, error) {
	cfg := eino.ReadAIConfig()
	return eino.RunParallelScreener(ctx, query, cfg)
}

func (uc *ScreenUsecase) Screen(ctx context.Context, query string) (*eino.ScreenResult, error) {
	cfg := eino.ReadAIConfig()
	return eino.RunScreener(ctx, query, cfg)
}