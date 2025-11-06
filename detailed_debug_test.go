package duckdb

import (
	"os"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// Test with detailed callback logging to see where SQL generation fails
func TestDetailedCallbackDebug(t *testing.T) {
	t.Log("=== Detailed Callback Debug Test ===")

	// Enable debug mode
	os.Setenv("GORM_DUCKDB_DEBUG", "1")
	defer os.Unsetenv("GORM_DUCKDB_DEBUG")

	// Setup DuckDB with extra debugging
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

	// Add custom callback to inspect statement building
	db.Callback().Create().Before("gorm:create").Register("debug:before_create", func(db *gorm.DB) {
		t.Logf("DEBUG: Before gorm:create callback")
		t.Logf("DEBUG: Statement.SQL: '%s'", db.Statement.SQL.String())
		t.Logf("DEBUG: Statement.Table: '%s'", db.Statement.Table)
		t.Logf("DEBUG: Statement.Schema: %+v", db.Statement.Schema)
		if db.Statement.Schema != nil {
			t.Logf("DEBUG: Schema.Table: '%s'", db.Statement.Schema.Table)
			t.Logf("DEBUG: Schema.Name: '%s'", db.Statement.Schema.Name)
		}
		t.Logf("DEBUG: Statement.Clauses: %+v", db.Statement.Clauses)
	})

	db.Callback().Create().After("gorm:create").Register("debug:after_create", func(db *gorm.DB) {
		t.Logf("DEBUG: After gorm:create callback")
		t.Logf("DEBUG: Statement.SQL: '%s'", db.Statement.SQL.String())
		t.Logf("DEBUG: Statement.Vars: %+v", db.Statement.Vars)
		t.Logf("DEBUG: Error: %v", db.Error)
		t.Logf("DEBUG: RowsAffected: %d", db.RowsAffected)
	})

	// Add callback to inspect statement at each step
	db.Callback().Create().Before("gorm:before_create").Register("debug:statement_inspect", func(db *gorm.DB) {
		t.Logf("DEBUG: Before gorm:before_create - SQL: '%s'", db.Statement.SQL.String())
	})

	db.Callback().Create().After("gorm:before_create").Register("debug:after_before_create", func(db *gorm.DB) {
		t.Logf("DEBUG: After gorm:before_create - SQL: '%s'", db.Statement.SQL.String())
	})

	type SimpleModel struct {
		ID   uint   `gorm:"primaryKey;autoIncrement"`
		Name string
	}

	// Migration should work
	err = db.AutoMigrate(&SimpleModel{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}
	t.Log("Migration completed successfully")

	// Create test record with detailed debugging
	model := SimpleModel{Name: "Test"}
	
	t.Log("About to call db.Create()...")
	result := db.Create(&model)
	t.Logf("Create result: Error=%v, RowsAffected=%d, ID=%d", 
		result.Error, result.RowsAffected, model.ID)
}

// Test GORM's internal statement building process
func TestStatementBuilding(t *testing.T) {
	t.Log("=== Statement Building Test ===")

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

	type SimpleModel struct {
		ID   uint   `gorm:"primaryKey;autoIncrement"`
		Name string
	}

	// Migration
	err = db.AutoMigrate(&SimpleModel{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Manually build and test the statement
	model := SimpleModel{Name: "Manual Test"}
	stmt := &gorm.Statement{DB: db}
	stmt.Parse(&model)
	
	t.Logf("Manual statement parse:")
	t.Logf("  Schema: %+v", stmt.Schema)
	t.Logf("  Table: %s", stmt.Table)
	
	// Try to build the SQL manually using GORM's clause system
	stmt.AddClause(clause.Insert{})
	stmt.AddClause(clause.Values{Columns: []clause.Column{{Name: "name"}}, Values: [][]interface{}{{"Manual Test"}}})
	
	// Build the statement
	stmt.Build("INSERT")
	
	t.Logf("Manual statement build:")
	t.Logf("  SQL: %s", stmt.SQL.String())
	t.Logf("  Vars: %+v", stmt.Vars)

	// Try executing the manually built statement
	if stmt.SQL.Len() > 0 {
		result, err := stmt.ConnPool.ExecContext(stmt.Context, stmt.SQL.String(), stmt.Vars...)
		if err != nil {
			t.Logf("Manual execution failed: %v", err)
		} else {
			affected, _ := result.RowsAffected()
			t.Logf("Manual execution succeeded: %d rows affected", affected)
		}
	}
}