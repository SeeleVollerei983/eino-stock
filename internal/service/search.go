package service

import (
	"net/http"

	bizsearch "eino-stock/internal/biz/search"
)

type SearchService struct {
	uc *bizsearch.SearchUsecase
}

func NewSearchService() *SearchService {
	return &SearchService{uc: bizsearch.NewSearchUsecase()}
}

func (s *SearchService) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" { writeError(w, 400, errMissingQuery); return }
	results, err := s.uc.Search(r.Context(), q)
	writeJSON(w, results, err)
}
