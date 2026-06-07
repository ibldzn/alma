package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/ibldzn/alma/internal/interfaces"
	"github.com/ibldzn/alma/internal/models"
	"github.com/ibldzn/alma/internal/types"
	"github.com/ibldzn/alma/internal/utils"
)

type SupermanService struct {
	SupermanRepo interfaces.ISupermanRepository
}

func NewSupermanService(repo interfaces.ISupermanRepository) *SupermanService {
	return &SupermanService{
		SupermanRepo: repo,
	}
}

func (s *SupermanService) GetSaldoNeracas(ctx context.Context, startDate, endDate string, accounts []string) ([]models.SaldoNeraca, error) {
	if err := validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	normalizedAccounts := normalizeAccounts(accounts)
	if len(normalizedAccounts) == 0 {
		return nil, fmt.Errorf("%w: accounts cannot be empty", types.ErrInvalidData)
	}

	rows, err := s.SupermanRepo.GetSaldoNeracas(ctx, startDate, endDate, normalizedAccounts)
	if err != nil {
		return nil, fmt.Errorf("get Superman saldo neracas: %w", err)
	}

	return rows, nil
}

func validateDateRange(startDate, endDate string) error {
	start, err := utils.ParseDateInJakarta(startDate)
	if err != nil {
		return fmt.Errorf("%w: start_date=%q: %v", types.ErrInvalidDateFormat, startDate, err)
	}

	end, err := utils.ParseDateInJakarta(endDate)
	if err != nil {
		return fmt.Errorf("%w: end_date=%q: %v", types.ErrInvalidDateFormat, endDate, err)
	}

	if end.Before(start) {
		return types.ErrInvalidDateRange
	}

	return nil
}

func normalizeAccounts(accounts []string) []string {
	seen := make(map[string]struct{}, len(accounts))
	normalized := make([]string, 0, len(accounts))

	for _, account := range accounts {
		account = strings.TrimSpace(account)
		if account == "" {
			continue
		}
		if _, exists := seen[account]; exists {
			continue
		}

		seen[account] = struct{}{}
		normalized = append(normalized, account)
	}

	return normalized
}
