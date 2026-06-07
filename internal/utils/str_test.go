package utils

import (
	"math"
	"testing"
)

func TestParseFloatField(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    float64
		wantErr bool
	}{
		{name: "english decimal with thousands", raw: "1,234.56", want: 1234.56},
		{name: "indonesian decimal with thousands", raw: "1.234,56", want: 1234.56},
		{name: "comma decimal", raw: "1234,56", want: 1234.56},
		{name: "comma thousands", raw: "1,234", want: 1234},
		{name: "blank", raw: "", want: 0},
		{name: "invalid", raw: "x", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFloatField(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(got-tt.want) > 1e-9 {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
