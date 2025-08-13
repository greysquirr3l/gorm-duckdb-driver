package duckdb

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// ErrorTranslator implements gorm.ErrorTranslator for DuckDB
type ErrorTranslator struct{}

// Translate converts DuckDB errors to GORM errors
func (et ErrorTranslator) Translate(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	errStrLower := strings.ToLower(errStr)

	// Handle DuckDB specific errors
	switch {
	case strings.Contains(errStrLower, "unique constraint"):
		return gorm.ErrDuplicatedKey
	case strings.Contains(errStrLower, "foreign key constraint"):
		return gorm.ErrForeignKeyViolated
	case strings.Contains(errStrLower, "check constraint"):
		return gorm.ErrCheckConstraintViolated
	case strings.Contains(errStrLower, "not null constraint"):
		return gorm.ErrInvalidValue
	case strings.Contains(errStrLower, "no such table"):
		return gorm.ErrRecordNotFound
	case strings.Contains(errStrLower, "no such column"):
		return gorm.ErrInvalidField
	case strings.Contains(errStrLower, "syntax error"):
		return gorm.ErrInvalidData
	case strings.Contains(errStrLower, "connection"):
		return gorm.ErrInvalidDB
	case strings.Contains(errStrLower, "database is locked"):
		return gorm.ErrInvalidDB
	}

	// Check for specific DuckDB error patterns
	if strings.Contains(errStrLower, "constraint") {
		return gorm.ErrInvalidValue
	}

	if strings.Contains(errStrLower, "invalid") || strings.Contains(errStrLower, "malformed") {
		return gorm.ErrInvalidData
	}

	// Default to the original error if no specific translation is found
	return err
}

// Common DuckDB error patterns
var (
	ErrUniqueConstraint  = errors.New("UNIQUE constraint failed")
	ErrForeignKey        = errors.New("FOREIGN KEY constraint failed")
	ErrCheckConstraint   = errors.New("CHECK constraint failed")
	ErrNotNullConstraint = errors.New("NOT NULL constraint failed")
	ErrNoSuchTable       = errors.New("no such table")
	ErrNoSuchColumn      = errors.New("no such column")
	ErrSyntaxError       = errors.New("syntax error")
	ErrDatabaseLocked    = errors.New("database is locked")
)

// IsSpecificError checks if an error matches a specific DuckDB error type
func IsSpecificError(err error, target error) bool {
	if err == nil || target == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	targetStr := strings.ToLower(target.Error())

	return strings.Contains(errStr, targetStr)
}

// IsDuplicateKeyError checks if the error is a duplicate key constraint violation
func IsDuplicateKeyError(err error) bool {
	return IsSpecificError(err, ErrUniqueConstraint)
}

// IsForeignKeyError checks if the error is a foreign key constraint violation
func IsForeignKeyError(err error) bool {
	return IsSpecificError(err, ErrForeignKey)
}

// IsNotNullError checks if the error is a not null constraint violation
func IsNotNullError(err error) bool {
	return IsSpecificError(err, ErrNotNullConstraint)
}

// IsTableNotFoundError checks if the error is a table not found error
func IsTableNotFoundError(err error) bool {
	return IsSpecificError(err, ErrNoSuchTable)
}

// IsColumnNotFoundError checks if the error is a column not found error
func IsColumnNotFoundError(err error) bool {
	return IsSpecificError(err, ErrNoSuchColumn)
}
