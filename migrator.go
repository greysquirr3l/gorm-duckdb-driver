package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

// Constants for repeated strings
const (
	schemaMain        = "main"
	notNullConstraint = " NOT NULL"
)

// normalizeTable splits and strips quotes from a table identifier which may be
// schema-qualified (e.g. "schema"."table" or schema.table). Returns schema
// (may be empty) and table name.
func normalizeTable(table string) string {
	if table == "" {
		return ""
	}
	// Remove escaped quotes/backticks
	t := strings.ReplaceAll(table, `\"`, "")
	t = strings.ReplaceAll(t, `\`+"`", "")
	t = strings.ReplaceAll(t, `"`, "")
	t = strings.Trim(t, "`\"")
	if parts := strings.SplitN(t, ".", 2); len(parts) == 2 {
		return strings.Trim(parts[1], "`\"")
	}
	return t
}

const (
	// SQL data type constants
	sqlTypeBigInt  = "BIGINT"
	sqlTypeInteger = "INTEGER"
)

// isAlreadyExistsError checks if an error indicates that an object already exists
func isAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "already exists") ||
		strings.Contains(errMsg, "duplicate")
}

// Migrator implements gorm.Migrator interface for DuckDB database.
type Migrator struct {
	migrator.Migrator
}

// isAutoIncrementField checks if a field is an auto-increment field
func (m Migrator) isAutoIncrementField(field *schema.Field) bool {
	return field.AutoIncrement || (!field.HasDefaultValue && field.DataType == schema.Uint)
}

// CurrentDatabase returns the current database name.
func (m Migrator) CurrentDatabase() (name string) {
	if m.DB == nil {
		return schemaMain
	}
	row := m.DB.Raw("SELECT current_database()").Row()
	if row == nil {
		return schemaMain
	}
	if err := row.Scan(&name); err != nil {
		return schemaMain
	}
	return
}

// FullDataTypeOf returns the full data type for a field including constraints.
// Override FullDataTypeOf to prevent GORM from adding duplicate PRIMARY KEY clauses
func (m Migrator) FullDataTypeOf(field *schema.Field) clause.Expr {
	// Get the base data type from our dialector
	dataType := m.Dialector.DataTypeOf(field)

	expr := clause.Expr{SQL: dataType}

	// For primary key fields, ensure clean type definition without duplicate PRIMARY KEY
	if field.PrimaryKey {
		// DuckDB doesn't support native AUTO_INCREMENT, so we use sequences to emulate this behavior for auto-increment primary keys
		// Check if this is an auto-increment field (no default value specified)
		if m.isAutoIncrementField(field) {
			// Use BIGINT with a default sequence value. If field.Schema is nil
			// try to derive the table name from m.DB.Statement.
			tableName := ""
			if field != nil && field.Schema != nil && field.Schema.Table != "" {
				tableName = field.Schema.Table
			} else if m.DB != nil && m.DB.Statement != nil && m.DB.Statement.Table != "" {
				tableName = m.DB.Statement.Table
			}

			if tableName != "" {
				expr.SQL = "BIGINT DEFAULT nextval('seq_" + strings.ToLower(tableName) + "_" + strings.ToLower(field.DBName) + "')"
			}
		} else {
			// Make sure the data type is clean for non-auto-increment primary keys
			upperDataType := strings.ToUpper(dataType)
			switch {
			case strings.Contains(upperDataType, sqlTypeBigInt):
				expr.SQL = sqlTypeBigInt
			case strings.Contains(upperDataType, sqlTypeInteger):
				expr.SQL = sqlTypeInteger
			default:
				expr.SQL = dataType
			}
		}

		// Add NOT NULL for primary keys
		expr.SQL += notNullConstraint

		// Do NOT add PRIMARY KEY here - let GORM handle it in the table definition
		return expr
	}

	// For non-primary key fields, add constraints
	if field.NotNull {
		expr.SQL += notNullConstraint
	}

	if field.Unique {
		expr.SQL += " UNIQUE"
	}

	// Handle defaults for non-primary key fields only
	if field.HasDefaultValue && (field.DefaultValueInterface != nil || field.DefaultValue != "") {
		if field.DefaultValueInterface != nil {
			defaultStmt := &gorm.Statement{Vars: []interface{}{field.DefaultValueInterface}}
			m.BindVarTo(defaultStmt, defaultStmt, field.DefaultValueInterface)
			expr.SQL += " DEFAULT " + m.Explain(defaultStmt.SQL.String(), field.DefaultValueInterface)
		} else if field.DefaultValue != "(-)" {
			expr.SQL += " DEFAULT " + field.DefaultValue
		}
	}

	if field.Comment != "" {
		expr.SQL += " COMMENT '" + field.Comment + "'"
	}

	return expr
}

