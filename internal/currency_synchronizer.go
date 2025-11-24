package internal

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type CurrencyService struct {
	repository CurrencyRepository
	source     CurrencyRateSource
}

func NewCurrencyService(repository CurrencyRepository, source CurrencyRateSource) *CurrencyService {
	return &CurrencyService{
		repository: repository,
		source:     source,
	}
}

func (c *CurrencyService) SaveCurrencyRatesForLastNDays(ctx context.Context, days int) error {
	currencies := strings.Split(strings.ToLower(os.Getenv("CURRENCIES")), ",")

	for _, base := range currencies {
		otherCurrencies := make([]Currency, len(currencies)-1)
		nextIndex := 0

		for _, cur := range currencies {
			if cur != base {
				otherCurrencies[nextIndex] = Currency(cur)
				nextIndex++
			}
		}

		for i := 0; i <= days; i++ {
			date := time.Now().AddDate(0, 0, -i)

			result, err := c.source.Get(Currency(base), otherCurrencies, date)
			if err != nil {
				return fmt.Errorf("failed to get currency rates for %s for %s: %w", base, date.Format("2006-01-02"), err)
			}

			for _, rate := range result {
				_, err := c.repository.Create(ctx, date, Currency(base), rate.Currency, rate.Rate)
				if err != nil {
					return fmt.Errorf("failed to create currency rate for %s-%s for %s: %w", base, rate.Currency, date, err)
				}
			}
		}
	}

	return nil
}
