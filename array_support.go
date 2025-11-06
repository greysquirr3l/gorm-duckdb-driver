package duckdb

import (
	"database/sql/driver"

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

// GORM DataType implementations
func (StringArray) GormDataType() string {
	return "VARCHAR[]"
}

func (IntArray) GormDataType() string {
	return "BIGINT[]"
}

func (FloatArray) GormDataType() string {
	return "DOUBLE[]"
}

// Value implementations for driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	values := a.Get()
	if values == nil {
		return nil, nil
	}
	return values, nil
}

func (a IntArray) Value() (driver.Value, error) {
	values := a.Get()
	if values == nil {
		return nil, nil
	}
	return values, nil
}

func (a FloatArray) Value() (driver.Value, error) {
	values := a.Get()
	if values == nil {
		return nil, nil
	}
	return values, nil
}

// Scan implementations for sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	return a.Composite.Scan(value)
}

func (a *IntArray) Scan(value interface{}) error {
	return a.Composite.Scan(value)
}

func (a *FloatArray) Scan(value interface{}) error {
	return a.Composite.Scan(value)
}

// Convenience constructors for backward compatibility
func NewStringArray(values []string) StringArray {
	var arr StringArray
	_ = arr.Scan(values)
	return arr
}

func NewIntArray(values []int64) IntArray {
	var arr IntArray
	_ = arr.Scan(values)
	return arr
}

func NewFloatArray(values []float64) FloatArray {
	var arr FloatArray
	_ = arr.Scan(values)
	return arr
}
