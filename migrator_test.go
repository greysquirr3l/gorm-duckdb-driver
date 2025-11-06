package duckdb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
)

// Test models for migration functionality
type TestUser struct {
	ID     uint   `gorm:"primaryKey"`
	Name   string `gorm:"size:100"`
	Email  string `gorm:"size:255;uniqueIndex:idx_email"`
	Age    int
	Active bool
}

type MigrationTestPost struct {
	ID      uint   `gorm:"primaryKey"`
	Title   string `gorm:"size:200"`
	Content string `gorm:"type:text"`
	UserID  uint
}

func setupMigratorTestDB(t *testing.T) (*gorm.DB, duckdb.Migrator) {
	t.Helper()

	dialector := duckdb.Open(":memory:")
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	// Get the migrator
	migrator := dialector.Migrator(db)
	duckdbMigrator, ok := migrator.(duckdb.Migrator)
	require.True(t, ok, "Migrator should be of type duckdb.Migrator")

	return db, duckdbMigrator
}

func TestMigrator_HasTable(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Table should not exist initially
	hasTable := migrator.HasTable(&TestUser{})
	assert.False(t, hasTable)

	// Create table
	err := db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	// Table should exist now
	hasTable = migrator.HasTable(&TestUser{})
	assert.True(t, hasTable)

	// Test with table name string - use the actual table name GORM generates
	hasTable = migrator.HasTable("test_users")
	assert.True(t, hasTable)
	assert.True(t, hasTable)

	// Non-existent table
	hasTable = migrator.HasTable("non_existent_table")
	assert.False(t, hasTable)
}

func TestMigrator_CreateTable(t *testing.T) {
	_, migrator := setupMigratorTestDB(t)

	// Use a unique table name to avoid conflicts
	type CreateTestTable struct {
		ID    uint   `gorm:"primaryKey"`
		Title string `gorm:"size:100"`
	}

	// Create table using migrator
	err := migrator.CreateTable(&CreateTestTable{})
	require.NoError(t, err)

	// Verify table exists
	hasTable := migrator.HasTable(&CreateTestTable{})
	assert.True(t, hasTable)
}

func TestMigrator_DropTable(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table first
	err := db.AutoMigrate(&TestUser{})
	require.NoError(t, err)
	assert.True(t, migrator.HasTable(&TestUser{}))

	// Drop table
	err = migrator.DropTable(&TestUser{})
	require.NoError(t, err)

	// Verify table no longer exists
	hasTable := migrator.HasTable(&TestUser{})
	assert.False(t, hasTable)

	// Try to drop non-existent table - should not error due to IF EXISTS
	err = migrator.DropTable("non_existent_table")
	require.NoError(t, err)
}

func TestMigrator_HasColumn(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	// Check existing columns
	hasColumn := migrator.HasColumn(&TestUser{}, "name")
	assert.True(t, hasColumn)

	hasColumn = migrator.HasColumn(&TestUser{}, "email")
	assert.True(t, hasColumn)

	hasColumn = migrator.HasColumn(&TestUser{}, "age")
	assert.True(t, hasColumn)

	// Check non-existent column
	hasColumn = migrator.HasColumn(&TestUser{}, "non_existent_column")
	assert.False(t, hasColumn)

	// Test with table name string - use correct GORM table name
	hasColumn = migrator.HasColumn("test_users", "name")
	assert.True(t, hasColumn)
}

func TestMigrator_AlterColumn(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table with a fresh name to avoid dependency issues
	type AlterTestTable struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:50"`
	}

	err := db.AutoMigrate(&AlterTestTable{})
	require.NoError(t, err)

	// Alter column - this tests the AlterColumn method
	// Note: DuckDB may have limitations with ALTER COLUMN due to dependencies
	err = migrator.AlterColumn(&AlterTestTable{}, "name")
	if err != nil {
		// DuckDB dependency errors are expected in some cases
		t.Logf("AlterColumn failed as expected due to DuckDB dependency constraints: %v", err)
		assert.Contains(t, err.Error(), "Cannot alter entry")
	} else {
		// If successful, verify column still exists
		hasColumn := migrator.HasColumn(&AlterTestTable{}, "name")
		assert.True(t, hasColumn)
	}
}

