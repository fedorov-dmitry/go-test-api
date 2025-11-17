package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fedorov-dmitry/go-test-api/internal"
	"github.com/fedorov-dmitry/go-test-api/internal/service"
)

type CurrencyRepository interface {
	Get(ctx context.Context, baseCurrency string, currency string, date time.Time) (internal.CurrencyRate, error)
	GetMany(ctx context.Context, baseCurrency string, date time.Time) ([]internal.CurrencyRate, error)
	Create(ctx context.Context, date time.Time, baseCurrency string, currency string, rate float64) (internal.CurrencyRate, error)
}

type Server struct {
	repo    CurrencyRepository
	service service.CurrencyService
}

func NewServer(repo CurrencyRepository, service service.CurrencyService) *Server {
	return &Server{repo: repo, service: service}
}

func (s *Server) Start() error {
	http.HandleFunc("/rates/historical", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		base := strings.ToLower(r.URL.Query().Get("base"))

		if base == "" {
			http.Error(w, "missing base query parameter", http.StatusBadRequest)
			return
		}

		dateStr := r.URL.Query().Get("date")

		var date time.Time
		var err error

		if dateStr == "" {
			http.Error(w, "missing `date` query parameter", http.StatusBadRequest)
			return
		} else {
			date, err = time.Parse("2006-01-02", dateStr)

			if err != nil {
				http.Error(w, "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
				return
			}
		}

		rates, err := s.repo.GetMany(ctx, base, date)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(rates) // handle?
	})

	http.HandleFunc("/rates/latest", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		base := strings.ToLower(r.URL.Query().Get("base"))
		currency := strings.ToLower(r.URL.Query().Get("currency"))

		if base == "" {
			http.Error(w, "missing `base` query parameter", http.StatusBadRequest)
			return
		}

		if currency == "" {
			http.Error(w, "missing `currency` query parameter", http.StatusBadRequest)
			return
		}

		rates, err := s.repo.Get(ctx, base, currency, time.Now())

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(rates) // handle?
	})

	err := http.ListenAndServe(":8080", nil) // handle?

	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
