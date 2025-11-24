package jsdelivrnet

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/fedorov-dmitry/go-test-api/internal"
)

type CurrencyRateSource struct{}

func NewCurrencyRateSource() *CurrencyRateSource {
	return &CurrencyRateSource{}
}

func (s *CurrencyRateSource) Get(baseCurrency internal.Currency, currencies []internal.Currency, date time.Time) ([]internal.CurrencyRate, error) {
	url := fmt.Sprintf("%s@%s/v1/currencies/%s.json", os.Getenv("API_BASE_URL"), date.Format("2006-01-02"), baseCurrency)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get rates for %s from external API: %w", baseCurrency, err)
	}

	defer resp.Body.Close() // do I need to handle an error here?

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad response from rates API. status: %s. response: %s", resp.Status, string(body))
	}

	var result map[internal.Currency]interface{}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode rates API response: %w", err)
	}

	currencyRates := make([]internal.CurrencyRate, len(currencies))

	for i, currency := range currencies {
		rate := result[baseCurrency].(map[string]interface{})[string(currency)].(float64)

		currencyRate := internal.CurrencyRate{}

		currencyRate.Rate = rate
		currencyRate.Base = baseCurrency
		currencyRate.Currency = currency
		currencyRate.Date = date

		currencyRates[i] = currencyRate
	}

	return currencyRates, nil
}
