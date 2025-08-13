//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/marcboeker/go-duckdb/v2"
)

func main() {
	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create sequence and table
	_, err = db.Exec("CREATE SEQUENCE seq_test_id START 1")
	if err != nil {
		log.Fatal("Create sequence:", err)
	}

	_, err = db.Exec("CREATE TABLE test (id BIGINT DEFAULT nextval('seq_test_id') NOT NULL PRIMARY KEY, name TEXT)")
	if err != nil {
		log.Fatal("Create table:", err)
	}

	// Test 1: Insert and get last insert ID
	fmt.Println("=== Test 1: Insert with LastInsertId ===")
	result, err := db.Exec("INSERT INTO test (name) VALUES (?)", "Test1")
	if err != nil {
		log.Fatal("Insert:", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("LastInsertId error: %v\n", err)
	} else {
		fmt.Printf("LastInsertId: %d\n", lastID)
	}

	// Test 2: Insert with RETURNING
	fmt.Println("\n=== Test 2: Insert with RETURNING ===")
	var returnedID int64
	err = db.QueryRow("INSERT INTO test (name) VALUES (?) RETURNING id", "Test2").Scan(&returnedID)
	if err != nil {
		fmt.Printf("RETURNING error: %v\n", err)
	} else {
		fmt.Printf("RETURNING ID: %d\n", returnedID)
	}

	// Test 3: Check what's actually in the table
	fmt.Println("\n=== Test 3: Current table contents ===")
	rows, err := db.Query("SELECT id, name FROM test ORDER BY id")
	if err != nil {
		log.Fatal("Query:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatal("Scan:", err)
		}
		fmt.Printf("ID: %d, Name: %s\n", id, name)
	}
}
