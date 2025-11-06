package duckdb

import (
	"os"
	"reflect"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// Custom CREATE callback to replace GORM's broken one
//nolint:gocyclo // Complex function handling comprehensive GORM CREATE integration
func customCreateCallback(db *gorm.DB) {
	debugLog("customCreateCallback called")
	
	if db.Error != nil {
		debugLog("customCreateCallback: early exit due to error: %v", db.Error)
		return
	}

	// Build INSERT statement manually
	if db.Statement.Schema != nil {
		debugLog("customCreateCallback: building INSERT for table %s", db.Statement.Schema.Table)
		
		// Get all fields that should be inserted
		var columns []clause.Column
		var values []interface{}
		
		for _, field := range db.Statement.Schema.Fields {
			// Skip auto-increment primary keys (they'll be handled by RETURNING)
			if field.PrimaryKey && field.AutoIncrement {
				debugLog("customCreateCallback: skipping auto-increment field %s", field.Name)
				continue
			}
			
			// Get field value from the model
			fieldValue := db.Statement.ReflectValue.FieldByName(field.Name).Interface()
			debugLog("customCreateCallback: adding field %s = %v", field.DBName, fieldValue)
			columns = append(columns, clause.Column{Name: field.DBName})
			values = append(values, fieldValue)
		}
		
		if len(columns) > 0 {
			// Add INSERT clause
			db.Statement.AddClause(clause.Insert{Table: clause.Table{Name: db.Statement.Table}})
			
			// Add VALUES clause
			db.Statement.AddClause(clause.Values{
				Columns: columns,
				Values:  [][]interface{}{values},
			})
			
			// For auto-increment fields, add RETURNING clause
			hasAutoIncrement := false
			var autoIncrementField *schema.Field
			for _, field := range db.Statement.Schema.Fields {
				if field.PrimaryKey && field.AutoIncrement {
					hasAutoIncrement = true
					autoIncrementField = field
					break
				}
			}
			
			if hasAutoIncrement {
				debugLog("customCreateCallback: adding RETURNING for field %s", autoIncrementField.DBName)
				db.Statement.AddClause(clause.Returning{
					Columns: []clause.Column{{Name: autoIncrementField.DBName}},
				})
			}
			
			// Build and execute the statement
			db.Statement.Build("INSERT", "VALUES", "RETURNING")
			debugLog("customCreateCallback: generated SQL: %s", db.Statement.SQL.String())
			debugLog("customCreateCallback: vars: %+v", any(db.Statement.Vars))
			
			// Execute the statement
			if hasAutoIncrement {
				// Use QueryRow for RETURNING
				var id interface{}
				err := db.Statement.ConnPool.QueryRowContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...).Scan(&id)
				if err != nil {
					db.Error = err
					debugLog("customCreateCallback: QueryRow failed: %v", err)
				} else {
					db.RowsAffected = 1
					debugLog("customCreateCallback: QueryRow succeeded, ID: %v", id)
					
					// Set the ID back to the model
					if autoIncrementField != nil {
						// Get the struct value (dereference pointer if needed)
						structValue := db.Statement.ReflectValue
						if structValue.Kind() == reflect.Ptr {
							structValue = structValue.Elem()
						}
						
						fieldValue := structValue.FieldByName(autoIncrementField.Name)
						debugLog("customCreateCallback: Setting field %s, Valid: %t, CanSet: %t, Kind: %s", 
							autoIncrementField.Name, fieldValue.IsValid(), fieldValue.CanSet(), fieldValue.Kind())
						
						if fieldValue.IsValid() && fieldValue.CanSet() {
							debugLog("customCreateCallback: ID value type: %T, value: %v", id, id)
							switch fieldValue.Kind() {
							case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
								var uintVal uint64
								switch v := id.(type) {
							case uint64:
								uintVal = v
							case int64:
								if v < 0 {
									debugLog("customCreateCallback: Negative int64 %d cannot be converted to uint64", v)
									return
								}
								uintVal = uint64(v)
							case int32:
								if v < 0 {
									debugLog("customCreateCallback: Negative int32 %d cannot be converted to uint64", v)
									return
								}
								uintVal = uint64(v)
							case int:
								if v < 0 {
									debugLog("customCreateCallback: Negative int %d cannot be converted to uint64", v)
									return
								}
								uintVal = uint64(v)
								default:
									debugLog("customCreateCallback: Could not convert ID %v (%T) to uint", id, id)
									return
								}
								fieldValue.SetUint(uintVal)
								debugLog("customCreateCallback: Set uint field to %d", uintVal)
							case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
								var intVal int64
								switch v := id.(type) {
							case int64:
								intVal = v
							case uint64:
								if v > 9223372036854775807 { // math.MaxInt64
									debugLog("customCreateCallback: uint64 %d exceeds int64 maximum", v)
									return
								}
								intVal = int64(v)
								case int32:
									intVal = int64(v)
								case int:
									intVal = int64(v)
								default:
									debugLog("customCreateCallback: Could not convert ID %v (%T) to int", id, id)
									return
								}
								fieldValue.SetInt(intVal)
								debugLog("customCreateCallback: Set int field to %d", intVal)
							}
						} else {
							debugLog("customCreateCallback: Cannot set field %s", autoIncrementField.Name)
						}
					}
				}
			} else {
				// Use Exec for non-returning operations
				result, err := db.Statement.ConnPool.ExecContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
				if err != nil {
					db.Error = err
					debugLog("customCreateCallback: Exec failed: %v", err)
				} else {
					affected, _ := result.RowsAffected()
					db.RowsAffected = affected
					debugLog("customCreateCallback: Exec succeeded, rows affected: %d", affected)
				}
			}
		} else {
			debugLog("customCreateCallback: no columns to insert")
		}
	} else {
		debugLog("customCreateCallback: no schema available")
	}
}

// Test with custom CREATE callback
func TestCustomCreateCallback(t *testing.T) {
	t.Log("=== Custom CREATE Callback Test ===")

	// Enable debug mode
	if err := os.Setenv("GORM_DUCKDB_DEBUG", "1"); err != nil {
		t.Fatalf("Failed to set debug environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("GORM_DUCKDB_DEBUG"); err != nil {
			t.Logf("Failed to unset debug environment variable: %v", err)
		}
	}()

	dialector := Dialector{
		Config: &Config{
			DSN: ":memory:",
		},
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatalf("Failed to open DuckDB: %v", err)
	}

	// Replace GORM's broken create callback with our custom one
	err = db.Callback().Create().Replace("gorm:create", customCreateCallback)
	if err != nil {
		t.Fatalf("Failed to register custom create callback: %v", err)
	}
	if err != nil {
		t.Fatalf("Failed to replace create callback: %v", err)
	}

	type SimpleModel struct {
		ID   uint   `gorm:"primaryKey;autoIncrement"`
		Name string
	}

	// Migration
	err = db.AutoMigrate(&SimpleModel{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Test create with custom callback
	model := SimpleModel{Name: "Custom Callback Test"}
	result := db.Create(&model)
	t.Logf("Create result: Error=%v, RowsAffected=%d, ID=%d", 
		result.Error, result.RowsAffected, model.ID)

	// Verify the record was actually created
	var count int64
	db.Model(&SimpleModel{}).Count(&count)
	t.Logf("Total records in table: %d", count)
	
	// Verify we can read it back
	var retrieved SimpleModel
	err = db.First(&retrieved).Error
	if err != nil {
		t.Logf("Failed to retrieve record: %v", err)
	} else {
		t.Logf("Retrieved record: ID=%d, Name=%s", retrieved.ID, retrieved.Name)
	}
}