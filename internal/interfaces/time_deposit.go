package interfaces

import "github.com/ibldzn/alma/internal/models"

type ITimeDepositRepository interface {
	GetTimeDepositHistory(startDate, endDate string) ([]models.TimeDeposit, error)
	UpsertTimeDeposits(timeDeposits []models.TimeDeposit) error
}
