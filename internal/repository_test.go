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

func TestCurrencyRepository_Get_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockCurrencyStorage(ctrl)
	repo := internal.NewCurrencyRepository(mockStorage)

	ctx := context.Background()
	base := internal.NewCurrency("usd")
	cur := internal.NewCurrency("eur")
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	expected := internal.CurrencyRate{
		Date:     date,
		Base:     base,
		Currency: cur,
		Rate:     0.92,
	}

	mockStorage.
		EXPECT().
		Get(ctx, base, cur, date).
		Return(expected, nil)

	got, err := repo.Get(ctx, base, cur, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("got %+v, want %+v", got, expected)
	}
}

func TestCurrencyRepository_Get_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockCurrencyStorage(ctrl)
	repo := internal.NewCurrencyRepository(mockStorage)

	ctx := context.Background()
	base := internal.NewCurrency("usd")
	cur := internal.NewCurrency("eur")
	date := time.Now()

	mockStorage.
		EXPECT().
		Get(ctx, base, cur, date).
		Return(internal.CurrencyRate{}, errors.New("db failure"))

	_, err := repo.Get(ctx, base, cur, date)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCurrencyRepository_GetMany_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockCurrencyStorage(ctrl)
	repo := internal.NewCurrencyRepository(mockStorage)

	ctx := context.Background()
	base := internal.NewCurrency("usd")
	date := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)
	expected := []internal.CurrencyRate{
		{Date: date, Base: base, Currency: internal.NewCurrency("eur"), Rate: 0.92},
		{Date: date, Base: base, Currency: internal.NewCurrency("jpy"), Rate: 145.1},
	}

	mockStorage.
		EXPECT().
		GetMany(ctx, base, date).
		Return(expected, nil)

	got, err := repo.GetMany(ctx, base, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(expected) {
		t.Fatalf("got len %d, want %d", len(got), len(expected))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("got[%d]=%+v, want %+v", i, got[i], expected[i])
		}
	}
}

func TestCurrencyRepository_Create_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockCurrencyStorage(ctrl)
	repo := internal.NewCurrencyRepository(mockStorage)

	ctx := context.Background()
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	base := internal.NewCurrency("usd")
	cur := internal.NewCurrency("eur")
	rate := 0.93

	mockStorage.
		EXPECT().
		Set(ctx, gomock.AssignableToTypeOf(internal.CurrencyRate{})).
		DoAndReturn(func(_ context.Context, r internal.CurrencyRate) error {
			if r.Date != date || r.Base != base || r.Currency != cur || r.Rate != rate {
				t.Fatalf("unexpected rate %+v", r)
			}
			return nil
		})

	got, err := repo.Create(ctx, date, base, cur, rate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Date != date || got.Base != base || got.Currency != cur || got.Rate != rate {
		t.Fatalf("got %+v", got)
	}
}

func TestCurrencyRepository_Create_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockCurrencyStorage(ctrl)
	repo := internal.NewCurrencyRepository(mockStorage)

	ctx := context.Background()
	date := time.Now()
	base := internal.NewCurrency("usd")
	cur := internal.NewCurrency("eur")

	mockStorage.
		EXPECT().
		Set(ctx, gomock.Any()).
		Return(errors.New("write failed"))

	_, err := repo.Create(ctx, date, base, cur, 0.9)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
