package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/ibldzn/alma/internal/constants"
	"github.com/ibldzn/alma/internal/interfaces"
	"github.com/ibldzn/alma/internal/models"
	"github.com/ibldzn/alma/internal/utils"
)

type TKSService struct {
	SupermanService interfaces.ISupermanService
}

func NewTKSService(supermanService interfaces.ISupermanService) *TKSService {
	return &TKSService{
		SupermanService: supermanService,
	}
}

func (s *TKSService) GetLDRHistory(ctx context.Context, startDate, endDate string) ([]models.LDRSummaryRow, error) {
	saldoNeracas, err := s.SupermanService.GetSaldoNeracas(ctx, startDate, endDate, ldrAccounts())
	if err != nil {
		return nil, fmt.Errorf("get LDR saldo neracas: %w", err)
	}

	return calculateLDRSummary(saldoNeracas), nil
}

func (s *TKSService) GetCashRatioHistory(ctx context.Context, startDate, endDate string) ([]models.CashRatioSummaryRow, error) {
	saldoNeracas, err := s.SupermanService.GetSaldoNeracas(ctx, startDate, endDate, cashRatioAccounts())
	if err != nil {
		return nil, fmt.Errorf("get cash ratio saldo neracas: %w", err)
	}

	return calculateCashRatioSummary(saldoNeracas), nil
}

func ldrAccounts() []string {
	accounts := make(
		[]string,
		0,
		len(constants.LDRBakiDebetAccounts)+
			len(constants.LDRFundingAccounts)+
			len(constants.LDRFundingExclusionAccounts),
	)
	accounts = append(accounts, constants.LDRBakiDebetAccounts...)
	accounts = append(accounts, constants.LDRFundingAccounts...)
	accounts = append(accounts, constants.LDRFundingExclusionAccounts...)

	return utils.Dedup(accounts)
}

func cashRatioAccounts() []string {
	accounts := make(
		[]string,
		0,
		len(constants.CashRatioAssetLiquidAccounts)+
			len(constants.CashRatioLiabilityShortTermAccounts),
	)
	accounts = append(accounts, constants.CashRatioAssetLiquidAccounts...)
	accounts = append(accounts, constants.CashRatioLiabilityShortTermAccounts...)

	return utils.Dedup(accounts)
}

func calculateLDRSummary(saldoNeracas []models.SaldoNeraca) []models.LDRSummaryRow {
	bakiDebetAccounts := toAccountSet(constants.LDRBakiDebetAccounts)
	fundingAccounts := toAccountSet(constants.LDRFundingAccounts)
	exclusionAccounts := toAccountSet(constants.LDRFundingExclusionAccounts)

	rowsByDate := make(map[string]*models.LDRSummaryRow)
	for _, saldoNeraca := range saldoNeracas {
		date := strings.TrimSpace(saldoNeraca.Date)
		if date == "" {
			continue
		}

		row, exists := rowsByDate[date]
		if !exists {
			row = &models.LDRSummaryRow{Date: date}
			rowsByDate[date] = row
		}

		account := strings.TrimSpace(saldoNeraca.NoAkun)
		balance := math.Abs(saldoNeraca.SaldoAkhir)
		switch {
		case bakiDebetAccounts[account]:
			row.BakiDebet += balance
		case fundingAccounts[account]:
			row.Simpanan += balance
		case exclusionAccounts[account]:
			row.Exclusions += balance
		}
	}

	rows := make([]models.LDRSummaryRow, 0, len(rowsByDate))
	for _, row := range rowsByDate {
		summary := *row
		summary.FundingBase = summary.Simpanan - summary.Exclusions
		if summary.FundingBase != 0 {
			summary.ConsolidatedLDR = summary.BakiDebet / summary.FundingBase * 100
		}

		rows = append(rows, summary)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Date < rows[j].Date
	})

	return rows
}

func calculateCashRatioSummary(saldoNeracas []models.SaldoNeraca) []models.CashRatioSummaryRow {
	assetLiquidAccounts := toAccountSet(constants.CashRatioAssetLiquidAccounts)
	liabilityShortTermAccounts := toAccountSet(constants.CashRatioLiabilityShortTermAccounts)

	rowsByDate := make(map[string]*models.CashRatioSummaryRow)
	for _, saldoNeraca := range saldoNeracas {
		date := strings.TrimSpace(saldoNeraca.Date)
		if date == "" {
			continue
		}

		row, exists := rowsByDate[date]
		if !exists {
			row = &models.CashRatioSummaryRow{Date: date}
			rowsByDate[date] = row
		}

		account := strings.TrimSpace(saldoNeraca.NoAkun)
		balance := math.Abs(saldoNeraca.SaldoAkhir)
		switch {
		case assetLiquidAccounts[account]:
			row.AssetLiquid += balance
		case liabilityShortTermAccounts[account]:
			row.LiabilityShortTerm += balance
		}
	}

	rows := make([]models.CashRatioSummaryRow, 0, len(rowsByDate))
	for _, row := range rowsByDate {
		summary := *row
		if summary.LiabilityShortTerm != 0 {
			summary.CashRatio = summary.AssetLiquid / summary.LiabilityShortTerm * 100
		}

		rows = append(rows, summary)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Date < rows[j].Date
	})

	return rows
}

func toAccountSet(accounts []string) map[string]bool {
	set := make(map[string]bool, len(accounts))
	for _, account := range accounts {
		account = strings.TrimSpace(account)
		if account != "" {
			set[account] = true
		}
	}

	return set
}
