package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/ibldzn/alma/internal/interfaces"
	"github.com/ibldzn/alma/internal/models"
	"github.com/ibldzn/alma/internal/utils"
)

type TimeDepositService struct {
	TimeDepositRepo interfaces.ITimeDepositRepository
	SupermanService interfaces.ISupermanService
}

func NewTimeDepositService(repo interfaces.ITimeDepositRepository, supermanService interfaces.ISupermanService) *TimeDepositService {
	return &TimeDepositService{
		TimeDepositRepo: repo,
		SupermanService: supermanService,
	}
}

func (s *TimeDepositService) GetTimeDepositHistory(ctx context.Context, startDate, endDate string) ([]models.TimeDeposit, error) {
	return s.TimeDepositRepo.GetTimeDepositHistory(ctx, startDate, endDate)
}

func (s *TimeDepositService) UpsertTimeDeposits(ctx context.Context, timeDeposits []models.TimeDeposit) error {
	return s.TimeDepositRepo.UpsertTimeDeposits(ctx, timeDeposits)
}

func (s *TimeDepositService) GetTimeDepositSummary(ctx context.Context, startDate, endDate string) ([]models.TimeDepositSummaryRow, error) {
	_, _, err := utils.ValidateDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	/*
		2312200	Time Deposit
		2312201	Time Deposit Compound
		2312202	Time Deposit ABP
		2312203	Time Deposit ABP Compound
	*/
	const (
		timeDepositProductsCount = 4
		timeDepositProductPrefix = "2312"
	)
	timeDepositProducts := make([]string, 0, timeDepositProductsCount)
	for i := range timeDepositProductsCount {
		timeDepositProducts = append(timeDepositProducts, fmt.Sprintf("%s%d", timeDepositProductPrefix, i+200))
	}

	saldoNeracas, err := s.SupermanService.GetSaldoNeracas(ctx, startDate, endDate, timeDepositProducts)
	if err != nil {
		return nil, err
	}

	results := make([]models.TimeDepositSummaryRow, len(saldoNeracas))
	for i, saldoNeraca := range saldoNeracas {
		account := strings.TrimSpace(saldoNeraca.NoAkun)
		productID := account[len(timeDepositProductPrefix):]
		results[i] = models.TimeDepositSummaryRow{
			Date:      saldoNeraca.Date,
			ProductID: productID,
			Balance:   saldoNeraca.SaldoAkhir,
		}
	}

	return results, nil
}
