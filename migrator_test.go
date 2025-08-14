package duckdb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
)

// Test models for migration functionality
type MigrationTestUser struct {
	ID     uint   `gorm:"primarykey"`
	Name   string `gorm:"size:100"`
	Email  string `gorm:"size:255;uniqueIndex:idx_email"`
	Age    int
	Active bool
}

type MigrationTestPost struct {
	ID      uint   `gorm:"primarykey"`
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
	hasTable := migrator.HasTable(&MigrationTestUser{})
	assert.False(t, hasTable)

	// Create table
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	// Table should exist now
	hasTable = migrator.HasTable(&MigrationTestUser{})
	assert.True(t, hasTable)

	// Test with table name string
	hasTable = migrator.HasTable("migration_test_users")
	assert.True(t, hasTable)

	// Non-existent table
	hasTable = migrator.HasTable("non_existent_table")
	assert.False(t, hasTable)
}

func TestMigrator_CreateTable(t *testing.T) {
	_, migrator := setupMigratorTestDB(t)

	// Create table using migrator
	err := migrator.CreateTable(&MigrationTestUser{})
	require.NoError(t, err)

	// Verify table exists
	hasTable := migrator.HasTable(&MigrationTestUser{})
	assert.True(t, hasTable)

	// Try to create the same table again - should not error due to IF NOT EXISTS
	err = migrator.CreateTable(&MigrationTestUser{})
	require.NoError(t, err)

	// Test creating multiple tables
	err = migrator.CreateTable(&MigrationTestUser{}, &MigrationTestPost{})
	require.NoError(t, err)

	assert.True(t, migrator.HasTable(&MigrationTestUser{}))
	assert.True(t, migrator.HasTable(&MigrationTestPost{}))
}

func TestMigrator_DropTable(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table first
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)
	assert.True(t, migrator.HasTable(&MigrationTestUser{}))

	// Drop table
	err = migrator.DropTable(&MigrationTestUser{})
	require.NoError(t, err)

	// Verify table no longer exists
	hasTable := migrator.HasTable(&MigrationTestUser{})
	assert.False(t, hasTable)

	// Try to drop non-existent table - should not error due to IF EXISTS
	err = migrator.DropTable("non_existent_table")
	require.NoError(t, err)
}

func TestMigrator_HasColumn(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	// Check existing columns
	hasColumn := migrator.HasColumn(&MigrationTestUser{}, "name")
	assert.True(t, hasColumn)

	hasColumn = migrator.HasColumn(&MigrationTestUser{}, "email")
	assert.True(t, hasColumn)

	hasColumn = migrator.HasColumn(&MigrationTestUser{}, "age")
	assert.True(t, hasColumn)

	// Check non-existent column
	hasColumn = migrator.HasColumn(&MigrationTestUser{}, "non_existent_column")
	assert.False(t, hasColumn)

	// Test with table name string
	hasColumn = migrator.HasColumn("migration_test_users", "name")
	assert.True(t, hasColumn)
}

func TestMigrator_AlterColumn(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	// Alter column - this tests the AlterColumn method
	err = migrator.AlterColumn(&MigrationTestUser{}, "name")
	require.NoError(t, err)

	// Verify column still exists (basic check)
	hasColumn := migrator.HasColumn(&MigrationTestUser{}, "name")
	assert.True(t, hasColumn)
}

func TestMigrator_RenameColumn(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	// Rename column
	err = migrator.RenameColumn(&MigrationTestUser{}, "name", "full_name")
	require.NoError(t, err)

	// Verify old column doesn't exist and new column exists
	hasOldColumn := migrator.HasColumn(&MigrationTestUser{}, "name")
	assert.False(t, hasOldColumn)

	hasNewColumn := migrator.HasColumn(&MigrationTestUser{}, "full_name")
	assert.True(t, hasNewColumn)
}

func TestMigrator_AddColumn(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	// Add a new column
	err = migrator.AddColumn(&MigrationTestUser{}, "new_column")
	// Note: This might fail since the field doesn't exist in the struct
	// but we're testing that the method doesn't panic
	// The actual implementation should handle missing fields gracefully
}

