package biz

import (
	"eino-stock/internal/biz/ai"
	"eino-stock/internal/biz/cron"
	"eino-stock/internal/biz/f10"
	"eino-stock/internal/biz/follow"
	"eino-stock/internal/biz/market"
	"eino-stock/internal/biz/prompt"
	"eino-stock/internal/biz/screen"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	market.NewMarketUsecase,
	screen.NewScreenUsecase,
	ai.NewScreenUsecase,
	follow.NewFollowUsecase,
	f10.NewF10Usecase,
	cron.NewCronUsecase,
	prompt.NewPromptUsecase,
)
