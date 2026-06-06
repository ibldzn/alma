package models

import (
	"github.com/ibldzn/alma/internal/adapters/utils"
)

type TimeDepositToday struct {
	ID                int     `db:"id"`
	Date              string  `db:"date"`
	BranchCode        string  `db:"branch_code"`
	ProductID         string  `db:"product_id"`
	ProductName       string  `db:"product_name"`
	AccountNo         string  `db:"account_no"`
	CustomerName      string  `db:"customer_name"`
	CIFNo             string  `db:"cif_no"`
	CertificateNo     string  `db:"certificate_no"`
	InterestRate      float64 `db:"interest_rate"`
	StartDate         string  `db:"start_date"`
	MaturityDate      string  `db:"maturity_date"`
	Term              string  `db:"term"`
	AutomaticRollover string  `db:"automatic_rollover"`
	CompoundInterest  string  `db:"compound_interest"`
	Currency          string  `db:"currency"`
	Nominal           float64 `db:"nominal"`
	InterestAccrual   float64 `db:"interest_accrual"`
	MarketingID       *string `db:"marketing_id"`
	CreatedAt         string  `db:"created_at"`
	UpdatedAt         string  `db:"updated_at"`
}

func (td *TimeDepositToday) FromCSV(headers, record []string) error {
	return utils.FromCSV(headers, record, td)
}
