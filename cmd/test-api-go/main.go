package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/fedorov-dmitry/go-test-api/internal"
	"github.com/fedorov-dmitry/go-test-api/internal/api"
	"github.com/fedorov-dmitry/go-test-api/internal/jsdelivrnet"
	"github.com/fedorov-dmitry/go-test-api/internal/logging"
	"github.com/fedorov-dmitry/go-test-api/internal/middleware"
	"github.com/fedorov-dmitry/go-test-api/internal/postgresql"
	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.SetOutput(os.Stdout)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	cfg := LoadConfig()

	pgxPool, err := pgxpool.New(ctx, cfg.ConnectionString)
	if err != nil {
		log.Fatalf("failed to create postgresql pool: %v", err)
	}

	storage := postgresql.NewCurrencyStorage(pgxPool)
	repository := internal.NewCurrencyRepository(storage)
	currencyRateSource := jsdelivrnet.NewCurrencyRateSource(cfg.CurrencyApiBaseUrl)

	currencyStrings := strings.Split(strings.ToLower(cfg.Currencies), ",")
	currencies := make([]internal.Currency, len(currencyStrings))
	for i, c := range currencyStrings {
		currencies[i] = internal.NewCurrency(c)
	}

	currencySynchronizer := internal.NewCurrencySynchronizer(*repository, currencyRateSource, currencies)

	err = currencySynchronizer.UpdateCurrencyRatesForTodayAndLastNDays(ctx, cfg.DaysLookBack)
	if err != nil {
		log.Printf("failed to save currency rates for last %v days: %v", cfg.DaysLookBack, err)
	}

	s, err := gocron.NewScheduler()

	if err != nil {
		log.Printf("failed to start scheduler for currency rate synchronizer: %v\n", err)
	} else {
		_, err := s.NewJob(
			gocron.CronJob(cfg.JobCron, false),
			gocron.NewTask(currencySynchronizer.UpdateCurrencyRatesForTodayAndLastNDays, ctx, 0),
		)

		if err != nil {
			log.Printf("failed to start currency rate synchronizer: %v\n", err)
		}

		s.Start()
	}

	logCh := make(chan middleware.RequestLog, 1000)
	logging.StartLoggingWorker(ctx, pgxPool, logCh)

	server := api.NewServer(repository, *currencySynchronizer, ctx, logCh, cfg.AppPort, cfg.ApiKey)

	err = server.Start()
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	select {
	case <-ctx.Done():
		log.Println("context cancelled, shutting down scheduler...")
	case sig := <-sigChan:
		log.Println("received signal:", sig)
	}

	err = s.Shutdown()
	if err != nil {
		log.Fatalf("failed to shutdown scheduler: %v", err)
	}

	log.Println("currency rate synchronizer scheduler stopped cleanly")
}

func LoadConfig() Config {
	cfg := Config{}

	daysLookBack, err := strconv.Atoi(os.Getenv("DAYS_LOOK_BACK"))
	if err != nil {
		log.Fatalf("failed to parse DAYS_LOOK_BACK env var: %v", err)
	}

	jobCron := os.Getenv("JOB_CRON")
	if jobCron == "" {
		jobCron = "* * * * *"
	}

	appPortStr := os.Getenv("APP_PORT")
	if appPortStr == "" {
		appPortStr = "8080"
	}
	appPort, err := strconv.Atoi(appPortStr)
	if err != nil {
		log.Fatalf("failed to parse APP_PORT env var: %v", err)
	}

	cfg.ConnectionString = os.Getenv("CONNECTION_STRING")
	cfg.CurrencyApiBaseUrl = os.Getenv("API_BASE_URL")
	cfg.Currencies = os.Getenv("CURRENCIES")
	cfg.DaysLookBack = daysLookBack
	cfg.JobCron = jobCron
	cfg.AppPort = appPort
	cfg.ApiKey = os.Getenv("API_KEY")

	return cfg
}
