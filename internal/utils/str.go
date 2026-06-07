package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ibldzn/alma/internal/types"
)

func ParseFloatField(raw string) (float64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, nil
	}

	s = strings.TrimSuffix(s, "%")
	s = strings.ReplaceAll(s, " ", "")

	// Support format:
	// 1234.56
	// 1,234.56
	// 1234,56
	// 1.234,56
	if strings.Contains(s, ",") && strings.Contains(s, ".") {
		lastComma := strings.LastIndex(s, ",")
		lastDot := strings.LastIndex(s, ".")

		if lastComma > lastDot {
			// Indonesian-style decimal: 1.234,56
			s = strings.ReplaceAll(s, ".", "")
			s = strings.ReplaceAll(s, ",", ".")
		} else {
			// English-style decimal: 1,234.56
			s = strings.ReplaceAll(s, ",", "")
		}
	} else if strings.Contains(s, ",") {
		// Assume comma is decimal separator: 6,75 => 6.75
		s = strings.ReplaceAll(s, ",", ".")
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: expected decimal number, got %q: %v", types.ErrInvalidData, raw, err)
	}

	return value, nil
}
