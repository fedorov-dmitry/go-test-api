package internal

import (
	"context"
	"fmt"
	"time"
)

type CurrencySynchronizer struct {
	repository CurrencyRepository
	source     CurrencyRateSource
	currencies []Currency
}

func NewCurrencySynchronizer(repository CurrencyRepository, source CurrencyRateSource, currencies []Currency) *CurrencySynchronizer {
	return &CurrencySynchronizer{
		repository: repository,
		source:     source,
		currencies: currencies,
	}
}

func (c *CurrencySynchronizer) UpdateCurrencyRatesForTodayAndLastNDays(ctx context.Context, days int) error {
	for _, base := range c.currencies {
		otherCurrencies := make([]Currency, 0, len(c.currencies)-1)
		for _, cur := range c.currencies {
			if cur != base {
				otherCurrencies = append(otherCurrencies, cur)
			}
		}

		for i := 0; i <= days; i++ {
			date := time.Now().AddDate(0, 0, -i)

			result, err := c.source.Get(base, otherCurrencies, date)
			if err != nil {
				return fmt.Errorf("failed to get currency rates for %s for %s: %w", base, date.Format("2006-01-02"), err)
			}

			for _, rate := range result {
				_, err := c.repository.Create(ctx, date, base, rate.Currency, rate.Rate)
				if err != nil {
					return fmt.Errorf("failed to create currency rate for %s-%s for %s: %w", base, rate.Currency, date, err)
				}
			}
		}
	}

	return nil
}
