package duckdb_test

import (
	"testing"

	"github.com/marcboeker/go-duckdb/v2"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	gormDuckdb "github.com/greysquirr3l/gorm-duckdb-driver"
)

// setupNativeArrayTestDB sets up a test database for native array testing
func setupNativeArrayTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dialector := gormDuckdb.Open(":memory:")
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	return db
}

// TestNativeArraySupport tests DuckDB's native array support via go-duckdb v2.4.3
func TestNativeArraySupport(t *testing.T) {
	db := setupNativeArrayTestDB(t)

	// Create table with native array columns
	err := db.Exec(`
		CREATE TABLE native_arrays (
			id INTEGER PRIMARY KEY,
			string_array VARCHAR[3],
			int_array INTEGER[4], 
			float_array DOUBLE[2]
		)
	`).Error
	require.NoError(t, err)

	// Insert data using native DuckDB array syntax
	err = db.Exec(`
		INSERT INTO native_arrays VALUES 
		(1, array['hello', 'world', 'test'], array[1, 2, 3, 4], array[1.1, 2.2]),
		(2, array['foo', 'bar', 'baz'], array[10, 20, 30, 40], array[3.3, 4.4])
	`).Error
	require.NoError(t, err)

	// Test DuckDB native array support
	t.Run("DuckDB Native Arrays", func(t *testing.T) {

		// Test reading with Raw().Scan() using Composite wrapper
		t.Run("Raw Scan with Composite", func(t *testing.T) {
			var (
				id          int
				stringArray duckdb.Composite[[3]string]
				intArray    duckdb.Composite[[4]int32]
				floatArray  duckdb.Composite[[2]float64]
			)

			row := db.Raw("SELECT id, string_array, int_array, float_array FROM native_arrays WHERE id = ?", 1).Row()
			err := row.Scan(&id, &stringArray, &intArray, &floatArray)
			require.NoError(t, err)

			require.Equal(t, 1, id)
			require.Equal(t, [3]string{"hello", "world", "test"}, stringArray.Get())
			require.Equal(t, [4]int32{1, 2, 3, 4}, intArray.Get())
			require.Equal(t, [2]float64{1.1, 2.2}, floatArray.Get())
		})

		// Test reading multiple rows
		t.Run("Multiple Rows", func(t *testing.T) {
			rows, err := db.Raw("SELECT id, string_array, int_array, float_array FROM native_arrays ORDER BY id").Rows()
			require.NoError(t, err)
			defer rows.Close()

			expectedStringArrays := [][3]string{
				{"hello", "world", "test"},
				{"foo", "bar", "baz"},
			}
			expectedIntArrays := [][4]int32{
				{1, 2, 3, 4},
				{10, 20, 30, 40},
			}
			expectedFloatArrays := [][2]float64{
				{1.1, 2.2},
				{3.3, 4.4},
			}

			i := 0
			for rows.Next() {
				var (
					id          int
					stringArray duckdb.Composite[[3]string]
					intArray    duckdb.Composite[[4]int32]
					floatArray  duckdb.Composite[[2]float64]
				)

				err := rows.Scan(&id, &stringArray, &intArray, &floatArray)
				require.NoError(t, err)

				require.Equal(t, i+1, id)
				require.Equal(t, expectedStringArrays[i], stringArray.Get())
				require.Equal(t, expectedIntArrays[i], intArray.Get())
				require.Equal(t, expectedFloatArrays[i], floatArray.Get())
				i++
			}
			require.Equal(t, 2, i)
		})

		// Test working with GORM struct using Raw()
		t.Run("GORM Struct with Raw", func(t *testing.T) {
			type NativeArrayRecord struct {
				ID          int                         `gorm:"primaryKey"`
				StringArray duckdb.Composite[[3]string] `gorm:"column:string_array"`
				IntArray    duckdb.Composite[[4]int32]  `gorm:"column:int_array"`
				FloatArray  duckdb.Composite[[2]float64] `gorm:"column:float_array"`
			}

			var record NativeArrayRecord
			err := db.Raw("SELECT * FROM native_arrays WHERE id = ?", 1).Scan(&record).Error
			require.NoError(t, err)

			require.Equal(t, 1, record.ID)
			require.Equal(t, [3]string{"hello", "world", "test"}, record.StringArray.Get())
			require.Equal(t, [4]int32{1, 2, 3, 4}, record.IntArray.Get())
			require.Equal(t, [2]float64{1.1, 2.2}, record.FloatArray.Get())
		})

		// Test dynamic array operations
		t.Run("Dynamic Array Operations", func(t *testing.T) {
			// Test array functions
			var result duckdb.Composite[[]int32]
			err := db.Raw("SELECT range(1, 6)").Scan(&result).Error
			require.NoError(t, err)
			require.Equal(t, []int32{1, 2, 3, 4, 5}, result.Get())

			// Test array slicing
			var sliced duckdb.Composite[[2]string]
			err = db.Raw("SELECT (array['a', 'b', 'c', 'd'])[2:3]").Scan(&sliced).Error
			require.NoError(t, err)
			require.Equal(t, [2]string{"b", "c"}, sliced.Get())
		})
	})
}

// TestCompositeTypeUsage demonstrates how to use duckdb.Composite for various scenarios
func TestCompositeTypeUsage(t *testing.T) {
	db := setupNativeArrayTestDB(t)

	t.Run("Composite with Lists", func(t *testing.T) {
		// Test dynamic lists (not fixed arrays)
		var result duckdb.Composite[[]string]
		err := db.Raw("SELECT ['hello', 'world', 'variable', 'length']").Scan(&result).Error
		require.NoError(t, err)
		require.Equal(t, []string{"hello", "world", "variable", "length"}, result.Get())
	})

	t.Run("Composite with Nested Structures", func(t *testing.T) {
		// Test nested arrays
		var result duckdb.Composite[[][]int32]
		err := db.Raw("SELECT [[1, 2], [3, 4], [5, 6]]").Scan(&result).Error
		require.NoError(t, err)
		expected := [][]int32{{1, 2}, {3, 4}, {5, 6}}
		require.Equal(t, expected, result.Get())
	})
}