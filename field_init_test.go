package duckdb_test

import (
	"fmt"
	"testing"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
)

// Test if the issue is struct field initialization
type TestStructWithDefaults struct {
	ID    uint `gorm:"primaryKey"`
	Array duckdb.StringArray `gorm:"column:array_col;default:[]"`
}

// Try with pointer field
type TestStructWithPointer struct {
	ID    uint `gorm:"primaryKey"`
	Array *duckdb.StringArray `gorm:"column:array_col"`
}

func TestFieldInitialization(t *testing.T) {
	db := setupTestDB(t)

	// Test 1: Regular field
	t.Run("Regular Field", func(t *testing.T) {
		err := db.AutoMigrate(&TestStructWithDefaults{})
		if err != nil {
			t.Fatal("AutoMigrate failed:", err)
		}

		original := TestStructWithDefaults{
			Array: duckdb.NewStringArray([]string{"test1", "test2"}),
		}

		if err := db.Create(&original).Error; err != nil {
			t.Fatal("Create failed:", err)
		}

		var retrieved TestStructWithDefaults
		if err := db.First(&retrieved, original.ID).Error; err != nil {
			t.Fatal("First failed:", err)
		}

		fmt.Printf("Regular field result: %v (len=%d)\n", retrieved.Array.Get(), len(retrieved.Array.Get()))
	})

	// Test 2: Pointer field
	t.Run("Pointer Field", func(t *testing.T) {
		err := db.AutoMigrate(&TestStructWithPointer{})
		if err != nil {
			t.Fatal("AutoMigrate failed:", err)
		}

		arr := duckdb.NewStringArray([]string{"test3", "test4"})
		original := TestStructWithPointer{
			Array: &arr,
		}

		if err := db.Create(&original).Error; err != nil {
			t.Fatal("Create failed:", err)
		}

		var retrieved TestStructWithPointer
		if err := db.First(&retrieved, original.ID).Error; err != nil {
			t.Fatal("First failed:", err)
		}

		if retrieved.Array != nil {
			fmt.Printf("Pointer field result: %v (len=%d)\n", retrieved.Array.Get(), len(retrieved.Array.Get()))
		} else {
			fmt.Println("Pointer field result: nil")
		}
	})
}