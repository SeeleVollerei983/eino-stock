package follow

import "context"

type FollowedStock struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type FollowRepo interface {
	List(ctx context.Context) ([]FollowedStock, error)
	Add(ctx context.Context, code, name string) error
	Remove(ctx context.Context, code string) error
}

type FollowUsecase struct {
	repo FollowRepo
}

func NewFollowUsecase(repo FollowRepo) *FollowUsecase {
	return &FollowUsecase{repo: repo}
}

func (uc *FollowUsecase) List(ctx context.Context) ([]FollowedStock, error) {
	return uc.repo.List(ctx)
}

func (uc *FollowUsecase) Add(ctx context.Context, code, name string) error {
	return uc.repo.Add(ctx, code, name)
}

func (uc *FollowUsecase) Remove(ctx context.Context, code string) error {
	return uc.repo.Remove(ctx, code)
}