// AlterColumn modifies a column definition in DuckDB, handling syntax limitations.
func (m Migrator) AlterColumn(value interface{}, field string) error {
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(field); field != nil {
				// For ALTER COLUMN, only use the base data type without defaults
				baseType := m.Dialector.DataTypeOf(field)

				// Clean the base type - remove any DEFAULT clauses
				baseType = strings.Split(baseType, " DEFAULT")[0]

				return m.DB.Exec(
					"ALTER TABLE ? ALTER COLUMN ? TYPE ?",
					m.CurrentTable(stmt), clause.Column{Name: field.DBName}, clause.Expr{SQL: baseType},
				).Error
			}
		}
		return fmt.Errorf("failed to look up field with name: %s", field)
	})
	if err != nil {
		return fmt.Errorf("failed to alter column: %w", err)
	}
	return nil
}

// RenameColumn renames a column in the database table.
func (m Migrator) RenameColumn(value interface{}, oldName, newName string) error {
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(oldName); field != nil {
				oldName = field.DBName
			}

			if field := stmt.Schema.LookUpField(newName); field != nil {
				newName = field.DBName
			}
		}

		return m.DB.Exec(
			"ALTER TABLE ? RENAME COLUMN ? TO ?",
			m.CurrentTable(stmt), clause.Column{Name: oldName}, clause.Column{Name: newName},
		).Error
	})
	if err != nil {
		return fmt.Errorf("failed to rename column: %w", err)
	}
	return nil
}

// RenameIndex renames an index in the database.
func (m Migrator) RenameIndex(value interface{}, oldName, newName string) error {
	err := m.RunWithValue(value, func(_ *gorm.Statement) error {
		return m.DB.Exec(
			"ALTER INDEX ? RENAME TO ?",
			clause.Column{Name: oldName}, clause.Column{Name: newName},
		).Error
	})
	if err != nil {
		return fmt.Errorf("failed to rename index: %w", err)
	}
	return nil
}

// DropIndex drops an index from the database.
func (m Migrator) DropIndex(value interface{}, name string) error {
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if idx := stmt.Schema.LookIndex(name); idx != nil {
				name = idx.Name
			}
		}

		return m.DB.Exec("DROP INDEX IF EXISTS ?", clause.Column{Name: name}).Error
	})
	if err != nil {
		return fmt.Errorf("failed to drop index: %w", err)
	}
	return nil
}

// DropConstraint drops a constraint from the database.
func (m Migrator) DropConstraint(value interface{}, name string) error {
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		constraint, table := m.GuessConstraintInterfaceAndTable(stmt, name)
		if constraint != nil {
			name = constraint.GetName()
		}
		return m.Migrator.DB.Exec("ALTER TABLE ? DROP CONSTRAINT ?", clause.Table{Name: table}, clause.Column{Name: name}).Error
	})
	if err != nil {
		return fmt.Errorf("failed to drop constraint: %w", err)
	}
	return nil
}

