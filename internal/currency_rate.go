package internal

import (
	"context"
	"fmt"
	"time"
)

type CurrencyRate struct {
	Date     time.Time
	Base     string
	Currency string
	Rate     float64
}

type CurrencyStorage interface {
	Get(ctx context.Context, baseCurrency string, currency string, date time.Time) (CurrencyRate, error)
	GetMany(ctx context.Context, baseCurrency string, date time.Time) ([]CurrencyRate, error)
	Set(ctx context.Context, currency CurrencyRate) error
}

type CurrencyRepository struct {
	storage CurrencyStorage
}

func NewCurrencyRepository(storage CurrencyStorage) *CurrencyRepository {
	return &CurrencyRepository{storage: storage}
}

func (repo *CurrencyRepository) Get(ctx context.Context, baseCurrency string, currency string, date time.Time) (CurrencyRate, error) {
	rate, err := repo.storage.Get(ctx, baseCurrency, currency, date)

	if err != nil {
		return CurrencyRate{}, fmt.Errorf("failed to get currency rates for %v: %w", baseCurrency, err)
	}

	return rate, nil
}

func (repo *CurrencyRepository) GetMany(ctx context.Context, baseCurrency string, date time.Time) ([]CurrencyRate, error) {
	currencies, err := repo.storage.GetMany(ctx, baseCurrency, date)

	if err != nil {
		return nil, fmt.Errorf("failed to get currency rates for %v: %w", baseCurrency, err)
	}

	return currencies, nil
}

func (repo *CurrencyRepository) Create(ctx context.Context, date time.Time, baseCurrency string, currency string, rate float64) (CurrencyRate, error) {
	currencyRate := CurrencyRate{
		Date:     date,
		Base:     baseCurrency,
		Currency: currency,
		Rate:     rate,
	}

	err := repo.storage.Set(ctx, currencyRate)

	if err != nil {
		return CurrencyRate{}, fmt.Errorf("failed to save currency rate %s-%s: %w", baseCurrency, currency, err)
	}

	return currencyRate, nil
}
