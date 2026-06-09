package service

import (
	"encoding/json"
	"net/http"
	"strconv"

	bizcron "eino-stock/internal/biz/cron"
	"eino-stock/internal/data"
)

type CronService struct {
	uc *bizcron.CronUsecase
}

func NewCronService(uc *bizcron.CronUsecase) *CronService {
	return &CronService{uc: uc}
}

func (s *CronService) RegisterHTTP(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/cron/list", s.list)
	mux.HandleFunc("POST /api/cron/create", s.create)
	mux.HandleFunc("POST /api/cron/update", s.update)
	mux.HandleFunc("POST /api/cron/delete", s.delete)
	mux.HandleFunc("POST /api/cron/enable", s.enable)
	mux.HandleFunc("POST /api/cron/execute", s.executeNow)
}

func (s *CronService) list(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.uc.List(r.Context())
	writeJSON(w, tasks, err)
}

func (s *CronService) create(w http.ResponseWriter, r *http.Request) {
	var t data.CronTask
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, 400, err)
		return
	}
	err := s.uc.Create(r.Context(), &t)
	writeJSON(w, map[string]string{"status": "ok"}, err)
}

func (s *CronService) update(w http.ResponseWriter, r *http.Request) {
	var t data.CronTask
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, 400, err)
		return
	}
	err := s.uc.Update(r.Context(), &t)
	writeJSON(w, map[string]string{"status": "ok"}, err)
}

func (s *CronService) delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	err := s.uc.Delete(r.Context(), uint(id))
	writeJSON(w, map[string]string{"status": "ok"}, err)
}

func (s *CronService) enable(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	enable := r.URL.Query().Get("enable") == "true"
	err := s.uc.Enable(r.Context(), uint(id), enable)
	writeJSON(w, map[string]string{"status": "ok"}, err)
}

func (s *CronService) executeNow(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	err := s.uc.ExecuteNow(r.Context(), uint(id))
	writeJSON(w, map[string]string{"status": "ok"}, err)
}
