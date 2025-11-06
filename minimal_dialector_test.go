package duckdb

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

// Minimal dialector to test basic GORM functionality
type MinimalDialector struct {
	DSN string
}

func (d MinimalDialector) Name() string {
	return "duckdb"
}

func (d MinimalDialector) Initialize(db *gorm.DB) error {
	return nil
}

func (d MinimalDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return migrator.Migrator{Config: migrator.Config{
		DB:                          db,
		Dialector:                   d,
		CreateIndexAfterCreateTable: true,
	}}
}

func (d MinimalDialector) DataTypeOf(field *schema.Field) string {
	switch field.DataType {
	case schema.String:
		return "TEXT"
	case schema.Int:
		return "INTEGER"
	default:
		return "TEXT"
	}
}

func (d MinimalDialector) DefaultValueOf(field *schema.Field) clause.Expression {
	return clause.Expr{SQL: "''"}
}

func (d MinimalDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	writer.WriteByte('?')
}

func (d MinimalDialector) QuoteTo(writer clause.Writer, str string) {
	writer.WriteByte('"')
	writer.WriteString(str)
	writer.WriteByte('"')
}

func (d MinimalDialector) Explain(sql string, vars ...interface{}) string {
	return fmt.Sprintf("EXPLAIN %s", sql)
}

// Test with minimal dialector
func TestMinimalDialector(t *testing.T) {
	t.Log("=== Minimal Dialector Test ===")

	// Enable debug mode
	os.Setenv("GORM_DUCKDB_DEBUG", "1")
	defer os.Unsetenv("GORM_DUCKDB_DEBUG")

	// Open raw database connection
	rawDB, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open raw DuckDB: %v", err)
	}
	defer rawDB.Close()

	// Create minimal dialector  
	dialector := MinimalDialector{DSN: ":memory:"}

	// Open GORM with minimal dialector
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Info),
		ConnPool:    rawDB,  // Use our raw connection
	})
	if err != nil {
		t.Fatalf("Failed to open GORM with minimal dialector: %v", err)
	}

	// Add debug callbacks
	db.Callback().Create().Before("gorm:create").Register("debug:minimal_before", func(db *gorm.DB) {
		t.Logf("[MINIMAL] Before gorm:create - SQL: '%s', Clauses: %+v", db.Statement.SQL.String(), db.Statement.Clauses)
	})

	db.Callback().Create().After("gorm:create").Register("debug:minimal_after", func(db *gorm.DB) {
		t.Logf("[MINIMAL] After gorm:create - SQL: '%s', Clauses: %+v", db.Statement.SQL.String(), db.Statement.Clauses)
	})

	// Simple model
	type SimpleModel struct {
		ID   int    `gorm:"primaryKey"`
		Name string
	}

	// Create table manually (skip migration)
	_, err = rawDB.Exec("CREATE TABLE simple_models (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test create
	model := SimpleModel{ID: 1, Name: "Minimal Test"}
	result := db.Create(&model)
	t.Logf("Create result: Error=%v, RowsAffected=%d", result.Error, result.RowsAffected)
}