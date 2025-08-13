package duckdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/marcboeker/go-duckdb/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

type Dialector struct {
	*Config
}

type Config struct {
	DriverName        string
	DSN               string
	Conn              gorm.ConnPool
	DefaultStringSize uint
}

func Open(dsn string) gorm.Dialector {
	return &Dialector{Config: &Config{DSN: dsn}} // Remove DriverName to use default custom driver
}

func New(config Config) gorm.Dialector {
	return &Dialector{Config: &config}
}

func (dialector Dialector) Name() string {
	return "duckdb"
}

func init() {
	sql.Register("duckdb-gorm", &convertingDriver{&duckdb.Driver{}})
}

// Custom driver that converts time pointers at the lowest level
type convertingDriver struct {
	driver.Driver
}

func (d *convertingDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.Driver.Open(name)
	if err != nil {
		return nil, err
	}
	return &convertingConn{conn}, nil
}

type convertingConn struct {
	driver.Conn
}

func (c *convertingConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.Conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &convertingStmt{stmt}, nil
}

func (c *convertingConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if prepCtx, ok := c.Conn.(driver.ConnPrepareContext); ok {
		stmt, err := prepCtx.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		return &convertingStmt{stmt}, nil
	}
	return c.Prepare(query)
}

func (c *convertingConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	// Convert to context-aware version - this is the recommended approach
	namedArgs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		namedArgs[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   arg,
		}
	}
	return c.ExecContext(context.Background(), query, namedArgs)
}

func (c *convertingConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if execCtx, ok := c.Conn.(driver.ExecerContext); ok {
		convertedArgs := convertNamedValues(args)
		return execCtx.ExecContext(ctx, query, convertedArgs)
	}
	// Fallback to non-context version
	values := make([]driver.Value, len(args))
	for i, arg := range args {
		values[i] = arg.Value
	}
	return c.Exec(query, values)
}

func (c *convertingConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	// Convert to context-aware version - this is the recommended approach
	namedArgs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		namedArgs[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   arg,
		}
	}
	return c.QueryContext(context.Background(), query, namedArgs)
}

func (c *convertingConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if queryCtx, ok := c.Conn.(driver.QueryerContext); ok {
		convertedArgs := convertNamedValues(args)
		return queryCtx.QueryContext(ctx, query, convertedArgs)
	}
	// Fallback to non-context version
	values := make([]driver.Value, len(args))
	for i, arg := range args {
		values[i] = arg.Value
	}
	return c.Query(query, values)
}

type convertingStmt struct {
	driver.Stmt
}

func (s *convertingStmt) Exec(args []driver.Value) (driver.Result, error) {
	// Convert to context-aware version - this is the recommended approach
	namedArgs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		namedArgs[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   arg,
		}
	}
	return s.ExecContext(context.Background(), namedArgs)
}

func (s *convertingStmt) Query(args []driver.Value) (driver.Rows, error) {
	// Convert to context-aware version - this is the recommended approach
	namedArgs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		namedArgs[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   arg,
		}
	}
	return s.QueryContext(context.Background(), namedArgs)
}

func (s *convertingStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if stmtCtx, ok := s.Stmt.(driver.StmtExecContext); ok {
		convertedArgs := convertNamedValues(args)
		return stmtCtx.ExecContext(ctx, convertedArgs)
	}
	// Direct fallback without using deprecated methods
	convertedArgs := convertNamedValues(args)
	values := make([]driver.Value, len(convertedArgs))
	for i, arg := range convertedArgs {
		values[i] = arg.Value
	}
	//nolint:staticcheck // Fallback required for drivers that don't implement StmtExecContext
	return s.Stmt.Exec(values)
}

func (s *convertingStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if stmtCtx, ok := s.Stmt.(driver.StmtQueryContext); ok {
		convertedArgs := convertNamedValues(args)
		return stmtCtx.QueryContext(ctx, convertedArgs)
	}
	// Direct fallback without using deprecated methods
	convertedArgs := convertNamedValues(args)
	values := make([]driver.Value, len(convertedArgs))
	for i, arg := range convertedArgs {
		values[i] = arg.Value
	}
	//nolint:staticcheck // Fallback required for drivers that don't implement StmtQueryContext
	return s.Stmt.Query(values)
}

// Convert driver.NamedValue slice
func convertNamedValues(args []driver.NamedValue) []driver.NamedValue {
	converted := make([]driver.NamedValue, len(args))

	for i, arg := range args {
		converted[i] = arg

		if timePtr, ok := arg.Value.(*time.Time); ok {
			if timePtr == nil {
				converted[i].Value = nil
			} else {
				converted[i].Value = *timePtr
			}
		} else if isSlice(arg.Value) {
			// Convert Go slices to DuckDB array format
			if arrayStr, err := formatSliceForDuckDB(arg.Value); err == nil {
				converted[i].Value = arrayStr
			}
		}
	}

	return converted
}

