package postgres

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// parseVariadic returns the first value from `vals` or the default value `def` if `vals` is empty.
func parseVariadic[T any](def T, vals ...T) T {
	if len(vals) > 0 {
		return vals[0]
	}
	return def
}

// isStruct checks if the type of T is a struct.
func isStruct[T any](_ ...T) bool {
	var v T
	val := reflect.Indirect(reflect.ValueOf(v))
	return val.Kind() == reflect.Struct
}

// compile replaces @placeholder in SQL query and converts '?' to numbered placeholders ($1, $2, ...).
func compile(query string, replacements ...string) string {
	return normalizePlaceholder(strings.NewReplacer(replacements...).Replace(query))
}

// structColumns extracts column names from the `db` struct tag, skipping unexported fields.
func structColumns(v any, only, exclude []string) []string {
	val := reflect.Indirect(reflect.ValueOf(v))
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	columns := make([]string, 0, typ.NumField())

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields.
		if !field.IsExported() {
			continue
		}

		// Extract valid 'db' tags (ignoring "-" and empty values).
		if tag, ok := field.Tag.Lookup("db"); ok && !skipped(tag, only, exclude) {
			columns = append(columns, quoteField(tag))
		}
	}
	return columns
}

// structValues extracts field values from a struct based on the `db` tag, excluding unexported fields.
func structValues(v any, only, exclude []string) []any {
	val := reflect.Indirect(reflect.ValueOf(v))
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	values := make([]any, 0, typ.NumField())

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields.
		if !field.IsExported() {
			continue
		}

		// Add values for valid 'db' tags (not "-" or empty).
		if tag, ok := field.Tag.Lookup("db"); ok && !skipped(tag, only, exclude) {
			values = append(values, val.Field(i).Interface())
		}
	}
	return values
}

// skipped checks if a struct field's `db` tag should be skipped based on `only` and `exclude` lists.
func skipped(tag string, only, exclude []string) bool {
	if tag == "-" || tag == "" ||
		(len(only) > 0 && !slices.Contains(only, tag)) ||
		(len(exclude) > 0 && slices.Contains(exclude, tag)) {
		return true
	}

	return false
}

// quoteField wraps a column name in double quotes for SQL compatibility.
func quoteField(field string) string {
	return fmt.Sprintf(`"%s"`, field)
}

// normalizePlaceholder converts '?' placeholders in SQL to PostgreSQL-style numbered parameters ($1, $2, ...).
func normalizePlaceholder(query string) string {
	var builder strings.Builder
	builder.Grow(len(query) + 10)
	counter := 0

	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			counter++
			builder.WriteString(fmt.Sprintf("$%d", counter))
		} else {
			builder.WriteByte(query[i])
		}
	}
	return builder.String()
}
