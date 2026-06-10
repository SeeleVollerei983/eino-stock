package service

import (
	"net/http"
	"strconv"
	"strings"

	bizmarket "eino-stock/internal/biz/market"
)

type MarketService struct {
	uc *bizmarket.MarketUsecase
}

func NewMarketService(uc *bizmarket.MarketUsecase) *MarketService {
	return &MarketService{uc: uc}
}

func (s *MarketService) RegisterHTTP(mux *http.ServeMux) {
	mux.HandleFunc("GET /market/stocks", s.searchStocks)
	mux.HandleFunc("GET /market/quote/", s.getQuote)
	mux.HandleFunc("GET /market/kline/", s.getKLines)
}

func (s *MarketService) searchStocks(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	stocks, err := s.uc.SearchStocks(r.Context(), r.URL.Query().Get("keyword"), limit)
	writeJSON(w, stocks, err)
}

func (s *MarketService) getQuote(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/api/market/quote/")
	if code == "" {
		writeError(w, http.StatusBadRequest, errMissingStockCode)
		return
	}
	quote, err := s.uc.GetQuote(r.Context(), code)
	writeJSON(w, quote, err)
}

func (s *MarketService) getKLines(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/api/market/kline/")
	code = strings.TrimSuffix(code, "/")
	if code == "" {
		writeError(w, http.StatusBadRequest, errMissingStockCode)
		return
	}
	ktype := bizmarket.KLineType(r.URL.Query().Get("ktype"))
	if ktype == "" {
		ktype = bizmarket.KLineDay
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 60
	}
	kLines, err := s.uc.GetKLines(r.Context(), code, ktype, limit)
	writeJSON(w, kLines, err)
}
