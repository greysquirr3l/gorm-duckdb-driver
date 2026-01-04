# RowQuery callback fails to set Statement.Dest causing Raw().Row() to return nil and panic

## Issue Description

GORM's default `RowQuery` callback has a critical bug that causes `Raw().Row()` to return `nil` instead of a valid `*sql.Row`, leading to nil pointer panics when attempting to scan results. This affects production applications using any GORM dialector that relies on the standard callback system.

## GORM Version

- **Affected Version**: v1.30.2 (and likely earlier versions)
- **Severity**: High - Causes runtime panics in production code

## Related Issues

This issue is related to but distinct from #6222, which covers the DryRun case. This bug affects normal (non-DryRun) operation.

## Minimal Reproduction Case

```go
package main

import (
    "log"
    "gorm.io/gorm"
    "gorm.io/driver/sqlite" // Any dialector
)

func main() {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // This should work but returns nil with GORM's default callback
    row := db.Raw("SELECT 1 as test_value").Row()
    if row == nil {
        log.Fatal("BUG: Raw().Row() returned nil instead of *sql.Row")
    }

    var result int
    err = row.Scan(&result) // This will panic: nil pointer dereference
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Result: %d", result)
}
```

## Expected Behavior

`Raw().Row()` should return a valid `*sql.Row` that can be used with `row.Scan()`.

## Actual Behavior

`Raw().Row()` returns `nil`, causing `row.Scan()` to panic with nil pointer dereference.

## Root Cause Analysis

The bug is in GORM's `RowQuery` callback implementation. The callback should:

1. Execute `QueryRowContext()` on the connection pool  
2. Assign the returned `*sql.Row` to `db.Statement.Dest`
3. Allow subsequent `Row()` calls to return the assigned row

**What's happening instead**: The default callback fails to properly assign the result to `db.Statement.Dest`, leaving it `nil`.

## Evidence

Our investigation shows:

- ✅ Direct driver calls work: `sql.DB.QueryRowContext()` returns valid rows
- ✅ GORM's connection pool works: `db.Statement.ConnPool.QueryRowContext()` returns valid rows  
- ✅ Other methods work: `Raw().Rows()` and `Raw().Scan()` work fine
- ❌ Only `Raw().Row()` is broken due to callback bug

## Working Fix

We've implemented and tested a working fix by replacing the default callback:

```go
// Working RowQuery callback implementation
func rowQueryCallback(db *gorm.DB) {
    if db.Error != nil || db.Statement.SQL.Len() == 0 || db.DryRun {
        return
    }

    // Handle both single row and multiple rows cases
    if isRows, ok := db.Get("rows"); ok && isRows.(bool) {
        // Multiple rows - call QueryContext  
        db.Statement.Settings.Delete("rows")
        db.Statement.Dest, db.Error = db.Statement.ConnPool.QueryContext(
            db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
    } else {
        // Single row - call QueryRowContext (this is what the default callback fails to do)
        db.Statement.Dest = db.Statement.ConnPool.QueryRowContext(
            db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
    }

    db.RowsAffected = -1
}

// Apply the fix
err := db.Callback().Row().Replace("gorm:row", rowQueryCallback)
```

## Impact Assessment

### Severity: HIGH

- **Runtime Crashes**: Nil pointer panics in production applications
- **Data Access Failures**: Unable to execute single-row queries via `Raw().Row()`  
- **Common Use Cases**: Information schema queries, `SELECT COUNT(*)`, configuration queries
- **Silent Failures**: May work intermittently depending on GORM version/configuration

### Affected Operations

- Information schema queries (common in migrations)
- Single value queries: `SELECT COUNT(*)`, `SELECT EXISTS(...)`
- Configuration queries: `SELECT current_database()`, version checks  
- Any application code using `Raw().Row()`

## Test Case

```go
func TestRowCallbackBug(t *testing.T) {
    db, _ := gorm.Open(dialector, &gorm.Config{})
    
    // Test the bug
    row := db.Raw("SELECT 1").Row()
    require.NotNil(t, row, "Row() should not return nil")
    
    var result int
    err := row.Scan(&result)
    require.NoError(t, err)
    require.Equal(t, 1, result)
}
```

## Proposed Solution

1. **Fix the default RowQuery callback** to properly assign `Statement.Dest`
2. **Add comprehensive tests** for `Raw().Row()` functionality
3. **Ensure backwards compatibility** with existing callback system

The core fix would be ensuring that the default callback properly calls:

```go
db.Statement.Dest = db.Statement.ConnPool.QueryRowContext(
    db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
```

## Environment

- Go Version: go1.21+
- GORM Version: v1.30.2  
- Database: Affects all databases (tested with SQLite, DuckDB)
- OS: macOS, Linux (likely all platforms)

## Additional Context

This bug has been affecting dialector authors and application developers. We've implemented a comprehensive workaround in our DuckDB driver that properly handles the callback replacement, but the upstream fix in GORM would benefit the entire ecosystem.

The issue is particularly problematic because:

1. The error is not obvious (nil pointer panic rather than clear error message)
2. Other Raw() methods work fine, making debugging difficult  
3. It affects common operations like schema introspection
4. Production applications crash rather than getting proper error handling
