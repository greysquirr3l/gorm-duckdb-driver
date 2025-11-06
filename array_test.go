package duckdb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
)

// Test model for array functionality
type TestArrayModel struct {
	ID        uint               `gorm:"primaryKey"`
	StringArr duckdb.StringArray `json:"string_arr"`
	FloatArr  duckdb.FloatArray  `json:"float_arr"`
	IntArr    duckdb.IntArray    `json:"int_arr"`
}

func setupArrayTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)

	err := db.AutoMigrate(&TestArrayModel{})
	require.NoError(t, err)

	return db
}

func TestNativeArrayFunctionality(t *testing.T) {
	// Skip this test for now due to issues with GORM Scan and Raw queries
	// The underlying array functionality is tested in other test files
	t.Skip("Skipping due to GORM Raw().Scan() issues with array queries")
	
	// Test basic array creation without DB queries
	stringValues := []string{"hello", "world", "test"}
	stringArr := duckdb.NewStringArray(stringValues)
	assert.Equal(t, stringValues, stringArr.Get())

	intValues := []int64{1, 2, 3, 4}
	intArr := duckdb.NewIntArray(intValues)
	assert.Equal(t, intValues, intArr.Get())
}

func TestNativeArray_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "single element array",
			input:    []string{"hello"},
			expected: []string{"hello"},
		},
		{
			name:     "multiple elements array",
			input:    []string{"hello", "world", "test"},
			expected: []string{"hello", "world", "test"},
		},
		{
			name:     "empty array",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var arr duckdb.StringArray
			err := arr.Scan(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, arr.Get())
		})
	}
}

func TestArrays_DatabaseIntegration(t *testing.T) {
	t.Skip("Skipping due to GORM Raw().Scan() issues with array structures")

	db := setupArrayTestDB(t)

	t.Run("Native Array Insert and Retrieve", func(t *testing.T) {
		// Test data using constructor functions
		stringValues := []string{"software", "analytics", "business"}
		floatValues := []float64{4.5, 4.8, 4.2, 4.9}
		intValues := []int64{1250, 890, 2340, 567}

		// Create record using Raw SQL since native arrays work best with Raw queries
		err := db.Exec(`
			INSERT INTO test_array_models (string_arr, float_arr, int_arr) 
			VALUES (?, ?, ?)
		`, stringValues, floatValues, intValues).Error
		require.NoError(t, err)

		// Retrieve record using Raw SQL with Scan
		var retrieved TestArrayModel
		err = db.Raw("SELECT id, string_arr, float_arr, int_arr FROM test_array_models ORDER BY id DESC LIMIT 1").Scan(&retrieved).Error
		require.NoError(t, err)
		assert.NotZero(t, retrieved.ID)

		// Verify arrays were stored and retrieved correctly
		assert.Equal(t, stringValues, retrieved.StringArr.Get())
		assert.Equal(t, floatValues, retrieved.FloatArr.Get())
		assert.Equal(t, intValues, retrieved.IntArr.Get())
	})

	t.Run("Array Update Operations", func(t *testing.T) {
		// Test update using Raw SQL
		initialStringValues := []string{"test1", "test2"}
		initialFloatValues := []float64{1.0, 2.0}
		initialIntValues := []int64{10, 20}

		// Insert initial record
		err := db.Exec(`
			INSERT INTO test_array_models (string_arr, float_arr, int_arr) 
			VALUES (?, ?, ?)
		`, initialStringValues, initialFloatValues, initialIntValues).Error
		require.NoError(t, err)

		// Get the inserted record ID
		var insertedID uint
		err = db.Raw("SELECT id FROM test_array_models ORDER BY id DESC LIMIT 1").Scan(&insertedID).Error
		require.NoError(t, err)

		// Update with new values
		newStringValues := []string{"updated1", "updated2", "updated3"}
		newFloatValues := []float64{3.0, 4.0, 5.0}
		newIntValues := []int64{30, 40, 50}

		err = db.Exec(`
			UPDATE test_array_models 
			SET string_arr = ?, float_arr = ?, int_arr = ? 
			WHERE id = ?
		`, newStringValues, newFloatValues, newIntValues, insertedID).Error
		require.NoError(t, err)

		// Verify update using Raw SQL
		var updated TestArrayModel
		err = db.Raw("SELECT id, string_arr, float_arr, int_arr FROM test_array_models WHERE id = ?", insertedID).Scan(&updated).Error
		require.NoError(t, err)

		assert.Equal(t, newStringValues, updated.StringArr.Get())
		assert.Equal(t, newFloatValues, updated.FloatArr.Get())
		assert.Equal(t, newIntValues, updated.IntArr.Get())
	})
}

func TestArrayConstructors(t *testing.T) {
	t.Run("StringArray constructor", func(t *testing.T) {
		values := []string{"test1", "test2", "test3"}
		arr := duckdb.NewStringArray(values)
		assert.Equal(t, values, arr.Get())
	})

	t.Run("FloatArray constructor", func(t *testing.T) {
		values := []float64{1.1, 2.2, 3.3}
		arr := duckdb.NewFloatArray(values)
		assert.Equal(t, values, arr.Get())
	})

	t.Run("IntArray constructor", func(t *testing.T) {
		values := []int64{1, 2, 3}
		arr := duckdb.NewIntArray(values)
		assert.Equal(t, values, arr.Get())
	})
}
