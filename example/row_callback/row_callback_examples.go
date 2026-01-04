// Example usage of GORM DuckDB driver with RowCallback workaround control
package main

import (
	"fmt"
	"log"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
	"gorm.io/gorm"
)

func main() {
	// Example 1: Default behavior (workaround enabled for current GORM versions)
	fmt.Println("=== Example 1: Default Behavior ===")
	db1, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var result1 int
	row1 := db1.Raw("SELECT 1").Row()
	if row1 != nil {
		row1.Scan(&result1)
		fmt.Printf("✅ Default: Got result = %d\n", result1)
	} else {
		fmt.Println("❌ Default: Raw().Row() returned nil")
	}

	// Example 2: Explicitly enable workaround (current recommended approach)
	fmt.Println("\n=== Example 2: Explicitly Enable Workaround ===")
	db2, err := gorm.Open(duckdb.OpenWithRowCallbackWorkaround(":memory:", true), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var result2 int
	row2 := db2.Raw("SELECT 2").Row()
	if row2 != nil {
		row2.Scan(&result2)
		fmt.Printf("✅ Enabled: Got result = %d\n", result2)
	} else {
		fmt.Println("❌ Enabled: Raw().Row() returned nil")
	}

	// Example 3: Advanced configuration
	fmt.Println("\n=== Example 3: Advanced Configuration ===")
	enabled := true
	config := &duckdb.Config{
		RowCallbackWorkaround: &enabled,
		DefaultStringSize:     512,
	}
	db3, err := gorm.Open(duckdb.OpenWithConfig(":memory:", config), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var result3 int
	row3 := db3.Raw("SELECT 3").Row()
	if row3 != nil {
		row3.Scan(&result3)
		fmt.Printf("✅ Advanced: Got result = %d\n", result3)
	} else {
		fmt.Println("❌ Advanced: Raw().Row() returned nil")
	}

	// Example 4: Future-proof - disable workaround for newer GORM versions
	fmt.Println("\n=== Example 4: Future-Proof (Workaround Disabled) ===")
	fmt.Println("This is for when GORM fixes the RowQuery callback bug...")

	db4, err := gorm.Open(duckdb.OpenWithRowCallbackWorkaround(":memory:", false), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var result4 int
	row4 := db4.Raw("SELECT 4").Row()
	if row4 != nil {
		row4.Scan(&result4)
		fmt.Printf("🎉 Future: GORM bug is fixed! Got result = %d\n", result4)
	} else {
		fmt.Println("⚠️  Future: Raw().Row() returned nil (GORM bug still exists)")
		fmt.Println("    When GORM v1.31+ fixes the bug, this should work!")
	}

	// Example 5: Demonstrate that Rows() always works
	fmt.Println("\n=== Example 5: Rows() Always Works ===")
	rows, err := db1.Raw("SELECT 5 UNION SELECT 6").Rows()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("✅ Multiple rows:")
	for rows.Next() {
		var val int
		rows.Scan(&val)
		fmt.Printf("   Row value: %d\n", val)
	}

	fmt.Println("\n🎯 All examples complete!")
	fmt.Println("\nUsage recommendations:")
	fmt.Println("• Current GORM versions: Use default behavior or explicitly enable workaround")
	fmt.Println("• Future GORM versions: Disable workaround when bug is fixed")
	fmt.Println("• See docs/GORM_ROW_CALLBACK_BUG_ANALYSIS.md for details")
}
