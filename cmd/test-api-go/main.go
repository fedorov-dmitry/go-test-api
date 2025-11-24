package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/fedorov-dmitry/go-test-api/internal"
	"github.com/fedorov-dmitry/go-test-api/internal/api"
	"github.com/fedorov-dmitry/go-test-api/internal/jsdelivrnet"
	"github.com/fedorov-dmitry/go-test-api/internal/postgresql"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	cfg := LoadConfig()

	pgxPool, err := pgxpool.New(ctx, cfg.ConnectionString)
	if err != nil {
		log.Fatalf("failed to create postgresql pool: %v", err)
	}

	storage := postgresql.NewCurrencyStorage(pgxPool)
	repository := internal.NewCurrencyRepository(storage)
	currencyRateSource := jsdelivrnet.NewCurrencyRateSource()
	currencyService := internal.NewCurrencyService(*repository, currencyRateSource)

	err = currencyService.SaveCurrencyRatesForLastNDays(ctx, cfg.DaysLookBack)
	if err != nil {
		fmt.Printf("failed to save currency rates for last %v days: %v", cfg.DaysLookBack, err)
	}

	server := api.NewServer(repository, *currencyService, ctx)

	err = server.Start()
	if err != nil {
		fmt.Printf("failed to start server: %v", err)
	}
}

func LoadConfig() Config {
	cfg := Config{}

	daysLookBack, err := strconv.Atoi(os.Getenv("DAYS_LOOK_BACK"))
	if err != nil {
		log.Fatalf("failed to parse DAYS_LOOK_BACK env var: %v", err)
	}

	cfg.ConnectionString = os.Getenv("CONNECTION_STRING")
	cfg.CurrencyApiBaseUrl = os.Getenv("API_BASE_URL")
	cfg.Currencies = os.Getenv("CURRENCIES")
	cfg.DaysLookBack = daysLookBack

	return cfg
}