// isSlice checks if a value is a slice (but not string or []byte)
func isSlice(v interface{}) bool {
	if v == nil {
		return false
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return false
	}

	// Don't treat strings or []byte as arrays
	switch v.(type) {
	case string, []byte:
		return false
	default:
		return true
	}
}

// Initialize implements gorm.Dialector
func (dialector Dialector) Initialize(db *gorm.DB) error {
	callbacks.RegisterDefaultCallbacks(db, &callbacks.Config{})

	// Override the create callback to use RETURNING for auto-increment fields
	db.Callback().Create().Before("gorm:create").Register("duckdb:before_create", beforeCreateCallback)
	db.Callback().Create().Replace("gorm:create", createCallback)

	if dialector.DefaultStringSize == 0 {
		dialector.DefaultStringSize = 256
	}

	if dialector.DriverName == "" {
		dialector.DriverName = "duckdb-gorm"
	}

	if dialector.Conn != nil {
		db.ConnPool = dialector.Conn
	} else {
		connPool, err := sql.Open(dialector.DriverName, dialector.DSN)
		if err != nil {
			return err
		}
		db.ConnPool = connPool
	}

	return nil
}

func (dialector Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return Migrator{
		migrator.Migrator{
			Config: migrator.Config{
				DB:                          db,
				Dialector:                   dialector,
				CreateIndexAfterCreateTable: true,
			},
		},
	}
}

func (dialector Dialector) DataTypeOf(field *schema.Field) string {
	switch field.DataType {
	case schema.Bool:
		return "BOOLEAN"
	case schema.Int:
		switch field.Size {
		case 8:
			return "TINYINT"
		case 16:
			return "SMALLINT"
		case 32:
			return "INTEGER"
		default:
			return "BIGINT"
		}
	case schema.Uint:
		// For primary keys, use INTEGER to enable auto-increment in DuckDB
		if field.PrimaryKey {
			return "INTEGER"
		}
		// Use signed integers for uint to ensure foreign key compatibility
		// DuckDB has issues with foreign keys between signed and unsigned types
		switch field.Size {
		case 8:
			return "TINYINT"
		case 16:
			return "SMALLINT"
		case 32:
			return "INTEGER"
		default:
			return "BIGINT"
		}
	case schema.Float:
		if field.Size == 32 {
			return "REAL"
		}
		return "DOUBLE"
	case schema.String:
		size := field.Size
		if size == 0 {
			if dialector.DefaultStringSize > 0 && dialector.DefaultStringSize <= 65535 {
				size = int(dialector.DefaultStringSize) //nolint:gosec // Safe conversion, bounds already checked
			} else {
				size = 256 // Safe default
			}
		}
		if size > 0 && size < 65536 {
			return fmt.Sprintf("VARCHAR(%d)", size)
		}
		return "TEXT"
	case schema.Time:
		return "TIMESTAMP"
	case schema.Bytes:
		return "BLOB"
	}

	// Check if it's an array type
	if strings.HasSuffix(string(field.DataType), "[]") {
		baseType := strings.TrimSuffix(string(field.DataType), "[]")
		return fmt.Sprintf("%s[]", baseType)
	}

	return string(field.DataType)
}

func (dialector Dialector) DefaultValueOf(field *schema.Field) clause.Expression {
	if field.HasDefaultValue && (field.DefaultValueInterface != nil || field.DefaultValue != "") {
		if field.DefaultValueInterface != nil {
			switch v := field.DefaultValueInterface.(type) {
			case bool:
				if v {
					return clause.Expr{SQL: "TRUE"}
				}
				return clause.Expr{SQL: "FALSE"}
			default:
				return clause.Expr{SQL: fmt.Sprintf("'%v'", field.DefaultValueInterface)}
			}
		} else if field.DefaultValue != "" && field.DefaultValue != "(-)" {
			if field.DataType == schema.Bool {
				if strings.ToLower(field.DefaultValue) == "true" {
					return clause.Expr{SQL: "TRUE"}
				}
				return clause.Expr{SQL: "FALSE"}
			}
			return clause.Expr{SQL: field.DefaultValue}
		}
	}
	return clause.Expr{}
}

func (dialector Dialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	_ = writer.WriteByte('?')
}

