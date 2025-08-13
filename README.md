# GORM DuckDB Driver

A comprehensive DuckDB driver for [GORM](https://gorm.io), following the same patterns and conventions used by other official GORM drivers.

## Features

- Full GORM compatibility with custom migrator
- Auto-migration support with DuckDB-specific optimizations
- All standard SQL operations (CRUD)
- Transaction support with savepoints
- Index management
- Constraint support including foreign keys
- Comprehensive data type mapping
- Connection pooling support
- Auto-increment support with sequences and RETURNING clause
- Array data type support

## Quick Start

### Install

**Step 1:** Add the dependencies to your project:

```bash
go get -u gorm.io/gorm
go get -u github.com/greysquirr3l/gorm-duckdb-driver
```

**Step 2:** Add a `replace` directive to your `go.mod` file:

```go
module your-project

go 1.24

require (
    github.com/greysquirr3l/gorm-duckdb-driver v0.0.0
    gorm.io/gorm v1.30.1
)

// Replace directive required since the driver isn't published yet
replace github.com/greysquirr3l/gorm-duckdb-driver => github.com/greysquirr3l/gorm-duckdb-driver v0.2.6
```

### For Local Development

If you're working with a local copy of this driver, use a local replace directive:

```go
// For local development - replace with your local path
replace github.com/greysquirr3l/gorm-duckdb-driver => ../../

// For published version - replace with specific version
replace github.com/greysquirr3l/gorm-duckdb-driver => github.com/greysquirr3l/gorm-duckdb-driver v0.2.6
```

**Step 3:** Run `go mod tidy` to update dependencies:

```bash
go mod tidy
```

### Connect to Database

```go
import (
  duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
  "gorm.io/gorm"
)

// In-memory database
db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{})

// File-based database
db, err := gorm.Open(duckdb.Open("test.db"), &gorm.Config{})

// With custom configuration
db, err := gorm.Open(duckdb.New(duckdb.Config{
  DSN: "test.db",
  DefaultStringSize: 256,
}), &gorm.Config{})
```

## Example Application

This repository includes a comprehensive example application demonstrating all key features:

### Comprehensive Example (`example/`)

A single, comprehensive example that demonstrates:

- **Array Support**: StringArray, FloatArray, IntArray with full CRUD operations
- **Auto-Increment**: Sequences with RETURNING clause for ID generation  
- **Migrations**: Schema evolution with DuckDB-specific optimizations
- **Time Handling**: Time fields with manual control and timezone considerations
- **Data Types**: Complete mapping of Go types to DuckDB types
- **ALTER TABLE Fixes**: Demonstrates resolved DuckDB syntax limitations
- **Advanced Queries**: Aggregations, analytics, and transaction support

```bash
cd example
go run main.go
```

**Features Demonstrated:**
- ✅ Arrays (StringArray, FloatArray, IntArray)
- ✅ Migrations and auto-increment with sequences  
- ✅ Time handling and various data types
- ✅ ALTER TABLE fixes for DuckDB syntax
- ✅ Basic CRUD operations
- ✅ Advanced queries and transactions

> **⚠️ Important:** The example application must be executed using `go run main.go` from within the `example/` directory. It uses an in-memory database for clean demonstration runs.
```

## Data Type Mapping

| Go Type | DuckDB Type |
|---------|-------------|
| bool | BOOLEAN |
| int8 | TINYINT |
| int16 | SMALLINT |
| int32 | INTEGER |
| int64 | BIGINT |
| uint8 | UTINYINT |
| uint16 | USMALLINT |
| uint32 | UINTEGER |
| uint64 | UBIGINT |
| float32 | REAL |
| float64 | DOUBLE |
| string | VARCHAR(n) / TEXT |
| time.Time | TIMESTAMP |
| []byte | BLOB |

## Usage Examples

### Define Models

```go
type User struct {
  ID        uint      `gorm:"primarykey"`
  Name      string    `gorm:"size:100;not null"`
  Email     string    `gorm:"size:255;uniqueIndex"`
  Age       uint8
  Birthday  *time.Time
  CreatedAt time.Time
  UpdatedAt time.Time
}
```

### Auto Migration

```go
db.AutoMigrate(&User{})
```

### CRUD Operations

```go
// Create
user := User{Name: "John", Email: "john@example.com", Age: 30}
db.Create(&user)

// Read
var user User
db.First(&user, 1)                 // find user with integer primary key
db.First(&user, "name = ?", "John") // find user with name John

// Update
db.Model(&user).Update("name", "John Doe")
db.Model(&user).Updates(User{Name: "John Doe", Age: 31})

// Delete
db.Delete(&user, 1)
```

### Advanced Queries

```go
// Where
db.Where("name = ?", "John").Find(&users)
db.Where("age > ?", 18).Find(&users)

// Order
db.Order("age desc, name").Find(&users)

// Limit & Offset
db.Limit(3).Find(&users)
db.Offset(3).Limit(3).Find(&users)

// Group & Having
db.Model(&User{}).Group("name").Having("count(id) > ?", 1).Find(&users)
```

### Transactions

```go
db.Transaction(func(tx *gorm.DB) error {
  // do some database operations in the transaction
  if err := tx.Create(&User{Name: "John"}).Error; err != nil {
    return err
  }
  
  if err := tx.Create(&User{Name: "Jane"}).Error; err != nil {
    return err
  }
  
  return nil
})
```

### Raw SQL

```go
// Raw SQL
db.Raw("SELECT id, name, age FROM users WHERE name = ?", "John").Scan(&users)

// Exec
db.Exec("UPDATE users SET age = ? WHERE name = ?", 30, "John")
```

## Migration Features

The DuckDB driver includes a custom migrator that handles DuckDB-specific SQL syntax and provides enhanced functionality:

### Auto-Increment Support

The driver implements auto-increment using DuckDB sequences with the RETURNING clause:

```go
type User struct {
    ID   uint   `gorm:"primarykey"`  // Automatically uses sequence + RETURNING
    Name string `gorm:"size:100;not null"`
}

// Creates: CREATE SEQUENCE seq_users_id START 1
// Table:   CREATE TABLE users (id BIGINT DEFAULT nextval('seq_users_id') NOT NULL, ...)
// Insert:  INSERT INTO users (...) VALUES (...) RETURNING "id"
```

### DuckDB-Specific ALTER TABLE Handling

The migrator correctly handles DuckDB's ALTER COLUMN syntax limitations:

```go
// The migrator automatically splits DEFAULT clauses from type changes
// DuckDB: ALTER TABLE users ALTER COLUMN name TYPE VARCHAR(200)  ✅
// Not:    ALTER TABLE users ALTER COLUMN name TYPE VARCHAR(200) DEFAULT 'value'  ❌
```

### Table Operations

```go
// Create table
db.Migrator().CreateTable(&User{})

// Drop table  
db.Migrator().DropTable(&User{})

// Check if table exists
db.Migrator().HasTable(&User{})

// Rename table
db.Migrator().RenameTable(&User{}, &Admin{})
```

### Column Operations

```go
// Add column
db.Migrator().AddColumn(&User{}, "nickname")

// Drop column
db.Migrator().DropColumn(&User{}, "nickname")

// Alter column
db.Migrator().AlterColumn(&User{}, "name")

// Check if column exists
db.Migrator().HasColumn(&User{}, "name")

// Rename column
db.Migrator().RenameColumn(&User{}, "name", "full_name")

// Get column types
columnTypes, _ := db.Migrator().ColumnTypes(&User{})
```

### Index Operations

```go
// Create index
db.Migrator().CreateIndex(&User{}, "idx_user_name")

// Drop index
db.Migrator().DropIndex(&User{}, "idx_user_name")

// Check if index exists
db.Migrator().HasIndex(&User{}, "idx_user_name")

// Rename index
db.Migrator().RenameIndex(&User{}, "old_idx", "new_idx")
```

### Constraint Operations

```go
// Create constraint
db.Migrator().CreateConstraint(&User{}, "fk_user_company")

// Drop constraint
db.Migrator().DropConstraint(&User{}, "fk_user_company")

// Check if constraint exists
db.Migrator().HasConstraint(&User{}, "fk_user_company")
```

## Configuration Options

```go
type Config struct {
    DriverName        string        // Driver name, default: "duckdb"
    DSN               string        // Database source name
    Conn              gorm.ConnPool // Custom connection pool
    DefaultStringSize uint          // Default size for VARCHAR columns, default: 256
}
```

## Notes

- DuckDB is an embedded analytical database that excels at OLAP workloads
- The driver supports both in-memory and file-based databases
- All standard GORM features are supported including associations, hooks, and scopes
- The driver follows DuckDB's SQL dialect and capabilities
- For production use, consider DuckDB's performance characteristics for your specific use case

## Known Limitations

While this driver provides full GORM compatibility, there are some DuckDB-specific considerations:

### ALTER TABLE Syntax

**Resolved in Current Version** ✅

Previous versions had issues with ALTER COLUMN statements containing DEFAULT clauses. This has been fixed in the custom migrator:

- **Before:** `ALTER TABLE users ALTER COLUMN name TYPE VARCHAR(200) DEFAULT 'value'` (syntax error)
- **After:** Split into separate `ALTER COLUMN ... TYPE ...` and default handling operations

### Migration Schema Validation

**Issue:** DuckDB's `PRAGMA table_info()` returns slightly different column metadata format than PostgreSQL/MySQL.

**Symptoms:**

- GORM AutoMigrate occasionally reports false schema differences
- Unnecessary migration attempts on startup  
- Warnings in logs about column type mismatches

**Example Warning:**

```text
[WARN] column type mismatch: expected 'VARCHAR', got 'STRING'
```

**Workaround:**

```go
// Disable automatic migration validation for specific cases
db.AutoMigrate(&YourModel{})
// Add manual validation if needed
```

**Impact:** Low - Cosmetic warnings, doesn't affect functionality

### Transaction Isolation Levels

**Issue:** DuckDB has limited transaction isolation level support compared to traditional databases.

**Symptoms:**

- `db.Begin().Isolation()` methods have limited options
- Some GORM transaction patterns may not work as expected
- Read phenomena behavior differs from PostgreSQL

**Workaround:**

```go
// Use simpler transaction patterns
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// Perform operations...
if err := tx.Commit().Error; err != nil {
    return err
}
```

**Impact:** Low - Simple transactions work fine, complex isolation scenarios need adjustment

### Time Pointer Conversion

**Issue:** Current implementation has limitations with `*time.Time` pointer conversion in some edge cases.

**Symptoms:**

- Potential issues when working with nullable time fields
- Some time pointer operations may not behave identically to other GORM drivers

**Workaround:**

```go
// Use time.Time instead of *time.Time when possible
type Model struct {
    ID        uint      `gorm:"primarykey"`
    CreatedAt time.Time // Preferred
    UpdatedAt time.Time // Preferred
    DeletedAt gorm.DeletedAt `gorm:"index"` // This works fine
}
```

**Impact:** Low - Standard GORM time handling works correctly

## Performance Considerations

- DuckDB is optimized for analytical workloads (OLAP) rather than transactional workloads (OLTP)
- For high-frequency write operations, consider batching or using traditional OLTP databases
- DuckDB excels at complex queries, aggregations, and read-heavy workloads
- For production use, consider DuckDB's performance characteristics for your specific use case

## Contributing

This DuckDB driver aims to become an official GORM driver. Contributions are welcome!

### Development Setup

```bash
git clone https://github.com/greysquirr3l/gorm-duckdb-driver.git
cd gorm-duckdb-driver
go mod tidy
```

### Running the Example

Test the comprehensive example application:

```bash
# Test all key features in one comprehensive example
cd example && go run main.go
```

> **📝 Note:** The example uses an in-memory database (`:memory:`) for clean demonstration runs. All data is cleaned up automatically when the program exits.

### Running Tests

```bash
# Run all tests
go test -v

# Run with coverage
go test -v -cover

# Run specific test
go test -v -run TestMigration
```

### Issue Reporting

Please use our [Issue Template](ISSUE_TEMPLATE.md) when reporting bugs. For common issues, check the `bugs/` directory for known workarounds.

### Submitting to GORM

This driver follows GORM's architecture and coding standards. Once stable and well-tested by the community, it will be submitted for inclusion in the official GORM drivers under `go-gorm/duckdb`.

Current status:

- ✅ Full GORM interface implementation
- ✅ Custom migrator with DuckDB-specific optimizations
- ✅ Auto-increment support with sequences and RETURNING clause
- ✅ ALTER TABLE syntax handling for DuckDB
- ✅ Comprehensive test suite and example applications
- ✅ Array data type support
- ✅ Foreign key constraint support
- ✅ Documentation and examples
- 🔄 Community testing phase
- ⏳ Awaiting official GORM integration

## License

This driver is released under the MIT License, consistent with GORM's licensing.
