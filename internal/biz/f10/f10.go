package f10

import (
	"context"

	"eino-stock/internal/infrastructure/f10"
)

type FinanceData struct {
	ReportDate string             `json:"reportDate"`
	Items      map[string]string  `json:"items"`
}

type F10Usecase struct {
	client *f10.Client
}

func NewF10Usecase(client *f10.Client) *F10Usecase {
	return &F10Usecase{client: client}
}

func (uc *F10Usecase) LatestFinance(ctx context.Context, code string) (*f10.GenericResp, error) {
	resp, err := uc.client.LatestFinance(code)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (uc *F10Usecase) QtrFinance(ctx context.Context, code string) (*f10.GenericResp, error) {
	return uc.client.QtrFinance(code)
}

func (uc *F10Usecase) HolderTrend(ctx context.Context, code string) (*f10.GenericResp, error) {
	return uc.client.HolderTrend(code)
}

func (uc *F10Usecase) OrgPredict(ctx context.Context, code string) (*f10.GenericResp, error) {
	return uc.client.OrgPredict(code)
}

func (uc *F10Usecase) PredictSummary(ctx context.Context, code string) (*f10.GenericResp, error) {
	return uc.client.PredictSummary(code)
}