// HasTable checks if a table exists in the database.
func (m Migrator) HasTable(value interface{}) bool {
	var count int64

	_ = m.RunWithValue(value, func(stmt *gorm.Statement) error {
		// Derive table identifier: prefer schema metadata, fall back to stmt.Table or CurrentTable
		tableIdentifier := ""
		if stmt != nil && stmt.Schema != nil && stmt.Schema.Table != "" {
			tableIdentifier = stmt.Schema.Table
		} else if stmt != nil && stmt.Table != "" {
			tableIdentifier = stmt.Table
		} else {
			tableIdentifier = fmt.Sprint(m.CurrentTable(stmt))
		}

		// Normalize table identifier to handle quoted and schema-qualified names
		tableName := normalizeTable(tableIdentifier)
		rows, err := m.DB.Raw(
			"SELECT count(*) FROM information_schema.tables WHERE lower(table_name) = lower(?) AND table_type = 'BASE TABLE'",
			tableName,
		).Rows()
		if err != nil {
			return fmt.Errorf("failed to query table existence for %s: %w", tableName, err)
		}
		if rows == nil {
			return fmt.Errorf("rows is nil for table existence query")
		}
		defer func() { _ = rows.Close() }()
		if rows.Next() {
			if err := rows.Scan(&count); err != nil {
				return fmt.Errorf("failed to scan table count: %w", err)
			}
		}
		return nil
	})

	return count > 0
}

// GetTables returns a list of all table names in the database.
func (m Migrator) GetTables() (tableList []string, err error) {
	rows, err := m.DB.Raw(
		"SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE'",
	).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query information_schema tables: %w", err)
	}
	if rows == nil {
		return nil, nil
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			tableList = append(tableList, name)
		}
	}
	return tableList, nil
}

// HasColumn checks if a column exists in the database table.
func (m Migrator) HasColumn(value interface{}, field string) bool {
	var count int64
	_ = m.RunWithValue(value, func(stmt *gorm.Statement) error {
		name := field
		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(field); field != nil {
				name = field.DBName
			}
		}

		// Derive table identifier similarly to HasTable
		tableIdentifier := ""
		if stmt != nil && stmt.Schema != nil && stmt.Schema.Table != "" {
			tableIdentifier = stmt.Schema.Table
		} else if stmt != nil && stmt.Table != "" {
			tableIdentifier = stmt.Table
		} else {
			tableIdentifier = fmt.Sprint(m.CurrentTable(stmt))
		}
		tableName := normalizeTable(tableIdentifier)
		rows, err := m.DB.Raw(
			"SELECT count(*) FROM information_schema.columns WHERE lower(table_name) = lower(?) AND lower(column_name) = lower(?)",
			tableName, name,
		).Rows()
		if err != nil {
			return nil
		}
		if rows == nil {
			return nil
		}
		defer func() { _ = rows.Close() }()
		if rows.Next() {
			if err := rows.Scan(&count); err != nil {
				return nil
			}
		}
		return nil
	})

	return count > 0
}

// HasIndex checks if an index exists in the database.
func (m Migrator) HasIndex(value interface{}, name string) bool {
	var count int64
	_ = m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if idx := stmt.Schema.LookIndex(name); idx != nil {
				name = idx.Name
			}
		}

		tableIdentifier := ""
		if stmt != nil && stmt.Schema != nil && stmt.Schema.Table != "" {
			tableIdentifier = stmt.Schema.Table
		} else if stmt != nil && stmt.Table != "" {
			tableIdentifier = stmt.Table
		} else {
			tableIdentifier = fmt.Sprint(m.CurrentTable(stmt))
		}
		tableName := normalizeTable(tableIdentifier)
		rows, err := m.DB.Raw(
			"SELECT count(*) FROM information_schema.statistics WHERE lower(table_name) = lower(?) AND lower(index_name) = lower(?)",
			tableName, name,
		).Rows()
		if err != nil {
			return nil
		}
		if rows == nil {
			return nil
		}
		defer func() { _ = rows.Close() }()
		if rows.Next() {
			if err := rows.Scan(&count); err != nil {
				return nil
			}
		}
		return nil
	})

	return count > 0
}

