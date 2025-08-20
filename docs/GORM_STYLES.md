# GORM Coding Style & Function Reference

## Overview

GORM is a full-featured ORM library for Go that aims to be developer-friendly. This document outlines the coding style conventions and provides a comprehensive function reference for working with GORM.

## Table of Contents

1. [Coding Style Guidelines](#coding-style-guidelines)
2. [Model Declaration](#model-declaration)
3. [Database Operations](#database-operations)
4. [Query Methods](#query-methods)
5. [Advanced Features](#advanced-features)
6. [Best Practices](#best-practices)

---

## Coding Style Guidelines

### General Conventions

- **CamelCase**: Use CamelCase for struct names and field names
- **snake_case**: GORM automatically converts struct names to snake_case for table names
- **Pluralization**: Table names are automatically pluralized (e.g., `User` → `users`)
- **Primary Key**: Use `ID` as the default primary key field name
- **Timestamps**: Use `CreatedAt` and `UpdatedAt` for automatic timestamp tracking

### Naming Conventions

```go
// ✅ Good - Follow Go naming conventions
type User struct {
    ID        uint      `gorm:"primaryKey"`
    FirstName string    `gorm:"column:first_name"`
    LastName  string    `gorm:"column:last_name"`
    Email     string    `gorm:"uniqueIndex"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// ❌ Avoid - Inconsistent naming
type user struct {
    id        uint
    firstName string
    lastName  string
}
```

### Error Handling Pattern

```go
// ✅ Always check for errors
result := db.First(&user, 1)
if result.Error != nil {
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        // Handle record not found
        return nil, fmt.Errorf("user not found")
    }
    return nil, result.Error
}

// ✅ Use method chaining with error checking
if err := db.Where("email = ?", email).First(&user).Error; err != nil {
    return nil, err
}
```

---

## Model Declaration

### Basic Model Structure

```go
type User struct {
    ID        uint           `gorm:"primaryKey"`
    Name      string         `gorm:"size:255;not null"`
    Email     string         `gorm:"uniqueIndex;size:255"`
    Age       int            `gorm:"check:age > 0"`
    Active    bool           `gorm:"default:true"`
    Profile   *string        `gorm:"size:1000"` // Nullable field
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

### Using gorm.Model

```go
// gorm.Model includes ID, CreatedAt, UpdatedAt, DeletedAt
type User struct {
    gorm.Model
    Name  string `gorm:"size:255;not null"`
    Email string `gorm:"uniqueIndex;size:255"`
}
```

### Field Tags Reference

| Tag | Description | Example |
|-----|-------------|---------|
| `column` | Specify column name | `gorm:"column:user_name"` |
| `type` | Specify column data type | `gorm:"type:varchar(255)"` |
| `size` | Specify column size | `gorm:"size:255"` |
| `primaryKey` | Mark as primary key | `gorm:"primaryKey"` |
| `unique` | Mark as unique | `gorm:"unique"` |
| `uniqueIndex` | Create unique index | `gorm:"uniqueIndex"` |
| `index` | Create index | `gorm:"index"` |
| `not null` | NOT NULL constraint | `gorm:"not null"` |
| `default` | Default value | `gorm:"default:true"` |
| `autoIncrement` | Auto increment | `gorm:"autoIncrement"` |
| `check` | Check constraint | `gorm:"check:age > 0"` |
| `->` | Read-only permission | `gorm:"->:false"` |
| `<-` | Write-only permission | `gorm:"<-:create"` |
| `-` | Ignore field | `gorm:"-"` |

### Embedded Structs

```go
type Author struct {
    Name  string
    Email string
}

type Blog struct {
    ID      int
    Author  Author `gorm:"embedded"`
    Title   string
    Content string
}

// With prefix
type BlogWithPrefix struct {
    ID      int
    Author  Author `gorm:"embedded;embeddedPrefix:author_"`
    Title   string
    Content string
}
```

---

## Database Operations

### Connection and Configuration

```go
import (
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
)

// Database connection
func InitDB() (*gorm.DB, error) {
    dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, err
    }
    
    // Auto migrate
    err = db.AutoMigrate(&User{})
    if err != nil {
        return nil, err
    }
    
    return db, nil
}
```

### CRUD Operations

#### Create

```go
// Create single record
user := User{Name: "John", Email: "john@example.com"}
result := db.Create(&user)
if result.Error != nil {
    // Handle error
}

// Create multiple records
users := []User{
    {Name: "John", Email: "john@example.com"},
    {Name: "Jane", Email: "jane@example.com"},
}
db.Create(&users)

// Create in batches
db.CreateInBatches(users, 100)
```

#### Read

```go
var user User

// Get first record
db.First(&user)

// Get by primary key
db.First(&user, 1)
db.First(&user, "id = ?", "string_primary_key")

// Get last record
db.Last(&user)

// Get all records
var users []User
db.Find(&users)

// Get record with conditions
db.Where("name = ?", "John").First(&user)
```

#### Update

```go
// Update single field
db.Model(&user).Update("name", "John Updated")

// Update multiple fields
db.Model(&user).Updates(User{Name: "John", Age: 30})

// Update with map
db.Model(&user).Updates(map[string]interface{}{
    "name": "John",
    "age":  30,
})

// Update with conditions
db.Model(&user).Where("active = ?", true).Update("name", "John")
```

#### Delete

```go
// Soft delete (if DeletedAt field exists)
db.Delete(&user, 1)

// Permanent delete
db.Unscoped().Delete(&user, 1)

// Delete with conditions
db.Where("age < ?", 18).Delete(&User{})
```

---

## Query Methods

### Basic Queries

#### Single Record Retrieval

```go
// First - ordered by primary key
db.First(&user)
db.First(&user, 1)                    // With primary key
db.First(&user, "id = ?", "uuid")     // String primary key

// Take - no ordering
db.Take(&user)

// Last - ordered by primary key desc
db.Last(&user)
```

#### Multiple Records

```go
var users []User

// Find all
db.Find(&users)

// Find with conditions
db.Where("age > ?", 18).Find(&users)

// Find with limit
db.Limit(10).Find(&users)

// Find with offset
db.Offset(5).Limit(10).Find(&users)
```

### Conditions

#### String Conditions

```go
// Simple condition
db.Where("name = ?", "John").Find(&users)

// Multiple conditions
db.Where("name = ? AND age > ?", "John", 18).Find(&users)

// IN condition
db.Where("name IN ?", []string{"John", "Jane"}).Find(&users)

// LIKE condition
db.Where("name LIKE ?", "%John%").Find(&users)
```

#### Struct Conditions

```go
// Struct condition (non-zero fields only)
db.Where(&User{Name: "John", Age: 20}).Find(&users)

// Map condition (includes zero values)
db.Where(map[string]interface{}{"name": "John", "age": 0}).Find(&users)
```

#### Advanced Conditions

```go
// NOT conditions
db.Not("name = ?", "John").Find(&users)

// OR conditions
db.Where("name = ?", "John").Or("name = ?", "Jane").Find(&users)

// Complex conditions with parentheses
db.Where(
    db.Where("name = ?", "John").Or("name = ?", "Jane"),
).Where("age > ?", 18).Find(&users)
```

### Ordering and Limiting

```go
// Order by single field
db.Order("age desc").Find(&users)

// Order by multiple fields
db.Order("age desc, name asc").Find(&users)

// Limit and offset
db.Limit(10).Offset(5).Find(&users)

// Distinct
db.Distinct("name", "age").Find(&users)
```

### Selecting Fields

```go
// Select specific fields
db.Select("name", "age").Find(&users)

// Select with expressions
db.Select("name", "age", "age * 2 as double_age").Find(&users)

// Omit fields
db.Omit("password").Find(&users)
```

### Aggregation

```go
// Count
var count int64
db.Model(&User{}).Where("age > ?", 18).Count(&count)

// Group by with having
type Result struct {
    Date  time.Time
    Total int
}

var results []Result
db.Model(&User{}).
    Select("date(created_at) as date, count(*) as total").
    Group("date(created_at)").
    Having("count(*) > ?", 10).
    Scan(&results)
```

### Joins

```go
type User struct {
    ID        uint
    Name      string
    CompanyID uint
    Company   Company
}

type Company struct {
    ID   uint
    Name string
}

// Inner join
db.Joins("Company").Find(&users)

// Left join with conditions
db.Joins("LEFT JOIN companies ON companies.id = users.company_id").
    Where("companies.name = ?", "Tech Corp").
    Find(&users)

// Join with preloading
db.Joins("Company").Where("companies.name = ?", "Tech Corp").Find(&users)
```

---

## Advanced Features

### Transactions

```go
// Manual transaction
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := tx.Create(&user1).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(&user2).Error; err != nil {
    tx.Rollback()
    return err
}

return tx.Commit().Error

// Transaction with closure
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user1).Error; err != nil {
        return err
    }
    
    if err := tx.Create(&user2).Error; err != nil {
        return err
    }
    
    return nil
})
```

### Hooks

```go
// BeforeCreate hook
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
    if u.Email == "" {
        return errors.New("email is required")
    }
    
    // Generate UUID
    u.ID = generateUUID()
    return
}

// AfterFind hook
func (u *User) AfterFind(tx *gorm.DB) (err error) {
    if u.Role == "" {
        u.Role = "user"
    }
    return
}
```

### Scopes

```go
// Define scope
func AgeGreaterThan(age int) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("age > ?", age)
    }
}

func ActiveUsers(db *gorm.DB) *gorm.DB {
    return db.Where("active = ?", true)
}

// Use scopes
db.Scopes(AgeGreaterThan(18), ActiveUsers).Find(&users)
```

### Method Chaining Categories

#### Chain Methods

- `Where`, `Or`, `Not`
- `Limit`, `Offset`
- `Order`, `Group`, `Having`
- `Joins`, `Preload`, `Eager`
- `Select`, `Omit`

#### Finisher Methods

- `Create`, `Save`, `Update`, `Delete`
- `First`, `Last`, `Take`, `Find`
- `Count`, `Pluck`, `Scan`

#### New Session Methods

- `Session`, `WithContext`
- `Debug`, `Clauses`

---

## Best Practices

### 1. Error Handling

```go
// ✅ Always check for specific errors
if err := db.Where("email = ?", email).First(&user).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return ErrUserNotFound
    }
    return err
}

// ✅ Use context for timeout control
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := db.WithContext(ctx).First(&user, id).Error; err != nil {
    return err
}
```

### 2. Performance Optimization

```go
// ✅ Use indexes for frequently queried fields
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Email string `gorm:"uniqueIndex"`
    Name  string `gorm:"index"`
}

// ✅ Use Select to limit fields
db.Select("id", "name", "email").Find(&users)

// ✅ Use batching for large operations
db.CreateInBatches(users, 100)

// ✅ Use Find instead of First when you don't need ordering
db.Limit(1).Find(&user) // Instead of db.First(&user)
```

### 3. Security

```go
// ✅ Always use parameterized queries
db.Where("email = ?", userInput).First(&user)

// ❌ Never use string concatenation
// db.Where("email = '" + userInput + "'").First(&user) // SQL injection risk

// ✅ Validate input before queries
if !isValidEmail(email) {
    return errors.New("invalid email format")
}
```

### 4. Model Design

```go
// ✅ Use appropriate data types
type User struct {
    ID          uint           `gorm:"primaryKey"`
    Email       string         `gorm:"uniqueIndex;size:255;not null"`
    Password    string         `gorm:"size:255;not null"`
    IsActive    bool           `gorm:"default:true"`
    LastLoginAt *time.Time     // Nullable timestamp
    Settings    datatypes.JSON `gorm:"type:jsonb"` // PostgreSQL JSON
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// ✅ Use proper field permissions
type User struct {
    ID       uint   `gorm:"primaryKey"`
    Name     string `gorm:"<-:create"` // Create-only
    Email    string `gorm:"<-"`        // Create and update
    ReadOnly string `gorm:"->"`        // Read-only
    Internal string `gorm:"-"`         // Ignored
}
```

### 5. Association Management

```go
// ✅ Define clear relationships
type User struct {
    ID       uint
    Name     string
    Posts    []Post    `gorm:"foreignKey:UserID"`
    Profile  Profile   `gorm:"foreignKey:UserID"`
    Roles    []Role    `gorm:"many2many:user_roles;"`
}

// ✅ Use preloading for related data
db.Preload("Posts").Preload("Profile").Find(&users)

// ✅ Use joins for filtering
db.Joins("Profile").Where("profiles.verified = ?", true).Find(&users)
```

### 6. Database Connection Management

```go
// ✅ Configure connection pool
sqlDB, err := db.DB()
if err != nil {
    return err
}

sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

### 7. Testing

```go
// ✅ Use test database
func setupTestDB() *gorm.DB {
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    db.AutoMigrate(&User{})
    return db
}

// ✅ Use transactions for test isolation
func TestCreateUser(t *testing.T) {
    db := setupTestDB()
    
    tx := db.Begin()
    defer tx.Rollback()
    
    user := &User{Name: "Test User", Email: "test@example.com"}
    err := tx.Create(user).Error
    
    assert.NoError(t, err)
    assert.NotZero(t, user.ID)
}
```

---

## Migration and Schema

### Auto Migration

```go
// Basic auto migration
db.AutoMigrate(&User{})

// Multiple models
db.AutoMigrate(&User{}, &Product{}, &Order{})

// With error handling
if err := db.AutoMigrate(&User{}); err != nil {
    log.Fatal("Failed to migrate database:", err)
}
```

### Manual Migration

```go
// Create table
db.Migrator().CreateTable(&User{})

// Add column
db.Migrator().AddColumn(&User{}, "Age")

// Drop column
db.Migrator().DropColumn(&User{}, "Age")

// Create index
db.Migrator().CreateIndex(&User{}, "Email")

// Drop index
db.Migrator().DropIndex(&User{}, "Email")
```

This reference document provides a comprehensive guide to GORM's coding style and functionality.
Follow these patterns and conventions to write maintainable, performant, and secure database code
with GORM.
