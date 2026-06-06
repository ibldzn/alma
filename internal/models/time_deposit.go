package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/ibldzn/alma/internal/adapters/utils"
	"github.com/ibldzn/alma/internal/constants"
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
	MarketingID       *string `db:"marketing_id" json:"marketing_id,omitempty"`
	CreatedAt         string  `db:"created_at" json:"created_at"`
	UpdatedAt         string  `db:"updated_at" json:"updated_at"`
}

type DwhTimeDeposit struct {
	// _RowHash          string    `db:"_row_hash" json:"-"`
	// SourceFile        *string   `db:"source_file" json:"source_file,omitempty"`
	// RowIndex          *int64    `db:"row_index" json:"row_index,omitempty"`
	AsOfDate          time.Time `db:"as_of_date" json:"as_of_date"`
	Date              string    `db:"date" json:"date"`
	BranchCode        string    `db:"branch_code" json:"branch_code"`
	ProductID         string    `db:"product_i_d" json:"product_id"`
	ProductName       string    `db:"product_name" json:"product_name"`
	AccountNo         string    `db:"account_no" json:"account_no"`
	CustomerName      string    `db:"customer_name" json:"customer_name"`
	CIFNo             string    `db:"c_i_f_no" json:"cif_no"`
	CertificateNo     string    `db:"certificate_no" json:"certificate_no"`
	InterestRate      string    `db:"interest_rate" json:"interest_rate"`
	StartDate         string    `db:"start_date" json:"start_date"`
	MaturityDate      string    `db:"maturity_date" json:"maturity_date"`
	Duration          string    `db:"duration" json:"duration"`
	AutomaticRollover string    `db:"automatic_rollover" json:"automatic_rollover"`
	CompoundInterest  string    `db:"compound_interest" json:"compound_interest"`
	Currency          string    `db:"currency" json:"currency"`
	Nominal           string    `db:"nominal" json:"nominal"`
	InterestAccrual   string    `db:"interest_accrual" json:"interest_accrual"`
	MarketingID       *string   `db:"marketing_i_d" json:"marketing_id,omitempty"`
	IngestedAt        time.Time `db:"ingested_at" json:"ingested_at"`
}

func (td *TimeDeposit) FromCSV(headers, record []string) error {
	return utils.FromCSV(headers, record, td)
}

func (dwh *DwhTimeDeposit) ToTimeDeposit() (TimeDeposit, error) {
	if dwh == nil {
		return TimeDeposit{}, fmt.Errorf("nil DwhTimeDeposit")
	}

	interestRate, err := utils.ParseFloatField(dwh.InterestRate)
	if err != nil {
		return TimeDeposit{}, err
	}

	nominal, err := utils.ParseFloatField(dwh.Nominal)
	if err != nil {
		return TimeDeposit{}, err
	}

	interestAccrual, err := utils.ParseFloatField(dwh.InterestAccrual)
	if err != nil {
		return TimeDeposit{}, err
	}

	marketingId := dwh.MarketingID
	if marketingId != nil {
		trimmed := strings.TrimSpace(*marketingId)
		marketingId = &trimmed
	}

	timestamp := dwh.IngestedAt.Format(constants.DateTimeFormat)

	return TimeDeposit{
		Date:              dwh.Date,
		BranchCode:        strings.TrimSpace(dwh.BranchCode),
		ProductID:         strings.TrimSpace(dwh.ProductID),
		ProductName:       strings.TrimSpace(dwh.ProductName),
		AccountNo:         strings.TrimSpace(dwh.AccountNo),
		CustomerName:      strings.TrimSpace(dwh.CustomerName),
		CIFNo:             strings.TrimSpace(dwh.CIFNo),
		CertificateNo:     strings.TrimSpace(dwh.CertificateNo),
		InterestRate:      interestRate,
		StartDate:         strings.TrimSpace(dwh.StartDate),
		MaturityDate:      strings.TrimSpace(dwh.MaturityDate),
		Term:              strings.TrimSpace(dwh.Duration),
		AutomaticRollover: strings.TrimSpace(dwh.AutomaticRollover),
		CompoundInterest:  strings.TrimSpace(dwh.CompoundInterest),
		Currency:          strings.TrimSpace(dwh.Currency),
		Nominal:           nominal,
		InterestAccrual:   interestAccrual,
		MarketingID:       marketingId,

		// Kalau DB lu handle created_at / updated_at otomatis,
		// sebaiknya field ini jangan ikut di INSERT.
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}, nil
}
