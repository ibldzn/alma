package services

import (
	"context"
	"math"
	"reflect"
	"testing"

	"github.com/ibldzn/alma/internal/models"
)

type fakeSupermanService struct {
	rows      []models.SaldoNeraca
	err       error
	startDate string
	endDate   string
	accounts  []string
}

func (f *fakeSupermanService) GetSaldoNeracas(ctx context.Context, startDate, endDate string, accounts []string) ([]models.SaldoNeraca, error) {
	f.startDate = startDate
	f.endDate = endDate
	f.accounts = append([]string(nil), accounts...)

	if f.err != nil {
		return nil, f.err
	}

	return f.rows, nil
}

func TestLDRServiceGetLDRHistoryCalculatesFormula(t *testing.T) {
	fake := &fakeSupermanService{
		rows: []models.SaldoNeraca{
			{Date: "2026-06-01", NoAkun: "121", SaldoAkhir: 40},
			{Date: "2026-06-01", NoAkun: "122", SaldoAkhir: 60},
			{Date: "2026-06-01", NoAkun: "221", SaldoAkhir: 50},
			{Date: "2026-06-01", NoAkun: "2312200", SaldoAkhir: 30},
			{Date: "2026-06-01", NoAkun: "2212111", SaldoAkhir: 20},
		},
	}
	service := NewLDRService(fake)

	rows, err := service.GetLDRHistory(context.Background(), "2026-06-01", "2026-06-01")
	if err != nil {
		t.Fatalf("GetLDRHistory returned error: %v", err)
	}

	expectedAccounts := []string{"121", "122", "221", "2312200", "2312201", "2212111", "2212116", "2212199"}
	if !reflect.DeepEqual(fake.accounts, expectedAccounts) {
		t.Fatalf("accounts = %v, want %v", fake.accounts, expectedAccounts)
	}

	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(rows))
	}

	row := rows[0]
	if row.BakiDebet != 100 {
		t.Fatalf("BakiDebet = %v, want 100", row.BakiDebet)
	}
	if row.Simpanan != 80 {
		t.Fatalf("Simpanan = %v, want 80", row.Simpanan)
	}
	if row.Exclusions != 20 {
		t.Fatalf("Exclusions = %v, want 20", row.Exclusions)
	}
	if row.FundingBase != 60 {
		t.Fatalf("FundingBase = %v, want 60", row.FundingBase)
	}
	if math.Abs(row.ConsolidatedLDR-166.6667) > 0.0001 {
		t.Fatalf("ConsolidatedLDR = %v, want about 166.6667", row.ConsolidatedLDR)
	}
}

func TestLDRServiceGetLDRHistoryReturnsZeroWhenFundingBaseZero(t *testing.T) {
	fake := &fakeSupermanService{
		rows: []models.SaldoNeraca{
			{Date: "2026-06-01", NoAkun: "121", SaldoAkhir: 100},
			{Date: "2026-06-01", NoAkun: "221", SaldoAkhir: 20},
			{Date: "2026-06-01", NoAkun: "2212111", SaldoAkhir: 20},
		},
	}
	service := NewLDRService(fake)

	rows, err := service.GetLDRHistory(context.Background(), "2026-06-01", "2026-06-01")
	if err != nil {
		t.Fatalf("GetLDRHistory returned error: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(rows))
	}

	row := rows[0]
	if row.FundingBase != 0 {
		t.Fatalf("FundingBase = %v, want 0", row.FundingBase)
	}
	if row.ConsolidatedLDR != 0 {
		t.Fatalf("ConsolidatedLDR = %v, want 0", row.ConsolidatedLDR)
	}
}

func TestLDRServiceGetLDRHistoryNormalizesSaldoSigns(t *testing.T) {
	fake := &fakeSupermanService{
		rows: []models.SaldoNeraca{
			{Date: "2026-06-01", NoAkun: "121", SaldoAkhir: -40},
			{Date: "2026-06-01", NoAkun: "122", SaldoAkhir: -60},
			{Date: "2026-06-01", NoAkun: "221", SaldoAkhir: -80},
			{Date: "2026-06-01", NoAkun: "2212111", SaldoAkhir: -20},
		},
	}
	service := NewLDRService(fake)

	rows, err := service.GetLDRHistory(context.Background(), "2026-06-01", "2026-06-01")
	if err != nil {
		t.Fatalf("GetLDRHistory returned error: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(rows))
	}

	row := rows[0]
	if row.BakiDebet != 100 {
		t.Fatalf("BakiDebet = %v, want 100", row.BakiDebet)
	}
	if row.Simpanan != 80 {
		t.Fatalf("Simpanan = %v, want 80", row.Simpanan)
	}
	if row.Exclusions != 20 {
		t.Fatalf("Exclusions = %v, want 20", row.Exclusions)
	}
	if row.FundingBase != 60 {
		t.Fatalf("FundingBase = %v, want 60", row.FundingBase)
	}
	if math.Abs(row.ConsolidatedLDR-166.6667) > 0.0001 {
		t.Fatalf("ConsolidatedLDR = %v, want about 166.6667", row.ConsolidatedLDR)
	}
}
