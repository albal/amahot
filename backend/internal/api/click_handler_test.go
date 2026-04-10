package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/albal/amahot/backend/internal/models"
)

// stubDealGetter implements dealGetter for tests.
type stubDealGetter struct {
	result models.Deal
	err    error
}

func (s *stubDealGetter) GetByID(_ context.Context, _ int64) (models.Deal, error) {
	return s.result, s.err
}

// stubClickInserter implements clickInserter for tests.
type stubClickInserter struct {
	inserted []models.Click
	err      error
}

func (s *stubClickInserter) Insert(_ context.Context, c models.Click) error {
	s.inserted = append(s.inserted, c)
	return s.err
}

func chiRoute(method, pattern string, handler http.HandlerFunc) http.Handler {
	r := chi.NewRouter()
	r.Method(method, pattern, handler)
	return r
}

func TestClickHandler_Record_OK(t *testing.T) {
	now := time.Now()
	dealStore := &stubDealGetter{
		result: models.Deal{
			ID:        7,
			Title:     "Great Deal",
			DealURL:   "https://www.amazon.co.uk/dp/B001?tag=prbox",
			ScrapedAt: now,
			UpdatedAt: now,
		},
	}
	clickStore := &stubClickInserter{}

	h := NewClickHandler(dealStore, clickStore)
	router := chiRoute(http.MethodPost, "/api/clicks/{id}", h.Record)

	req := httptest.NewRequest(http.MethodPost, "/api/clicks/7", nil)
	req.Header.Set("User-Agent", "TestBrowser/1.0")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp clickResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.RedirectURL != "https://www.amazon.co.uk/dp/B001?tag=prbox" {
		t.Errorf("redirect_url = %q, want Amazon URL", resp.RedirectURL)
	}
}

func TestClickHandler_Record_LogsClick(t *testing.T) {
	now := time.Now()
	dealStore := &stubDealGetter{result: models.Deal{ID: 5, DealURL: "https://amazon.co.uk/dp/B?tag=prbox", ScrapedAt: now, UpdatedAt: now}}
	clickStore := &stubClickInserter{}

	h := NewClickHandler(dealStore, clickStore)
	router := chiRoute(http.MethodPost, "/api/clicks/{id}", h.Record)

	req := httptest.NewRequest(http.MethodPost, "/api/clicks/5", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://amahot.tsew.com/")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if len(clickStore.inserted) != 1 {
		t.Fatalf("expected 1 click inserted, got %d", len(clickStore.inserted))
	}
	click := clickStore.inserted[0]
	if click.DealID != 5 {
		t.Errorf("click.DealID = %d, want 5", click.DealID)
	}
	if click.UserAgent != "Mozilla/5.0" {
		t.Errorf("click.UserAgent = %q, want Mozilla/5.0", click.UserAgent)
	}
	if click.Referer != "https://amahot.tsew.com/" {
		t.Errorf("click.Referer = %q", click.Referer)
	}
}

func TestClickHandler_Record_InvalidID(t *testing.T) {
	h := NewClickHandler(&stubDealGetter{}, &stubClickInserter{})
	router := chiRoute(http.MethodPost, "/api/clicks/{id}", h.Record)

	req := httptest.NewRequest(http.MethodPost, "/api/clicks/notanumber", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestClickHandler_Record_DealNotFound(t *testing.T) {
	dealStore := &stubDealGetter{err: errors.New("no rows")}
	h := NewClickHandler(dealStore, &stubClickInserter{})
	router := chiRoute(http.MethodPost, "/api/clicks/{id}", h.Record)

	req := httptest.NewRequest(http.MethodPost, "/api/clicks/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestClickHandler_Record_ClickInsertFailureDoesNotBlock(t *testing.T) {
	// Even if the click insert fails, the redirect URL is still returned.
	now := time.Now()
	dealStore := &stubDealGetter{result: models.Deal{ID: 1, DealURL: "https://amazon.co.uk/dp/B?tag=prbox", ScrapedAt: now, UpdatedAt: now}}
	clickStore := &stubClickInserter{err: errors.New("db timeout")}

	h := NewClickHandler(dealStore, clickStore)
	router := chiRoute(http.MethodPost, "/api/clicks/{id}", h.Record)

	req := httptest.NewRequest(http.MethodPost, "/api/clicks/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 even on click insert failure", w.Code)
	}
}

func TestRealIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 10.0.0.1")
	got := realIP(req)
	if got != "1.2.3.4" {
		t.Errorf("realIP = %q, want 1.2.3.4", got)
	}
}

func TestRealIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "5.6.7.8")
	got := realIP(req)
	if got != "5.6.7.8" {
		t.Errorf("realIP = %q, want 5.6.7.8", got)
	}
}

func TestRealIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "9.10.11.12:45678"
	got := realIP(req)
	if got != "9.10.11.12" {
		t.Errorf("realIP = %q, want 9.10.11.12", got)
	}
}
