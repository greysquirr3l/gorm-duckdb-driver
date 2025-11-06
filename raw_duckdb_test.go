package duckdb_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/marcboeker/go-duckdb/v2"
	"github.com/stretchr/testify/require"
)

func TestRawDuckDBDriver(t *testing.T) {
	// Test the raw DuckDB driver directly
	sql.Register("test-duckdb", &duckdb.Driver{})
	
	db, err := sql.Open("test-duckdb", ":memory:")
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Failed to close database: %v", err)
		}
	}()

	// Create sequence and table (same pattern as GORM)
	_, err = db.ExecContext(context.Background(), "CREATE SEQUENCE IF NOT EXISTS seq_test_users_id START 1")
	require.NoError(t, err)
	
	_, err = db.ExecContext(context.Background(), `CREATE TABLE test_users (
		id INTEGER DEFAULT nextval('seq_test_users_id') PRIMARY KEY,
		name VARCHAR(100),
		email VARCHAR(255)
	)`)
	require.NoError(t, err)

	// Insert data
	result, err := db.ExecContext(context.Background(), "INSERT INTO test_users (name, email) VALUES (?, ?)", "John Doe", "john@example.com")
	require.NoError(t, err)
	
	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	t.Logf("Rows affected: %d", rowsAffected)

	// Try to get last insert ID
	lastID, err := result.LastInsertId()
	if err != nil {
		t.Logf("LastInsertId not supported: %v", err)
	} else {
		t.Logf("Last insert ID: %d", lastID)
	}

	// Query data back
	rows, err := db.QueryContext(context.Background(), "SELECT id, name, email FROM test_users")
	require.NoError(t, err)
	defer func() {
		if err := rows.Close(); err != nil {
			t.Logf("Failed to close rows: %v", err)
		}
	}()

	var users []struct {
		ID    int
		Name  string
		Email string
	}

	for rows.Next() {
		var user struct {
			ID    int
			Name  string
			Email string
		}
		err := rows.Scan(&user.ID, &user.Name, &user.Email)
		require.NoError(t, err)
		users = append(users, user)
	}

	require.NoError(t, rows.Err())
	t.Logf("Retrieved users: %+v", users)
	require.Len(t, users, 1)
	require.Equal(t, "John Doe", users[0].Name)
}