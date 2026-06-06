package repositories

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ibldzn/alma/internal/adapters/utils"
	"github.com/ibldzn/alma/internal/constants"
	"github.com/ibldzn/alma/internal/models"
	"github.com/ibldzn/alma/internal/types"
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

func (r *SavingRepository) GetSavingHistory(ctx context.Context, startDate, endDate string) ([]models.Saving, error) {
	start, err := time.Parse(constants.DateFormat, startDate)
	if err != nil {
		return nil, err
	}

	end, err := time.Parse(constants.DateFormat, endDate)
	if err != nil {
		return nil, err
	}

	if end.Before(start) {
		return nil, types.ErrInvalidDateRange
	}

	dbFields, err := utils.DBFields[models.DwhSaving]()
	if err != nil {
		return nil, err
	}

	dwhQuery := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE as_of_date BETWEEN ? AND ?
		ORDER BY as_of_date
	`,
		strings.Join(dbFields, ", "),
		constants.SavingsHistoryTable,
	)
	var savings []models.DwhSaving
	if err = r.DwhDB.SelectContext(ctx, &savings, dwhQuery, startDate, endDate); err != nil {
		return nil, err
	}

	results := make([]models.Saving, len(savings))
	for i, td := range savings {
		results[i], err = td.ToSaving()
		if err != nil {
			return nil, types.ErrUnableToMapDwhToAppModel
		}
	}

	today := time.Now().In(time.FixedZone(constants.AsiaJakarta, 7*60*60))
	if utils.IsDateEqual(today, end) {
		appQuery := fmt.Sprintf(`
			SELECT %s
			FROM %s
			WHERE date = ?
		`,
			strings.Join(dbFields, ", "),
			constants.SavingsTodayTable,
		)
		var savingsToday []models.Saving
		if err = r.AppDB.SelectContext(
			ctx,
			&savingsToday,
			appQuery,
			today.Format(constants.DateFormat),
		); err != nil {
			return nil, err
		}
		results = append(results, savingsToday...)
	}

	return results, nil
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
