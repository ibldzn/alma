package services

import (
	"context"

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

func (s *SavingService) GetSavingHistory(ctx context.Context, date string) ([]models.Saving, error) {
	return s.SavingRepo.GetSavingHistory(ctx, date)
}

func (s *SavingService) UpsertSavings(ctx context.Context, savings []models.Saving) error {
	return s.SavingRepo.UpsertSavings(ctx, savings)
}

func (s *SavingService) GetSavingSummary(ctx context.Context, startDate, endDate string) ([]models.SavingSummaryRow, error) {
	return s.SavingRepo.GetSavingSummary(ctx, startDate, endDate)
}
