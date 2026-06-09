package search

import (
	"context"

	"eino-stock/internal/infrastructure/search"
)

type SearchUsecase struct {
	client *search.Client
}

func NewSearchUsecase() *SearchUsecase {
	return &SearchUsecase{client: search.NewClient()}
}

func (uc *SearchUsecase) Search(ctx context.Context, query string) ([]search.Result, error) {
	return uc.client.Search(ctx, query)
}
