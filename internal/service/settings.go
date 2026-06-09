package service

import (
	"fmt"
	"net/http"

	"eino-stock/internal/data"
)

// SettingsService 系统设置服务。
type SettingsService struct {
	repo *data.SettingRepo
}

func NewSettingsService(repo *data.SettingRepo) *SettingsService {
	return &SettingsService{repo: repo}
}

func (s *SettingsService) Get(w http.ResponseWriter, r *http.Request) {
	settings, err := s.repo.GetAll(r.Context())
	writeJSON(w, settings, err)
}

func (s *SettingsService) Set(w http.ResponseWriter, r *http.Request) {
	key, value := r.URL.Query().Get("key"), r.URL.Query().Get("value")
	if key == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("missing key"))
		return
	}
	err := s.repo.Set(r.Context(), key, value)
	writeJSON(w, map[string]string{"status": "ok"}, err)
}