// HasConstraint checks if a constraint exists in the database.
func (m Migrator) HasConstraint(value interface{}, name string) bool {
	var count int64
	_ = m.RunWithValue(value, func(stmt *gorm.Statement) error {
		constraint, table := m.GuessConstraintInterfaceAndTable(stmt, name)
		if constraint != nil {
			name = constraint.GetName()
		}

		// Normalize guessed table name as well
		tableIdentifier := ""
		if table != "" {
			tableIdentifier = table
		} else {
			tableIdentifier = fmt.Sprint(m.CurrentTable(stmt))
		}
		tableName := normalizeTable(tableIdentifier)

		rows, err := m.DB.Raw(
			"SELECT count(*) FROM information_schema.table_constraints WHERE lower(table_name) = lower(?) AND lower(constraint_name) = lower(?)",
			tableName, name,
		).Rows()
		if err != nil {
			return nil
		}
		if rows == nil {
			return nil
		}
		defer func() { _ = rows.Close() }()
		if rows.Next() {
			if err := rows.Scan(&count); err != nil {
				return nil
			}
		}
		return nil
	})

	return count > 0
}

// CreateView creates a database view.
func (m Migrator) CreateView(name string, option gorm.ViewOption) error {
	if option.Query == nil {
		return gorm.ErrSubQueryRequired
	}

	sql := new(strings.Builder)
	sql.WriteString("CREATE ")
	if option.Replace {
		sql.WriteString("OR REPLACE ")
	}
	sql.WriteString("VIEW ")
	m.QuoteTo(sql, name)
	sql.WriteString(" AS ")

	m.DB.Statement.AddVar(sql, option.Query)

	if option.CheckOption != "" {
		sql.WriteString(" ")
		sql.WriteString(option.CheckOption)
	}

	return m.DB.Exec(m.Explain(sql.String(), m.DB.Statement.Vars...)).Error
}

// DropView drops a database view.
func (m Migrator) DropView(name string) error {
	return m.DB.Exec("DROP VIEW IF EXISTS ?", clause.Table{Name: name}).Error
}

// GetTypeAliases returns type aliases for the given database type name.
func (m Migrator) GetTypeAliases(databaseTypeName string) []string {
	aliases := map[string][]string{
		"boolean":   {"bool"},
		"tinyint":   {"int8"},
		"smallint":  {"int16"},
		"integer":   {"int", "int32"},
		"bigint":    {"int64"},
		"utinyint":  {"uint8"},
		"usmallint": {"uint16"},
		"uinteger":  {"uint", "uint32"},
		"ubigint":   {"uint64"},
		"real":      {"float32"},
		"double":    {"float64", "float"},
		"varchar":   {"string"},
		"text":      {"string"},
		"blob":      {"bytes"},
		"timestamp": {"time"},
	}

	return aliases[databaseTypeName]
}

