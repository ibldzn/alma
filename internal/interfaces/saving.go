package interfaces

import (
	"context"

	"github.com/ibldzn/alma/internal/models"
)

type ISavingRepository interface {
	GetSavingHistory(ctx context.Context, date string) ([]models.Saving, error)
	UpsertSavings(ctx context.Context, savings []models.Saving) error
	GetSavingSummary(ctx context.Context, startDate, endDate string) ([]models.SavingSummaryRow, error)
}

type ISavingService interface {
	GetSavingHistory(ctx context.Context, date string) ([]models.Saving, error)
	UpsertSavings(ctx context.Context, savings []models.Saving) error
	GetSavingSummary(ctx context.Context, startDate, endDate string) ([]models.SavingSummaryRow, error)
}
