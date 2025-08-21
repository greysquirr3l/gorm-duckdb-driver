package duckdb_test

import (
	"database/sql/driver"
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

func TestStringArray_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    duckdb.StringArray
		expected string
	}{
		{
			name:     "empty array",
			input:    duckdb.StringArray{},
			expected: "[]",
		},
		{
			name:     "single element",
			input:    duckdb.StringArray{"hello"},
			expected: `["hello"]`,
		},
		{
			name:     "multiple elements",
			input:    duckdb.StringArray{"hello", "world", "test"},
			expected: `["hello","world","test"]`,
		},
		{
			name:     "elements with special characters",
			input:    duckdb.StringArray{"hello\"world", "test,comma", "newline\n"},
			expected: `["hello\"world","test,comma","newline\n"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.input.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestStringArray_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected duckdb.StringArray
		wantErr  bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "empty array string",
			input:    "[]",
			expected: duckdb.StringArray{},
			wantErr:  false,
		},
		{
			name:     "single element array",
			input:    `["hello"]`,
			expected: duckdb.StringArray{"hello"},
			wantErr:  false,
		},
		{
			name:     "multiple elements array",
			input:    `["hello","world","test"]`,
			expected: duckdb.StringArray{"hello", "world", "test"},
			wantErr:  false,
		},
		{
			name:     "array with spaces",
			input:    `["hello", "world", "test"]`,
			expected: duckdb.StringArray{"hello", "world", "test"},
			wantErr:  false,
		},
		{
			name:     "byte slice input",
			input:    []byte(`["hello","world"]`),
			expected: duckdb.StringArray{"hello", "world"},
			wantErr:  false,
		},
		{
			name:     "string slice input",
			input:    []string{"hello", "world"},
			expected: duckdb.StringArray{"hello", "world"},
			wantErr:  false,
		},
		{
			name:     "interface slice input",
			input:    []interface{}{"hello", "world"},
			expected: duckdb.StringArray{"hello", "world"},
			wantErr:  false,
		},
		{
			name:    "invalid json",
			input:   `["invalid json`,
			wantErr: true,
		},
		{
			name:    "non-string array element",
			input:   `[123, "hello"]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var arr duckdb.StringArray
			err := arr.Scan(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, arr)
		})
	}
}

func TestFloatArray_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    duckdb.FloatArray
		expected string
	}{
		{
			name:     "empty array",
			input:    duckdb.FloatArray{},
			expected: "[]",
		},
		{
			name:     "single element",
			input:    duckdb.FloatArray{3.14},
			expected: "[3.14]",
		},
		{
			name:     "multiple elements",
			input:    duckdb.FloatArray{1.1, 2.2, 3.3},
			expected: "[1.1,2.2,3.3]",
		},
		{
			name:     "with zero and negative",
			input:    duckdb.FloatArray{0.0, -1.5, 2.7},
			expected: "[0,-1.5,2.7]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.input.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestFloatArray_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected duckdb.FloatArray
		wantErr  bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "empty array string",
			input:    "[]",
			expected: duckdb.FloatArray{},
			wantErr:  false,
		},
		{
			name:     "single element array",
			input:    "[3.14]",
			expected: duckdb.FloatArray{3.14},
			wantErr:  false,
		},
		{
			name:     "multiple elements array",
			input:    "[1.1, 2.2, 3.3]",
			expected: duckdb.FloatArray{1.1, 2.2, 3.3},
			wantErr:  false,
		},
		{
			name:     "byte slice input",
			input:    []byte("[1.5, 2.5]"),
			expected: duckdb.FloatArray{1.5, 2.5},
			wantErr:  false,
		},
		{
			name:     "float slice input",
			input:    []float64{1.1, 2.2},
			expected: duckdb.FloatArray{1.1, 2.2},
			wantErr:  false,
		},
		{
			name:     "interface slice input",
			input:    []interface{}{1.1, 2.2},
			expected: duckdb.FloatArray{1.1, 2.2},
			wantErr:  false,
		},
		{
			name:    "invalid json",
			input:   "[invalid json",
			wantErr: true,
		},
		{
			name:    "non-numeric array element",
			input:   `["hello", 1.5]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var arr duckdb.FloatArray
			err := arr.Scan(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, arr)
		})
	}
}

