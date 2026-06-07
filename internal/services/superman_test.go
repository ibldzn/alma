package services

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ibldzn/alma/internal/models"
	"github.com/ibldzn/alma/internal/types"
)

type fakeSupermanRepository struct {
	called    bool
	rows      []models.SaldoNeraca
	err       error
	startDate string
	endDate   string
	accounts  []string
}

func (f *fakeSupermanRepository) GetSaldoNeracas(ctx context.Context, startDate, endDate string, accounts []string) ([]models.SaldoNeraca, error) {
	f.called = true
	f.startDate = startDate
	f.endDate = endDate
	f.accounts = append([]string(nil), accounts...)

	if f.err != nil {
		return nil, f.err
	}

	return f.rows, nil
}

func TestSupermanServiceGetSaldoNeracasNormalizesAccounts(t *testing.T) {
	repo := &fakeSupermanRepository{
		rows: []models.SaldoNeraca{
			{Date: "2026-06-01", NoAkun: "121", SaldoAkhir: 100},
		},
	}
	service := NewSupermanService(repo)

	rows, err := service.GetSaldoNeracas(
		context.Background(),
		"2026-06-01",
		"2026-06-30",
		[]string{" 121", "121", "", " 122 "},
	)
	if err != nil {
		t.Fatalf("GetSaldoNeracas returned error: %v", err)
	}

	if !repo.called {
		t.Fatal("repository was not called")
	}

	expectedAccounts := []string{"121", "122"}
	if !reflect.DeepEqual(repo.accounts, expectedAccounts) {
		t.Fatalf("accounts = %v, want %v", repo.accounts, expectedAccounts)
	}

	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(rows))
	}
}

func TestSupermanServiceGetSaldoNeracasRejectsInvalidStartDate(t *testing.T) {
	repo := &fakeSupermanRepository{}
	service := NewSupermanService(repo)

	_, err := service.GetSaldoNeracas(context.Background(), "2026-99-01", "2026-06-30", []string{"121"})
	if !errors.Is(err, types.ErrInvalidDateFormat) {
		t.Fatalf("error = %v, want ErrInvalidDateFormat", err)
	}
	if repo.called {
		t.Fatal("repository called for invalid date")
	}
}

func TestSupermanServiceGetSaldoNeracasRejectsInvalidDateRange(t *testing.T) {
	repo := &fakeSupermanRepository{}
	service := NewSupermanService(repo)

	_, err := service.GetSaldoNeracas(context.Background(), "2026-06-30", "2026-06-01", []string{"121"})
	if !errors.Is(err, types.ErrInvalidDateRange) {
		t.Fatalf("error = %v, want ErrInvalidDateRange", err)
	}
	if repo.called {
		t.Fatal("repository called for invalid date range")
	}
}

func TestSupermanServiceGetSaldoNeracasRejectsEmptyAccounts(t *testing.T) {
	repo := &fakeSupermanRepository{}
	service := NewSupermanService(repo)

	_, err := service.GetSaldoNeracas(context.Background(), "2026-06-01", "2026-06-30", []string{" ", ""})
	if !errors.Is(err, types.ErrInvalidData) {
		t.Fatalf("error = %v, want ErrInvalidData", err)
	}
	if repo.called {
		t.Fatal("repository called for empty accounts")
	}
}
