package duckdb

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Simple logger for testing
func newTestLogger(prefix string) logger.Interface {
	return logger.New(
		&testWriter{prefix: prefix},
		logger.Config{
			SlowThreshold: 0, // Log all SQL
			LogLevel:      logger.Info,
			Colorful:      false,
		},
	)
}

type testWriter struct {
	prefix string
}

func (w *testWriter) Printf(_ string, _ ...interface{}) {
	// Just print to test output - simplified for this test
}
func TestDialectorComparison(t *testing.T) {
	t.Log("=== Dialector Comparison Test ===")

	// Setup SQLite
	sqliteDialector := sqlite.Open(":memory:")
	sqliteDB, err := gorm.Open(sqliteDialector, &gorm.Config{
		Logger: newTestLogger("SQLITE"),
	})
	if err != nil {
		t.Fatalf("Failed to open SQLite: %v", err)
	}

	// Setup DuckDB
	duckdbDialector := Dialector{
		Config: &Config{
			DSN: ":memory:",
		},
	}
	duckdbDB, err := gorm.Open(duckdbDialector, &gorm.Config{
		Logger: newTestLogger("DUCKDB"),
	})
	if err != nil {
		t.Fatalf("Failed to open DuckDB: %v", err)
	}

	// Compare Name() method
	t.Logf("SQLite Name(): %s", sqliteDialector.Name())
	t.Logf("DuckDB Name(): %s", duckdbDialector.Name())

	// Compare QuoteTo methods by writing to a string builder
	var sqliteQuote, duckdbQuote strings.Builder
	
	sqliteDialector.QuoteTo(&sqliteQuote, "test_table")
	duckdbDialector.QuoteTo(&duckdbQuote, "test_table")
	
	t.Logf("SQLite QuoteTo('test_table'): %s", sqliteQuote.String())
	t.Logf("DuckDB QuoteTo('test_table'): %s", duckdbQuote.String())

	// Test with complex identifier
	sqliteQuote.Reset()
	duckdbQuote.Reset()
	
	sqliteDialector.QuoteTo(&sqliteQuote, "test.column")
	duckdbDialector.QuoteTo(&duckdbQuote, "test.column")
	
	t.Logf("SQLite QuoteTo('test.column'): %s", sqliteQuote.String())
	t.Logf("DuckDB QuoteTo('test.column'): %s", duckdbQuote.String())

	// Create test tables with both
	type TestModel struct {
		ID   uint   `gorm:"primaryKey;autoIncrement"`
		Name string `gorm:"size:100"`
	}

	t.Log("\n=== Migration Test ===")
	err = sqliteDB.AutoMigrate(&TestModel{})
	if err != nil {
		t.Errorf("SQLite migration failed: %v", err)
	} else {
		t.Log("SQLite migration: SUCCESS")
	}

	err = duckdbDB.AutoMigrate(&TestModel{})
	if err != nil {
		t.Errorf("DuckDB migration failed: %v", err)
	} else {
		t.Log("DuckDB migration: SUCCESS")
	}

	// Test creating records
	t.Log("\n=== Create Test ===")
	
	sqliteModel := TestModel{Name: "SQLite Test"}
	result := sqliteDB.Create(&sqliteModel)
	t.Logf("SQLite Create - Error: %v, RowsAffected: %d, ID: %d", 
		result.Error, result.RowsAffected, sqliteModel.ID)

	duckdbModel := TestModel{Name: "DuckDB Test"}
	result = duckdbDB.Create(&duckdbModel)
	t.Logf("DuckDB Create - Error: %v, RowsAffected: %d, ID: %d", 
		result.Error, result.RowsAffected, duckdbModel.ID)
}

// Custom writer to capture clause output
type clauseWriter struct {
	content strings.Builder
}

func (w *clauseWriter) WriteByte(b byte) error {
	if err := w.content.WriteByte(b); err != nil {
		return fmt.Errorf("failed to write byte: %w", err)
	}
	return nil
}

func (w *clauseWriter) WriteString(s string) (int, error) {
	n, err := w.content.WriteString(s)
	if err != nil {
		return n, fmt.Errorf("failed to write string: %w", err)
	}
	return n, nil
}

func (w *clauseWriter) String() string {
	return w.content.String()
}

// Test the actual clause building process
func TestClauseBuilding(t *testing.T) {
	t.Log("=== Clause Building Test ===")

	// Test SQLite clause building
	sqliteDialector := sqlite.Open(":memory:")
	var sqliteWriter clauseWriter
	sqliteDialector.QuoteTo(&sqliteWriter, "users")
	t.Logf("SQLite clause building for 'users': %s", sqliteWriter.String())

	// Test DuckDB clause building
	duckdbDialector := Dialector{
		Config: &Config{
			DSN: ":memory:",
		},
	}
	var duckdbWriter clauseWriter
	duckdbDialector.QuoteTo(&duckdbWriter, "users")
	t.Logf("DuckDB clause building for 'users': %s", duckdbWriter.String())

	// Test both with quotes and special characters
	testCases := []string{
		"simple",
		"table.column", 
		`"quoted"`,
		"user-name",
		"user name",
	}

	for _, testCase := range testCases {
		sqliteWriter.content.Reset()
		duckdbWriter.content.Reset()
		
		sqliteDialector.QuoteTo(&sqliteWriter, testCase)
		duckdbDialector.QuoteTo(&duckdbWriter, testCase)
		
		t.Logf("Input: %s", testCase)
		t.Logf("  SQLite: %s", sqliteWriter.String())
		t.Logf("  DuckDB: %s", duckdbWriter.String())
	}
}

// Test if the issue is with DataTypeOf
func TestDataTypeComparison(t *testing.T) {
	t.Log("=== DataType Comparison Test ===")

	// We can't easily test SQLite's DataTypeOf without accessing internals,
	// but we can test our DuckDB implementation
	duckdbDialector := Dialector{
		Config: &Config{
			DSN: ":memory:",
		},
	}
	
	// Create a mock schema field for testing
	// This is a simplified test - in real GORM this would be more complex
	t.Logf("DuckDB Dialector Name: %s", duckdbDialector.Name())
}

func TestDebugSQL(t *testing.T) {
	t.Log("=== Debug SQL Generation ===")
	
	// Enable debug mode
	if err := os.Setenv("GORM_DUCKDB_DEBUG", "1"); err != nil {
		t.Fatalf("Failed to set debug environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("GORM_DUCKDB_DEBUG"); err != nil {
			t.Logf("Failed to unset debug environment variable: %v", err)
		}
	}()

	// Test with DuckDB
	duckdbDialector := Dialector{
		Config: &Config{
			DSN: ":memory:",
		},
	}
	db, err := gorm.Open(duckdbDialector, &gorm.Config{
		Logger: newTestLogger("DEBUG_DUCKDB"),
	})
	if err != nil {
		t.Fatalf("Failed to open DuckDB: %v", err)
	}

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

	// Now try create - this should show us exactly what's happening
	model := SimpleModel{Name: "Test"}
	
	t.Log("About to call db.Create()...")
	result := db.Create(&model)
	t.Logf("Create result: Error=%v, RowsAffected=%d, ID=%d", 
		result.Error, result.RowsAffected, model.ID)
		
	// Let's also try a raw SQL insert to verify the connection works
	t.Log("Testing raw SQL...")
	err = db.Exec("INSERT INTO simple_models (name) VALUES (?)", "Raw Test").Error
	if err != nil {
		t.Logf("Raw SQL failed: %v", err)
	} else {
		t.Log("Raw SQL succeeded")
	}
}