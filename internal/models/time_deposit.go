package models

import (
	"github.com/ibldzn/alma/internal/adapters/utils"
)

type TimeDeposit struct {
	ID                int     `db:"id" json:"-"`
	Date              string  `db:"date" json:"date"`
	BranchCode        string  `db:"branch_code" json:"branch_code"`
	ProductID         string  `db:"product_id" json:"product_id"`
	ProductName       string  `db:"product_name" json:"product_name"`
	AccountNo         string  `db:"account_no" json:"account_no"`
	CustomerName      string  `db:"customer_name" json:"customer_name"`
	CIFNo             string  `db:"cif_no" json:"cif_no"`
	CertificateNo     string  `db:"certificate_no" json:"certificate_no"`
	InterestRate      float64 `db:"interest_rate" json:"interest_rate"`
	StartDate         string  `db:"start_date" json:"start_date"`
	MaturityDate      string  `db:"maturity_date" json:"maturity_date"`
	Term              string  `db:"term" json:"term"`
	AutomaticRollover string  `db:"automatic_rollover" json:"automatic_rollover"`
	CompoundInterest  string  `db:"compound_interest" json:"compound_interest"`
	Currency          string  `db:"currency" json:"currency"`
	Nominal           float64 `db:"nominal" json:"nominal"`
	InterestAccrual   float64 `db:"interest_accrual" json:"interest_accrual"`
	MarketingID       *string `db:"marketing_id" json:"marketing_id"`
	CreatedAt         string  `db:"created_at" json:"created_at"`
	UpdatedAt         string  `db:"updated_at" json:"updated_at"`
}

func (td *TimeDeposit) FromCSV(headers, record []string) error {
	return utils.FromCSV(headers, record, td)
}
