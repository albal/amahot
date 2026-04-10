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
	"github.com/albal/amahot/backend/internal/repository"
)

// stubDealStore implements dealStore for tests.
type stubDealStore struct {
	listResult repository.ListResult
	listErr    error
	getResult  models.Deal
	getErr     error
}

func (s *stubDealStore) ListPaginated(_ context.Context, _ repository.ListParams) (repository.ListResult, error) {
	return s.listResult, s.listErr
}

func (s *stubDealStore) GetByID(_ context.Context, _ int64) (models.Deal, error) {
	return s.getResult, s.getErr
}

func TestDealsHandler_List_OK(t *testing.T) {
	now := time.Now()
	store := &stubDealStore{
		listResult: repository.ListResult{
			Deals: []models.Deal{
				{ID: 1, Title: "Headphones", Temperature: 342, DealURL: "https://amazon.co.uk/dp/B001?tag=prbox", ScrapedAt: now, UpdatedAt: now},
				{ID: 2, Title: "Keyboard", Temperature: 200, DealURL: "https://amazon.co.uk/dp/B002?tag=prbox", ScrapedAt: now, UpdatedAt: now},
			},
			Total:   2,
			HasMore: false,
		},
	}

	h := NewDealsHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/api/deals?limit=20&offset=0", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp dealsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Errorf("len(data) = %d, want 2", len(resp.Data))
	}
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}
	if resp.HasMore {
		t.Error("has_more should be false")
	}
}

func TestDealsHandler_List_EmptyIsArray(t *testing.T) {
	// Nil slice from repo must serialise as [] not null
	store := &stubDealStore{
		listResult: repository.ListResult{Deals: nil, Total: 0, HasMore: false},
	}

	h := NewDealsHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/api/deals", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	body := w.Body.String()
	if body == "" {
		t.Fatal("empty body")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	data, ok := resp["data"]
	if !ok {
		t.Fatal("missing data field")
	}
	arr, ok := data.([]interface{})
	if !ok {
		t.Fatalf("data should be array, got %T", data)
	}
	if len(arr) != 0 {
		t.Errorf("expected empty array, got len=%d", len(arr))
	}
}

func TestDealsHandler_List_LimitClamped(t *testing.T) {
	var capturedParams repository.ListParams
	store := &stubDealStore{}
	store.listResult.Deals = []models.Deal{}

	// Override to capture params
	h := &DealsHandler{repo: &capturingStore{
		captured: &capturedParams,
		inner:    store,
	}}

	req := httptest.NewRequest(http.MethodGet, "/api/deals?limit=999&offset=0", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if capturedParams.Limit > 50 {
		t.Errorf("limit should be clamped to 50, got %d", capturedParams.Limit)
	}
}

// capturingStore wraps a dealStore and records the ListParams it receives.
type capturingStore struct {
	captured *repository.ListParams
	inner    dealStore
}

func (c *capturingStore) ListPaginated(ctx context.Context, p repository.ListParams) (repository.ListResult, error) {
	*c.captured = p
	return c.inner.ListPaginated(ctx, p)
}

func (c *capturingStore) GetByID(ctx context.Context, id int64) (models.Deal, error) {
	return c.inner.GetByID(ctx, id)
}

func TestDealsHandler_List_RepoError(t *testing.T) {
	store := &stubDealStore{listErr: errors.New("db down")}
	h := NewDealsHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/api/deals", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestDealsHandler_GetOne_OK(t *testing.T) {
	now := time.Now()
	store := &stubDealStore{
		getResult: models.Deal{ID: 42, Title: "Widget", Temperature: 150, ScrapedAt: now, UpdatedAt: now},
	}
	h := NewDealsHandler(store)

	// Use chi router so URLParam works
	r := chi.NewRouter()
	r.Get("/api/deals/{id}", h.GetOne)

	req := httptest.NewRequest(http.MethodGet, "/api/deals/42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var deal models.Deal
	if err := json.NewDecoder(w.Body).Decode(&deal); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if deal.ID != 42 {
		t.Errorf("deal.ID = %d, want 42", deal.ID)
	}
}

func TestDealsHandler_GetOne_InvalidID(t *testing.T) {
	h := NewDealsHandler(&stubDealStore{})
	r := chi.NewRouter()
	r.Get("/api/deals/{id}", h.GetOne)

	req := httptest.NewRequest(http.MethodGet, "/api/deals/notanumber", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestDealsHandler_GetOne_NotFound(t *testing.T) {
	store := &stubDealStore{getErr: errors.New("not found")}
	h := NewDealsHandler(store)
	r := chi.NewRouter()
	r.Get("/api/deals/{id}", h.GetOne)

	req := httptest.NewRequest(http.MethodGet, "/api/deals/99", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestDealsHandler_ContentType(t *testing.T) {
	store := &stubDealStore{listResult: repository.ListResult{Deals: []models.Deal{}}}
	h := NewDealsHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/api/deals", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}
