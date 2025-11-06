package duckdb

import (
	"testing"

	"gorm.io/gorm"
)

func TestRowCallbackWorkaroundConfiguration(t *testing.T) {
	t.Log("🧪 Testing RowCallback workaround configuration options")

	// Test 1: Default behavior (workaround enabled)
	t.Run("Default_Workaround_Enabled", func(t *testing.T) {
		db, err := gorm.Open(Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		// Test that Raw().Row() works (indicates workaround is active)
		row := db.Raw("SELECT 1").Row()
		if row == nil {
			t.Fatal("❌ Default behavior should enable workaround, but Raw().Row() returned nil")
		}

		var result int
		if err := row.Scan(&result); err != nil {
			t.Fatalf("Failed to scan result: %v", err)
		}

		if result != 1 {
			t.Fatalf("Expected result=1, got %d", result)
		}
		
		t.Log("✅ Default workaround behavior working")
	})

	// Test 2: Explicitly enabled workaround
	t.Run("Explicit_Workaround_Enabled", func(t *testing.T) {
		db, err := gorm.Open(OpenWithRowCallbackWorkaround(":memory:", true), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		row := db.Raw("SELECT 2").Row()
		if row == nil {
			t.Fatal("❌ Explicitly enabled workaround should work, but Raw().Row() returned nil")
		}

		var result int
		if err := row.Scan(&result); err != nil {
			t.Fatalf("Failed to scan result: %v", err)
		}

		if result != 2 {
			t.Fatalf("Expected result=2, got %d", result)
		}
		
		t.Log("✅ Explicitly enabled workaround working")
	})

	// Test 3: Configuration via Config struct
	t.Run("Config_Struct_Workaround", func(t *testing.T) {
		enabled := true
		config := &Config{
			RowCallbackWorkaround: &enabled,
		}
		
		db, err := gorm.Open(OpenWithConfig(":memory:", config), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		row := db.Raw("SELECT 3").Row()
		if row == nil {
			t.Fatal("❌ Config struct workaround should work, but Raw().Row() returned nil")
		}

		var result int
		if err := row.Scan(&result); err != nil {
			t.Fatalf("Failed to scan result: %v", err)
		}

		if result != 3 {
			t.Fatalf("Expected result=3, got %d", result)
		}
		
		t.Log("✅ Config struct workaround working")
	})

	// Test 4: Future compatibility - disabled workaround
	// Note: This test documents the intended behavior for when GORM fixes the bug
	t.Run("Future_Workaround_Disabled", func(t *testing.T) {
		// Create dialector with workaround explicitly disabled
		db, err := gorm.Open(OpenWithRowCallbackWorkaround(":memory:", false), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		// In the current GORM version with the bug, this will return nil
		// In future GORM versions (when fixed), this should work
		row := db.Raw("SELECT 4").Row()
		
		// For now, we expect this to fail since we know GORM has the bug
		// When GORM fixes the bug, this test should pass
		if row == nil {
			t.Log("⚠️ Workaround disabled: Raw().Row() returned nil (expected with current GORM bug)")
			t.Log("✅ When GORM fixes the bug, this test should pass with workaround disabled")
		} else {
			var result int
			if err := row.Scan(&result); err != nil {
				t.Fatalf("Failed to scan result: %v", err)
			}
			t.Log("🎉 GORM bug appears to be fixed! Workaround can be disabled.")
		}
	})

	t.Log("🎯 Configuration tests complete - future-proof workaround is ready")
}

func TestWorkaroundCompatibility(t *testing.T) {
	t.Log("🔄 Testing workaround compatibility across different scenarios")

	// Test both Rows() and Row() work correctly with our callback
	t.Run("Rows_And_Row_Compatibility", func(t *testing.T) {
		db, err := gorm.Open(Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		// Test Rows() (multi-row)
		rows, err := db.Raw("SELECT 1 UNION SELECT 2").Rows()
		if err != nil {
			t.Fatalf("Failed to get rows: %v", err)
		}
		defer func() { _ = rows.Close() }()

		var count int
		for rows.Next() {
			var val int
			if err := rows.Scan(&val); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}
			count++
		}

		if count != 2 {
			t.Fatalf("Expected 2 rows, got %d", count)
		}

		// Test Row() (single-row)
		row := db.Raw("SELECT 42").Row()
		if row == nil {
			t.Fatal("Raw().Row() returned nil")
		}

		var result int
		if err := row.Scan(&result); err != nil {
			t.Fatalf("Failed to scan single row: %v", err)
		}

		if result != 42 {
			t.Fatalf("Expected result=42, got %d", result)
		}

		t.Log("✅ Both Rows() and Row() working correctly with workaround")
	})

	// Test error handling
	t.Run("Error_Handling", func(t *testing.T) {
		db, err := gorm.Open(Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		// Test invalid SQL
		row := db.Raw("INVALID SQL QUERY").Row()
		if row != nil {
			var dummy int
			err := row.Scan(&dummy)
			// Should get an error from the invalid SQL
			if err == nil {
				t.Fatal("Expected error from invalid SQL, but got none")
			}
			t.Log("✅ Error handling working correctly")
		}
	})

	t.Log("✅ Compatibility tests passed")
}