// ColumnTypes returns comprehensive column type information for the given value
func (m Migrator) ColumnTypes(value interface{}) ([]gorm.ColumnType, error) {
	var columnTypes []gorm.ColumnType

	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		// Get table name using the same approach as HasTable
		tableIdentifier := ""
		if stmt != nil && stmt.Schema != nil && stmt.Schema.Table != "" {
			tableIdentifier = stmt.Schema.Table
		} else if stmt != nil && stmt.Table != "" {
			tableIdentifier = stmt.Table
		} else {
			// Get table name from current table - handle different return types
			currentTable := m.CurrentTable(stmt)
			switch v := currentTable.(type) {
			case string:
				tableIdentifier = v
			case clause.Table:
				tableIdentifier = v.Name
			default:
				// If we can't get a table name, try to get all tables and use the first one
				if tables, gErr := m.GetTables(); gErr == nil && len(tables) > 0 {
					tableIdentifier = tables[0]
				}
			}
		}

		if tableIdentifier == "" {
			return nil
		}

		// Normalize the table identifier
		tableName := normalizeTable(tableIdentifier)

		// Build query for this table
		query := `
			SELECT
				c.column_name,
				c.data_type,
				CASE
					WHEN c.character_maximum_length IS NOT NULL THEN c.data_type || '(' || c.character_maximum_length || ')'
					WHEN c.numeric_precision IS NOT NULL AND c.numeric_scale IS NOT NULL THEN c.data_type || '(' || c.numeric_precision || ',' || c.numeric_scale || ')'
					WHEN c.numeric_precision IS NOT NULL THEN c.data_type || '(' || c.numeric_precision || ')'
					ELSE c.data_type
				END as column_type,
				CASE WHEN c.is_nullable = 'YES' THEN true ELSE false END as nullable,
				c.column_default,
				COALESCE(pk.is_primary_key, false) as is_primary_key,
				CASE WHEN c.column_default LIKE '%nextval%' OR c.column_default LIKE '%seq_%' THEN true ELSE false END as is_auto_increment,
				c.character_maximum_length,
				c.numeric_precision,
				c.numeric_scale,
				COALESCE(uk.is_unique, false) as is_unique,
				'' as column_comment
			FROM information_schema.columns c
			LEFT JOIN (
				SELECT kcu.column_name, true as is_primary_key
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
				WHERE tc.constraint_type = 'PRIMARY KEY' AND lower(tc.table_name) = lower(?)
			) pk ON c.column_name = pk.column_name
			LEFT JOIN (
				SELECT kcu.column_name, true as is_unique
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
				WHERE tc.constraint_type = 'UNIQUE' AND lower(tc.table_name) = lower(?)
			) uk ON c.column_name = uk.column_name
			WHERE lower(c.table_name) = lower(?)
			ORDER BY c.ordinal_position
		`

		args := []interface{}{tableName, tableName, tableName}

		rows, err := m.DB.Raw(query, args...).Rows()

		if err != nil {
			return fmt.Errorf("failed to execute foreign key check query: %w", err)
		}
		if rows == nil {
			return nil
		}
		defer func() { _ = rows.Close() }()

		var found int
		for rows.Next() {
			found++
			var (
				columnName, dataType, columnTypeStr, columnComment  string
				columnDefault                                       sql.NullString
				isNullable, isPrimaryKey, isAutoIncrement, isUnique bool
				charMaxLength, numericPrecision, numericScale       sql.NullInt64
			)

			if scanErr := rows.Scan(
				&columnName, &dataType, &columnTypeStr, &isNullable, &columnDefault,
				&isPrimaryKey, &isAutoIncrement, &charMaxLength, &numericPrecision,
				&numericScale, &isUnique, &columnComment,
			); scanErr != nil {
				// Skip malformed rows but continue processing others
				continue
			}

			columnType := &migrator.ColumnType{
				NameValue:          sql.NullString{String: columnName, Valid: true},
				DataTypeValue:      sql.NullString{String: dataType, Valid: true},
				ColumnTypeValue:    sql.NullString{String: columnTypeStr, Valid: true},
				NullableValue:      sql.NullBool{Bool: isNullable, Valid: true},
				PrimaryKeyValue:    sql.NullBool{Bool: isPrimaryKey, Valid: true},
				AutoIncrementValue: sql.NullBool{Bool: isAutoIncrement, Valid: true},
				UniqueValue:        sql.NullBool{Bool: isUnique, Valid: true},
				CommentValue:       sql.NullString{String: columnComment, Valid: columnComment != ""},
				DefaultValueValue:  columnDefault,
				ScanTypeValue:      reflect.TypeOf(""), // Default to string type for safety
			}

			// Set length information defensively
			if charMaxLength.Valid {
				columnType.LengthValue = charMaxLength
			} else {
				// Prefer schema metadata if available (use closure stmt)
				if stmt != nil && stmt.Schema != nil {
					// Look up field by column name
					if f := stmt.Schema.LookUpField(columnName); f != nil && f.Size > 0 {
						columnType.LengthValue = sql.NullInt64{Int64: int64(f.Size), Valid: true}
					} else {
						// Try to parse from column_type string as fallback
						if idx := strings.Index(columnTypeStr, "("); idx > 0 {
							// naive parse between parentheses
							end := strings.Index(columnTypeStr[idx+1:], ")")
							if end > 0 {
								if l, parseErr := strconv.ParseInt(columnTypeStr[idx+1:idx+1+end], 10, 64); parseErr == nil {
									columnType.LengthValue = sql.NullInt64{Int64: l, Valid: true}
								}
							}
						}
					}
				}
			}

			// Set decimal size information
			if numericPrecision.Valid {
				columnType.DecimalSizeValue = numericPrecision
				if numericScale.Valid {
					columnType.ScaleValue = numericScale
				}
			}

			columnTypes = append(columnTypes, columnType)
		}

		return rows.Err()
	})

	if err != nil {
		return columnTypes, fmt.Errorf("failed to get column types: %w", err)
	}
	return columnTypes, nil
}

