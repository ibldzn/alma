package services

import (
	"context"
	"fmt"

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
	_, _, err := utils.ValidateDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	normalizedAccounts := utils.Dedup(accounts)
	if len(normalizedAccounts) == 0 {
		return nil, fmt.Errorf("%w: accounts cannot be empty", types.ErrInvalidData)
	}

	rows, err := s.SupermanRepo.GetSaldoNeracas(ctx, startDate, endDate, normalizedAccounts)
	if err != nil {
		return nil, fmt.Errorf("get Superman saldo neracas: %w", err)
	}

	return rows, nil
}
