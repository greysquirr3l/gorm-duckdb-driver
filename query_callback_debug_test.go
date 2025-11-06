package duckdb

import (
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestQueryCallbackDebug(t *testing.T) {
	t.Log("=== Query Callback Debug Test ===")

	// Open database with maximum debug logging
	dialector := Dialector{
		Config: &Config{
			DSN: ":memory:",
		},
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Define our test model
	type User struct {
		ID       uint      `gorm:"primaryKey;autoIncrement" json:"id"`
		Name     string    `gorm:"type:VARCHAR(100);not null" json:"name"`
		Email    string    `gorm:"type:VARCHAR(255)" json:"email"`
		Age      int8      `gorm:"type:TINYINT" json:"age"`
		Birthday *time.Time `gorm:"type:TIMESTAMP" json:"birthday"`
	}

	// Auto-migrate to create the table
	err = db.AutoMigrate(&User{})
	if err != nil {
		t.Fatalf("Failed to auto-migrate: %v", err)
	}

	// Create a test record
	birthday := time.Date(1993, 1, 1, 0, 0, 0, 0, time.UTC)
	user := User{
		Name:     "Query Test User",
		Email:    "query@example.com",
		Age:      25,
		Birthday: &birthday,
	}

	result := db.Create(&user)
	t.Logf("Create result: Error=%v, RowsAffected=%d, User.ID=%d", result.Error, result.RowsAffected, user.ID)

	// Inspect the query callback registration
	t.Log("\n=== Callback Investigation ===")
	
	// Check if query callback exists
	queryCallback := db.Callback().Query()
	if queryCallback == nil {
		t.Log("Query callback processor is nil!")
	} else {
		t.Log("Query callback processor exists")
	}

	// Try to manually trigger GORM Find with detailed debugging
	t.Log("\n=== Manual GORM Find Call ===")
	
	var users []User
	
	// Try different approaches to see which one generates SQL
	t.Log("Attempting db.Find(&users)...")
	findResult := db.Find(&users)
	t.Logf("Find result: Error=%v, RowsAffected=%d, Count=%d", findResult.Error, findResult.RowsAffected, len(users))

	t.Log("Attempting db.Limit(10).Find(&users)...")
	users = []User{} // reset
	limitResult := db.Limit(10).Find(&users)
	t.Logf("Limit Find result: Error=%v, RowsAffected=%d, Count=%d", limitResult.Error, limitResult.RowsAffected, len(users))

	t.Log("Attempting db.Where(\"id > ?\", 0).Find(&users)...")
	users = []User{} // reset
	whereResult := db.Where("id > ?", 0).Find(&users)
	t.Logf("Where Find result: Error=%v, RowsAffected=%d, Count=%d", whereResult.Error, whereResult.RowsAffected, len(users))

	// Check if any specific method works
	t.Log("\n=== Testing Direct SQL Generation ===")
	
	// Test what SQL GORM would generate
	var dryRunUsers []User
	dryRunDB := db.Session(&gorm.Session{DryRun: true})
	
	t.Log("Testing DryRun mode to see SQL generation...")
	dryRunResult := dryRunDB.Find(&dryRunUsers)
	t.Logf("DryRun Find result: Error=%v, Statement.SQL=%s", dryRunResult.Error, dryRunResult.Statement.SQL.String())
	t.Logf("DryRun Statement.Vars=%v", dryRunResult.Statement.Vars)
}