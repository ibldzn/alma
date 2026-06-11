package repositories

import (
	"context"

	"github.com/ibldzn/alma/internal/models"
	"github.com/jmoiron/sqlx"
)

type EdapemRepository struct {
	EdapemDB *sqlx.DB
}

func NewEdapemRepository(edapemDB *sqlx.DB) *EdapemRepository {
	return &EdapemRepository{
		EdapemDB: edapemDB,
	}
}

func (r *EdapemRepository) GetTotalDapemByType(ctx context.Context, startDate, endDate, dapemType string) (models.EdapemSummaryRow, error) {
	var total models.EdapemSummaryRow
	query := `
		WITH x AS (
		SELECT
			STR_TO_DATE(bulan_dapem, '%Y%m%d') AS tgl_dapem,
			COUNT(*) AS total_pensiunan
			FROM payroll_dapem_masters
			WHERE jenis_dapem = ?
			GROUP BY tgl_dapem
		)
		SELECT *
		FROM x
		WHERE tgl_dapem BETWEEN ? AND ?;
	`
	err := r.EdapemDB.GetContext(ctx, &total, query, dapemType, startDate, endDate)
	return total, err
}
