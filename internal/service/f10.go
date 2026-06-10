package service

import (
	"net/http"

	bizf10 "eino-stock/internal/biz/f10"
)

type F10Service struct {
	uc *bizf10.F10Usecase
}

func NewF10Service(uc *bizf10.F10Usecase) *F10Service {
	return &F10Service{uc: uc}
}

func (s *F10Service) RegisterHTTP(mux *http.ServeMux) {
	mux.HandleFunc("GET /f10/latest-finance", s.latestFinance)
	mux.HandleFunc("GET /f10/qtr-finance", s.qtrFinance)
	mux.HandleFunc("GET /f10/holder-trend", s.holderTrend)
	mux.HandleFunc("GET /f10/org-predict", s.orgPredict)
	mux.HandleFunc("GET /f10/predict-summary", s.predictSummary)
}

func (s *F10Service) latestFinance(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" { writeError(w, 400, errMissingStockCode); return }
	resp, err := s.uc.LatestFinance(r.Context(), code)
	writeJSON(w, resp, err)
}

func (s *F10Service) qtrFinance(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" { writeError(w, 400, errMissingStockCode); return }
	resp, err := s.uc.QtrFinance(r.Context(), code)
	writeJSON(w, resp, err)
}

func (s *F10Service) holderTrend(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" { writeError(w, 400, errMissingStockCode); return }
	resp, err := s.uc.HolderTrend(r.Context(), code)
	writeJSON(w, resp, err)
}

func (s *F10Service) orgPredict(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" { writeError(w, 400, errMissingStockCode); return }
	resp, err := s.uc.OrgPredict(r.Context(), code)
	writeJSON(w, resp, err)
}

func (s *F10Service) predictSummary(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" { writeError(w, 400, errMissingStockCode); return }
	resp, err := s.uc.PredictSummary(r.Context(), code)
	writeJSON(w, resp, err)
}
