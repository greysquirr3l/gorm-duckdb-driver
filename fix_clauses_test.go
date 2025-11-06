package duckdb

import (
	"os"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// Test if we can fix the nil Clauses map by manually initializing it
func TestFixStatementClauses(t *testing.T) {
	t.Log("=== Fix Statement Clauses Test ===")

	// Enable debug mode
	os.Setenv("GORM_DUCKDB_DEBUG", "1")
	defer os.Unsetenv("GORM_DUCKDB_DEBUG")

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

	// Add a callback to initialize Statement.Clauses if it's nil
	db.Callback().Create().Before("gorm:create").Register("fix:init_clauses", func(db *gorm.DB) {
		t.Logf("Before fix - Clauses: %+v", db.Statement.Clauses)
		if db.Statement.Clauses == nil {
			t.Log("Initializing nil Clauses map")
			db.Statement.Clauses = make(map[string]clause.Clause)
		}
		t.Logf("After fix - Clauses: %+v", db.Statement.Clauses)
	})

	// Add callback to inspect statement after GORM's create
	db.Callback().Create().After("gorm:create").Register("debug:after_gorm_create", func(db *gorm.DB) {
		t.Logf("After GORM create:")
		t.Logf("  SQL: '%s'", db.Statement.SQL.String())
		t.Logf("  Clauses: %+v", db.Statement.Clauses)
		t.Logf("  Error: %v", db.Error)
		t.Logf("  RowsAffected: %d", db.RowsAffected)
	})

	type SimpleModel struct {
		ID   uint   `gorm:"primaryKey;autoIncrement"`
		Name string
	}

	// Migration
	err = db.AutoMigrate(&SimpleModel{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Test create
	model := SimpleModel{Name: "Test with fixed clauses"}
	result := db.Create(&model)
	t.Logf("Create result: Error=%v, RowsAffected=%d, ID=%d", 
		result.Error, result.RowsAffected, model.ID)
}

// Test if the issue is specifically with our dialector or GORM version
func TestGORMVersion(t *testing.T) {
	t.Log("=== GORM Version Test ===")

	// Test if we can reproduce the issue with a minimal setup
	dialector := Dialector{
		Config: &Config{
			DSN: ":memory:",
		},
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open DuckDB: %v", err)
	}

	// Manually create a statement like GORM does
	stmt := &gorm.Statement{DB: db}
	
	// This should not panic if GORM is working correctly
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Panic during statement operations: %v", r)
		}
	}()

	type TestModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string
	}

	model := TestModel{Name: "Test"}
	
	// Test statement parsing
	err = stmt.Parse(&model)
	if err != nil {
		t.Logf("Parse failed: %v", err)
	} else {
		t.Logf("Parse succeeded: Table=%s, Schema=%s", stmt.Table, stmt.Schema.Name)
	}

	// Check if Clauses is nil after Parse
	t.Logf("Clauses after Parse: %+v (nil? %t)", stmt.Clauses, stmt.Clauses == nil)
	
	// Try to access Statement fields that should be initialized
	t.Logf("Statement.DB: %T", stmt.DB)
	t.Logf("Statement.Dest: %+v", stmt.Dest)
	t.Logf("Statement.Model: %+v", stmt.Model)
}