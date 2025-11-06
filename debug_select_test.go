package duckdb

import (
	"os"
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Test to debug SELECT operations after CREATE works
func TestSelectDebug(t *testing.T) {
	t.Log("=== SELECT Debug Test ===")

	// Enable debug mode
	os.Setenv("GORM_DUCKDB_DEBUG", "1")
	defer os.Unsetenv("GORM_DUCKDB_DEBUG")

	dialector := Dialector{
		Config: &Config{
			DSN: ":memory:",
		},
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatalf("Failed to open DuckDB: %v", err)
	}

	type User struct {
		ID       uint   `gorm:"primaryKey;autoIncrement"`
		Name     string `gorm:"size:100;not null"`
		Email    string `gorm:"size:255"`
		Age      uint8
		Birthday time.Time
	}

	// Migration
	err = db.AutoMigrate(&User{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Create a user (we know this works now)
	user := User{
		Name:     "Debug User",
		Email:    "debug@example.com",
		Age:      25,
		Birthday: time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	t.Log("=== Creating User ===")
	result := db.Create(&user)
	t.Logf("Create result: Error=%v, RowsAffected=%d, User.ID=%d",
		result.Error, result.RowsAffected, user.ID)

	// Test direct SQL query first
	t.Log("=== Testing Raw SQL Query ===")
	rows, err := db.Raw("SELECT * FROM users").Rows()
	if err != nil {
		t.Logf("Raw SQL failed: %v", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
			var id uint
			var name, email string
			var age uint8
			var birthday time.Time
			err := rows.Scan(&id, &name, &email, &age, &birthday)
			if err != nil {
				t.Logf("Row scan failed: %v", err)
			} else {
				t.Logf("Raw query row %d: ID=%d, Name=%s, Email=%s, Age=%d",
					count, id, name, email, age)
			}
		}
		t.Logf("Total rows from raw query: %d", count)
	}

	// Test GORM Find
	t.Log("=== Testing GORM Find ===")
	var users []User
	result = db.Find(&users)
	t.Logf("Find result: Error=%v, RowsAffected=%d, Count=%d",
		result.Error, result.RowsAffected, len(users))
	for i, u := range users {
		t.Logf("Find row %d: ID=%d, Name=%s, Email=%s, Age=%d",
			i+1, u.ID, u.Name, u.Email, u.Age)
	}

	// Test GORM First with specific ID
	t.Log("=== Testing GORM First by ID ===")
	var foundUser User
	result = db.First(&foundUser, user.ID)
	t.Logf("First result: Error=%v, RowsAffected=%d", result.Error, result.RowsAffected)
	t.Logf("Found user: ID=%d, Name=%s, Email=%s, Age=%d",
		foundUser.ID, foundUser.Name, foundUser.Email, foundUser.Age)

	// Test GORM Where clause
	t.Log("=== Testing GORM Where ===")
	var whereUser User
	result = db.Where("name = ?", "Debug User").First(&whereUser)
	t.Logf("Where result: Error=%v, RowsAffected=%d", result.Error, result.RowsAffected)
	t.Logf("Where user: ID=%d, Name=%s, Email=%s, Age=%d",
		whereUser.ID, whereUser.Name, whereUser.Email, whereUser.Age)
}
