package main

import (
	"fmt"
	"log"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
	"gorm.io/gorm"
)

func main() {
	// Open GORM database connection
	gormDB, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	
	db, err := gormDB.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table with array column
	_, err = db.Exec(`CREATE TABLE test_array (
		id INTEGER PRIMARY KEY,
		str_arr TEXT[]
	)`)
	if err != nil {
		log.Fatal("Create table:", err)
	}

	// Insert array data using different formats
	formats := []string{
		"['software', 'analytics', 'business']",
		"[\"software\", \"analytics\", \"business\"]",
	}

	for i, format := range formats {
		_, err = db.Exec(fmt.Sprintf("INSERT INTO test_array (id, str_arr) VALUES (%d, %s)", i+1, format))
		if err != nil {
			log.Printf("Insert format %s failed: %v", format, err)
		}
	}

	// Query back and see what we get
	rows, err := db.Query("SELECT id, str_arr FROM test_array")
	if err != nil {
		log.Fatal("Query:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var strArr interface{}
		
		err := rows.Scan(&id, &strArr)
		if err != nil {
			log.Fatal("Scan:", err)
		}
		
		fmt.Printf("ID: %d, Array: %v (type: %T)\n", id, strArr, strArr)
		
		// Try as string and []byte
		if s, ok := strArr.(string); ok {
			fmt.Printf("  As string: %q\n", s)
		}
		if b, ok := strArr.([]byte); ok {
			fmt.Printf("  As []byte: %q\n", string(b))
		}
	}
}