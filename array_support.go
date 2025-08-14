package duckdb

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// Helper function to parse array string representation
func parseArrayString(s string) []string {
	s = strings.TrimSpace(s)

	// Handle empty array
	if s == "[]" || s == "" {
		return []string{}
	}

	// Remove brackets
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = s[1 : len(s)-1]
	}

	if strings.TrimSpace(s) == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}

	return result
}

// StringArray represents a DuckDB TEXT[] array type
type StringArray []string

// Value implements driver.Valuer interface for StringArray
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}

	if len(a) == 0 {
		return "[]", nil
	}

	elements := make([]string, 0, len(a))
	for _, s := range a {
		// Escape single quotes in strings
		escaped := strings.ReplaceAll(s, "'", "''")
		elements = append(elements, fmt.Sprintf("'%s'", escaped))
	}

	return "[" + strings.Join(elements, ", ") + "]", nil
}

// Scan implements sql.Scanner interface for StringArray
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		return a.scanFromString(v)
	case []byte:
		return a.scanFromString(string(v))
	case []interface{}:
		return a.scanFromSlice(v)
	default:
		// Try JSON unmarshaling as fallback
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("cannot scan %T into StringArray", value)
		}
		if err := json.Unmarshal(data, a); err != nil {
			return fmt.Errorf("failed to unmarshal JSON data into StringArray: %w", err)
		}
		return nil
	}
}

func (a *StringArray) scanFromString(s string) error {
	s = strings.TrimSpace(s)

	// Handle empty array
	if s == "[]" || s == "" {
		*a = StringArray{}
		return nil
	}

	// Remove brackets
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = s[1 : len(s)-1]
	}

	if strings.TrimSpace(s) == "" {
		*a = StringArray{}
		return nil
	}

	// Simple CSV parsing - this could be enhanced for complex cases
	parts := strings.Split(s, ",")
	result := make(StringArray, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Remove quotes if present
		if strings.HasPrefix(part, "'") && strings.HasSuffix(part, "'") {
			part = part[1 : len(part)-1]
			// Unescape single quotes
			part = strings.ReplaceAll(part, "''", "'")
		} else if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") {
			part = part[1 : len(part)-1]
			// Unescape double quotes
			part = strings.ReplaceAll(part, "\"\"", "\"")
		}
		result = append(result, part)
	}

	*a = result
	return nil
}

func (a *StringArray) scanFromSlice(slice []interface{}) error {
	result := make(StringArray, 0, len(slice))
	for _, item := range slice {
		result = append(result, fmt.Sprintf("%v", item))
	}
	*a = result
	return nil
}

// IntArray represents a DuckDB INTEGER[] array type
type IntArray []int64

// Value implements driver.Valuer interface for IntArray
func (a IntArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}

	if len(a) == 0 {
		return "[]", nil
	}

	elements := make([]string, 0, len(a))
	for _, i := range a {
		elements = append(elements, fmt.Sprintf("%d", i))
	}

	return "[" + strings.Join(elements, ", ") + "]", nil
}

// Scan implements sql.Scanner interface for IntArray
func (a *IntArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		return a.scanFromString(v)
	case []byte:
		return a.scanFromString(string(v))
	case []interface{}:
		return a.scanFromSlice(v)
	default:
		return fmt.Errorf("cannot scan %T into IntArray", value)
	}
}

func (a *IntArray) scanFromString(s string) error {
	parts := parseArrayString(s)

	if len(parts) == 0 {
		*a = IntArray{}
		return nil
	}

	result := make(IntArray, 0, len(parts))
	for _, part := range parts {
		var i int64
		if _, err := fmt.Sscanf(part, "%d", &i); err != nil {
			return fmt.Errorf("cannot parse '%s' as integer: %w", part, err)
		}
		result = append(result, i)
	}

	*a = result
	return nil
}

func (a *IntArray) scanFromSlice(slice []interface{}) error {
	result := make(IntArray, 0, len(slice))
	for _, item := range slice {
		switch v := item.(type) {
		case int64:
			result = append(result, v)
		case int:
			result = append(result, int64(v))
		case float64:
			result = append(result, int64(v))
		default:
			var i int64
			if _, err := fmt.Sscanf(fmt.Sprintf("%v", item), "%d", &i); err != nil {
				return fmt.Errorf("cannot convert %T to int64: %w", item, err)
			}
			result = append(result, i)
		}
	}
	*a = result
	return nil
}

// FloatArray represents a DuckDB DOUBLE[] array type
type FloatArray []float64

// Value implements driver.Valuer interface for FloatArray
func (a FloatArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}

	if len(a) == 0 {
		return "[]", nil
	}

	elements := make([]string, 0, len(a))
	for _, f := range a {
		elements = append(elements, fmt.Sprintf("%g", f))
	}

	return "[" + strings.Join(elements, ", ") + "]", nil
}

// Scan implements sql.Scanner interface for FloatArray
func (a *FloatArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		return a.scanFromString(v)
	case []byte:
		return a.scanFromString(string(v))
	case []interface{}:
		return a.scanFromSlice(v)
	default:
		return fmt.Errorf("cannot scan %T into FloatArray", value)
	}
}

func (a *FloatArray) scanFromString(s string) error {
	parts := parseArrayString(s)

	if len(parts) == 0 {
		*a = FloatArray{}
		return nil
	}

	result := make(FloatArray, 0, len(parts))
	for _, part := range parts {
		var f float64
		if _, err := fmt.Sscanf(part, "%g", &f); err != nil {
			return fmt.Errorf("cannot parse '%s' as float: %w", part, err)
		}
		result = append(result, f)
	}

	*a = result
	return nil
}

func (a *FloatArray) scanFromSlice(slice []interface{}) error {
	result := make(FloatArray, 0, len(slice))
	for _, item := range slice {
		switch v := item.(type) {
		case float64:
			result = append(result, v)
		case float32:
			result = append(result, float64(v))
		case int64:
			result = append(result, float64(v))
		case int:
			result = append(result, float64(v))
		default:
			var f float64
			if _, err := fmt.Sscanf(fmt.Sprintf("%v", item), "%g", &f); err != nil {
				return fmt.Errorf("cannot convert %T to float64: %w", item, err)
			}
			result = append(result, f)
		}
	}
	*a = result
	return nil
}

// GormDataType implements the GormDataTypeInterface for StringArray
func (StringArray) GormDataType() string {
	return "TEXT[]"
}

// GormDataType implements the GormDataTypeInterface for IntArray
func (IntArray) GormDataType() string {
	return "BIGINT[]"
}

// GormDataType implements the GormDataTypeInterface for FloatArray
func (FloatArray) GormDataType() string {
	return "DOUBLE[]"
}
