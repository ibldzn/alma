package interfaces

import (
	"context"

	"github.com/ibldzn/alma/internal/models"
)

type ITKSService interface {
	GetLDRHistory(ctx context.Context, startDate, endDate string) ([]models.LDRSummaryRow, error)
	GetCashRatioHistory(ctx context.Context, startDate, endDate string) ([]models.CashRatioSummaryRow, error)
}
