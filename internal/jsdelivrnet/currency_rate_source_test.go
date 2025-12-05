package jsdelivrnet_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fedorov-dmitry/go-test-api/internal"
	"github.com/fedorov-dmitry/go-test-api/internal/jsdelivrnet"
)

func TestCurrencyRateSource_Get_Success(t *testing.T) {
	t.Parallel()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// respond with {"usd": {"eur": 0.92, "jpy": 145.1}}
		resp := map[string]map[string]float64{
			"usd": {
				"eur": 0.92,
				"jpy": 145.1,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	srv := httptest.NewServer(h)
	defer srv.Close()

	source := jsdelivrnet.NewCurrencyRateSource(srv.URL + "/")

	date := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)
	base := internal.NewCurrency("usd")
	targets := []internal.Currency{internal.NewCurrency("eur"), internal.NewCurrency("jpy")}

	rates, err := source.Get(base, targets, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rates) != 2 {
		t.Fatalf("expected 2 rates, got %d", len(rates))
	}
	if rates[0].Base != base || rates[0].Currency != internal.NewCurrency("eur") || rates[0].Rate != 0.92 || !rates[0].Date.Equal(date) {
		t.Fatalf("unexpected first rate: %+v", rates[0])
	}
	if rates[1].Base != base || rates[1].Currency != internal.NewCurrency("jpy") || rates[1].Rate != 145.1 || !rates[1].Date.Equal(date) {
		t.Fatalf("unexpected second rate: %+v", rates[1])
	}
}

func TestCurrencyRateSource_Get_BadStatus(t *testing.T) {
	t.Parallel()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad upstream", http.StatusBadGateway)
	})
	srv := httptest.NewServer(h)
	defer srv.Close()

	source := jsdelivrnet.NewCurrencyRateSource(srv.URL + "/")

	_, err := source.Get(internal.NewCurrency("usd"), []internal.Currency{internal.NewCurrency("eur")}, time.Now())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCurrencyRateSource_Get_BadJSON(t *testing.T) {
	t.Parallel()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{not-json"))
	})
	srv := httptest.NewServer(h)
	defer srv.Close()

	source := jsdelivrnet.NewCurrencyRateSource(srv.URL + "/")

	_, err := source.Get(internal.NewCurrency("usd"), []internal.Currency{internal.NewCurrency("eur")}, time.Now())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
