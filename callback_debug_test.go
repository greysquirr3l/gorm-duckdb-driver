package duckdb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCallbacksRegistered(t *testing.T) {
	// Setup database
	dialector := duckdb.OpenWithRowCallbackWorkaround(":memory:", false)
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Check what create callbacks are registered
	createProcessor := db.Callback().Create()
	t.Logf("Create processor: %+v", createProcessor)

	// Check what row callbacks are registered
	rowProcessor := db.Callback().Row()
	t.Logf("Row processor: %+v", rowProcessor)
}