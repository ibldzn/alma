package services

import (
	"context"

	"github.com/ibldzn/alma/internal/interfaces"
	"github.com/ibldzn/alma/internal/models"
)

type TimeDepositService struct {
	TimeDepositRepo interfaces.ITimeDepositRepository
}

func (s *TimeDepositService) GetTimeDepositHistory(ctx context.Context, startDate, endDate string) ([]models.TimeDeposit, error) {
	return s.TimeDepositRepo.GetTimeDepositHistory(ctx, startDate, endDate)
}

func (s *TimeDepositService) UpsertTimeDeposits(ctx context.Context, timeDeposits []models.TimeDeposit) error {
	return s.TimeDepositRepo.UpsertTimeDeposits(ctx, timeDeposits)
}

func (s *TimeDepositService) GetTimeDepositSummary(ctx context.Context, startDate, endDate string) (map[string]float64, error) {
	timeDeposits, err := s.TimeDepositRepo.GetTimeDepositHistory(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	summary := make(map[string]float64)
	for _, td := range timeDeposits {
		product, exists := summary[td.ProductID]
		if !exists {
			summary[td.ProductID] = td.Nominal
		} else {
			summary[td.ProductID] = product + td.Nominal
		}
	}

	return summary, nil
}
