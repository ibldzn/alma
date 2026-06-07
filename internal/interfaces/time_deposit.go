package interfaces

import (
	"context"

	"github.com/ibldzn/alma/internal/models"
)

type ITimeDepositRepository interface {
	GetTimeDepositHistory(ctx context.Context, startDate, endDate string) ([]models.TimeDeposit, error)
	UpsertTimeDeposits(ctx context.Context, timeDeposits []models.TimeDeposit) error
}

type ITimeDepositService interface {
	GetTimeDepositHistory(ctx context.Context, startDate, endDate string) ([]models.TimeDeposit, error)
	UpsertTimeDeposits(ctx context.Context, timeDeposits []models.TimeDeposit) error
	GetTimeDepositSummary(ctx context.Context, startDate, endDate string) ([]models.TimeDepositSummaryRow, error)
}