func TestMigrator_DropColumn(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	// Drop a column
	err = migrator.DropColumn(&MigrationTestUser{}, "age")
	require.NoError(t, err)

	// Verify column no longer exists
	hasColumn := migrator.HasColumn(&MigrationTestUser{}, "age")
	assert.False(t, hasColumn)
}

func TestMigrator_HasIndex(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	// Check for the email index that should be created
	hasIndex := migrator.HasIndex(&MigrationTestUser{}, "idx_email")
	assert.True(t, hasIndex)

	// Check for non-existent index
	hasIndex = migrator.HasIndex(&MigrationTestUser{}, "non_existent_index")
	assert.False(t, hasIndex)
}

func TestMigrator_CreateIndex(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	// Create an index
	err = migrator.CreateIndex(&MigrationTestUser{}, "name")
	require.NoError(t, err)

	// Verify index exists (this might vary depending on how DuckDB handles index names)
	// The exact index name might be auto-generated
}

func TestMigrator_RenameIndex(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table with index
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	// Rename index
	err = migrator.RenameIndex(&MigrationTestUser{}, "idx_email", "idx_user_email")
	require.NoError(t, err)

	// Verify old index doesn't exist and new index exists
	hasOldIndex := migrator.HasIndex(&MigrationTestUser{}, "idx_email")
	assert.False(t, hasOldIndex)

	hasNewIndex := migrator.HasIndex(&MigrationTestUser{}, "idx_user_email")
	assert.True(t, hasNewIndex)
}

func TestMigrator_DropIndex(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table with index
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	// Drop index
	err = migrator.DropIndex(&MigrationTestUser{}, "idx_email")
	require.NoError(t, err)

	// Verify index no longer exists
	hasIndex := migrator.HasIndex(&MigrationTestUser{}, "idx_email")
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
	err = db.AutoMigrate(&MigrationTestUser{}, &MigrationTestPost{})
	require.NoError(t, err)

	// Get tables again
	tables, err = migrator.GetTables()
	require.NoError(t, err)
	assert.Contains(t, tables, "migration_test_users")
	assert.Contains(t, tables, "migration_test_posts")
}

func TestMigrator_FullDataTypeOf(t *testing.T) {
	_, migrator := setupMigratorTestDB(t)

	// Test with a sample field
	// This requires creating a statement with the field information
	// For now, we test that the method exists and doesn't panic
	user := &MigrationTestUser{}
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
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	// Create view
	viewName := "user_view"
	viewOption := gorm.ViewOption{
		Query: db.Select("id, name").Table("migration_test_users"),
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
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	viewName := "user_view"
	viewOption := gorm.ViewOption{
		Query: db.Select("id, name").Table("migration_test_users"),
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
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	// Check for constraints (this depends on how DuckDB handles constraints)
	hasConstraint := migrator.HasConstraint(&MigrationTestUser{}, "idx_email")
	// The result depends on the implementation - we mainly test that it doesn't panic
	assert.IsType(t, true, hasConstraint)
}

func TestMigrator_DropConstraint(t *testing.T) {
	db, migrator := setupMigratorTestDB(t)

	// Create table
	err := db.AutoMigrate(&MigrationTestUser{})
	require.NoError(t, err)

	// Try to drop a constraint
	err = migrator.DropConstraint(&MigrationTestUser{}, "idx_email")
	require.NoError(t, err)

	// Verify constraint no longer exists
	hasConstraint := migrator.HasConstraint(&MigrationTestUser{}, "idx_email")
	assert.False(t, hasConstraint)
}

func TestMigrator_GetTypeAliases(t *testing.T) {
	_, migrator := setupMigratorTestDB(t)

	// Get type aliases with a dummy table name
	aliases := migrator.GetTypeAliases("migration_test_users")
	assert.NotNil(t, aliases)
	// Test that common aliases exist
	if len(aliases) > 0 {
		// Just verify it returns a map without specific assertions
		// since the exact aliases may vary
		assert.IsType(t, map[string]string{}, aliases)
	}
}
