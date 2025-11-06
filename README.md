# GORM DuckDB Driver

[![Tests](https://img.shields.io/badge/tests-passing-brightgreen.svg)](https://github.com/greysquirr3l/gorm-duckdb-driver) [![Coverage](https://img.shields.io/badge/coverage-67.7%25-yellow.svg)](https://github.com/greysquirr3l/gorm-duckdb-driver)

A comprehensive DuckDB driver for [GORM](https://gorm.io), featuring native array support and complete GORM v2 compliance.

## Features

- **Complete GORM Compliance** - Full GORM v2 interface implementation with all required methods
- **Native Array Support** - First-class array types using DuckDB's native `Composite[T]` wrappers  
- **Advanced DuckDB Types** - 19 sophisticated types including JSON, Decimal, UUID, ENUM, UNION, and more
- **Production Ready** - Auto-increment with sequences, comprehensive error handling, connection pooling
- **Extension Management** - Built-in system for loading and managing DuckDB extensions
- **Schema Introspection** - Complete metadata access with ColumnTypes() and TableType() interfaces
- **High Performance** - DuckDB-optimized configurations and native type operations

## Quick Start

### Install

**Step 1:** Add dependencies to your project:

```bash
go get -u gorm.io/gorm
go get -u github.com/greysquirr3l/gorm-duckdb-driver
```

**Step 2:** Add replace directive to your `go.mod`:

```go
module your-project

go 1.24

require (
    github.com/greysquirr3l/gorm-duckdb-driver v0.6.1
    gorm.io/gorm v1.31.1
)

// Replace directive for latest release
replace github.com/greysquirr3l/gorm-duckdb-driver => github.com/greysquirr3l/gorm-duckdb-driver v0.6.1
```

**Step 3:** Run `go mod tidy`:

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

// With configuration
db, err := gorm.Open(duckdb.New(duckdb.Config{
  DSN: "test.db",
  DefaultStringSize: 256,
}), &gorm.Config{})
```

## Native Array Support

The driver provides native DuckDB array support using `duckdb.Composite[T]` wrappers, offering significant performance improvements over custom implementations.

### Key Features

- **Native Performance**: Uses DuckDB's built-in array types with 79% code reduction (371→77 lines)
- **Type Safety**: Full Go type safety with `duckdb.Composite[T]` wrappers
- **Array Functions**: Access to DuckDB's native array functions like `range()`, `array_length()`, `array_has()`
- **GORM Integration**: Implements `GormDataType()`, `driver.Valuer`, and `sql.Scanner` interfaces

### Usage

```go
// Define model with array fields
type Product struct {
    ID         uint                `gorm:"primaryKey"`
    Name       string              `gorm:"size:100"`
    Categories duckdb.StringArray  `json:"categories"`
    Scores     duckdb.FloatArray   `json:"scores"`
    ViewCounts duckdb.IntArray     `json:"view_counts"`
}

// Create arrays
product := Product{
    Categories: duckdb.NewStringArray([]string{"software", "analytics"}),
    Scores:     duckdb.NewFloatArray([]float64{4.5, 4.8, 4.2}),
    ViewCounts: duckdb.NewIntArray([]int64{1250, 890, 2340}),
}

// Insert and query using Raw SQL (recommended for arrays)
err := db.Exec(`INSERT INTO products (name, categories, scores, view_counts) VALUES (?, ?, ?, ?)`,
    product.Name, product.Categories.Get(), product.Scores.Get(), product.ViewCounts.Get()).Error

// Retrieve using Raw SQL with native array functions
var result Product
err = db.Raw(`SELECT * FROM products WHERE array_length(categories) > ?`, 1).Scan(&result).Error

// Access array data
categories := result.Categories.Get() // Returns []string
scores := result.Scores.Get()         // Returns []float64
counts := result.ViewCounts.Get()     // Returns []int64
```

### Important Notes

- **Use Raw SQL**: Native arrays work best with `Raw().Scan()` rather than GORM ORM methods (`First()`, `Find()`)
- **Float Arrays**: May return `duckdb.Decimal` types due to DuckDB's native behavior
- **Performance**: Native implementation provides significant performance benefits over JSON serialization

## Advanced DuckDB Types

The driver supports 19 advanced DuckDB types for comprehensive analytical capabilities:

### Core Types
- **Arrays**: StringArray, IntArray, FloatArray with native `Composite[T]` support
- **JSON/Document**: JSONType for flexible document storage
- **Precision Math**: DecimalType, HugeIntType (128-bit integers)
- **Temporal**: IntervalType, TimestampTZType with timezone support
- **Identifiers**: UUIDType for unique identifiers

### Advanced Types
- **Structured Data**: StructType, MapType, ListType for complex hierarchies
- **Variants**: ENUMType, UNIONType for type-safe variants
- **Binary**: BLOBType, BitStringType for binary data
- **Spatial**: GEOMETRYType with WKT support
- **Performance**: QueryHintType, PerformanceMetricsType for optimization

### Usage Example
```go
type AdvancedModel struct {
    ID        uint                `gorm:"primaryKey"`
    Config    duckdb.JSONType     `gorm:"type:json"`
    Price     duckdb.DecimalType  `gorm:"type:decimal(10,2)"`
    UUID      duckdb.UUIDType     `gorm:"type:uuid"`
    Tags      duckdb.StringArray  `json:"tags"`
    Location  duckdb.GEOMETRYType `gorm:"type:geometry"`
}
```

## Extension Management

Built-in DuckDB extension management for enhanced functionality:

```go
// Create database with extension support
db, err := gorm.Open(duckdb.OpenWithExtensions(":memory:", &duckdb.ExtensionConfig{
  AutoInstall:       true,
  PreloadExtensions: []string{"json", "parquet"},
  Timeout:           30 * time.Second,
}), &gorm.Config{})

// Get extension manager
manager, err := duckdb.GetExtensionManager(db)

// Load specific extensions
err = manager.LoadExtension("spatial")

// Use helper functions
helper := duckdb.NewExtensionHelper(manager)
err = helper.EnableAnalytics()    // json, parquet, fts, autocomplete
err = helper.EnableDataFormats()  // json, parquet, csv, excel, arrow
err = helper.EnableSpatial()      // spatial extension
```

## Basic Usage

### Define Models

```go
type User struct {
  ID        uint      `gorm:"primaryKey"`
  Name      string    `gorm:"size:100"`
  Email     string    `gorm:"uniqueIndex"`
  Age       int
  CreatedAt time.Time
  UpdatedAt time.Time
}
```

### CRUD Operations

```go
// Auto-migrate schema
db.AutoMigrate(&User{})

// Create
user := User{Name: "John", Email: "john@example.com", Age: 30}
db.Create(&user)

// Read
var user User
db.First(&user, 1)
db.Where("name = ?", "John").Find(&users)

// Update
db.Model(&user).Update("age", 31)

// Delete
db.Delete(&user, 1)
```

### Advanced Queries

```go
// Raw SQL with arrays
var products []Product
db.Raw("SELECT * FROM products WHERE array_has(categories, ?)", "software").Scan(&products)

// Analytical queries
db.Raw(`
    SELECT 
        category,
        COUNT(*) as count,
        AVG(price) as avg_price
    FROM products, UNNEST(categories) AS t(category)
    GROUP BY category
    ORDER BY count DESC
`).Scan(&results)
```

## Configuration Options

```go
config := duckdb.Config{
    DSN:               "production.db",
    DefaultStringSize: 256,
}

gormConfig := &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
}

db, err := gorm.Open(duckdb.New(config), gormConfig)

// Configure connection pool
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

## Migration Features

The driver includes a custom migrator with DuckDB-specific optimizations:

### Auto-Increment Support

```go
type User struct {
    ID   uint   `gorm:"primaryKey"`  // Uses sequence + RETURNING
    Name string `gorm:"size:100"`
}

// Creates: CREATE SEQUENCE seq_users_id START 1
// Insert:  INSERT INTO users (...) VALUES (...) RETURNING "id"
```

### Schema Operations

```go
// Migrator operations
db.Migrator().CreateTable(&User{})
db.Migrator().AddColumn(&User{}, "nickname")
db.Migrator().CreateIndex(&User{}, "idx_user_email")

// Check operations
hasTable := db.Migrator().HasTable(&User{})
hasColumn := db.Migrator().HasColumn(&User{}, "email")
```

## Error Translation

Comprehensive error handling with DuckDB-specific error patterns:

```go
// Automatic error translation
if err := db.Create(&user).Error; err != nil {
    if duckdb.IsDuplicateKeyError(err) {
        // Handle unique constraint violation
    }
    if duckdb.IsForeignKeyError(err) {
        // Handle foreign key violation
    }
}
```

## Examples

The repository includes comprehensive examples in the `example/` directory demonstrating:

- Basic CRUD operations
- Array usage patterns  
- Advanced type system
- Extension management
- Migration best practices
- Performance optimization

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Security

For security issues, please see [SECURITY.md](SECURITY.md) for reporting instructions.