// TableType returns comprehensive table type information
func (m Migrator) TableType(value interface{}) (gorm.TableType, error) {
	var result *migrator.TableType

	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		// Use Rows() and defensive scanning to avoid nil-row panics
		query := `
			SELECT
				table_schema,
				table_name,
				table_type,
				COALESCE(table_comment, '') as table_comment
			FROM information_schema.tables
			WHERE lower(table_name) = lower(?)
		`

		rows, err := m.DB.Raw(query, stmt.Table).Rows()
		if err != nil {
			return nil
		}
		if rows == nil {
			return nil
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var schemaName, tableName, tableTypeStr, tableComment string
			if scanErr := rows.Scan(&schemaName, &tableName, &tableTypeStr, &tableComment); scanErr != nil {
				continue
			}

			result = &migrator.TableType{
				SchemaValue:  schemaName,
				NameValue:    tableName,
				TypeValue:    tableTypeStr,
				CommentValue: sql.NullString{String: tableComment, Valid: tableComment != ""},
			}
			// Only need first matching row
			break
		}

		return rows.Err()
	})

	if result == nil {
		if err != nil {
			return nil, fmt.Errorf("failed to get table type information: %w", err)
		}
		return nil, fmt.Errorf("table type not found")
	}

	if err != nil {
		return result, fmt.Errorf("failed to get table type information: %w", err)
	}
	return result, nil
}

// Index represents a database index with its properties.
type Index struct {
	TableName   string
	IndexName   string
	ColumnNames []string
	IsUnique    bool
	IsPrimary   bool
	Options     string
}

// Table returns the table name for the index.
func (idx Index) Table() string {
	return idx.TableName
}

// Name returns the name of the index.
func (idx Index) Name() string {
	return idx.IndexName
}

// Columns returns the columns that make up the index.
func (idx Index) Columns() []string {
	return idx.ColumnNames
}

// PrimaryKey returns whether this index is a primary key.
func (idx Index) PrimaryKey() (isPrimaryKey bool, ok bool) {
	return idx.IsPrimary, true
}

// Unique returns whether this index enforces uniqueness.
func (idx Index) Unique() (unique bool, ok bool) {
	return idx.IsUnique, true
}

// Option returns additional options for the index.
func (idx Index) Option() string {
	return idx.Options
}

