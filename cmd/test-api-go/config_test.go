package main

import (
	"testing"
)

func TestLoadConfig_ParsesEnv(t *testing.T) {
	t.Setenv("DAYS_LOOK_BACK", "2")
	t.Setenv("JOB_CRON", "0 * * * *")
	t.Setenv("APP_PORT", "8088")
	t.Setenv("CONNECTION_STRING", "postgresql://user:pass@localhost:5432/db")
	t.Setenv("API_BASE_URL", "http://example")
	t.Setenv("CURRENCIES", "USD,EUR,JPY")
	t.Setenv("API_KEY", "secret")

	cfg := LoadConfig()

	if cfg.DaysLookBack != 2 {
		t.Fatalf("DaysLookBack=%d, want 2", cfg.DaysLookBack)
	}
	if cfg.JobCron != "0 * * * *" {
		t.Fatalf("JobCron=%s", cfg.JobCron)
	}
	if cfg.AppPort != 8088 {
		t.Fatalf("AppPort=%d, want 8088", cfg.AppPort)
	}
	if cfg.ConnectionString != "postgresql://user:pass@localhost:5432/db" {
		t.Fatalf("ConnectionString=%s", cfg.ConnectionString)
	}
	if cfg.CurrencyApiBaseUrl != "http://example" {
		t.Fatalf("CurrencyApiBaseUrl=%s", cfg.CurrencyApiBaseUrl)
	}
	if cfg.Currencies != "USD,EUR,JPY" {
		t.Fatalf("Currencies=%s", cfg.Currencies)
	}
	if cfg.ApiKey != "secret" {
		t.Fatalf("ApiKey=%s", cfg.ApiKey)
	}
}
