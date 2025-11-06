package duckdb

import (
	"os"
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Test to debug what SQL GORM is generating for Find/First operations
func TestGORMSelectSQL(t *testing.T) {
	t.Log("=== GORM SELECT SQL Debug Test ===")

	// Enable debug mode
	if err := os.Setenv("GORM_DUCKDB_DEBUG", "1"); err != nil {
		t.Fatalf("Failed to set debug environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("GORM_DUCKDB_DEBUG"); err != nil {
			t.Logf("Failed to unset debug environment variable: %v", err)
		}
	}()

	dialector := Dialector{
		Config: &Config{
			DSN: ":memory:",
		},
	}

	// Use GORM's Info level logging to see generated SQL
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatalf("Failed to open DuckDB: %v", err)
	}

	type User struct {
		ID       uint      `gorm:"primaryKey;autoIncrement"`
		Name     string    `gorm:"size:100;not null"`
		Email    string    `gorm:"size:255"`
		Age      uint8
		Birthday time.Time
	}

	// Migration
	err = db.AutoMigrate(&User{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Create a user first
	user := User{
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      30,
		Birthday: time.Date(1993, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	result := db.Create(&user)
	t.Logf("Create result: Error=%v, RowsAffected=%d, User.ID=%d", 
		result.Error, result.RowsAffected, user.ID)

	// Verify the data exists with raw SQL
	var count int64
	db.Raw("SELECT COUNT(*) FROM users").Scan(&count)
	t.Logf("Raw count query result: %d rows in users table", count)

	// Test GORM Find - this will show the generated SQL in logs
	t.Log("\n=== Testing GORM Find (watch for generated SQL) ===")
	var users []User
	result = db.Find(&users)
	t.Logf("Find result: Error=%v, RowsAffected=%d, Count=%d", 
		result.Error, result.RowsAffected, len(users))

	// Test GORM First - this will show the generated SQL in logs
	t.Log("\n=== Testing GORM First (watch for generated SQL) ===")
	var foundUser User
	result = db.First(&foundUser)
	t.Logf("First result: Error=%v, RowsAffected=%d", result.Error, result.RowsAffected)
	t.Logf("Found user: ID=%d, Name=%s, Email=%s", foundUser.ID, foundUser.Name, foundUser.Email)

	// Test GORM First with ID - this will show the generated SQL in logs
	t.Log("\n=== Testing GORM First by ID (watch for generated SQL) ===")
	var foundByID User
	result = db.First(&foundByID, user.ID)
	t.Logf("First by ID result: Error=%v, RowsAffected=%d", result.Error, result.RowsAffected)
	t.Logf("Found by ID: ID=%d, Name=%s, Email=%s", foundByID.ID, foundByID.Name, foundByID.Email)

	// Test to see if the issue is with result scanning
	t.Log("\n=== Testing Raw SQL with GORM Scan ===")
	var scanUser User
	result = db.Raw("SELECT id, name, email, age, birthday FROM users LIMIT 1").Scan(&scanUser)
	t.Logf("Raw scan result: Error=%v, RowsAffected=%d", result.Error, result.RowsAffected)
	t.Logf("Scanned user: ID=%d, Name=%s, Email=%s", scanUser.ID, scanUser.Name, scanUser.Email)
}