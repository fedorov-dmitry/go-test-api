package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fedorov-dmitry/go-test-api/internal"
)

type CurrencyService struct {
	repository internal.CurrencyRepository
	baseUrl    string
	apiKey     string
	currencies []string
}

func NewCurrencyService(repository internal.CurrencyRepository, baseUrl string, apiKey string, currencies []string) *CurrencyService {
	return &CurrencyService{
		repository: repository,
		baseUrl:    baseUrl,
		apiKey:     apiKey,
		currencies: currencies,
	}
}

func (c *CurrencyService) SaveCurrencyRatesForLastNDays(ctx context.Context, days int) error {
	for _, base := range c.currencies {
		otherCurrencies := make([]string, len(c.currencies)-1)
		nextIndex := 0

		for _, cur := range c.currencies {
			if cur != base {
				otherCurrencies[nextIndex] = cur
				nextIndex++
			}
		}

		for i := 0; i <= days; i++ {
			date := time.Now().AddDate(0, 0, -i)

			result, err := c.getCurrencyRates(base, date)

			if err != nil {
				return fmt.Errorf("failed to get currency rates for %s for %s: %w", base, date.Format("2006-01-02"), err)
			}

			for _, cur := range otherCurrencies {
				rate := result[base].(map[string]interface{})[cur].(float64)
				_, err := c.repository.Create(ctx, date, base, cur, rate)

				if err != nil {
					// TODO
				}
			}
		}
	}

	return nil
}

func (c *CurrencyService) getCurrencyRates(baseCurrency string, date time.Time) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s@%s/v1/currencies/%s.json", c.baseUrl, date.Format("2006-01-02"), baseCurrency)
	resp, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("failed to get rates for %s from external API: %w", baseCurrency, err)
	}

	defer resp.Body.Close() // do I need to handle an error here?

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad response from rates API. status: %s. response: %s", resp.Status, string(body))
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("failed to decode rates API response: %w", err)
	}

	return result, nil
}
