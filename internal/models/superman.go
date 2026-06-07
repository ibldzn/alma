package models

type SaldoNeraca struct {
	Date       string  `db:"date" json:"date"`
	NoAkun     string  `db:"noakun" json:"noakun"`
	SaldoAkhir float64 `db:"saldoakhir" json:"saldoakhir"`
}
