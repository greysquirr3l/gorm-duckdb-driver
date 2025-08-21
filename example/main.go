package main

import (
	"fmt"
	"log"
	"time"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
	"gorm.io/gorm"
)

// User model demonstrating basic GORM features
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Email     string    `gorm:"size:255;uniqueIndex" json:"email"`
	Age       uint8     `json:"age"`
	Birthday  time.Time `json:"birthday"`
	CreatedAt time.Time `gorm:"autoCreateTime:false" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime:false" json:"updated_at"`
}

// Post model demonstrating simple relationships
type Post struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Tag model demonstrating auto-increment
type Tag struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"size:50;uniqueIndex" json:"name"`
}

// Product model demonstrating DuckDB array support
type Product struct {
	ID          uint               `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string             `gorm:"size:100;not null" json:"name"`
	Price       float64            `json:"price"`
	Description string             `json:"description"`
	Categories  duckdb.StringArray `json:"categories"`  // Array support
	Scores      duckdb.FloatArray  `json:"scores"`      // Float array support
	ViewCounts  duckdb.IntArray    `json:"view_counts"` // Int array support
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func main() {
	fmt.Println("🦆 GORM DuckDB Driver - Comprehensive Example")
	fmt.Println("==============================================")
	fmt.Println("This example demonstrates:")
	fmt.Println("• Arrays (StringArray, FloatArray, IntArray)")
	fmt.Println("• Migrations and auto-increment with sequences")
	fmt.Println("• Time handling and various data types")
	fmt.Println("• ALTER TABLE fixes for DuckDB syntax")
	fmt.Println("• Basic CRUD operations")
	fmt.Println("")

	// Initialize database (use in-memory for clean runs)
	db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("✅ Connected to DuckDB (in-memory)")

	// Migrate the schema
	fmt.Println("🔧 Auto-migrating database schema...")
	err = db.AutoMigrate(&User{}, &Post{}, &Tag{}, &Product{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	fmt.Println("✅ Schema migration completed")

	// Demonstrate core features
	demonstrateBasicCRUD(db)
	demonstrateArrayFeatures(db)
	demonstrateAdvancedQueries(db)

	fmt.Println("\n🎉 Example completed successfully!")
	fmt.Println("📝 Note: Using in-memory database - data will be cleaned up automatically")
}

func demonstrateBasicCRUD(db *gorm.DB) {
	fmt.Println("\n📝 Basic CRUD Operations")
	fmt.Println("------------------------")
	fmt.Println("Demonstrating: Create, Read, Update, Delete operations")
	fmt.Println("Features: Auto-increment IDs, manual timestamps, unique constraints")

	// Create sample users with manual timestamps
	now := time.Now()
	birthday := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	users := []User{
		{
			Name:      "Alice Johnson",
			Email:     "alice@example.com",
			Age:       25,
			Birthday:  birthday,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Name:      "Bob Smith",
			Email:     "bob@example.com",
			Age:       30,
			Birthday:  time.Time{}, // Zero time for no birthday
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Name:      "Charlie Brown",
			Email:     "charlie@example.com",
			Age:       35,
			Birthday:  time.Time{}, // Zero time for no birthday
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// Create users individually to demonstrate auto-increment
	fmt.Printf("Creating %d users...\n", len(users))
	for i, user := range users {
		result := db.Create(&user)
		if result.Error != nil {
			log.Printf("Error creating user %d: %v", i+1, result.Error)
			continue
		}
		users[i] = user // Update with generated ID
		fmt.Printf("  ✅ Created: %s (ID: %d)\n", user.Name, user.ID)
	}

	// Read operations
	var allUsers []User
	db.Find(&allUsers)
	fmt.Printf("\n👥 Found %d users in database:\n", len(allUsers))

	// Show basic user info
	for _, user := range allUsers {
		fmt.Printf("  • %s (Age: %d, Email: %s)\n", user.Name, user.Age, user.Email)
	}

	// Update operation
	if len(users) > 0 {
		result := db.Model(&users[0]).Update("age", 26)
		if result.Error != nil {
			log.Printf("Error updating user: %v", result.Error)
		} else {
			fmt.Printf("\n✏️ Updated %s's age to 26\n", users[0].Name)
		}
	}

	// Delete operation (soft delete if applicable)
	if len(users) > 2 {
		result := db.Delete(&users[2])
		if result.Error != nil {
			log.Printf("Error deleting user: %v", result.Error)
		} else {
			fmt.Printf("🗑️ Deleted user: %s\n", users[2].Name)
		}
	}

	// Verify final count
	var finalCount int64
	db.Model(&User{}).Count(&finalCount)
	fmt.Printf("📊 Final user count: %d\n", finalCount)
}

func demonstrateArrayFeatures(db *gorm.DB) {
	fmt.Println("\n🎨 Array Features Demonstration")
	fmt.Println("-------------------------------")
	fmt.Println("Demonstrating: StringArray, FloatArray, IntArray support")
	fmt.Println("Features: Array creation, retrieval, and updates")

	// Create products with arrays
	now := time.Now()
	products := []Product{
		{
			Name:        "Analytics Software",
			Price:       299.99,
			Description: "Advanced data analytics platform",
			Categories:  duckdb.StringArray{"software", "analytics", "business"},
			Scores:      duckdb.FloatArray{4.5, 4.8, 4.2, 4.9},
			ViewCounts:  duckdb.IntArray{1250, 890, 2340, 567},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "Gaming Laptop",
			Price:       1299.99,
			Description: "High-performance gaming laptop",
			Categories:  duckdb.StringArray{"electronics", "computers", "gaming"},
			Scores:      duckdb.FloatArray{4.7, 4.9, 4.6},
			ViewCounts:  duckdb.IntArray{3200, 2100, 4500},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// Create products individually
	fmt.Printf("Creating %d products with arrays...\n", len(products))
	for i, product := range products {
		result := db.Create(&product)
		if result.Error != nil {
			log.Printf("Error creating product %d: %v", i+1, result.Error)
			continue
		}
		products[i] = product // Update with generated ID
		fmt.Printf("  ✅ Created: %s (ID: %d)\n", product.Name, product.ID)
	}

	// Retrieve and display arrays
	var retrievedProducts []Product
	db.Find(&retrievedProducts)

	fmt.Printf("\n📦 Products with array data:\n")
	for _, product := range retrievedProducts {
		fmt.Printf("\n• %s ($%.2f)\n", product.Name, product.Price)
		fmt.Printf("  Categories: %v\n", []string(product.Categories))
		fmt.Printf("  Scores: %v\n", []float64(product.Scores))
		fmt.Printf("  View Counts: %v\n", []int64(product.ViewCounts))
	}

	// Update arrays
	if len(retrievedProducts) > 0 {
		product := &retrievedProducts[0]
		originalCategories := len(product.Categories)

		// Add new elements to arrays
		product.Categories = append(product.Categories, "premium")
		product.Scores = append(product.Scores, 5.0)
		product.ViewCounts = append(product.ViewCounts, 1000)

		updateResult := db.Save(product)
		if updateResult.Error != nil {
			log.Printf("Error updating product arrays: %v", updateResult.Error)
		} else {
			fmt.Printf("\n✏️ Updated arrays for: %s\n", product.Name)
			fmt.Printf("  Categories: %d → %d elements: %v\n",
				originalCategories, len(product.Categories), []string(product.Categories))
		}
	}

	// Final count
	var productCount int64
	db.Model(&Product{}).Count(&productCount)
	fmt.Printf("\n📊 Total products: %d\n", productCount)
}

func demonstrateAdvancedQueries(db *gorm.DB) {
	fmt.Println("\n� Advanced Queries and Features")
	fmt.Println("--------------------------------")
	fmt.Println("Demonstrating: Complex queries, aggregations, transactions")

	// Create some tags for demonstration
	tags := []Tag{
		{Name: "go"},
		{Name: "database"},
		{Name: "tutorial"},
		{Name: "example"},
	}

	fmt.Printf("Creating %d tags...\n", len(tags))
	for i := range tags {
		result := db.Create(&tags[i])
		if result.Error != nil {
			log.Printf("Error creating tag %s: %v", tags[i].Name, result.Error)
			continue
		}
		fmt.Printf("  ✅ Created tag: %s (ID: %d)\n", tags[i].Name, tags[i].ID)
	}

	// Demonstrate analytical queries on products
	fmt.Println("\n💰 Price Analysis:")

	var expensiveProducts []Product
	db.Where("price > ?", 500.0).Find(&expensiveProducts)
	fmt.Printf("  • Found %d products over $500\n", len(expensiveProducts))

	// Calculate average price
	var avgPrice float64
	err := db.Model(&Product{}).Select("AVG(price)").Row().Scan(&avgPrice)
	if err != nil {
		log.Printf("Error calculating average price: %v", err)
		avgPrice = 0
	}
	fmt.Printf("  • Average product price: $%.2f\n", avgPrice)

	// Count by age groups
	fmt.Println("\n👥 User Demographics:")
	type UserStat struct {
		AgeGroup string
		Count    int64
	}

	var userStats []UserStat
	db.Model(&User{}).
		Select("CASE WHEN age < 30 THEN 'Young' ELSE 'Mature' END as age_group, COUNT(*) as count").
		Group("age_group").
		Scan(&userStats)

	for _, stat := range userStats {
		fmt.Printf("  • %s: %d users\n", stat.AgeGroup, stat.Count)
	}

	// Demonstrate transaction
	fmt.Println("\n💳 Transaction Example:")

	err = db.Transaction(func(tx *gorm.DB) error {
		// Create a post within transaction
		post := Post{
			Title:   "Transaction Test Post",
			Content: "This post was created within a database transaction",
			UserID:  1, // Assuming user ID 1 exists
		}

		if err := tx.Create(&post).Error; err != nil {
			return err // This will rollback the transaction
		}

		fmt.Printf("  ✅ Created post in transaction: %s (ID: %d)\n", post.Title, post.ID)
		return nil
	})

	if err != nil {
		fmt.Printf("  ❌ Transaction failed: %v\n", err)
	} else {
		fmt.Println("  ✅ Transaction completed successfully")
	}

	// Final database state
	var userCount, postCount, tagCount, productCount int64
	db.Model(&User{}).Count(&userCount)
	db.Model(&Post{}).Count(&postCount)
	db.Model(&Tag{}).Count(&tagCount)
	db.Model(&Product{}).Count(&productCount)

	fmt.Printf("\n📊 Final Database State:\n")
	fmt.Printf("  • Users: %d\n", userCount)
	fmt.Printf("  • Posts: %d\n", postCount)
	fmt.Printf("  • Tags: %d\n", tagCount)
	fmt.Printf("  • Products: %d\n", productCount)
}
