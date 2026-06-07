package utils

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func FromCSV[T any](headers, record []string, target *T) error {
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	if len(record) == 0 {
		return fmt.Errorf("empty record")
	}

	if len(headers) != len(record) {
		return fmt.Errorf(
			"header and record length mismatch: headers=%d, record=%d",
			len(headers),
			len(record),
		)
	}

	// Buat lookup column_name => value agar tidak perlu nested loop.
	values := make(map[string]string, len(headers))

	for i, header := range headers {
		header = strings.TrimSpace(header)
		if header == "" {
			continue
		}

		values[header] = strings.TrimSpace(record[i])
	}

	targetValue := reflect.ValueOf(target).Elem()
	targetType := targetValue.Type()

	for i := 0; i < targetValue.NumField(); i++ {
		structField := targetType.Field(i)

		dbTag := structField.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		rawValue, exists := values[dbTag]
		if !exists {
			// Kolom seperti id, created_at, atau updated_at
			// boleh tidak tersedia dalam CSV.
			continue
		}

		if err := setField(targetValue.Field(i), rawValue); err != nil {
			return fmt.Errorf("failed parsing column %q: %w", dbTag, err)
		}
	}

	return nil
}

func TransposeCSV(data []byte, separator rune) (headers []string, records [][]string, err error) {
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = separator
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	rows, err := r.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("parsing CSV data: %w", err)
	}

	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("CSV data is empty")
	}

	headers = rows[0]
	records = rows[1:]

	return headers, records, nil
}

func setField(field reflect.Value, rawValue string) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	// Handle pointer seperti *string.
	if field.Kind() == reflect.Pointer {
		if rawValue == "" {
			field.SetZero()
			return nil
		}

		value := reflect.New(field.Type().Elem())

		if err := setField(value.Elem(), rawValue); err != nil {
			return err
		}

		field.Set(value)

		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(rawValue)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rawValue == "" {
			field.SetInt(0)
			return nil
		}

		value, err := strconv.ParseInt(rawValue, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("expected integer, got %q: %w", rawValue, err)
		}

		field.SetInt(value)

	case reflect.Float32, reflect.Float64:
		value, err := ParseFloatField(rawValue)
		if err != nil {
			return fmt.Errorf("expected decimal number, got %q: %w", rawValue, err)
		}

		field.SetFloat(value)

	case reflect.Bool:
		if rawValue == "" {
			field.SetBool(false)
			return nil
		}

		value, err := parseCSVBool(rawValue)
		if err != nil {
			return err
		}

		field.SetBool(value)

	default:
		return fmt.Errorf("unsupported field type %s", field.Type())
	}

	return nil
}

func parseCSVBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "ya":
		return true, nil

	case "0", "false", "no", "n", "tidak":
		return false, nil

	default:
		return false, fmt.Errorf("expected boolean, got %q", value)
	}
}
