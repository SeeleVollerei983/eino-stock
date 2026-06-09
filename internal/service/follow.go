package service

import (
	"fmt"
	"net/http"

	"eino-stock/internal/biz/follow"
)

type FollowService struct {
	uc *follow.FollowUsecase
}

func NewFollowService(uc *follow.FollowUsecase) *FollowService {
	return &FollowService{uc: uc}
}

func (s *FollowService) List(w http.ResponseWriter, r *http.Request) {
	stocks, err := s.uc.List(r.Context())
	writeJSON(w, stocks, err)
}

func (s *FollowService) Add(w http.ResponseWriter, r *http.Request) {
	code, name := r.URL.Query().Get("code"), r.URL.Query().Get("name")
	if code == "" || name == "" { writeError(w, 400, fmt.Errorf("missing code or name")); return }
	writeJSON(w, map[string]string{"status": "ok"}, s.uc.Add(r.Context(), code, name))
}

func (s *FollowService) Remove(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" { writeError(w, 400, fmt.Errorf("missing code")); return }
	writeJSON(w, map[string]string{"status": "ok"}, s.uc.Remove(r.Context(), code))
}