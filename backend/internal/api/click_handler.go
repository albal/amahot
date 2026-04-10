package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/albal/amahot/backend/internal/models"
)

// dealGetter is the subset of repository.DealRepo used by ClickHandler.
type dealGetter interface {
	GetByID(ctx context.Context, id int64) (models.Deal, error)
}

// clickInserter is the subset of repository.ClickRepo used by ClickHandler.
type clickInserter interface {
	Insert(ctx context.Context, c models.Click) error
}

type clickResponse struct {
	RedirectURL string `json:"redirect_url"`
}

type ClickHandler struct {
	deals  dealGetter
	clicks clickInserter
}

func NewClickHandler(deals dealGetter, clicks clickInserter) *ClickHandler {
	return &ClickHandler{deals: deals, clicks: clicks}
}

func (h *ClickHandler) Record(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	deal, err := h.deals.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "deal not found", http.StatusNotFound)
		return
	}

	click := models.Click{
		DealID:    id,
		IPAddress: realIP(r),
		UserAgent: r.UserAgent(),
		Referer:   r.Referer(),
	}

	// Best-effort: don't fail the redirect if click logging fails.
	_ = h.clicks.Insert(r.Context(), click)

	writeJSON(w, http.StatusOK, clickResponse{RedirectURL: deal.DealURL})
}

// realIP extracts the real client IP, respecting X-Forwarded-For set by Caddy.
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		addr = addr[:idx]
	}
	return strings.Trim(addr, "[]")
}
