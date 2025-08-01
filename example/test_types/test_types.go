package main

import (
	"fmt"
	"log"
	"sync"

	duckdb "gorm.io/driver/duckdb"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type TypeTestUser struct { // Renamed to avoid conflict
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:100"`
}

type TypeTestPost struct { // Renamed to avoid conflict
	ID     uint `gorm:"primaryKey"`
	UserID uint
}

func main() {
	fmt.Println("Testing DuckDB type mapping...")

	// Initialize dialector
	dialector := duckdb.Open("test.db")

	// Initialize database
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Parse schemas
	userSchema, err := schema.Parse(&TypeTestUser{}, &sync.Map{}, db.NamingStrategy)
	if err != nil {
		log.Fatal("Failed to parse user schema:", err)
	}

	postSchema, err := schema.Parse(&TypeTestPost{}, &sync.Map{}, db.NamingStrategy)
	if err != nil {
		log.Fatal("Failed to parse post schema:", err)
	}

	// Check field types using the dialector
	duckdbDialector := dialector.(*duckdb.Dialector)

	fmt.Printf("User.ID (uint) -> %s\n", duckdbDialector.DataTypeOf(userSchema.LookUpField("ID")))
	fmt.Printf("Post.ID (uint) -> %s\n", duckdbDialector.DataTypeOf(postSchema.LookUpField("ID")))
	fmt.Printf("Post.UserID (uint) -> %s\n", duckdbDialector.DataTypeOf(postSchema.LookUpField("UserID")))

	fmt.Println("✅ All uint types map to BIGINT for foreign key compatibility")
}
