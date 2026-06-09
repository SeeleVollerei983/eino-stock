package service

import (
	"net/http"
	"strconv"

	bizscreen "eino-stock/internal/biz/screen"
)

type ScreenService struct {
	uc *bizscreen.ScreenUsecase
}

func NewScreenService(uc *bizscreen.ScreenUsecase) *ScreenService {
	return &ScreenService{uc: uc}
}

func (s *ScreenService) RegisterHTTP(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/screen/bk/", s.searchBk)
	mux.HandleFunc("GET /api/screen/etf/", s.searchETF)
	mux.HandleFunc("GET /api/screen/hot-strategy", s.hotStrategy)
}

func (s *ScreenService) searchBk(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Path[len("/api/screen/bk/"):]
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if pageSize <= 0 {
		pageSize = 20
	}
	items, err := s.uc.SearchBk(r.Context(), keyword, pageSize)
	writeJSON(w, items, err)
}

func (s *ScreenService) searchETF(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Path[len("/api/screen/etf/"):]
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if pageSize <= 0 {
		pageSize = 20
	}
	items, err := s.uc.SearchETF(r.Context(), keyword, pageSize)
	writeJSON(w, items, err)
}

func (s *ScreenService) hotStrategy(w http.ResponseWriter, r *http.Request) {
	items, err := s.uc.HotStrategy(r.Context())
	writeJSON(w, items, err)
}