func TestIntArray_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    duckdb.IntArray
		expected string
	}{
		{
			name:     "empty array",
			input:    duckdb.IntArray{},
			expected: "[]",
		},
		{
			name:     "single element",
			input:    duckdb.IntArray{42},
			expected: "[42]",
		},
		{
			name:     "multiple elements",
			input:    duckdb.IntArray{1, 2, 3},
			expected: "[1,2,3]",
		},
		{
			name:     "with zero and negative",
			input:    duckdb.IntArray{0, -5, 10},
			expected: "[0,-5,10]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.input.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestIntArray_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected duckdb.IntArray
		wantErr  bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "empty array string",
			input:    "[]",
			expected: duckdb.IntArray{},
			wantErr:  false,
		},
		{
			name:     "single element array",
			input:    "[42]",
			expected: duckdb.IntArray{42},
			wantErr:  false,
		},
		{
			name:     "multiple elements array",
			input:    "[1, 2, 3]",
			expected: duckdb.IntArray{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "byte slice input",
			input:    []byte("[10, 20]"),
			expected: duckdb.IntArray{10, 20},
			wantErr:  false,
		},
		{
			name:     "int slice input",
			input:    []int64{1, 2},
			expected: duckdb.IntArray{1, 2},
			wantErr:  false,
		},
		{
			name:     "interface slice input",
			input:    []interface{}{1, 2},
			expected: duckdb.IntArray{1, 2},
			wantErr:  false,
		},
		{
			name:    "invalid json",
			input:   "[invalid json",
			wantErr: true,
		},
		{
			name:    "non-numeric array element",
			input:   `["hello", 123]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var arr duckdb.IntArray
			err := arr.Scan(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, arr)
		})
	}
}

func TestMinimalArray_Value(t *testing.T) {
	// MinimalArray is not implemented - skipping these tests
	t.Skip("MinimalArray not implemented")
}

func TestMinimalArray_Scan(t *testing.T) {
	// MinimalArray is not implemented - skipping these tests
	t.Skip("MinimalArray not implemented")
}

func TestArrays_GormDataType(t *testing.T) {
	tests := []struct {
		name     string
		array    interface{ GormDataType() string }
		expected string
	}{
		{
			name:     "StringArray",
			array:    &duckdb.StringArray{},
			expected: "VARCHAR[]",
		},
		{
			name:     "FloatArray",
			array:    &duckdb.FloatArray{},
			expected: "DOUBLE[]",
		},
		{
			name:     "IntArray",
			array:    &duckdb.IntArray{},
			expected: "BIGINT[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataType := tt.array.GormDataType()
			assert.Equal(t, tt.expected, dataType)
		})
	}
}

func TestArrays_DatabaseIntegration(t *testing.T) {
	db := setupArrayTestDB(t)

	// Test data
	model := TestArrayModel{
		StringArr: duckdb.StringArray{"software", "analytics", "business"},
		FloatArr:  duckdb.FloatArray{4.5, 4.8, 4.2, 4.9},
		IntArr:    duckdb.IntArray{1250, 890, 2340, 567},
	}

	// Create record
	err := db.Create(&model).Error
	require.NoError(t, err)
	assert.NotZero(t, model.ID)

	// Retrieve record
	var retrieved TestArrayModel
	err = db.First(&retrieved, model.ID).Error
	require.NoError(t, err)

	// Verify arrays were stored and retrieved correctly
	assert.Equal(t, model.StringArr, retrieved.StringArr)
	assert.Equal(t, model.FloatArr, retrieved.FloatArr)
	assert.Equal(t, model.IntArr, retrieved.IntArr)

	// Test update
	retrieved.StringArr = append(retrieved.StringArr, "premium")
	retrieved.FloatArr = append(retrieved.FloatArr, 5.0)
	retrieved.IntArr = append(retrieved.IntArr, 1000)

	err = db.Save(&retrieved).Error
	require.NoError(t, err)

	// Verify update
	var updated TestArrayModel
	err = db.First(&updated, model.ID).Error
	require.NoError(t, err)

	assert.Equal(t, 4, len(updated.StringArr))
	assert.Equal(t, "premium", updated.StringArr[3])
	assert.Equal(t, 5, len(updated.FloatArr))
	assert.Equal(t, 5.0, updated.FloatArr[4])
	assert.Equal(t, 5, len(updated.IntArr))
	assert.Equal(t, int64(1000), updated.IntArr[4])
}

func TestArrays_EmptyAndNilHandling(t *testing.T) {
	db := setupArrayTestDB(t)

	// Test with empty arrays
	model := TestArrayModel{
		StringArr: duckdb.StringArray{},
		FloatArr:  duckdb.FloatArray{},
		IntArr:    duckdb.IntArray{},
	}

	err := db.Create(&model).Error
	require.NoError(t, err)

	var retrieved TestArrayModel
	err = db.First(&retrieved, model.ID).Error
	require.NoError(t, err)

	assert.Equal(t, 0, len(retrieved.StringArr))
	assert.Equal(t, 0, len(retrieved.FloatArr))
	assert.Equal(t, 0, len(retrieved.IntArr))

	// Test with nil arrays
	model2 := TestArrayModel{
		StringArr: nil,
		FloatArr:  nil,
		IntArr:    nil,
	}

	err = db.Create(&model2).Error
	require.NoError(t, err)

	var retrieved2 TestArrayModel
	err = db.First(&retrieved2, model2.ID).Error
	require.NoError(t, err)

	// Arrays should be nil after retrieval
	assert.Nil(t, retrieved2.StringArr)
	assert.Nil(t, retrieved2.FloatArr)
	assert.Nil(t, retrieved2.IntArr)
}

func TestArrays_ErrorCases(t *testing.T) {
	t.Run("StringArray invalid scan types", func(t *testing.T) {
		var arr duckdb.StringArray

		// Test unsupported type
		err := arr.Scan(123)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot scan")

		// Test invalid interface slice element
		err = arr.Scan([]interface{}{"valid", 123})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("FloatArray invalid scan types", func(t *testing.T) {
		var arr duckdb.FloatArray

		// Test unsupported type
		err := arr.Scan("not a number")
		assert.Error(t, err)

		// Test invalid interface slice element
		err = arr.Scan([]interface{}{1.5, "not a number"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("IntArray invalid scan types", func(t *testing.T) {
		var arr duckdb.IntArray

		// Test unsupported type
		err := arr.Scan("not a number")
		assert.Error(t, err)

		// Test invalid interface slice element
		err = arr.Scan([]interface{}{123, "not a number"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})
}

func TestArrays_DriverValueInterface(t *testing.T) {
	// Test that arrays implement driver.Valuer interface
	var _ driver.Valuer = (*duckdb.StringArray)(nil)
	var _ driver.Valuer = (*duckdb.FloatArray)(nil)
	var _ driver.Valuer = (*duckdb.IntArray)(nil)

	// Test that arrays implement sql.Scanner interface
	var _ interface{ Scan(interface{}) error } = (*duckdb.StringArray)(nil)
	var _ interface{ Scan(interface{}) error } = (*duckdb.FloatArray)(nil)
	var _ interface{ Scan(interface{}) error } = (*duckdb.IntArray)(nil)
}
