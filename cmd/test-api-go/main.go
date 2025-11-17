package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/fedorov-dmitry/go-test-api/internal"
	"github.com/fedorov-dmitry/go-test-api/internal/api"
	"github.com/fedorov-dmitry/go-test-api/internal/postgresql"
	"github.com/fedorov-dmitry/go-test-api/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	cfg := LoadConfig()

	pgxPool, err := pgxpool.New(ctx, cfg.ConnectionString)

	if err != nil {
		panic(fmt.Errorf("failed to create postgresql pool: %w", err))
	}

	storage := postgresql.NewCurrencyStorage(pgxPool)
	repository := internal.NewCurrencyRepository(storage)
	currencies := strings.Split(strings.ToLower(cfg.Currencies), ",")
	currencyService := service.NewCurrencyService(*repository, cfg.CurrencyApiBaseUrl, cfg.ApiKey, currencies)

	err = currencyService.SaveCurrencyRatesForLastNDays(ctx, cfg.DaysLookBack)

	if err != nil {
		fmt.Printf("failed to save currency rates for last %v days: %v", cfg.DaysLookBack, err)
	}

	server := api.NewServer(repository, *currencyService)
	err = server.Start()

	if err != nil {
		fmt.Printf("failed to start server: %v", err)
	}
}

func LoadConfig() Config {
	cfg := Config{}

	connectionString := flag.String("connection-string", "postgresql://postgres:password@localhost:5432/currencies", "Connection string for Postgresql")
	baseUrl := flag.String("api-base-url", "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api", "Base URL for currency rates API")
	apiKey := flag.String("api-key", "4a7a2a118a1216b06dc88819dd824eb7", "API key for currency rates API")
	currencies := flag.String("currencies", "EUR,USD,RUB,JPY", "Available currencies")
	daysLookBack := flag.Int("days-look-back", 1, "How many days to look back")

	cfg.ConnectionString = *connectionString
	cfg.CurrencyApiBaseUrl = *baseUrl
	cfg.ApiKey = *apiKey
	cfg.Currencies = *currencies
	cfg.DaysLookBack = *daysLookBack

	return cfg
}
