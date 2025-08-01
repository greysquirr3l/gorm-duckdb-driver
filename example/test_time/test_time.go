package main

import (
	"fmt"
	"log"
	"time"

	duckdb "gorm.io/driver/duckdb"
	"gorm.io/gorm"
)

type TimeTestUser struct { // Renamed to avoid conflict
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:100"`
	CreatedAt time.Time `gorm:"autoCreateTime:false"` // Disable auto-create to control manually
}

func main() {
	fmt.Println("Testing time.Time handling...")

	// Initialize database
	db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Migrate
	err = db.AutoMigrate(&TimeTestUser{})
	if err != nil {
		log.Fatal("Failed to migrate:", err)
	}

	// Try to create a simple user with manual timestamp
	now := time.Now()
	user := TimeTestUser{
		ID:        1,
		Name:      "Test User",
		CreatedAt: now, // Set manually
	}

	result := db.Create(&user)
	if result.Error != nil {
		log.Printf("Error: %v", result.Error)
	} else {
		fmt.Printf("✅ Successfully created user: %+v\n", user)
	}
}
