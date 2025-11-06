package duckdb

import (
	"os"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Test CREATE with a model that doesn't have auto-increment
func TestCreateWithoutAutoIncrement(t *testing.T) {
	t.Log("=== Create Without Auto-Increment Test ===")

	// Enable debug mode
	if err := os.Setenv("GORM_DUCKDB_DEBUG", "1"); err != nil {
		t.Fatalf("Failed to set debug environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("GORM_DUCKDB_DEBUG"); err != nil {
			t.Logf("Failed to unset debug environment variable: %v", err)
		}
	}()

	dialector := Dialector{
		Config: &Config{
			DSN: ":memory:",
		},
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatalf("Failed to open DuckDB: %v", err)
	}

	// Add debug callbacks
	if err := db.Callback().Create().Before("gorm:create").Register("debug:before_create_simple", func(db *gorm.DB) {
		t.Logf("[DEBUG] Before gorm:create - SQL: '%s', Clauses: %+v", db.Statement.SQL.String(), db.Statement.Clauses)
	}); err != nil {
		t.Logf("Failed to register debug callback: %v", err)
	}

	if err := db.Callback().Create().After("gorm:create").Register("debug:after_create_simple", func(db *gorm.DB) {
		t.Logf("[DEBUG] After gorm:create - SQL: '%s', Clauses: %+v", db.Statement.SQL.String(), db.Statement.Clauses)
	}); err != nil {
		t.Logf("Failed to register debug callback: %v", err)
	}

	// Simple model without auto-increment
	type SimpleModel struct {
		ID   int    `gorm:"primaryKey"`  // Manual primary key, no auto-increment
		Name string
	}

	// Migration
	err = db.AutoMigrate(&SimpleModel{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}
	t.Log("Migration completed successfully")

	// Test create with manual ID
	model := SimpleModel{ID: 1, Name: "Manual ID Test"}
	result := db.Create(&model)
	t.Logf("Create result: Error=%v, RowsAffected=%d", result.Error, result.RowsAffected)
}

// Test CREATE with string primary key
func TestCreateWithStringPK(t *testing.T) {
	t.Log("=== Create With String Primary Key Test ===")

	// Enable debug mode
	if err := os.Setenv("GORM_DUCKDB_DEBUG", "1"); err != nil {
		t.Fatalf("Failed to set debug environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("GORM_DUCKDB_DEBUG"); err != nil {
			t.Logf("Failed to unset debug environment variable: %v", err)
		}
	}()

	dialector := Dialector{
		Config: &Config{
			DSN: ":memory:",
		},
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatalf("Failed to open DuckDB: %v", err)
	}

	// Add debug callbacks
	//nolint:errcheck,gosec // Debug callbacks - errors not critical for test functionality
	db.Callback().Create().Before("gorm:create").Register("debug:before_create_string", func(db *gorm.DB) {
		t.Logf("[DEBUG] Before gorm:create - SQL: '%s', Clauses: %+v", db.Statement.SQL.String(), db.Statement.Clauses)
	})

	//nolint:errcheck,gosec // Debug callbacks - errors not critical for test functionality
	db.Callback().Create().After("gorm:create").Register("debug:after_create_string", func(db *gorm.DB) {
		t.Logf("[DEBUG] After gorm:create - SQL: '%s', Clauses: %+v", db.Statement.SQL.String(), db.Statement.Clauses)
	})

	// Model with string primary key (no auto-increment)
	type Product struct {
		Code string `gorm:"primaryKey;size:10"`  // String primary key
		Name string `gorm:"size:100"`
	}

	// Migration
	err = db.AutoMigrate(&Product{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}
	t.Log("Migration completed successfully")

	// Test create with string primary key
	product := Product{Code: "PROD001", Name: "Test Product"}
	result := db.Create(&product)
	t.Logf("Create result: Error=%v, RowsAffected=%d", result.Error, result.RowsAffected)
}

// Test CREATE with no primary key at all
func TestCreateNoPK(t *testing.T) {
	t.Log("=== Create With No Primary Key Test ===")

	// Enable debug mode
	if err := os.Setenv("GORM_DUCKDB_DEBUG", "1"); err != nil {
		t.Fatalf("Failed to set debug environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("GORM_DUCKDB_DEBUG"); err != nil {
			t.Logf("Failed to unset debug environment variable: %v", err)
		}
	}()

	dialector := Dialector{
		Config: &Config{
			DSN: ":memory:",
		},
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatalf("Failed to open DuckDB: %v", err)
	}

	// Add debug callbacks
	//nolint:errcheck,gosec // Debug callbacks - errors not critical for test functionality
	db.Callback().Create().Before("gorm:create").Register("debug:before_create_nopk", func(db *gorm.DB) {
		t.Logf("[DEBUG] Before gorm:create - SQL: '%s', Clauses: %+v", db.Statement.SQL.String(), db.Statement.Clauses)
	})

	//nolint:errcheck,gosec // Debug callbacks - errors not critical for test functionality
	db.Callback().Create().After("gorm:create").Register("debug:after_create_nopk", func(db *gorm.DB) {
		t.Logf("[DEBUG] After gorm:create - SQL: '%s', Clauses: %+v", db.Statement.SQL.String(), db.Statement.Clauses)
	})

	// Simple model with no primary key
	type LogEntry struct {
		Message   string `gorm:"size:255"`
		Timestamp int64
	}

	// Migration
	err = db.AutoMigrate(&LogEntry{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}
	t.Log("Migration completed successfully")

	// Test create with no primary key
	entry := LogEntry{Message: "Test log entry", Timestamp: 1234567890}
	result := db.Create(&entry)
	t.Logf("Create result: Error=%v, RowsAffected=%d", result.Error, result.RowsAffected)
}