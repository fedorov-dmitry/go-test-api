package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fedorov-dmitry/go-test-api/internal"
	apimocks "github.com/fedorov-dmitry/go-test-api/internal/api/mocks"
	"github.com/fedorov-dmitry/go-test-api/internal/middleware"
	"go.uber.org/mock/gomock"
)

func TestCurrentRatesHandler_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := apimocks.NewMockCurrencyRepository(ctrl)
	ctx := context.Background()
	logCh := make(chan middleware.RequestLog, 1)

	s := NewServer(mockRepo, internal.CurrencySynchronizer{}, ctx, logCh, 0, "k")

	req := httptest.NewRequest(http.MethodGet, "/rates/latest?base=usd&currency=eur", nil)
	rr := httptest.NewRecorder()

	base := internal.NewCurrency("usd")
	cur := internal.NewCurrency("eur")

	mockRepo.
		EXPECT().
		Get(ctx, base, cur, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ internal.Currency, _ internal.Currency, _ time.Time) (internal.CurrencyRate, error) {
			return internal.CurrencyRate{
				Date:     time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				Base:     base,
				Currency: cur,
				Rate:     0.91,
			}, nil
		})

	s.currentRatesHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rr.Code)
	}
	var got internal.CurrencyRate
	if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Base != base || got.Currency != cur || got.Rate != 0.91 {
		t.Fatalf("unexpected body: %+v", got)
	}
}

func TestCurrentRatesHandler_MissingBase(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := apimocks.NewMockCurrencyRepository(ctrl)
	ctx := context.Background()
	logCh := make(chan middleware.RequestLog, 1)
	s := NewServer(mockRepo, internal.CurrencySynchronizer{}, ctx, logCh, 0, "k")

	req := httptest.NewRequest(http.MethodGet, "/rates/latest?currency=eur", nil)
	rr := httptest.NewRecorder()

	s.currentRatesHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestCurrentRatesHandler_RepoError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := apimocks.NewMockCurrencyRepository(ctrl)
	ctx := context.Background()
	logCh := make(chan middleware.RequestLog, 1)
	s := NewServer(mockRepo, internal.CurrencySynchronizer{}, ctx, logCh, 0, "k")

	req := httptest.NewRequest(http.MethodGet, "/rates/latest?base=usd&currency=eur", nil)
	rr := httptest.NewRecorder()

	mockRepo.
		EXPECT().
		Get(ctx, internal.NewCurrency("usd"), internal.NewCurrency("eur"), gomock.Any()).
		Return(internal.CurrencyRate{}, errors.New("db error"))

	s.currentRatesHandler(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", rr.Code)
	}
}

func TestHistoricalRatesHandler_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := apimocks.NewMockCurrencyRepository(ctrl)
	ctx := context.Background()
	logCh := make(chan middleware.RequestLog, 1)
	s := NewServer(mockRepo, internal.CurrencySynchronizer{}, ctx, logCh, 0, "k")

	req := httptest.NewRequest(http.MethodGet, "/rates/historical?base=usd&date=2025-01-13", nil)
	rr := httptest.NewRecorder()

	date := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)
	mockRepo.
		EXPECT().
		GetMany(ctx, internal.NewCurrency("usd"), date).
		Return([]internal.CurrencyRate{
			{Date: date, Base: internal.NewCurrency("usd"), Currency: internal.NewCurrency("eur"), Rate: 0.92},
		}, nil)

	s.historicalRatesHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rr.Code)
	}
	var got []internal.CurrencyRate
	if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 1 || got[0].Currency != internal.NewCurrency("eur") {
		t.Fatalf("unexpected body: %+v", got)
	}
}

func TestHistoricalRatesHandler_InvalidDate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := apimocks.NewMockCurrencyRepository(ctrl)
	ctx := context.Background()
	logCh := make(chan middleware.RequestLog, 1)
	s := NewServer(mockRepo, internal.CurrencySynchronizer{}, ctx, logCh, 0, "k")

	req := httptest.NewRequest(http.MethodGet, "/rates/historical?base=usd&date=13-01-2025", nil)
	rr := httptest.NewRecorder()

	s.historicalRatesHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestGetHandlers_WithMiddleware(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := apimocks.NewMockCurrencyRepository(ctrl)
	ctx := context.Background()
	logCh := make(chan middleware.RequestLog, 10)
	s := NewServer(mockRepo, internal.CurrencySynchronizer{}, ctx, logCh, 0, "secret")

	// Setup repo expectation for authorized request
	mockRepo.
		EXPECT().
		Get(ctx, internal.NewCurrency("usd"), internal.NewCurrency("eur"), gomock.Any()).
		Return(internal.CurrencyRate{Base: internal.NewCurrency("usd"), Currency: internal.NewCurrency("eur"), Rate: 1.0}, nil).
		Times(1)

	ts := httptest.NewServer(s.getHandlers())
	defer ts.Close()

	// Unauthorized request
	resp, err := http.Get(ts.URL + "/rates/latest?base=usd&currency=eur")
	if err != nil {
		t.Fatalf("http get: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized status %d, want 401", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// Authorized request
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/rates/latest?base=usd&currency=eur", nil)
	req.Header.Set("Authorization", "secret")
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http do: %v", err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("authorized status %d, want 200", resp2.StatusCode)
	}
	_ = resp2.Body.Close()

	// Logging middleware should emit a log entry
	select {
	case entry := <-logCh:
		if entry.Path != "/rates/latest" {
			t.Fatalf("unexpected log path: %s", entry.Path)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected a log entry, got none")
	}
}
