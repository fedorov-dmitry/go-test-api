package internal_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fedorov-dmitry/go-test-api/internal"
	"github.com/fedorov-dmitry/go-test-api/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestCurrencySynchronizer_UpdateCurrencyRatesForToday_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	// mocks for source and storage (through repository)
	mockSource := mocks.NewMockCurrencyRateSource(ctrl)
	mockStorage := mocks.NewMockCurrencyStorage(ctrl)
	repo := internal.NewCurrencyRepository(mockStorage)

	currencies := []internal.Currency{
		internal.NewCurrency("usd"),
		internal.NewCurrency("eur"),
	}
	s := internal.NewCurrencySynchronizer(*repo, mockSource, currencies)

	// Expect 2 calls to source.Get (base=usd with [eur], base=eur with [usd])
	mockSource.
		EXPECT().
		Get(internal.NewCurrency("usd"), gomock.Any(), gomock.Any()).
		DoAndReturn(func(base internal.Currency, targets []internal.Currency, date time.Time) ([]internal.CurrencyRate, error) {
			if base != internal.NewCurrency("usd") {
				t.Fatalf("unexpected base: %s", base)
			}
			return []internal.CurrencyRate{
				{Base: base, Currency: internal.NewCurrency("eur"), Rate: 0.92, Date: date},
			}, nil
		})
	mockSource.
		EXPECT().
		Get(internal.NewCurrency("eur"), gomock.Any(), gomock.Any()).
		DoAndReturn(func(base internal.Currency, targets []internal.Currency, date time.Time) ([]internal.CurrencyRate, error) {
			if base != internal.NewCurrency("eur") {
				t.Fatalf("unexpected base: %s", base)
			}
			return []internal.CurrencyRate{
				{Base: base, Currency: internal.NewCurrency("usd"), Rate: 1.08, Date: date},
			}, nil
		})

	// Expect 2 Set calls via repository.Create
	mockStorage.
		EXPECT().
		Set(ctx, gomock.AssignableToTypeOf(internal.CurrencyRate{})).
		Times(2).
		Return(nil)

	if err := s.UpdateCurrencyRatesForTodayAndLastNDays(ctx, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCurrencySynchronizer_SourceError_Propagates(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockSource := mocks.NewMockCurrencyRateSource(ctrl)
	mockStorage := mocks.NewMockCurrencyStorage(ctrl)
	repo := internal.NewCurrencyRepository(mockStorage)

	currencies := []internal.Currency{
		internal.NewCurrency("usd"),
		internal.NewCurrency("eur"),
	}
	s := internal.NewCurrencySynchronizer(*repo, mockSource, currencies)

	mockSource.
		EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("source down"))

	err := s.UpdateCurrencyRatesForTodayAndLastNDays(ctx, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
