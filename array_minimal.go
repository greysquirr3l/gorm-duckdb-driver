// Package duckdb provides a GORM driver for DuckDB database.
// This file contains minimal array support for basic DuckDB array operations.
package duckdb

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
)

// formatSliceForDuckDB converts a Go slice to DuckDB array literal syntax
func formatSliceForDuckDB(value interface{}) (string, error) {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice {
		return "", fmt.Errorf("expected slice, got %T", value)
	}

	if v.Len() == 0 {
		return "[]", nil
	}

	var elements []string
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		switch elem.Kind() {
		case reflect.Float32, reflect.Float64:
			elements = append(elements, fmt.Sprintf("%g", elem.Float()))
		case reflect.String:
			// Escape single quotes in strings
			str := strings.ReplaceAll(elem.String(), "'", "''")
			elements = append(elements, fmt.Sprintf("'%s'", str))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			elements = append(elements, fmt.Sprintf("%d", elem.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			elements = append(elements, fmt.Sprintf("%d", elem.Uint()))
		case reflect.Bool:
			if elem.Bool() {
				elements = append(elements, "true")
			} else {
				elements = append(elements, "false")
			}
		default:
			return "", fmt.Errorf("unsupported slice element type: %v", elem.Kind())
		}
	}

	return "[" + strings.Join(elements, ", ") + "]", nil
}

// ArrayLiteral wraps a Go slice to be formatted as a DuckDB array literal
type ArrayLiteral struct {
	Data interface{}
}

// Value implements driver.Valuer for DuckDB array literals
func (al ArrayLiteral) Value() (driver.Value, error) {
	if al.Data == nil {
		return "[]", nil
	}

	return formatSliceForDuckDB(al.Data)
}

// SimpleArrayScanner provides basic array scanning functionality
type SimpleArrayScanner struct {
	Target interface{} // Pointer to slice
}

// Scan implements sql.Scanner for basic array types
//
//nolint:gocyclo // Complex parsing logic is acceptable for array utility
func (sas *SimpleArrayScanner) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	// Handle Go slice types directly (DuckDB returns []interface{})
	if slice, ok := value.([]interface{}); ok {
		targetValue := reflect.ValueOf(sas.Target)
		if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Slice {
			return fmt.Errorf("target must be pointer to slice")
		}

		sliceType := targetValue.Elem().Type()
		elemType := sliceType.Elem()
		result := reflect.MakeSlice(sliceType, len(slice), len(slice))

		for i, elem := range slice {
			elemValue := result.Index(i)

			switch elemType.Kind() {
			case reflect.Float64:
				// Handle both float32 and float64 from DuckDB
				switch f := elem.(type) {
				case float64:
					elemValue.SetFloat(f)
				case float32:
					elemValue.SetFloat(float64(f))
				default:
					return fmt.Errorf("expected float32/float64, got %T at index %d", elem, i)
				}
			case reflect.String:
				if s, ok := elem.(string); ok {
					elemValue.SetString(s)
				} else {
					return fmt.Errorf("expected string, got %T at index %d", elem, i)
				}
			case reflect.Int64:
				// Handle various integer types from DuckDB
				switch i := elem.(type) {
				case int64:
					elemValue.SetInt(i)
				case int32:
					elemValue.SetInt(int64(i))
				case int:
					elemValue.SetInt(int64(i))
				default:
					return fmt.Errorf("expected integer type, got %T at index %d", elem, i)
				}
			case reflect.Bool:
				if b, ok := elem.(bool); ok {
					elemValue.SetBool(b)
				} else {
					return fmt.Errorf("expected bool, got %T at index %d", elem, i)
				}
			default:
				return fmt.Errorf("unsupported target element type: %v", elemType.Kind())
			}
		}

		targetValue.Elem().Set(result)
		return nil
	}

	// Fallback: Handle string representations of arrays
	var arrayStr string
	switch v := value.(type) {
	case string:
		arrayStr = v
	case []byte:
		arrayStr = string(v)
	default:
		return fmt.Errorf("cannot scan %T into SimpleArrayScanner", value)
	}

	// Parse DuckDB array format: [1.0, 2.0, 3.0] or [item1, item2, item3]
	arrayStr = strings.TrimSpace(arrayStr)
	if !strings.HasPrefix(arrayStr, "[") || !strings.HasSuffix(arrayStr, "]") {
		return fmt.Errorf("invalid array format: %s", arrayStr)
	}

	// Remove brackets
	content := arrayStr[1 : len(arrayStr)-1]
	content = strings.TrimSpace(content)

	if content == "" {
		// Empty array
		targetValue := reflect.ValueOf(sas.Target)
		if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Slice {
			return fmt.Errorf("target must be pointer to slice")
		}
		targetValue.Elem().Set(reflect.MakeSlice(targetValue.Elem().Type(), 0, 0))
		return nil
	}

	// Split elements and parse based on target type
	elements := strings.Split(content, ",")
	targetValue := reflect.ValueOf(sas.Target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("target must be pointer to slice")
	}

	sliceType := targetValue.Elem().Type()
	elemType := sliceType.Elem()
	result := reflect.MakeSlice(sliceType, len(elements), len(elements))

	for i, elemStr := range elements {
		elemStr = strings.TrimSpace(elemStr)
		elemValue := result.Index(i)

		switch elemType.Kind() {
		case reflect.Float64:
			var f float64
			if _, err := fmt.Sscanf(elemStr, "%f", &f); err != nil {
				return fmt.Errorf("failed to parse float: %s", elemStr)
			}
			elemValue.SetFloat(f)
		case reflect.String:
			// Remove quotes if present
			if strings.HasPrefix(elemStr, "'") && strings.HasSuffix(elemStr, "'") {
				elemStr = elemStr[1 : len(elemStr)-1]
				elemStr = strings.ReplaceAll(elemStr, "''", "'") // Unescape quotes
			}
			elemValue.SetString(elemStr)
		case reflect.Int64:
			var i int64
			if _, err := fmt.Sscanf(elemStr, "%d", &i); err != nil {
				return fmt.Errorf("failed to parse int: %s", elemStr)
			}
			elemValue.SetInt(i)
		case reflect.Bool:
			var b bool
			if _, err := fmt.Sscanf(elemStr, "%t", &b); err != nil {
				return fmt.Errorf("failed to parse bool: %s", elemStr)
			}
			elemValue.SetBool(b)
		default:
			return fmt.Errorf("unsupported target element type: %v", elemType.Kind())
		}
	}

	targetValue.Elem().Set(result)
	return nil
}
