package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/fedorov-dmitry/go-test-api/internal"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CurrencyStorage struct {
	pgPool *pgxpool.Pool
}

func NewCurrencyStorage(pgPool *pgxpool.Pool) *CurrencyStorage {
	return &CurrencyStorage{pgPool: pgPool}
}

func (c *CurrencyStorage) Get(ctx context.Context, baseCurrency string, currency string, date time.Time) (internal.CurrencyRate, error) {
	sql := `
SELECT date, base, currency, rate FROM app.currency_rates
WHERE base = $1
  AND currency = $2
  AND date = $3`

	res := c.pgPool.QueryRow(ctx, sql, baseCurrency, currency, date.Format("2006-01-02"))
	rates := internal.CurrencyRate{}

	err := res.Scan(&rates.Date, &rates.Base, &rates.Currency, &rates.Rate)

	if err != nil {
		return internal.CurrencyRate{}, fmt.Errorf("failed to get currency rates for %s-%s: %w", baseCurrency, currency, err)
	}

	return rates, nil
}

func (c *CurrencyStorage) GetMany(ctx context.Context, baseCurrency string, date time.Time) ([]internal.CurrencyRate, error) {
	sql := `
SELECT date, base, currency, rate FROM app.currency_rates
WHERE base = $1
AND date = $2`

	rows, err := c.pgPool.Query(ctx, sql, baseCurrency, date.Format("2006-01-02"))

	if err != nil {
		return nil, fmt.Errorf("failed to fetch currency rates for %s: %w", baseCurrency, err)
	}

	rates := make([]internal.CurrencyRate, 0)

	for rows.Next() {
		rate := internal.CurrencyRate{}
		err = rows.Scan(&rate.Date, &rate.Base, &rate.Currency, &rate.Rate)

		if err != nil {
			return nil, fmt.Errorf("failed to fetch currency rates for %s: %w", baseCurrency, err)
		}

		rates = append(rates, rate)
	}

	return rates, nil
}

func (c *CurrencyStorage) Set(ctx context.Context, rate internal.CurrencyRate) error {
	sql := `
INSERT INTO app.currency_rates
VALUES ($1, $2, $3, $4)
ON CONFLICT (date, base, currency)
DO UPDATE SET
   rate = EXCLUDED.rate;`

	_, err := c.pgPool.Exec(ctx, sql, rate.Date, rate.Base, rate.Currency, rate.Rate)

	if err != nil {
		return fmt.Errorf("failed to save currency rate for %s-%s: %w", rate.Base, rate.Currency, err)
	}

	return nil
}
