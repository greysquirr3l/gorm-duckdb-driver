# GORM DuckDB Driver - Comprehensive Example

This example demonstrates the full capabilities of the GORM DuckDB driver, showcasing all major features and fixes implemented in this driver.

## Features Demonstrated

### ✅ Array Support

- **StringArray**: Categories field in Product model
- **FloatArray**: Scores field in Product model  
- **IntArray**: ViewCounts field in Product model
- Array creation, retrieval, and updates

### ✅ Auto-Increment with Sequences

- Automatic sequence generation (`seq_tablename_id`)
- RETURNING clause for ID retrieval
- Works across all models (User, Post, Tag, Product)

### ✅ Migrations and Schema Management

- Auto-migration support
- Custom migrator with DuckDB-specific optimizations
- ALTER TABLE syntax fixes for DuckDB

### ✅ Data Types and Time Handling

- Various numeric types (uint, uint8, float64)
- String fields with size constraints
- Time fields (time.Time) with manual control
- Proper type mapping for DuckDB

### ✅ CRUD Operations

- Create with auto-increment IDs
- Read operations with filtering
- Update operations (single field and multiple fields)
- Delete operations
- Batch operations

### ✅ Advanced Features

- Complex queries with WHERE, GROUP BY, aggregations
- Transactions with rollback support
- Analytical queries (AVG, COUNT, CASE statements)
- Database state reporting

## Key Fixes Demonstrated

### ALTER TABLE Syntax Fix

**Problem**: DuckDB doesn't support `ALTER COLUMN ... TYPE ... DEFAULT ...` syntax
**Solution**: Custom migrator splits DEFAULT clauses from type changes
**Result**: ✅ No more "syntax error at or near 'DEFAULT'" errors

### Auto-Increment Implementation

**Problem**: DuckDB doesn't have native AUTO_INCREMENT
**Solution**: Custom sequences with RETURNING clause
**Implementation**: 

```sql
CREATE SEQUENCE seq_users_id START 1
CREATE TABLE users (id BIGINT DEFAULT nextval('seq_users_id') NOT NULL, ...)
INSERT INTO users (...) VALUES (...) RETURNING "id"
```

### Array Type Support

**Problem**: Go doesn't have native array types for DuckDB
**Solution**: Custom array types with proper serialization
**Types**: StringArray, FloatArray, IntArray

## Running the Example

```bash
cd example
go run main.go
```

**Note**: This example uses an in-memory database (`:memory:`), so all data is cleaned up automatically when the program exits.

## Output

The example produces detailed output showing:

1. **Connection and Migration**: Database setup and schema creation
2. **CRUD Operations**: User creation, reading, updating, and deletion
3. **Array Operations**: Product creation with arrays and array updates
4. **Advanced Queries**: Analytics, demographics, and transaction examples
5. **Final State**: Summary of all created records

## Model Definitions

### User Model

```go
type User struct {
    ID        uint      `gorm:"primaryKey;autoIncrement"`
    Name      string    `gorm:"size:100;not null"`
    Email     string    `gorm:"size:255;uniqueIndex"`
    Age       uint8
    Birthday  time.Time
    CreatedAt time.Time `gorm:"autoCreateTime:false"`
    UpdatedAt time.Time `gorm:"autoUpdateTime:false"`
}
```

### Product Model (with Arrays)

```go
type Product struct {
    ID          uint               `gorm:"primaryKey;autoIncrement"`
    Name        string             `gorm:"size:100;not null"`
    Price       float64
    Description string
    Categories  duckdb.StringArray // Array support
    Scores      duckdb.FloatArray  // Float array support
    ViewCounts  duckdb.IntArray    // Int array support
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

## Performance Notes

- DuckDB excels at analytical workloads (OLAP)
- Arrays are stored efficiently in DuckDB's columnar format
- Transactions are lightweight but have different isolation semantics than traditional RDBMS
- Auto-increment sequences perform well for moderate insert rates

## Limitations Addressed

1. **Relationship Complexity**: This example focuses on core functionality rather than complex GORM relationships
2. **DuckDB-Specific Syntax**: All SQL generation respects DuckDB's dialect limitations
3. **Array Operations**: Advanced array querying would require raw SQL for complex operations

## Next Steps

After running this example, you can:

1. Modify the models to test your specific use cases
2. Add more complex queries using raw SQL
3. Test with file-based databases instead of in-memory
4. Explore DuckDB's analytical capabilities with larger datasets

This example serves as both a test suite and a reference implementation for the GORM DuckDB driver.
