package duckdb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SimpleUser struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"not null"`
	Email string
	Age   int
}

func TestSimpleCreate(t *testing.T) {
	// Setup database
	dialector := duckdb.Open(":memory:")
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		DryRun: false,
	})
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&SimpleUser{})
	require.NoError(t, err)

	t.Log("=== Starting Create operation ===")
	
	// Create a simple user
	user := SimpleUser{
		Name:  "Test User",
		Email: "test@example.com", 
		Age:   25,
	}
	
	result := db.Create(&user)
	t.Logf("Create result: Error=%v, RowsAffected=%d", result.Error, result.RowsAffected)
	t.Logf("User after create: %+v", user)
	
	require.NoError(t, result.Error)
}