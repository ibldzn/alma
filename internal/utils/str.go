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
	// 1,234
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
		parts := strings.Split(s, ",")
		if isCommaThousands(parts) {
			s = strings.Join(parts, "")
		} else {
			// Indonesian-style decimal: 6,75
			s = strings.ReplaceAll(s, ",", ".")
		}
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: expected decimal number, got %q: %v", types.ErrInvalidData, raw, err)
	}

	return value, nil
}

func isCommaThousands(parts []string) bool {
	if len(parts) < 2 {
		return false
	}

	first := parts[0]
	if first == "" {
		return false
	}
	if first[0] == '-' || first[0] == '+' {
		first = first[1:]
	}
	if first == "" || len(first) > 3 || !allDigits(first) {
		return false
	}

	for _, part := range parts[1:] {
		if len(part) != 3 || !allDigits(part) {
			return false
		}
	}

	return true
}

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
