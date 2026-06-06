package services

import (
	"context"

	"github.com/ibldzn/alma/internal/constants"
	"github.com/ibldzn/alma/internal/interfaces"
	"github.com/ibldzn/alma/internal/models"
)

type SavingService struct {
	SavingRepo interfaces.ISavingRepository
}

func NewSavingService(repo interfaces.ISavingRepository) *SavingService {
	return &SavingService{
		SavingRepo: repo,
	}
}

func (s *SavingService) GetSavingHistory(ctx context.Context, startDate, endDate string) ([]models.Saving, error) {
	return s.SavingRepo.GetSavingHistory(ctx, startDate, endDate)
}

func (s *SavingService) UpsertSavings(ctx context.Context, savings []models.Saving) error {
	return s.SavingRepo.UpsertSavings(ctx, savings)
}

func (s *SavingService) GetSavingSummary(ctx context.Context, startDate, endDate string) (map[string]float64, error) {
	savings, err := s.SavingRepo.GetSavingHistory(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	summary := make(map[string]float64)
	for _, saving := range savings {
		if saving.ProductID == constants.TabInternalProductID {
			continue
		}

		product, exists := summary[saving.ProductID]
		if !exists {
			summary[saving.ProductID] = saving.CreditBalance
		} else {
			summary[saving.ProductID] = product + saving.CreditBalance
		}
	}

	return summary, nil
}
