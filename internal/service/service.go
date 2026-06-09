package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewMarketService,
	NewScreenService,
	NewAIService,
	NewFollowService,
	NewSettingsService,
	NewF10Service,
	NewCronService,
	NewSearchService,
)