// GetIndexes returns comprehensive index information for the given value
func (m Migrator) GetIndexes(value interface{}) ([]gorm.Index, error) {
	var indexes []gorm.Index

	err := m.RunWithValue(value, func(_ *gorm.Statement) error {
		// DuckDB may not have complete information_schema.statistics support
		// For now, return empty indexes to avoid errors
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}
	return indexes, nil
}

// BuildIndexOptions builds index options for DuckDB
func (m Migrator) BuildIndexOptions(opts []schema.IndexOption, stmt *gorm.Statement) (results []interface{}) {
	for _, opt := range opts {
		str := stmt.Quote(opt.DBName)
		if opt.Expression != "" {
			str = opt.Expression
		} else if opt.Length > 0 {
			str += fmt.Sprintf("(%d)", opt.Length)
		}

		if opt.Collate != "" {
			str += " COLLATE " + opt.Collate
		}

		if opt.Sort != "" {
			str += " " + opt.Sort
		}
		results = append(results, clause.Expr{SQL: str})
	}
	return
}

// CreateTable overrides the default CreateTable to handle DuckDB-specific auto-increment sequences
func (m Migrator) CreateTable(values ...interface{}) error {
	for _, value := range values {
		if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
			// Get the underlying SQL database connection
			sqlDB, err := m.DB.DB()
			if err != nil {
				return fmt.Errorf("failed to get underlying database: %w", err)
			}

			// Step 1: Create sequences for auto-increment fields
			if stmt.Schema != nil {
				for _, field := range stmt.Schema.Fields {
					if field.PrimaryKey && (field.AutoIncrement || (!field.HasDefaultValue && field.DataType == schema.Uint)) {
						sequenceName := "seq_" + strings.ToLower(stmt.Schema.Table) + "_" + strings.ToLower(field.DBName)
					createSeqSQL := fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS %s START 1", sequenceName)
					_, err := sqlDB.ExecContext(context.Background(), createSeqSQL)
						if err != nil {
							// Ignore "already exists" errors
							if !isAlreadyExistsError(err) {
								return fmt.Errorf("failed to create sequence %s: %w", sequenceName, err)
							}
						}
					}
				}
			}

			// Step 2: Generate CREATE TABLE SQL manually instead of relying on parent migrator
			tableName := stmt.Schema.Table
			if tableName == "" {
				tableName = stmt.Table
			}

			var columns []string
			var primaryKeys []string

			for _, field := range stmt.Schema.Fields {
				columnDef := fmt.Sprintf(`"%s"`, field.DBName)

				// Add data type
				columnDef += " " + m.Dialector.DataTypeOf(field)

				// Add constraints
				if field.NotNull {
					columnDef += " NOT NULL"
				}
				if field.PrimaryKey {
					primaryKeys = append(primaryKeys, fmt.Sprintf(`"%s"`, field.DBName))
				}
				if field.Unique {
					columnDef += " UNIQUE"
				}

				// Handle auto-increment by setting default to nextval
				if field.PrimaryKey && (field.AutoIncrement || (!field.HasDefaultValue && field.DataType == schema.Uint)) {
					sequenceName := "seq_" + strings.ToLower(stmt.Schema.Table) + "_" + strings.ToLower(field.DBName)
					columnDef += fmt.Sprintf(" DEFAULT nextval('%s')", sequenceName)
				}

				columns = append(columns, columnDef)
			}

			// Build CREATE TABLE statement
			createSQL := fmt.Sprintf(`CREATE TABLE "%s" (%s`, tableName, strings.Join(columns, ","))

			// Add primary key constraint
			if len(primaryKeys) > 0 {
				createSQL += fmt.Sprintf(",PRIMARY KEY (%s)", strings.Join(primaryKeys, ","))
			}

			createSQL += ")"

			// Step 3: Execute CREATE TABLE using the underlying SQL connection
			_, err = sqlDB.ExecContext(context.Background(), createSQL)
			if err != nil {
				return fmt.Errorf("failed to create table %s: %w", tableName, err)
			}

			return nil
		}); err != nil {
			return fmt.Errorf("failed to create table for value: %w", err)
		}
	}
	return nil
}