func (dialector Dialector) QuoteTo(writer clause.Writer, str string) {
	var (
		underQuoted, selfQuoted bool
		continuousBacktick      int8
		shiftDelimiter          int8
	)

	for _, v := range []byte(str) {
		switch v {
		case '"':
			continuousBacktick++
			if continuousBacktick == 2 {
				_, _ = writer.WriteString(`""`)
				continuousBacktick = 0
			}
		case '.':
			if continuousBacktick > 0 || !selfQuoted {
				shiftDelimiter = 0
				underQuoted = false
				continuousBacktick = 0
				_ = writer.WriteByte('"')
			}
			_ = writer.WriteByte(v)
			continue
		default:
			if shiftDelimiter-continuousBacktick <= 0 && !underQuoted {
				_ = writer.WriteByte('"')
				underQuoted = true
				if selfQuoted = continuousBacktick > 0; selfQuoted {
					continuousBacktick -= 1
				}
			}

			for ; continuousBacktick > 0; continuousBacktick -= 1 {
				_, _ = writer.WriteString(`""`)
			}

			_ = writer.WriteByte(v)
		}
		shiftDelimiter++
	}

	if continuousBacktick > 0 && !selfQuoted {
		_, _ = writer.WriteString(`""`)
	}
	_ = writer.WriteByte('"')
}

func (dialector Dialector) Explain(sql string, vars ...interface{}) string {
	return logger.ExplainSQL(sql, nil, `"`, vars...)
}

func (dialector Dialector) SavePoint(tx *gorm.DB, name string) error {
	return tx.Exec("SAVEPOINT " + name).Error
}

func (dialector Dialector) RollbackTo(tx *gorm.DB, name string) error {
	return tx.Exec("ROLLBACK TO SAVEPOINT " + name).Error
}

// beforeCreateCallback prepares the statement for auto-increment handling
func beforeCreateCallback(db *gorm.DB) {
	// Nothing special needed here, just ensuring the statement is prepared
}

// createCallback handles INSERT operations with RETURNING for auto-increment fields
func createCallback(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	if db.Statement.Schema != nil {
		var hasAutoIncrement bool
		var autoIncrementField *schema.Field

		// Check if we have auto-increment primary key
		for _, field := range db.Statement.Schema.PrimaryFields {
			if field.AutoIncrement {
				hasAutoIncrement = true
				autoIncrementField = field
				break
			}
		}

		if hasAutoIncrement {
			// Build custom INSERT with RETURNING
			sql, vars := buildInsertSQL(db, autoIncrementField)
			if sql != "" {
				// Execute with RETURNING to get the auto-generated ID
				var id int64
				if err := db.Raw(sql, vars...).Row().Scan(&id); err != nil {
					db.AddError(err)
					return
				}

				// Set the ID in the model using GORM's ReflectValue
				if db.Statement.ReflectValue.IsValid() && db.Statement.ReflectValue.CanAddr() {
					modelValue := db.Statement.ReflectValue

					if idField := modelValue.FieldByName(autoIncrementField.Name); idField.IsValid() && idField.CanSet() {
						// Handle different integer types
						switch idField.Kind() {
						case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
							idField.SetUint(uint64(id))
						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
							idField.SetInt(id)
						}
					}
				}

				db.Statement.RowsAffected = 1
				return
			}
		}
	}

	// Fall back to default behavior for non-auto-increment cases
	if db.Statement.SQL.String() == "" {
		db.Statement.Build("INSERT")
	}

	if result, err := db.Statement.ConnPool.ExecContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...); err != nil {
		db.AddError(err)
	} else {
		if rows, _ := result.RowsAffected(); rows > 0 {
			db.Statement.RowsAffected = rows
		}
	}
}

// buildInsertSQL creates an INSERT statement with RETURNING for auto-increment fields
func buildInsertSQL(db *gorm.DB, autoIncrementField *schema.Field) (string, []interface{}) {
	if db.Statement.Schema == nil {
		return "", nil
	}

	var fields []string
	var placeholders []string
	var values []interface{}

	// Build field list excluding auto-increment field
	for _, field := range db.Statement.Schema.Fields {
		if field.DBName == autoIncrementField.DBName {
			continue // Skip auto-increment field
		}

		// Get the value for this field
		fieldValue := db.Statement.ReflectValue.FieldByName(field.Name)
		if !fieldValue.IsValid() {
			continue
		}

		// For optional fields, skip zero values
		if field.HasDefaultValue && fieldValue.Kind() != reflect.String && fieldValue.IsZero() {
			continue
		}

		fields = append(fields, db.Statement.Quote(field.DBName))
		placeholders = append(placeholders, "?")
		values = append(values, fieldValue.Interface())
	}

	if len(fields) == 0 {
		return "", nil
	}

	tableName := db.Statement.Quote(db.Statement.Table)
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s",
		tableName,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
		db.Statement.Quote(autoIncrementField.DBName))

	return sql, values
}
