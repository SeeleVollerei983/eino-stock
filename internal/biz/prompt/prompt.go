package prompt

import (
	"context"

	"eino-stock/internal/data"
)

type PromptUsecase struct {
	repo *data.PromptRepo
}

func NewPromptUsecase(repo *data.PromptRepo) *PromptUsecase {
	return &PromptUsecase{repo: repo}
}

func (uc *PromptUsecase) List(ctx context.Context) ([]data.PromptTemplate, error) {
	return uc.repo.List(ctx)
}
func (uc *PromptUsecase) GetByID(ctx context.Context, id uint) (*data.PromptTemplate, error) {
	return uc.repo.GetByID(ctx, id)
}
func (uc *PromptUsecase) Create(ctx context.Context, p *data.PromptTemplate) error {
	return uc.repo.Create(ctx, p)
}
func (uc *PromptUsecase) Update(ctx context.Context, p *data.PromptTemplate) error {
	return uc.repo.Update(ctx, p)
}
func (uc *PromptUsecase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}
