package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ibldzn/alma/internal/types"
)

// DBFields reflects on a struct and returns the values of fields with `db` tags,
// or a map[string]interface{} and returns the keys.
func DBFields[T any]() ([]string, error) {
	t := reflect.TypeFor[T]()
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: found %s", types.ErrExpectedStructOrMap, t.Kind().String())
	}

	fields := make([]string, 0, t.NumField())

	for field := range t.Fields() {
		tag := field.Tag.Get("db")
		if tag == "" || tag == "-" {
			continue
		}

		name, _, _ := strings.Cut(tag, ",")
		if name == "" || name == "-" {
			continue
		}

		fields = append(fields, name)
	}

	return fields, nil
}

func GenerateUpdateSetClause(fields []string, excludedFields ...string) string {
	excluded := make(map[string]bool, len(excludedFields))
	for _, field := range excludedFields {
		excluded[field] = true
	}

	sets := make([]string, 0, len(fields))
	for _, field := range fields {
		if excluded[field] {
			continue
		}
		sets = append(sets, fmt.Sprintf("%s = VALUES(%s)", field, field))
	}

	return strings.Join(sets, ", ")
}

func GenerateNamedPlaceholders(fields []string) string {
	placeholders := make([]string, len(fields))

	for i, field := range fields {
		placeholders[i] = ":" + field
	}

	return strings.Join(placeholders, ", ")
}

func ParseOptionalStringPtrField(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	return &trimmed
}
