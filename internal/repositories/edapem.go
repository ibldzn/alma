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
		SELECT
			? AS date,
			COUNT(*) AS total_customer
		FROM payroll_dapem_masters
		WHERE jenis_dapem = ?
			AND STR_TO_DATE(bulan_dapem, '%Y%m%d') BETWEEN ? AND ?;
	`
	err := r.EdapemDB.GetContext(ctx, &total, query, startDate, dapemType, startDate, endDate)
	return total, err
}
