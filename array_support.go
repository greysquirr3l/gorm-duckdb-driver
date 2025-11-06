package duckdb

import (
	"database/sql/driver"
	"fmt"

	"github.com/marcboeker/go-duckdb/v2"
)

// StringArray represents a DuckDB string array using native Composite type
type StringArray struct {
	duckdb.Composite[[]string]
}

// FloatArray represents a DuckDB float array using native Composite type  
type FloatArray struct {
	duckdb.Composite[[]float64]
}

// IntArray represents a DuckDB integer array using native Composite type
type IntArray struct {
	duckdb.Composite[[]int64]
}

// GormDataType returns the GORM data type for StringArray.
// GormDataType returns the GORM data type for StringArray.
func (StringArray) GormDataType() string {
	return "VARCHAR[]"
}

// GormDataType returns the GORM data type for IntArray.
func (IntArray) GormDataType() string {
	return "BIGINT[]"
}

// GormDataType returns the GORM data type for FloatArray.
func (FloatArray) GormDataType() string {
	return "DOUBLE[]"
}

// Value implementations for driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	values := a.Get()
	if values == nil {
		return []string{}, nil // Return empty slice instead of nil
	}
	return values, nil
}

// Value implements driver.Valuer interface for IntArray.
func (a IntArray) Value() (driver.Value, error) {
	values := a.Get()
	if values == nil {
		return []int64{}, nil // Return empty slice instead of nil
	}
	return values, nil
}

// Value implements driver.Valuer interface for FloatArray.
func (a FloatArray) Value() (driver.Value, error) {
	values := a.Get()
	if values == nil {
		return []float64{}, nil // Return empty slice instead of nil
	}
	return values, nil
}

// Scan implementations for sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	if err := a.Composite.Scan(value); err != nil {
		return fmt.Errorf("failed to scan string array: %w", err)
	}
	return nil
}

// Scan implements sql.Scanner interface for IntArray.
func (a *IntArray) Scan(value interface{}) error {
	if err := a.Composite.Scan(value); err != nil {
		return fmt.Errorf("failed to scan int array: %w", err)
	}
	return nil
}

// Scan implements sql.Scanner interface for FloatArray.
func (a *FloatArray) Scan(value interface{}) error {
	if err := a.Composite.Scan(value); err != nil {
		return fmt.Errorf("failed to scan float array: %w", err)
	}
	return nil
}

// NewStringArray creates a new StringArray from a slice of strings.
func NewStringArray(values []string) StringArray {
	var arr StringArray
	_ = arr.Scan(values)
	return arr
}

// NewIntArray creates a new IntArray from a slice of int64 values.
func NewIntArray(values []int64) IntArray {
	var arr IntArray
	_ = arr.Scan(values)
	return arr
}

// NewFloatArray creates a new FloatArray from a slice of float64 values.
func NewFloatArray(values []float64) FloatArray {
	var arr FloatArray
	_ = arr.Scan(values)
	return arr
}
