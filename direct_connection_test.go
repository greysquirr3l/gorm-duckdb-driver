package duckdb_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestConnectionDirect(t *testing.T) {
	// Test if our driver registration works
	db, err := sql.Open("duckdb-gorm", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Test basic query
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	require.NoError(t, err)
	require.Equal(t, 1, result)
	
	t.Log("Direct SQL connection works")

	// Now test GORM with explicit connection
	gormDB, err := gorm.Open(&duckdb.Dialector{
		Config: &duckdb.Config{
			Conn: db, // Use the working connection directly
		},
	}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	require.NoError(t, err)

	// Simple create test
	type DirectUser struct {
		ID   uint   `gorm:"primaryKey"`
		Name string
	}

	err = gormDB.AutoMigrate(&DirectUser{})
	require.NoError(t, err)

	user := DirectUser{Name: "Direct Test"}
	result2 := gormDB.Create(&user)
	
	t.Logf("Direct GORM Create result: Error=%v, RowsAffected=%d", result2.Error, result2.RowsAffected)
	t.Logf("Direct User after create: %+v", user)
}