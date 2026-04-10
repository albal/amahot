package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/albal/amahot/backend/internal/models"
	"github.com/albal/amahot/backend/internal/repository"
)

// dealStore is the subset of repository.DealRepo used by DealsHandler.
// Defining it here keeps the handler decoupled from the concrete repo type
// and makes it straightforward to inject a stub in tests.
type dealStore interface {
	ListPaginated(ctx context.Context, p repository.ListParams) (repository.ListResult, error)
	GetByID(ctx context.Context, id int64) (models.Deal, error)
}

type dealsResponse struct {
	Data    []models.Deal `json:"data"`
	Total   int           `json:"total"`
	Limit   int           `json:"limit"`
	Offset  int           `json:"offset"`
	HasMore bool          `json:"has_more"`
}

type DealsHandler struct {
	repo dealStore
}

func NewDealsHandler(repo dealStore) *DealsHandler {
	return &DealsHandler{repo: repo}
}

func (h *DealsHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	result, err := h.repo.ListPaginated(r.Context(), repository.ListParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := result.Deals
	if data == nil {
		data = []models.Deal{}
	}

	writeJSON(w, http.StatusOK, dealsResponse{
		Data:    data,
		Total:   result.Total,
		Limit:   limit,
		Offset:  offset,
		HasMore: result.HasMore,
	})
}

func (h *DealsHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	deal, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, deal)
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
