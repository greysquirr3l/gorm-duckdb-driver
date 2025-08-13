package duckdb_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
)

type User struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"size:100;not null"`
	Email     string `gorm:"size:255;uniqueIndex"`
	Age       uint8
	Birthday  time.Time `gorm:"autoCreateTime:false"`
	CreatedAt time.Time `gorm:"autoCreateTime:false"`
	UpdatedAt time.Time `gorm:"autoUpdateTime:false"`
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dialector := duckdb.Open(":memory:")
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&User{})
	require.NoError(t, err)

	return db
}

func TestDialector(t *testing.T) {
	dialector := duckdb.Open(":memory:")
	assert.Equal(t, "duckdb", dialector.Name())
}

func TestConnection(t *testing.T) {
	db := setupTestDB(t)

	// Test that the connection works
	sqlDB, err := db.DB()
	require.NoError(t, err)

	err = sqlDB.Ping()
	assert.NoError(t, err)
}

func TestBasicCRUD(t *testing.T) {
	db := setupTestDB(t)

	// Create
	user := User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		Birthday: time.Date(1993, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	err := db.Create(&user).Error
	require.NoError(t, err)
	assert.NotZero(t, user.ID)

	// Read
	var foundUser User
	err = db.First(&foundUser, user.ID).Error
	require.NoError(t, err)
	assert.Equal(t, user.Name, foundUser.Name)
	assert.Equal(t, user.Email, foundUser.Email)

	// Update
	err = db.Model(&foundUser).Update("age", 31).Error
	require.NoError(t, err)

	// Verify update
	err = db.First(&foundUser, user.ID).Error
	require.NoError(t, err)
	assert.Equal(t, uint8(31), foundUser.Age)

	// Delete
	err = db.Delete(&foundUser).Error
	require.NoError(t, err)

	// Verify deletion
	err = db.First(&foundUser, user.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestTransaction(t *testing.T) {
	db := setupTestDB(t)

	// Test successful transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		user1 := User{Name: "Alice", Email: "alice@example.com", Age: 25}
		if err := tx.Create(&user1).Error; err != nil {
			return err
		}

		user2 := User{Name: "Bob", Email: "bob@example.com", Age: 28}
		if err := tx.Create(&user2).Error; err != nil {
			return err
		}

		return nil
	})
	require.NoError(t, err)

	// Verify both users were created
	var count int64
	err = db.Model(&User{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestErrorTranslator(t *testing.T) {
	db := setupTestDB(t)

	// Create a user
	user := User{Name: "John", Email: "john@test.com", Age: 25}
	err := db.Create(&user).Error
	require.NoError(t, err)

	// Try to create another user with the same email (should violate unique constraint)
	duplicateUser := User{Name: "Jane", Email: "john@test.com", Age: 30}
	err = db.Create(&duplicateUser).Error

	// Should get a GORM error (the exact error type depends on the translator implementation)
	assert.Error(t, err)
}

func TestDataTypes(t *testing.T) {
	db := setupTestDB(t)

	user := User{
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      25,
		Birthday: time.Date(1998, 5, 15, 0, 0, 0, 0, time.UTC),
	}

	err := db.Create(&user).Error
	require.NoError(t, err)

	var retrieved User
	err = db.First(&retrieved, user.ID).Error
	require.NoError(t, err)

	assert.Equal(t, user.Name, retrieved.Name)
	assert.Equal(t, user.Email, retrieved.Email)
	assert.Equal(t, user.Age, retrieved.Age)

	// Check that timestamps are approximately equal (within a second)
	assert.WithinDuration(t, user.Birthday, retrieved.Birthday, time.Second)
}