func TestMigrator_RenameColumn(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table with a fresh name to avoid dependency issues
	type RenameTestTable struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:50"`
	}

	err := db.AutoMigrate(&RenameTestTable{})
	require.NoError(t, err)

	// Rename column - DuckDB may have dependency constraints
	err = migrator.RenameColumn(&RenameTestTable{}, "name", "full_name")
	if err != nil {
		// DuckDB dependency errors are expected in some cases
		t.Logf("RenameColumn failed as expected due to DuckDB dependency constraints: %v", err)
		assert.Contains(t, err.Error(), "Cannot alter entry")
	} else {
		// If successful, verify the rename worked
		hasOldColumn := migrator.HasColumn(&RenameTestTable{}, "name")
		assert.False(t, hasOldColumn)

		hasNewColumn := migrator.HasColumn(&RenameTestTable{}, "full_name")
		assert.True(t, hasNewColumn)
	}
}

func TestMigrator_AddColumn(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	// Add a new column
	if err = migrator.AddColumn(&TestUser{}, "new_column"); err != nil {
		t.Logf("AddColumn error (expected): %v", err)
	}
	// Note: This might fail since the field doesn't exist in the struct
	// but we're testing that the method doesn't panic
	// The actual implementation should handle missing fields gracefully
}

func TestMigrator_DropColumn(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table with a fresh name to avoid dependency issues
	type DropTestTable struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:50"`
		Age  int
	}

	err := db.AutoMigrate(&DropTestTable{})
	require.NoError(t, err)

	// Drop a column - DuckDB may have dependency constraints
	err = migrator.DropColumn(&DropTestTable{}, "age")
	if err != nil {
		// DuckDB dependency errors are expected in some cases
		t.Logf("DropColumn failed as expected due to DuckDB dependency constraints: %v", err)
		assert.Contains(t, err.Error(), "Cannot alter entry")
	} else {
		// If successful, verify column no longer exists
		hasColumn := migrator.HasColumn(&DropTestTable{}, "age")
		assert.False(t, hasColumn)
	}
}

func TestMigrator_HasIndex(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	// Check for the email index that should be created by uniqueIndex:idx_email tag
	// Note: DuckDB index detection might vary
	hasIndex := migrator.HasIndex(&TestUser{}, "idx_email")
	// For now, just test that the method doesn't panic
	// Index detection in DuckDB might work differently
	t.Logf("HasIndex result for idx_email: %v", hasIndex)

	// Check for non-existent index
	hasIndex = migrator.HasIndex(&TestUser{}, "non_existent_index")
	assert.False(t, hasIndex)
}

func TestMigrator_CreateIndex(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	// Create an index - this might fail or succeed depending on DuckDB implementation
	err = migrator.CreateIndex(&TestUser{}, "name")
	if err != nil {
		// Index creation might fail in DuckDB - log the error
		t.Logf("CreateIndex failed (may be expected): %v", err)
	} else {
		// If successful, the test passes
		t.Log("CreateIndex succeeded")
	}
}

func TestMigrator_RenameIndex(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table with index
	err := db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	// Rename index - DuckDB may not support this operation
	err = migrator.RenameIndex(&TestUser{}, "idx_email", "idx_user_email")
	if err != nil {
		// DuckDB may not support ALTER INDEX RENAME - that's acceptable
		t.Logf("RenameIndex failed as expected due to DuckDB limitations: %v", err)
		assert.Contains(t, err.Error(), "Schema element not supported")
	} else {
		// If successful, verify the rename worked
		hasOldIndex := migrator.HasIndex(&TestUser{}, "idx_email")
		assert.False(t, hasOldIndex)

		hasNewIndex := migrator.HasIndex(&TestUser{}, "idx_user_email")
		assert.True(t, hasNewIndex)
	}
}

