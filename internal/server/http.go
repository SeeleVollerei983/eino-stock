package server

import (
	"net/http"

	"eino-stock/internal/conf"
	"eino-stock/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

func NewHTTPServer(c *conf.Server, market *service.MarketService, screen *service.ScreenService, aisvc *service.AIService, fsvc *service.FollowService, settings *service.SettingsService, logger log.Logger, f10svc *service.F10Service, cronsvc *service.CronService, searchsvc *service.SearchService, promptsvc *service.PromptService) *kratoshttp.Server {
	var opts = []kratoshttp.ServerOption{kratoshttp.Middleware(recovery.Recovery())}
	if c.Http.Network != "" { opts = append(opts, kratoshttp.Network(c.Http.Network)) }
	if c.Http.Addr != "" { opts = append(opts, kratoshttp.Address(c.Http.Addr)) }
	if c.Http.Timeout != nil { opts = append(opts, kratoshttp.Timeout(c.Http.Timeout.AsDuration())) }
	srv := kratoshttp.NewServer(opts...)

	srv.HandleFunc("/api/ai/screen", aisvc.Screen)
	srv.HandleFunc("/api/ai/parallel", aisvc.ParallelScreen)
	srv.HandleFunc("/api/search", searchsvc.Search)
	srv.HandleFunc("/api/ai/chat", aisvc.ChatStream)

	apiMux := http.NewServeMux()
	market.RegisterHTTP(apiMux)
	screen.RegisterHTTP(apiMux)
	f10svc.RegisterHTTP(apiMux)
	cronsvc.RegisterHTTP(apiMux)
	promptsvc.RegisterHTTP(apiMux)
	// Tool routes (bypassing AI)
	srv.HandleFunc("/api/tool/screen", service.ToolScreen)
	srv.HandleFunc("/api/tool/screen-v2", service.ToolScreenV2)
	srv.HandleFunc("/api/tool/minute", service.ToolMinute)
	srv.HandleFunc("/api/tool/detail", service.ToolDetail)
	srv.HandleFunc("/api/tool/notice", service.ToolNotice)
	srv.HandleFunc("/api/tool/report", service.ToolReport)
	srv.HandleFunc("/api/tool/global-indexes", service.ToolGlobalIndexes)
	srv.HandleFunc("/api/tool/hot-plates", service.ToolHotPlates)
	srv.HandleFunc("/api/tool/long-tiger", service.ToolLongTiger)
	srv.HandleFunc("/api/tool/industry-valuation", service.ToolIndustryValuation)
	srv.HandleFunc("/api/tool/industry-money-rank", service.ToolIndustryMoneyRank)
	srv.HandleFunc("/api/tool/news-list", service.ToolNewsList)
	srv.HandleFunc("/api/tool/economic-data", service.ToolEconomicData)
	srv.HandleFunc("/api/tool/mutual-top10", service.ToolMutualTop10)
	// Follow routes
	srv.HandleFunc("/api/follow/list", fsvc.List)
	srv.HandleFunc("/api/follow/add", fsvc.Add)
	srv.HandleFunc("/api/follow/remove", fsvc.Remove)
	// Market global-indexes & telegraph (registered on srv to avoid HandlePrefix stripping)
	srv.HandleFunc("/api/market/global-indexes", service.GlobalIndexesJSON)
	srv.HandleFunc("/api/market/telegraph-list", service.TelegraphJSON)

	// Market statistics & hot topics
	srv.HandleFunc("/api/market/today-statistic", service.MarketStatisticJSON)
	srv.HandleFunc("/api/market/recent-statistic", service.RecentMarketStatisticJSON)
	srv.HandleFunc("/api/market/hot-topic", service.HotTopicJSON)
	srv.HandleFunc("/api/market/hot-event", service.HotEventJSON)
	srv.HandleFunc("/api/market/hot-stock", service.HotStockJSON)

	// Settings routes
	srv.HandleFunc("/api/settings/get", settings.Get)
	srv.HandleFunc("/api/settings/set", settings.Set)
	srv.HandlePrefix("/api/", apiMux)

	registerWebUI(srv)
	return srv
}

