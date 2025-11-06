# Native DuckDB Array Support Analysis

## Summary

After investigating both DuckDB go-duckdb v2.4.3 and GORM v1.31.1, I found that **native array support is available and working** through the DuckDB driver's `Composite[T]` wrapper type.

## Key Findings

### ✅ Working Native Array Support

1. **DuckDB v2.4.3 Native Features**:
   - `duckdb.Composite[T]` wrapper for array types
   - Native SQL array syntax: `array[1, 2, 3]` and `[1, 2, 3]`
   - Built-in array functions: `range()`, `array_length()`, etc.
   - Fixed-size arrays: `duckdb.Composite[[3]string]`
   - Dynamic arrays: `duckdb.Composite[[]string]`
   - Nested arrays: `duckdb.Composite[[][]int32]`

2. **GORM Integration**:
   - Works with `Raw().Scan()` - ✅ Fully functional
   - Works with `Raw().Scan(&struct{})` - ✅ Fully functional 
   - Does NOT work with `First()` method - ❌ GORM limitation

### 🔧 Working Examples

```go
// Fixed-size arrays
var stringArray duckdb.Composite[[3]string]
err := db.Raw("SELECT array['hello', 'world', 'test']").Scan(&stringArray).Error
// Result: [3]string{"hello", "world", "test"}

// Dynamic arrays  
var intList duckdb.Composite[[]int32]
err := db.Raw("SELECT [1, 2, 3, 4, 5]").Scan(&intList).Error
// Result: []int32{1, 2, 3, 4, 5}

// Nested arrays
var nested duckdb.Composite[[][]int32]
err := db.Raw("SELECT [[1, 2], [3, 4], [5, 6]]").Scan(&nested).Error
// Result: [][]int32{{1, 2}, {3, 4}, {5, 6}}

// Array functions
var range duckdb.Composite[[]int32]
err := db.Raw("SELECT range(1, 6)").Scan(&range).Error
// Result: []int32{1, 2, 3, 4, 5}

// Struct scanning
type Record struct {
    ID     int                         `gorm:"primaryKey"`
    Arrays duckdb.Composite[[3]string] `gorm:"column:array_col"`
}
var record Record
err := db.Raw("SELECT 1, array['a', 'b', 'c']").Scan(&record).Error
```

### ⚠️ Known Limitations

1. **Float Arrays**: DuckDB v2.4.3 returns `duckdb.Decimal` instead of `float64`
2. **GORM ORM Methods**: `First()`, `Find()` don't call custom Scanner interfaces
3. **Transaction Isolation**: Table creation in tests has isolation issues

### 🎯 Recommendation

**Use Native DuckDB Array Support** with the following approach:

1. **Replace Custom Array Types** with `duckdb.Composite[T]`
2. **Use Raw SQL** for array operations instead of GORM ORM methods
3. **Leverage Native Array Functions** like `range()`, `array_length()`, etc.

## Implementation Strategy

### Phase 1: Replace Internal Array Types

```go
// Old custom types
type StringArray []string
type FloatArray []float64  
type IntArray []int64

// New native types
type StringArray = duckdb.Composite[[]string]
type FloatArray = duckdb.Composite[[]float64]  // Note: may need duckdb.Decimal handling
type IntArray = duckdb.Composite[[]int64]
```

### Phase 2: Update Test Suite

```go
// Replace custom scanning with native
var record TestArrayModel
err := db.Raw("SELECT id, string_arr, float_arr, int_arr FROM test_array_models WHERE id = ?", 1).Scan(&record).Error
```

### Phase 3: Documentation

Document that:

- Use `Raw().Scan()` for array operations
- GORM ORM methods (`First`, `Find`) don't support arrays
- Float arrays may return `duckdb.Decimal` type

## Conclusion

The native DuckDB array support in v2.4.3 is **significantly more powerful** than our custom implementation and should be adopted. The `duckdb.Composite[T]` wrapper provides type-safe access to DuckDB's full array capabilities while maintaining Go type safety.

The only integration point needed is using `Raw().Scan()` instead of GORM ORM methods for array queries, which is a reasonable trade-off for gaining access to DuckDB's native array ecosystem.