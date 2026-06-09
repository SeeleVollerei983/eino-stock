package service

import (
	"encoding/json"
	"net/http"
	"strconv"

	bizprompt "eino-stock/internal/biz/prompt"
	"eino-stock/internal/data"
)

type PromptService struct {
	uc *bizprompt.PromptUsecase
}

func NewPromptService(uc *bizprompt.PromptUsecase) *PromptService {
	return &PromptService{uc: uc}
}

func (s *PromptService) RegisterHTTP(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/prompt/list", s.list)
	mux.HandleFunc("POST /api/prompt/create", s.create)
	mux.HandleFunc("POST /api/prompt/update", s.update)
	mux.HandleFunc("POST /api/prompt/delete", s.delete)
}

func (s *PromptService) list(w http.ResponseWriter, r *http.Request) {
	list, err := s.uc.List(r.Context())
	writeJSON(w, list, err)
}

func (s *PromptService) create(w http.ResponseWriter, r *http.Request) {
	var p data.PromptTemplate
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, 400, err); return
	}
	err := s.uc.Create(r.Context(), &p)
	writeJSON(w, map[string]string{"status": "ok", "id": strconv.Itoa(int(p.ID))}, err)
}

func (s *PromptService) update(w http.ResponseWriter, r *http.Request) {
	var p data.PromptTemplate
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, 400, err); return
	}
	err := s.uc.Update(r.Context(), &p)
	writeJSON(w, map[string]string{"status": "ok"}, err)
}

func (s *PromptService) delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	err := s.uc.Delete(r.Context(), uint(id))
	writeJSON(w, map[string]string{"status": "ok"}, err)
}
