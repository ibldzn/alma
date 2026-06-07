package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ibldzn/alma/internal/adapters/fincloud"
	"github.com/ibldzn/alma/internal/constants"
	"github.com/ibldzn/alma/internal/interfaces"
	"github.com/ibldzn/alma/internal/models"
	"github.com/ibldzn/alma/internal/utils"
)

type FincloudService struct {
	Config             fincloud.Config
	Credentials        fincloud.Credentials
	TimeDepositService interfaces.ITimeDepositService
	SavingService      interfaces.ISavingService
}

func NewFincloudService(
	config fincloud.Config,
	credentials fincloud.Credentials,
	timeDepositService interfaces.ITimeDepositService,
	savingService interfaces.ISavingService,
) *FincloudService {
	return &FincloudService{
		Config:             config,
		Credentials:        credentials,
		TimeDepositService: timeDepositService,
		SavingService:      savingService,
	}
}

func (s *FincloudService) initialize(ctx context.Context) (*fincloud.Client, *fincloud.Session, error) {
	client, err := fincloud.NewClient(s.Config)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.Login(ctx, s.Credentials)
	if err != nil {
		return nil, nil, err
	}

	return client, session, nil
}

func (s *FincloudService) SyncTimeDeposits(ctx context.Context) error {
	client, session, err := s.initialize(ctx)
	if err != nil {
		return err
	}
	defer client.Logout(ctx, session.ID)

	ctx = fincloud.WithFincloudSessionID(ctx, session.ID)

	report, err := client.DownloadReport(ctx, "Time Deposit Account Balance Detail Today", "")
	if err != nil {
		return err
	}

	timeDeposits, err := prepareDataFromReport(report, func(headers, record []string) (models.TimeDeposit, error) {
		var td models.TimeDeposit
		err := td.FromCSV(headers, record)
		return td, err
	})
	if err != nil {
		return err
	}

	return s.TimeDepositService.UpsertTimeDeposits(ctx, timeDeposits)
}

func (s *FincloudService) SyncSavings(ctx context.Context) error {
	client, session, err := s.initialize(ctx)
	if err != nil {
		return err
	}
	defer client.Logout(ctx, session.ID)

	ctx = fincloud.WithFincloudSessionID(ctx, session.ID)

	report, err := client.DownloadReport(ctx, "Savings Balance Details Report Today", "")
	if err != nil {
		return err
	}

	savings, err := prepareDataFromReport(report, func(headers, record []string) (models.Saving, error) {
		var s models.Saving
		s.Date = utils.GetTodayInJakarta().Format(constants.DateFormat)
		err := s.FromCSV(headers, record)
		return s, err
	})
	if err != nil {
		return err
	}

	return s.SavingService.UpsertSavings(ctx, savings)
}

func (s *FincloudService) KickOffScheduleSync(ctx context.Context) error {
	if err := s.SyncTimeDeposits(ctx); err != nil {
		log.Printf("initial sync of time deposits failed: %v", err)
	}

	if err := s.SyncSavings(ctx); err != nil {
		log.Printf("initial sync of savings failed: %v", err)
	}

	ticker := time.NewTicker(3 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.SyncTimeDeposits(ctx); err != nil {
				log.Printf("scheduled sync of time deposits failed: %v", err)
			}
			if err := s.SyncSavings(ctx); err != nil {
				log.Printf("scheduled sync of savings failed: %v", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func prepareDataFromReport[T any](report string, parser func(headers, record []string) (T, error)) ([]T, error) {
	headers, records, err := utils.TransposeCSV([]byte(report), '|')
	if err != nil {
		return nil, err
	}

	results := make([]T, len(records))
	for i, record := range records {
		result, err := parser(headers, record)
		if err != nil {
			return nil, fmt.Errorf("parse report row %d: %w", i+1, err)
		}
		results[i] = result
	}

	return results, nil
}
