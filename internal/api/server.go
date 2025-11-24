package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fedorov-dmitry/go-test-api/internal"
)

type CurrencyRepository interface {
	Get(ctx context.Context, baseCurrency internal.Currency, currency internal.Currency, date time.Time) (internal.CurrencyRate, error)
	GetMany(ctx context.Context, baseCurrency internal.Currency, date time.Time) ([]internal.CurrencyRate, error)
	Create(ctx context.Context, date time.Time, baseCurrency internal.Currency, currency internal.Currency, rate float64) (internal.CurrencyRate, error)
}

type Server struct {
	repo        CurrencyRepository
	service     internal.CurrencyService
	mainContext context.Context
}

func NewServer(repo CurrencyRepository, service internal.CurrencyService, mainContext context.Context) *Server {
	return &Server{repo: repo, service: service, mainContext: mainContext}
}

func (s *Server) Start() error {
	err := http.ListenAndServe(":"+os.Getenv("APP_PORT"), s.getHandlers()) // handle?
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) getHandlers() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/rates/historical", authorizationMiddleware(http.HandlerFunc(s.historicalRatesHandler)))
	mux.Handle("/rates/latest", authorizationMiddleware(http.HandlerFunc(s.currentRatesHandler)))

	return mux
}

func authorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("Authorization")
		if apiKey != os.Getenv("API_KEY") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
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
