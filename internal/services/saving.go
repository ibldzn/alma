package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/ibldzn/alma/internal/interfaces"
	"github.com/ibldzn/alma/internal/models"
	"github.com/ibldzn/alma/internal/utils"
)

type SavingService struct {
	SavingRepo      interfaces.ISavingRepository
	SupermanService interfaces.ISupermanService
}

func NewSavingService(repo interfaces.ISavingRepository, supermanService interfaces.ISupermanService) *SavingService {
	return &SavingService{
		SavingRepo:      repo,
		SupermanService: supermanService,
	}
}

func (s *SavingService) GetSavingHistory(ctx context.Context, date string) ([]models.Saving, error) {
	return s.SavingRepo.GetSavingHistory(ctx, date)
}

func (s *SavingService) UpsertSavings(ctx context.Context, savings []models.Saving) error {
	return s.SavingRepo.UpsertSavings(ctx, savings)
}

func (s *SavingService) GetSavingSummary(ctx context.Context, startDate, endDate string) ([]models.SavingSummaryRow, error) {
	_, _, err := utils.ValidateDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	/*
		2212101	Savings Saving Account - main
		2212102	Savings Pension ASN
		2212103	Savings Pension Taspen
		2212104	Savings Simpel
		2212105	Savings Friend
		2212106	Savings SISETO PERORANGAN
		2212107	Savings IBADAH
		2212108	Savings QURBAN
		2212109	Savings TOUR
		2212110	Savings Tabungan Emas
		2212111	Savings ABP
		2212112	Savings TASIRA
		2212113	Savings Dpentas Vaganza
		2212114	Savings Mandiri Siswa
		2212115	Savings Tamastra Berjangka
		2212116	Savings SISETO ABP
	*/
	const (
		savingProductsCount = 16
		savingProductPrefix = "2212"
	)
	savingProducts := make([]string, 0, savingProductsCount)
	for i := 1; i <= savingProductsCount; i++ {
		savingProducts = append(savingProducts, fmt.Sprintf("%s%d", savingProductPrefix, i+100))
	}

	saldoNeracas, err := s.SupermanService.GetSaldoNeracas(ctx, startDate, endDate, savingProducts)
	if err != nil {
		return nil, err
	}

	results := make([]models.SavingSummaryRow, len(saldoNeracas))
	for i, saldoNeraca := range saldoNeracas {
		account := strings.TrimSpace(saldoNeraca.NoAkun)
		productID := account[len(savingProductPrefix):]
		results[i] = models.SavingSummaryRow{
			Date:      saldoNeraca.Date,
			ProductID: productID,
			Balance:   saldoNeraca.SaldoAkhir,
		}
	}

	return results, nil
}
