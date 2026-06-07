package models

type LDRSummaryRow struct {
	Date            string  `db:"date" json:"date"`
	BakiDebet       float64 `db:"baki_debet" json:"baki_debet"`
	Simpanan        float64 `db:"simpanan" json:"simpanan"`
	Exclusions      float64 `db:"exclusions" json:"exclusions"`
	FundingBase     float64 `db:"funding_base" json:"funding_base"`
	ConsolidatedLDR float64 `db:"consolidated_ldr" json:"consolidated_ldr"`
}
