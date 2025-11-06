package duckdb

import (
	"os"
	"reflect"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Test to see what's happening inside GORM's create process
//nolint:errcheck,gosec // Test function with debug callbacks
func TestGORMCreateProcess(t *testing.T) {
	t.Log("=== GORM Create Process Test ===")

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

	// Add callback before gorm:before_create to inspect initial state
	db.Callback().Create().Before("gorm:before_create").Register("debug:start", func(db *gorm.DB) {
		t.Logf("[DEBUG:START] Statement state:")
		t.Logf("  Dest: %+v", db.Statement.Dest)
		t.Logf("  Model: %+v", db.Statement.Model)
		t.Logf("  ReflectValue: %+v", db.Statement.ReflectValue)
		if db.Statement.ReflectValue.IsValid() {
			t.Logf("  ReflectValue.Kind: %s", db.Statement.ReflectValue.Kind())
			t.Logf("  ReflectValue.Type: %s", db.Statement.ReflectValue.Type())
		}
		t.Logf("  Schema: %+v", db.Statement.Schema)
		t.Logf("  Table: %s", db.Statement.Table)
		t.Logf("  SQL: %s", db.Statement.SQL.String())
		t.Logf("  Clauses: %+v", db.Statement.Clauses)
	})

	// Add callback after gorm:before_create
	db.Callback().Create().After("gorm:before_create").Register("debug:before_create", func(db *gorm.DB) {
		t.Logf("[DEBUG:BEFORE_CREATE] After gorm:before_create:")
		t.Logf("  SQL: %s", db.Statement.SQL.String())
		t.Logf("  Clauses: %+v", db.Statement.Clauses)
		t.Logf("  Error: %v", db.Error)
	})

	// Add callback before gorm:create to see state just before core create
	db.Callback().Create().Before("gorm:create").Register("debug:pre_create", func(db *gorm.DB) {
		t.Logf("[DEBUG:PRE_CREATE] Before gorm:create:")
		t.Logf("  SQL: %s", db.Statement.SQL.String())
		t.Logf("  Clauses: %+v", db.Statement.Clauses)
		t.Logf("  Error: %v", db.Error)
		
		// Let's check if the model is valid for creation
		if db.Statement.Schema != nil {
			t.Logf("  Schema.Table: %s", db.Statement.Schema.Table)
			t.Logf("  Schema.Fields: %d fields", len(db.Statement.Schema.Fields))
			for _, field := range db.Statement.Schema.Fields {
				t.Logf("    Field: %s, DBName: %s, PrimaryKey: %t", field.Name, field.DBName, field.PrimaryKey)
			}
		}
		
		// Check if we have a valid destination
		if db.Statement.Dest != nil {
			destType := reflect.TypeOf(db.Statement.Dest)
			destValue := reflect.ValueOf(db.Statement.Dest)
			t.Logf("  Dest type: %s", destType)
			t.Logf("  Dest value: %+v", destValue)
		}
	})

	// Add callback after gorm:create to see what happened
	db.Callback().Create().After("gorm:create").Register("debug:post_create", func(db *gorm.DB) {
		t.Logf("[DEBUG:POST_CREATE] After gorm:create:")
		t.Logf("  SQL: %s", db.Statement.SQL.String())
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

	// Test create with detailed logging
	model := SimpleModel{Name: "Detailed Debug Test"}
	t.Log("\n=== Starting Create Operation ===")
	result := db.Create(&model)
	t.Logf("\n=== Final Result ===")
	t.Logf("Error: %v", result.Error)
	t.Logf("RowsAffected: %d", result.RowsAffected) 
	t.Logf("Model ID: %d", model.ID)
}