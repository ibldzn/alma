package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/ibldzn/alma/internal/adapters/utils"
)

type Saving struct {
	ID                    int       `db:"id" json:"-"`
	Date                  string    `db:"date" json:"date"`
	BranchCode            string    `db:"branch_code" json:"branch_code"`
	ProductID             string    `db:"product_id" json:"product_id"`
	ProductName           string    `db:"product_name" json:"product_name"`
	AccountNo             string    `db:"account_no" json:"account_no"`
	CustomerName          string    `db:"customer_name" json:"customer_name"`
	CIFNo                 string    `db:"cif_no" json:"cif_no"`
	AccAltNo              *string   `db:"acc_alt_no" json:"acc_alt_no,omitempty"`
	InterestRate          float64   `db:"interest_rate" json:"interest_rate"`
	AccountStatus         string    `db:"account_status" json:"account_status"`
	AccountRegisteredDate *string   `db:"account_registered_date" json:"account_registered_date,omitempty"`
	Currency              string    `db:"currency" json:"currency"`
	DebitBalance          float64   `db:"debit_balance" json:"debit_balance"`
	CreditBalance         float64   `db:"credit_balance" json:"credit_balance"`
	AccrueInterestDebit   float64   `db:"accrue_interest_debit" json:"accrue_interest_debit"`
	AccrueInterestCredit  float64   `db:"accrue_interest_credit" json:"accrue_interest_credit"`
	MarketingID           *string   `db:"marketing_id" json:"marketing_id,omitempty"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time `db:"updated_at" json:"updated_at"`
}

type DwhSaving struct {
	// RowHash    string     `db:"_row_hash" json:"-"`
	// SourceFile *string    `db:"source_file" json:"source_file,omitempty"`
	// RowIndex   *int64     `db:"row_index" json:"row_index,omitempty"`
	AsOfDate              *time.Time `db:"as_of_date" json:"as_of_date,omitempty"`
	BranchCode            *string    `db:"branch" json:"branch_code,omitempty"`
	ProductID             *string    `db:"product_i_d" json:"product_id,omitempty"`
	AccountNo             *string    `db:"account_no" json:"account_no,omitempty"`
	CustomerName          *string    `db:"customer_name" json:"customer_name,omitempty"`
	CIFNo                 *string    `db:"c_i_f_no" json:"cif_no,omitempty"`
	AccAltNo              *string    `db:"acc_alt_nor" json:"acc_alt_no,omitempty"`
	InterestRate          *string    `db:"interest_rate" json:"interest_rate,omitempty"`
	AccountStatus         *string    `db:"account_status" json:"account_status,omitempty"`
	AccountRegisteredDate *string    `db:"account_registered_date" json:"account_registered_date,omitempty"`
	Currency              *string    `db:"currency" json:"currency,omitempty"`
	DebitBalance          *string    `db:"debit_balance" json:"debit_balance,omitempty"`
	CreditBalance         *string    `db:"credit_balance" json:"credit_balance,omitempty"`
	AccrueInterestDebit   *string    `db:"accrue_interest_debit" json:"accrue_interest_debit,omitempty"`
	AccrueInterestCredit  *string    `db:"accrue_interest_credit" json:"accrue_interest_credit,omitempty"`
	MarketingID           *string    `db:"marketing_i_d" json:"marketing_id,omitempty"`

	IngestedAt time.Time `db:"ingested_at" json:"ingested_at"`
}

type SavingSummaryRow struct {
	Date               string  `json:"date"`
	ProductID          string  `json:"product_id"`
	TotalCreditBalance float64 `json:"total_credit_balance"`
}

func (s *Saving) FromCSV(headers, record []string) error {
	return utils.FromCSV(headers, record, s)
}

func (dwh *DwhSaving) ToSaving() (Saving, error) {
	if dwh == nil {
		return Saving{}, fmt.Errorf("nil DwhSavings")
	}

	optionalString := func(field *string) string {
		if field == nil {
			return ""
		}
		return *field
	}

	interestRate, err := utils.ParseFloatField(optionalString(dwh.InterestRate))
	if err != nil {
		return Saving{}, fmt.Errorf("parse interest_rate: %w", err)
	}

	debitBalance, err := utils.ParseFloatField(optionalString(dwh.DebitBalance))
	if err != nil {
		return Saving{}, fmt.Errorf("parse debit_balance: %w", err)
	}

	creditBalance, err := utils.ParseFloatField(optionalString(dwh.CreditBalance))
	if err != nil {
		return Saving{}, fmt.Errorf("parse credit_balance: %w", err)
	}

	accrueInterestDebit, err := utils.ParseFloatField(optionalString(dwh.AccrueInterestDebit))
	if err != nil {
		return Saving{}, fmt.Errorf("parse accrue_interest_debit: %w", err)
	}

	accrueInterestCredit, err := utils.ParseFloatField(optionalString(dwh.AccrueInterestCredit))
	if err != nil {
		return Saving{}, fmt.Errorf("parse accrue_interest_credit: %w", err)
	}

	return Saving{
		BranchCode:            strings.TrimSpace(optionalString(dwh.BranchCode)),
		ProductID:             strings.TrimSpace(optionalString(dwh.ProductID)),
		ProductName:           strings.TrimSpace(optionalString(dwh.ProductID)), // DWH tidak punya field product_name, jadi pakai product_id saja
		AccountNo:             strings.TrimSpace(optionalString(dwh.AccountNo)),
		CustomerName:          strings.TrimSpace(optionalString(dwh.CustomerName)),
		CIFNo:                 strings.TrimSpace(optionalString(dwh.CIFNo)),
		AccAltNo:              utils.ParseOptionalStringPtrField(dwh.AccAltNo),
		InterestRate:          interestRate,
		AccountStatus:         strings.TrimSpace(optionalString(dwh.AccountStatus)),
		AccountRegisteredDate: utils.ParseOptionalStringPtrField(dwh.AccountRegisteredDate),
		Currency:              strings.TrimSpace(optionalString(dwh.Currency)),
		DebitBalance:          debitBalance,
		CreditBalance:         creditBalance,
		AccrueInterestDebit:   accrueInterestDebit,
		AccrueInterestCredit:  accrueInterestCredit,
		MarketingID:           utils.ParseOptionalStringPtrField(dwh.MarketingID),

		// Kalau DB handle created_at / updated_at otomatis,
		// field ini tetap harus di-exclude dari INSERT/UPSERT.
		CreatedAt: dwh.IngestedAt,
		UpdatedAt: dwh.IngestedAt,
	}, nil
}
