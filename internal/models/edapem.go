package models

type EdapemSummaryRow struct {
	Date          string `db:"date" json:"date"`
	TotalCustomer int    `db:"total_customer" json:"total_customer"`
}
