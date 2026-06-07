package repositories

import (
	"context"
	"fmt"

	"github.com/ibldzn/alma/internal/constants"
	"github.com/ibldzn/alma/internal/models"
	"github.com/jmoiron/sqlx"
)

type SupermanRepository struct {
	SupermanDB *sqlx.DB
}

func NewSupermanRepository(supermanDB *sqlx.DB) *SupermanRepository {
	return &SupermanRepository{
		SupermanDB: supermanDB,
	}
}

func (r *SupermanRepository) GetSaldoNeracas(ctx context.Context, startDate, endDate string, accounts []string) ([]models.SaldoNeraca, error) {
	if len(accounts) == 0 {
		return []models.SaldoNeraca{}, nil
	}

	query := fmt.Sprintf(`
		SELECT
			DATE_FORMAT(tanggal, '%%Y-%%m-%%d') AS date,
			TRIM(noakun) AS noakun,
			saldoakhir AS saldoakhir
		FROM %s
		WHERE tanggal BETWEEN ? AND ?
			AND TRIM(noakun) IN (?)
		ORDER BY tanggal, noakun
	`,
		constants.SupermanSaldoNeracaTable,
	)

	query, args, err := sqlx.In(query, startDate, endDate, accounts)
	if err != nil {
		return nil, fmt.Errorf("build Superman saldo neraca query: %w", err)
	}

	query = r.SupermanDB.Rebind(query)

	var rows []models.SaldoNeraca
	if err := r.SupermanDB.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf(
			"select Superman saldo neracas start_date=%s end_date=%s accounts=%d: %w",
			startDate,
			endDate,
			len(accounts),
			err,
		)
	}

	return rows, nil
}
