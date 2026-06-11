package interfaces

import (
	"context"

	"github.com/ibldzn/alma/internal/models"
)

type IEdapemRepository interface {
	GetTotalDapemByType(ctx context.Context, startDate, endDate, dapemType string) (models.EdapemSummaryRow, error)
}

type IEdapemService interface {
	GetTotalDapemByType(ctx context.Context, startDate, endDate, dapemType string) (models.EdapemSummaryRow, error)
}
