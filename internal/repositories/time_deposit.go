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

type TimeDepositRepository struct {
	AppDB *sqlx.DB
	DwhDB *sqlx.DB
}

func NewTimeDepositRepository(appDB, dwhDB *sqlx.DB) *TimeDepositRepository {
	return &TimeDepositRepository{
		AppDB: appDB,
		DwhDB: dwhDB,
	}
}

func (r *TimeDepositRepository) GetTimeDepositHistory(ctx context.Context, startDate, endDate string) ([]models.TimeDeposit, error) {
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

	dbFields, err := utils.DBFields[models.DwhTimeDeposit]()
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
		constants.TimeDepositHistoryTable,
	)
	var timeDeposits []models.DwhTimeDeposit
	if err = r.DwhDB.SelectContext(ctx, &timeDeposits, dwhQuery, startDate, endDate); err != nil {
		return nil, err
	}

	results := make([]models.TimeDeposit, len(timeDeposits))
	for i, td := range timeDeposits {
		results[i], err = td.ToTimeDeposit()
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
			constants.TimeDepositTodayTable,
		)
		var tdToday []models.TimeDeposit
		if err = r.AppDB.SelectContext(
			ctx,
			&tdToday,
			appQuery,
			today.Format(constants.DateFormat),
		); err != nil {
			return nil, err
		}
		results = append(results, tdToday...)
	}

	return results, nil
}

func (r *TimeDepositRepository) UpsertTimeDeposits(ctx context.Context, timeDeposits []models.TimeDeposit) error {
	if len(timeDeposits) == 0 {
		return nil
	}

	dbFields, err := utils.DBFields[models.TimeDeposit]()
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
		constants.TimeDepositTodayTable,
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

	for i := range timeDeposits {
		if _, err := stmt.ExecContext(ctx, &timeDeposits[i]); err != nil {
			return fmt.Errorf(
				"upsert time deposit row %d account_no=%s: %w",
				i,
				timeDeposits[i].AccountNo,
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
