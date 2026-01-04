package duckdb

import (
	"context"
	"database/sql"
	"testing"
	"time"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DebugUser struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	CreatedAt time.Time
}

func TestDebugTableStructure(t *testing.T) {
	db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Enable SQL logging
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Test 1: Direct sql.DB QueryRowContext (this should work)
	t.Logf("=== Test 1: Direct QueryRowContext ===")
	row := sqlDB.QueryRowContext(context.Background(), "SELECT 1")
	var result1 int
	if err := row.Scan(&result1); err != nil {
		t.Errorf("Direct QueryRowContext failed: %v", err)
	} else {
		t.Logf("Direct QueryRowContext returned: %d", result1)
	}

	// Test 2: GORM Raw().Row() (this fails)
	t.Logf("=== Test 2: GORM Raw().Row() ===")

	// Let's debug GORM's internal state
	rawTx := db.Raw("SELECT 1")
	t.Logf("Raw tx Statement.SQL: %s", rawTx.Statement.SQL.String())
	t.Logf("Raw tx Statement.Vars: %v", rawTx.Statement.Vars)
	t.Logf("Raw tx DryRun: %t", rawTx.DryRun)
	t.Logf("Raw tx Statement.Context: %v", rawTx.Statement.Context)
	t.Logf("Raw tx Statement.ConnPool type: %T", rawTx.Statement.ConnPool)

	// Now call Row() and see what happens
	gormRow := rawTx.Row()
	t.Logf("GORM Row() returned: %v (nil: %t)", gormRow, gormRow == nil)
	t.Logf("Raw tx Statement.Dest type: %T", rawTx.Statement.Dest)
	t.Logf("Raw tx Statement.Dest value: %v", rawTx.Statement.Dest)

	// Test 6: Custom Row Callback
	t.Logf("=== Test 6: Custom Row Callback ===")
	// Replace GORM's row callback with our custom debugging version
	if err := db.Callback().Row().Replace("gorm:row", CustomRowQuery); err != nil {
		t.Errorf("Failed to replace row callback: %v", err)
	}

	// Now try GORM Raw().Row() with our custom callback
	t.Logf("Testing GORM with custom row callback...")
	customRawTx := db.Raw("SELECT 1")
	t.Logf("Custom raw tx Statement.SQL: %s", customRawTx.Statement.SQL.String())

	customRow := customRawTx.Row()
	t.Logf("Custom GORM Row() returned: %v (nil: %t)", customRow, customRow == nil)
	t.Logf("Custom raw tx Statement.Dest type: %T", customRawTx.Statement.Dest)
	t.Logf("Custom raw tx Statement.Dest value: %v", customRawTx.Statement.Dest)

	// Test 5: Debug Callback Execution
	t.Logf("=== Test 5: Debug Callback Execution ===")
	// Create a fresh transaction to debug the callback execution
	debugTx := db.Raw("SELECT 1")

	// Manually set the "rows" setting to false (same as Row() does)
	debugTx = debugTx.Set("rows", false)

	// Check Statement.Dest before callback execution
	t.Logf("Before callback - Statement.Dest: %v (type: %T)", debugTx.Statement.Dest, debugTx.Statement.Dest)

	// Execute the Row callback manually
	debugTx = debugTx.Callback().Row().Execute(debugTx)

	// Check Statement.Dest after callback execution
	t.Logf("After callback - Statement.Dest: %v (type: %T, nil: %t)", debugTx.Statement.Dest, debugTx.Statement.Dest, debugTx.Statement.Dest == nil)
	t.Logf("After callback - Statement.Error: %v", debugTx.Error)

	// Test 4: Simulate GORM's exact call
	t.Logf("=== Test 4: Simulate GORM's exact call ===")
	rawTx2 := db.Raw("SELECT 1")

	// Manually call what GORM should call
	t.Logf("Calling ConnPool.QueryRowContext with exact same parameters...")
	ctx := rawTx2.Statement.Context
	sqlStr := rawTx2.Statement.SQL.String()
	vars := rawTx2.Statement.Vars

	t.Logf("Context: %v, SQL: %s, Vars: %v", ctx, sqlStr, vars)

	connPool := rawTx2.Statement.ConnPool.(*sql.DB)
	manualRow := connPool.QueryRowContext(ctx, sqlStr, vars...)

	t.Logf("Manual QueryRowContext returned: %v (nil: %t)", manualRow, manualRow == nil)

	if manualRow != nil {
		var manualResult int
		if err := manualRow.Scan(&manualResult); err != nil {
			t.Errorf("Manual QueryRowContext scan failed: %v", err)
		} else {
			t.Logf("Manual QueryRowContext returned: %d", manualResult)
		}

		// Test assignment persistence
		t.Logf("=== Testing Assignment Persistence ===")
		t.Logf("Before assignment - Statement.Dest: %v", rawTx2.Statement.Dest)
		rawTx2.Statement.Dest = manualRow
		t.Logf("After assignment - Statement.Dest: %v (nil: %t)", rawTx2.Statement.Dest, rawTx2.Statement.Dest == nil)

		// Try type assertion like Row() does
		if row, ok := rawTx2.Statement.Dest.(*sql.Row); ok && row != nil {
			t.Logf("Type assertion successful: %v", row)
		} else {
			t.Logf("Type assertion failed: ok=%t, row=%v", ok, row)
		}
	}

	// Test 3: Check what GORM is actually doing - let's get the ConnPool
	t.Logf("=== Test 3: Check GORM ConnPool ===")
	// Get GORM's internal connection pool
	rawDB := db.Statement.ConnPool
	t.Logf("GORM ConnPool type: %T", rawDB)

	// Try to call QueryRowContext directly on GORM's ConnPool
	connPoolDB := rawDB.(*sql.DB)
	t.Logf("GORM ConnPool has QueryRowContext, calling it directly...")
	directRow := connPoolDB.QueryRowContext(context.Background(), "SELECT 1")
	t.Logf("Direct ConnPool QueryRowContext returned: %v (nil: %t)", directRow, directRow == nil)
	if directRow != nil {
		var result3 int
		if err := directRow.Scan(&result3); err != nil {
			t.Errorf("Direct ConnPool QueryRowContext scan failed: %v", err)
		} else {
			t.Logf("Direct ConnPool QueryRowContext returned: %d", result3)
		}
	}
}
