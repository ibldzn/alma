package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/ibldzn/alma/internal/adapters/utils"
)

type Saving struct {
	ID                    int        `db:"id" json:"-"`
	Date                  string     `db:"date" json:"date"`
	BranchCode            string     `db:"branch_code" json:"branch_code"`
	ProductID             string     `db:"product_id" json:"product_id"`
	ProductName           string     `db:"product_name" json:"product_name"`
	AccountNo             string     `db:"account_no" json:"account_no"`
	CustomerName          string     `db:"customer_name" json:"customer_name"`
	CIFNo                 string     `db:"cif_no" json:"cif_no"`
	AccAltNo              *string    `db:"acc_alt_no" json:"acc_alt_no,omitempty"`
	InterestRate          float64    `db:"interest_rate" json:"interest_rate"`
	AccountStatus         string     `db:"account_status" json:"account_status"`
	AccountRegisteredDate *time.Time `db:"account_registered_date" json:"account_registered_date,omitempty"`
	Currency              string     `db:"currency" json:"currency"`
	DebitBalance          float64    `db:"debit_balance" json:"debit_balance"`
	CreditBalance         float64    `db:"credit_balance" json:"credit_balance"`
	AccrueInterestDebit   float64    `db:"accrue_interest_debit" json:"accrue_interest_debit"`
	AccrueInterestCredit  float64    `db:"accrue_interest_credit" json:"accrue_interest_credit"`
	MarketingID           *string    `db:"marketing_id" json:"marketing_id,omitempty"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

type DwhSaving struct {
	// _RowHash          string     `db:"_row_hash" json:"-"`
	// SourceFile        *string    `db:"source_file" json:"source_file,omitempty"`
	// RowIndex          *int64     `db:"row_index" json:"row_index,omitempty"`
	AsOfDate              time.Time `db:"as_of_date" json:"as_of_date"`
	Date                  string    `db:"date" json:"date"`
	BranchCode            string    `db:"branch_code" json:"branch_code"`
	ProductID             string    `db:"product_i_d" json:"product_id"`
	ProductName           string    `db:"product_name" json:"product_name"`
	AccountNo             string    `db:"account_no" json:"account_no"`
	CustomerName          string    `db:"customer_name" json:"customer_name"`
	CIFNo                 string    `db:"c_i_f_no" json:"cif_no"`
	AccAltNo              *string   `db:"acc_alt_nor" json:"acc_alt_no,omitempty"`
	InterestRate          string    `db:"interest_rate" json:"interest_rate"`
	AccountStatus         string    `db:"account_status" json:"account_status"`
	AccountRegisteredDate *string   `db:"account_registered_date" json:"account_registered_date,omitempty"`
	Currency              string    `db:"currency" json:"currency"`
	DebitBalance          string    `db:"debit_balance" json:"debit_balance"`
	CreditBalance         string    `db:"credit_balance" json:"credit_balance"`
	AccrueInterestDebit   string    `db:"accrue_interest_debit" json:"accrue_interest_debit"`
	AccrueInterestCredit  string    `db:"accrue_interest_credit" json:"accrue_interest_credit"`
	MarketingID           *string   `db:"marketing_i_d" json:"marketing_id,omitempty"`
	IngestedAt            time.Time `db:"ingested_at" json:"ingested_at"`
}

func (s *Saving) FromCSV(headers, record []string) error {
	return utils.FromCSV(headers, record, s)
}

func (dwh *DwhSaving) ToSaving() (Saving, error) {
	if dwh == nil {
		return Saving{}, fmt.Errorf("nil DwhSavings")
	}

	interestRate, err := utils.ParseFloatField(dwh.InterestRate)
	if err != nil {
		return Saving{}, fmt.Errorf("parse interest_rate: %w", err)
	}

	debitBalance, err := utils.ParseFloatField(dwh.DebitBalance)
	if err != nil {
		return Saving{}, fmt.Errorf("parse debit_balance: %w", err)
	}

	creditBalance, err := utils.ParseFloatField(dwh.CreditBalance)
	if err != nil {
		return Saving{}, fmt.Errorf("parse credit_balance: %w", err)
	}

	accrueInterestDebit, err := utils.ParseFloatField(dwh.AccrueInterestDebit)
	if err != nil {
		return Saving{}, fmt.Errorf("parse accrue_interest_debit: %w", err)
	}

	accrueInterestCredit, err := utils.ParseFloatField(dwh.AccrueInterestCredit)
	if err != nil {
		return Saving{}, fmt.Errorf("parse accrue_interest_credit: %w", err)
	}

	regDate := dwh.AccountRegisteredDate
	accountRegisteredDate := time.Time{}
	if regDate != nil {
		accountRegisteredDate, err = utils.ParseOptionalTimeField(*regDate)
		if err != nil {
			return Saving{}, fmt.Errorf("parse account_registered_date: %w", err)
		}
	}

	return Saving{
		Date:                  dwh.Date,
		BranchCode:            strings.TrimSpace(dwh.BranchCode),
		ProductID:             strings.TrimSpace(dwh.ProductID),
		ProductName:           strings.TrimSpace(dwh.ProductName),
		AccountNo:             strings.TrimSpace(dwh.AccountNo),
		CustomerName:          strings.TrimSpace(dwh.CustomerName),
		CIFNo:                 strings.TrimSpace(dwh.CIFNo),
		AccAltNo:              utils.ParseOptionalStringPtrField(dwh.AccAltNo),
		InterestRate:          interestRate,
		AccountStatus:         strings.TrimSpace(dwh.AccountStatus),
		AccountRegisteredDate: &accountRegisteredDate,
		Currency:              strings.TrimSpace(dwh.Currency),
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
