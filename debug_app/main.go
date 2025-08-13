//go:build ignore

package main

import (
	"fmt"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
)

type TestUser struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"size:100;not null"`
}

func main() {
	db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Check what SQL is generated
	fmt.Println("Migrating schema...")
	err = db.AutoMigrate(&TestUser{})
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("Creating user...")
	user := TestUser{Name: "Test"}
	err = db.Create(&user).Error
	if err != nil {
		log.Fatal("Create failed:", err)
	}

	fmt.Printf("Created user with ID: %d\n", user.ID)
}
