package utils

import "testing"

type csvFloatRow struct {
	Amount float64 `db:"amount"`
}

func TestFromCSVUsesParseFloatField(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want float64
	}{
		{name: "comma decimal", raw: "6,75", want: 6.75},
		{name: "comma thousands", raw: "1,234", want: 1234},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var row csvFloatRow
			if err := FromCSV([]string{"amount"}, []string{tt.raw}, &row); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if row.Amount != tt.want {
				t.Fatalf("got %v, want %v", row.Amount, tt.want)
			}
		})
	}
}
