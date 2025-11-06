package duckdb_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite" // Use SQLite to compare
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQLiteTestUser struct {
	ID   uint   `gorm:"primaryKey"`
	Name string
}

func TestGORMWithSQLite(t *testing.T) {
	// Test the same operation with SQLite to verify GORM behavior
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	require.NoError(t, err)

	// Migrate
	err = db.AutoMigrate(&SQLiteTestUser{})
	require.NoError(t, err)

	t.Log("=== Starting SQLite Create operation ===")
	
	// Create
	user := SQLiteTestUser{Name: "Test User"}
	result := db.Create(&user)
	
	t.Logf("SQLite Create result: Error=%v, RowsAffected=%d", result.Error, result.RowsAffected)
	t.Logf("SQLite User after create: %+v", user)
	
	require.NoError(t, result.Error)
	require.Equal(t, int64(1), result.RowsAffected)
	require.NotZero(t, user.ID)
}