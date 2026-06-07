package repositories

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ibldzn/alma/internal/constants"
	"github.com/ibldzn/alma/internal/models"
	"github.com/ibldzn/alma/internal/types"
	"github.com/ibldzn/alma/internal/utils"
	"github.com/jmoiron/sqlx"
)

type SavingRepository struct {
	AppDB *sqlx.DB
	DwhDB *sqlx.DB
}

func NewSavingRepository(appDB, dwhDB *sqlx.DB) *SavingRepository {
	return &SavingRepository{
		AppDB: appDB,
		DwhDB: dwhDB,
	}
}

func (r *SavingRepository) GetSavingHistory(ctx context.Context, date string) ([]models.Saving, error) {
	d, err := time.Parse(constants.DateFormat, date)
	if err != nil {
		return nil, err
	}

	today := utils.GetTodayInJakarta()
	if d.Before(today) {
		dwhSavings, err := r.getDwhSavings(ctx, date)
		if err != nil {
			return nil, err
		}

		results := make([]models.Saving, len(dwhSavings))
		for i, s := range dwhSavings {
			results[i], err = s.ToSaving()
			if err != nil {
				return nil, types.ErrUnableToMapDwhToAppModel
			}
		}

		return results, nil
	}

	return r.getAppSavings(ctx, date)
}

func (r *SavingRepository) UpsertSavings(ctx context.Context, savings []models.Saving) error {
	if len(savings) == 0 {
		return nil
	}

	dbFields, err := utils.DBFields[models.Saving]()
	if err != nil {
		return err
	}

	dbFields = slices.DeleteFunc(dbFields, func(field string) bool {
		switch field {
		case "id", "created_at", "updated_at":
			return true
		default:
			return false
		}
	})

	query := fmt.Sprintf(`
		INSERT INTO %s (%s)
		VALUES (%s)
		ON DUPLICATE KEY UPDATE %s
	`,
		constants.SavingsTodayTable,
		strings.Join(dbFields, ", "),
		utils.GenerateNamedPlaceholders(dbFields),
		utils.GenerateUpdateSetClause(dbFields, "date", "account_no"),
	)

	tx, err := r.AppDB.Beginx()
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range savings {
		if _, err := stmt.ExecContext(ctx, &savings[i]); err != nil {
			return fmt.Errorf(
				"upsert saving row %d account_no=%s: %w",
				i,
				savings[i].AccountNo,
				err,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	committed = true

	return nil
}

func (r *SavingRepository) GetSavingSummary(ctx context.Context, startDate, endDate string) ([]models.SavingSummaryRow, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *SavingRepository) getDwhSavings(ctx context.Context, date string) ([]models.DwhSaving, error) {
	dbFields, err := utils.DBFields[models.DwhSaving]()
	if err != nil {
		return nil, err
	}

	dwhQuery := fmt.Sprintf(`
			SELECT %s
			FROM %s
			WHERE as_of_date = ?
			ORDER BY as_of_date
		`,
		strings.Join(dbFields, ", "),
		constants.SavingsHistoryTable,
	)
	var savings []models.DwhSaving
	if err = r.DwhDB.SelectContext(ctx, &savings, dwhQuery, date); err != nil {
		return nil, err
	}

	return savings, nil
}

func (r *SavingRepository) getAppSavings(ctx context.Context, date string) ([]models.Saving, error) {
	asOfDate, err := time.Parse(constants.DateFormat, date)
	if err != nil {
		return nil, err
	}

	dbFields, err := utils.DBFields[models.Saving]()
	if err != nil {
		return nil, err
	}

	appQuery := fmt.Sprintf(`
			SELECT %s
			FROM %s
			WHERE date = ?
		`,
		strings.Join(dbFields, ", "),
		constants.SavingsTodayTable,
	)
	var savings []models.Saving
	if err = r.AppDB.SelectContext(
		ctx,
		&savings,
		appQuery,
		asOfDate.Format(constants.DateFormat),
	); err != nil {
		return nil, err
	}

	return savings, nil
}
