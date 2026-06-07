package interfaces

import (
	"context"

	"github.com/ibldzn/alma/internal/models"
)

type ISupermanRepository interface {
	GetSaldoNeracas(ctx context.Context, startDate, endDate string, accounts []string) ([]models.SaldoNeraca, error)
}

type ISupermanService interface {
	GetSaldoNeracas(ctx context.Context, startDate, endDate string, accounts []string) ([]models.SaldoNeraca, error)
}