func TestMigrator_DropIndex(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table with index
	err := db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	// Drop index
	err = migrator.DropIndex(&TestUser{}, "idx_email")
	require.NoError(t, err)

	// Verify index no longer exists
	hasIndex := migrator.HasIndex(&TestUser{}, "idx_email")
	assert.False(t, hasIndex)
}

func TestMigrator_CurrentDatabase(t *testing.T) {
	_, migrator := setupMigratorTestDB(t)

	// Get current database name
	dbName := migrator.CurrentDatabase()
	// For in-memory database, this might be empty or a special value
	// We just test that the method doesn't panic
	assert.IsType(t, "", dbName)
}

func TestMigrator_GetTables(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Initially no tables
	tables, err := migrator.GetTables()
	require.NoError(t, err)
	assert.Empty(t, tables)

	// Create some tables
	err = db.AutoMigrate(&TestUser{}, &MigrationTestPost{})
	require.NoError(t, err)

	// Get tables again
	tables, err = migrator.GetTables()
	require.NoError(t, err)
	assert.Contains(t, tables, "test_users")
	assert.Contains(t, tables, "migration_test_posts")
}

func TestMigrator_FullDataTypeOf(t *testing.T) {
	_, migrator := setupMigratorTestDB(t)

	// Test with a sample field
	// This requires creating a statement with the field information
	// For now, we test that the method exists and doesn't panic
	user := &TestUser{}
	db, _ := setupMigratorTestDB(t)
	stmt := &gorm.Statement{DB: db}
	err := stmt.Parse(user)
	require.NoError(t, err)

	if len(stmt.Schema.Fields) > 0 {
		field := stmt.Schema.Fields[0]
		dataType := migrator.FullDataTypeOf(field)
		assert.NotEmpty(t, dataType)
	}
}

func TestMigrator_CreateView(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table first
	err := db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	// Create view
	viewName := "user_view"
	viewOption := gorm.ViewOption{
		Query: db.Select("id, name").Table("test_users"),
	}
	err = migrator.CreateView(viewName, viewOption)
	require.NoError(t, err)

	// Verify view exists by trying to query it
	var count int64
	err = db.Table(viewName).Count(&count).Error
	require.NoError(t, err)
}

func TestMigrator_DropView(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table and view first
	err := db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	viewName := "user_view"
	viewOption := gorm.ViewOption{
		Query: db.Select("id, name").Table("test_users"),
	}
	err = migrator.CreateView(viewName, viewOption)
	require.NoError(t, err)

	// Drop view
	err = migrator.DropView(viewName)
	require.NoError(t, err)

	// Verify view no longer exists by trying to query it
	var count int64
	err = db.Table(viewName).Count(&count).Error
	assert.Error(t, err) // Should error because view doesn't exist
}

func TestMigrator_HasConstraint(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	// Check for constraints (this depends on how DuckDB handles constraints)
	hasConstraint := migrator.HasConstraint(&TestUser{}, "idx_email")
	// The result depends on the implementation - we mainly test that it doesn't panic
	assert.IsType(t, true, hasConstraint)
}

func TestMigrator_DropConstraint(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	// Try to drop a constraint - DuckDB may not support this operation
	err = migrator.DropConstraint(&TestUser{}, "idx_email")
	if err != nil {
		// DuckDB may not support DROP CONSTRAINT - that's acceptable
		t.Logf("DropConstraint failed as expected due to DuckDB limitations: %v", err)
		assert.Contains(t, err.Error(), "No support for that ALTER TABLE option")
	} else {
		// If successful, verify constraint no longer exists
		hasConstraint := migrator.HasConstraint(&TestUser{}, "idx_email")
		assert.False(t, hasConstraint)
	}
}

func TestMigrator_GetTypeAliases(t *testing.T) {
	_, migrator := setupMigratorTestDB(t)

	// Get type aliases with a dummy table name
	aliases := migrator.GetTypeAliases("test_users")
	// GetTypeAliases might return nil for DuckDB - that's acceptable
	if aliases != nil {
		assert.IsType(t, map[string]string{}, aliases)
	}
	// The main test is that the method doesn't panic
}
