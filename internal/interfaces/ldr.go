package interfaces

import (
	"context"

	"github.com/ibldzn/alma/internal/models"
)

type ILDRService interface {
	GetLDRHistory(ctx context.Context, startDate, endDate string) ([]models.LDRSummaryRow, error)
}
