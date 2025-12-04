package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fedorov-dmitry/go-test-api/internal"
	"github.com/fedorov-dmitry/go-test-api/internal/middleware"
)

type CurrencyRepository interface {
	Get(ctx context.Context, baseCurrency internal.Currency, currency internal.Currency, date time.Time) (internal.CurrencyRate, error)
	GetMany(ctx context.Context, baseCurrency internal.Currency, date time.Time) ([]internal.CurrencyRate, error)
	Create(ctx context.Context, date time.Time, baseCurrency internal.Currency, currency internal.Currency, rate float64) (internal.CurrencyRate, error)
}

type Server struct {
	repo        CurrencyRepository
	service     internal.CurrencySynchronizer
	mainContext context.Context
	logCh       chan<- middleware.RequestLog
	port        int
	apiKey      string
}

func NewServer(repo CurrencyRepository, service internal.CurrencySynchronizer, mainContext context.Context, logCh chan<- middleware.RequestLog, port int, apiKey string) *Server {
	return &Server{repo: repo, service: service, mainContext: mainContext, logCh: logCh, port: port, apiKey: apiKey}
}

func (s *Server) Start() error {
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.getHandlers()) // handle?
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) getHandlers() *http.ServeMux {
	mux := http.NewServeMux()

	wrap := func(h http.HandlerFunc) http.Handler {
		return middleware.RequestLoggingMiddleware(
			s.logCh,
			middleware.AuthorizationMiddleware(s.apiKey, h),
		)
	}

	mux.Handle("/rates/historical", wrap(s.historicalRatesHandler))
	mux.Handle("/rates/latest", wrap(s.currentRatesHandler))

	return mux
}

func (s *Server) currentRatesHandler(w http.ResponseWriter, r *http.Request) {
	base := internal.NewCurrency(r.URL.Query().Get("base"))
	if base == "" {
		http.Error(w, "missing `base` query parameter", http.StatusBadRequest)
		return
	}

	currency := internal.NewCurrency(r.URL.Query().Get("currency"))
	if currency == "" {
		http.Error(w, "missing `currency` query parameter", http.StatusBadRequest)
		return
	}

	rates, err := s.repo.Get(s.mainContext, base, currency, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(rates) // handle?
}

func (s *Server) historicalRatesHandler(w http.ResponseWriter, r *http.Request) {
	base := internal.NewCurrency(r.URL.Query().Get("base"))
	if base == "" {
		http.Error(w, "missing base query parameter", http.StatusBadRequest)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "missing `date` query parameter", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	rates, err := s.repo.GetMany(s.mainContext, internal.Currency(base), date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(rates) // handle?
}
