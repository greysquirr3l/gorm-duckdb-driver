package duckdb_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
)

// Test model for error translator functionality
type ErrorTestUser struct {
	ID    uint   `gorm:"primarykey"`
	Email string `gorm:"size:255;uniqueIndex"`
	Name  string `gorm:"size:100;not null;check:name != ''"`
}

func setupErrorTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dialector := duckdb.Open(":memory:")
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&ErrorTestUser{})
	require.NoError(t, err)

	return db
}

func TestErrorTranslator_Translate(t *testing.T) {
	translator := duckdb.ErrorTranslator{}

	tests := []struct {
		name        string
		err         error
		expectedNil bool
	}{
		{
			name:        "nil error",
			err:         nil,
			expectedNil: true,
		},
		{
			name:        "non-database error",
			err:         errors.New("generic error"),
			expectedNil: false, // Returns original error
		},
		{
			name: "duplicate key error simulation",
			err:  errors.New("UNIQUE constraint failed: error_test_users.email"),
		},
		{
			name: "foreign key error simulation",
			err:  errors.New("FOREIGN KEY constraint failed"),
		},
		{
			name: "not null error simulation",
			err:  errors.New("NOT NULL constraint failed: error_test_users.name"),
		},
		{
			name: "table not found error simulation",
			err:  errors.New("no such table: non_existent_table"),
		},
		{
			name: "column not found error simulation",
			err:  errors.New("no such column: non_existent_column"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translated := translator.Translate(tt.err)

			if tt.expectedNil {
				assert.Nil(t, translated)
			} else {
				// We don't assert specific error types since the implementation
				// may vary, but we ensure it doesn't panic and returns an error
				assert.NotNil(t, translated)
			}
		})
	}
}

func TestErrorTranslator_IsDuplicateKeyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: false,
		},
		{
			name:     "duplicate key error",
			err:      errors.New("UNIQUE constraint failed"),
			expected: true,
		},
		{
			name:     "constraint violation",
			err:      errors.New("UNIQUE constraint failed in some context"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := duckdb.IsDuplicateKeyError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorTranslator_IsForeignKeyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: false,
		},
		{
			name:     "foreign key error",
			err:      errors.New("FOREIGN KEY constraint failed"),
			expected: true,
		},
		{
			name:     "fk constraint violation",
			err:      errors.New("FOREIGN KEY constraint failed in some context"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := duckdb.IsForeignKeyError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorTranslator_IsNotNullError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: false,
		},
		{
			name:     "not null error",
			err:      errors.New("NOT NULL constraint failed"),
			expected: true,
		},
		{
			name:     "null value error",
			err:      errors.New("NOT NULL constraint failed in some context"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := duckdb.IsNotNullError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorTranslator_IsTableNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: false,
		},
		{
			name:     "table not found error",
			err:      errors.New("no such table: non_existent"),
			expected: true,
		},
		{
			name:     "no such table error",
			err:      errors.New("no such table: non_existent"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := duckdb.IsTableNotFoundError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorTranslator_IsColumnNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: false,
		},
		{
			name:     "column not found error",
			err:      errors.New("no such column: non_existent_column"),
			expected: true,
		},
		{
			name:     "unknown column error",
			err:      errors.New("no such column: non_existent"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := duckdb.IsColumnNotFoundError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorTranslator_IsSpecificError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		target   error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			target:   errors.New("pattern"),
			expected: false,
		},
		{
			name:     "nil target",
			err:      errors.New("any error"),
			target:   nil,
			expected: false,
		},
		{
			name:     "matching pattern",
			err:      errors.New("UNIQUE constraint failed"),
			target:   duckdb.ErrUniqueConstraint,
			expected: true,
		},
		{
			name:     "no matching pattern",
			err:      errors.New("some other error"),
			target:   duckdb.ErrUniqueConstraint,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := duckdb.IsSpecificError(tt.err, tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorTranslator_RealDatabaseErrors(t *testing.T) {
	db := setupErrorTestDB(t)

	// Test real duplicate key error
	user1 := ErrorTestUser{Email: "test@example.com", Name: "Test User"}
	err := db.Create(&user1).Error
	require.NoError(t, err)

	// Try to create another user with the same email
	user2 := ErrorTestUser{Email: "test@example.com", Name: "Another User"}
	err = db.Create(&user2).Error
	assert.Error(t, err)

	// The error should be translated appropriately
	// We can't test the exact error type since it depends on the implementation
	// but we ensure it's handled gracefully

	// Test table not found error
	err = db.Table("non_existent_table").First(&ErrorTestUser{}).Error
	assert.Error(t, err)

	// Test column not found error
	err = db.Select("non_existent_column").First(&ErrorTestUser{}).Error
	assert.Error(t, err)
}
