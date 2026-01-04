package duckdb

import (
	"testing"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
	"gorm.io/gorm"
)

func TestRowCallbackFix(t *testing.T) {
	// Test that our RowQuery callback fix resolves the nil Row issue
	db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	t.Log("🧪 Testing GORM Raw().Row() with our callback fix")

	// Test 1: Simple SELECT 1 query
	var result int
	row := db.Raw("SELECT 1").Row()
	if row == nil {
		t.Fatal("❌ GORM Raw().Row() returned nil - callback fix failed!")
	}

	err = row.Scan(&result)
	if err != nil {
		t.Fatalf("❌ Failed to scan result: %v", err)
	}

	if result != 1 {
		t.Fatalf("❌ Expected result=1, got result=%d", result)
	}

	t.Log("✅ Test 1: Raw().Row() with SELECT 1 - SUCCESS")

	// Test 2: Query with parameter
	var result2 string
	row2 := db.Raw("SELECT ? as test_value", "hello world").Row()
	if row2 == nil {
		t.Fatal("❌ GORM Raw().Row() with parameter returned nil!")
	}

	err = row2.Scan(&result2)
	if err != nil {
		t.Fatalf("❌ Failed to scan result2: %v", err)
	}

	if result2 != "hello world" {
		t.Fatalf("❌ Expected result2='hello world', got result2='%s'", result2)
	}

	t.Log("✅ Test 2: Raw().Row() with parameter - SUCCESS")

	// Test 3: Query information_schema (common failing case)
	var count int
	row3 := db.Raw("SELECT count(*) FROM information_schema.tables WHERE table_name = ?", "non_existent_table").Row()
	if row3 == nil {
		t.Fatal("❌ GORM Raw().Row() with information_schema query returned nil!")
	}

	err = row3.Scan(&count)
	if err != nil {
		t.Fatalf("❌ Failed to scan count: %v", err)
	}

	if count != 0 {
		t.Fatalf("❌ Expected count=0, got count=%d", count)
	}

	t.Log("✅ Test 3: Raw().Row() with information_schema query - SUCCESS")

	t.Log("🎉 ALL TESTS PASSED - RowQuery callback fix is working perfectly!")
}
