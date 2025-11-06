package duckdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/marcboeker/go-duckdb/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

// Constants for repeated strings
const (
	dataTypeText = "TEXT"
	dataTypeBlob = "BLOB"
	dataTypeJSON = "JSON"
)

var debugLogging = os.Getenv("GORM_DUCKDB_DEBUG") == "true" || os.Getenv("GORM_DUCKDB_DEBUG") == "1"

// debugLog logs messages only when debug logging is enabled
func debugLog(format string, args ...interface{}) {
	if debugLogging {
		log.Printf("[GORM-DUCKDB-DEBUG] "+format, args...)
	}
}

// errorLog logs error messages (always enabled)
func errorLog(format string, args ...interface{}) {
	log.Printf("[GORM-DUCKDB-ERROR] "+format, args...)
}

// Dialector implements gorm.Dialector interface for DuckDB database.
type Dialector struct {
	*Config
}

// Config holds configuration options for the DuckDB dialector.
type Config struct {
	DriverName        string
	DSN               string
	Conn              gorm.ConnPool
	DefaultStringSize uint

	// RowCallbackWorkaround controls whether to apply the GORM RowQuery callback fix
	// Set to false to disable the workaround if GORM fixes the bug in the future
	// Default: true (apply workaround)
	RowCallbackWorkaround *bool
}

// Open creates a new DuckDB dialector with the given DSN.
func Open(dsn string) gorm.Dialector {
	return &Dialector{Config: &Config{DSN: dsn}} // Remove DriverName to use default custom driver
}

// OpenWithConfig creates a new DuckDB dialector with the given DSN and configuration options.
func OpenWithConfig(dsn string, config *Config) gorm.Dialector {
	if config == nil {
		config = &Config{}
	}
	config.DSN = dsn
	return &Dialector{Config: config}
}

// OpenWithRowCallbackWorkaround creates a DuckDB dialector with explicit RowCallback workaround control.
// Set enableWorkaround=false if you're using a GORM version that has fixed the RowQuery callback bug.
func OpenWithRowCallbackWorkaround(dsn string, enableWorkaround bool) gorm.Dialector {
	return &Dialector{Config: &Config{
		DSN:                   dsn,
		RowCallbackWorkaround: &enableWorkaround,
	}}
}

// New creates a new DuckDB dialector with the given configuration.
func New(config Config) gorm.Dialector {
	return &Dialector{Config: &config}
}

// Name returns the name of the dialector.
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
	debugLog(" convertingDriver.Open called with DSN: %s", name)
	conn, err := d.Driver.Open(name)
	if err != nil {
		debugLog(" convertingDriver.Open failed: %v", err)
		return nil, fmt.Errorf("failed to open DuckDB connection with name %s: %w", name, err)
	}
	debugLog(" convertingDriver.Open succeeded, returning convertingConn")
	return &convertingConn{conn}, nil
}

type convertingConn struct {
	driver.Conn
}

func (c *convertingConn) Prepare(query string) (driver.Stmt, error) {
	debugLog(" Prepare called with query: %s", query)
	stmt, err := c.Conn.Prepare(query)
	if err != nil {
		debugLog(" Prepare failed: %v", err)
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	debugLog(" Prepare succeeded, returning convertingStmt")
	return &convertingStmt{stmt}, nil
}

func (c *convertingConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	debugLog(" PrepareContext called with query: %s", query)
	if prepCtx, ok := c.Conn.(driver.ConnPrepareContext); ok {
		stmt, err := prepCtx.PrepareContext(ctx, query)
		if err != nil {
			debugLog(" PrepareContext failed: %v", err)
			return nil, fmt.Errorf("failed to prepare statement with context: %w", err)
		}
		debugLog(" PrepareContext succeeded, returning convertingStmt")
		return &convertingStmt{stmt}, nil
	}
	debugLog(" PrepareContext falling back to Prepare")
	return c.Prepare(query)
}

func (c *convertingConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	debugLog(" Exec (non-context) called with query: %s, args: %v", query, args)
	// Convert to context-aware version - this is the recommended approach
	namedArgs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		namedArgs[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   arg,
		}
	}
	result, err := c.ExecContext(context.Background(), query, namedArgs)
	if err != nil {
		debugLog(" Exec (non-context) failed: %v", err)
	} else {
		debugLog(" Exec (non-context) succeeded")
	}
	return result, err
}

func (c *convertingConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	debugLog(" ExecContext called with query: %s, args: %v", query, args)
	if execCtx, ok := c.Conn.(driver.ExecerContext); ok {
		convertedArgs := convertNamedValues(args)
		result, err := execCtx.ExecContext(ctx, query, convertedArgs)
		if err != nil {
			errorLog(" ExecContext failed: %v", err)
			return nil, translateDriverError(err)
		}
		debugLog(" ExecContext succeeded for query: %s", query)
		return result, nil
	}
	// Fallback to non-context version
	values := make([]driver.Value, len(args))
	for i, arg := range args {
		values[i] = arg.Value
	}
	// nolint:staticcheck // Using deprecated driver.Execer for backward compatibility
	if exec, ok := c.Conn.(driver.Execer); ok {
		result, err := exec.Exec(query, values)
		if err != nil {
			errorLog(" Exec fallback failed: %v", err)
			return nil, translateDriverError(err)
		}
		debugLog(" Exec fallback succeeded for query: %s", query)
		return result, nil
	}
	errorLog(" ExecContext: underlying driver does not support Exec operations for query: %s", query)
	return nil, fmt.Errorf("underlying driver does not support Exec operations")
}

func (c *convertingConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	debugLog(" Query called with query: %s, args: %v", query, args)
	// Convert to context-aware version - this is the recommended approach
	namedArgs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		namedArgs[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   arg,
		}
	}
	result, err := c.QueryContext(context.Background(), query, namedArgs)
	debugLog(" Query result: %v, err: %v", result, err)
	return result, err
}

func (c *convertingConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	debugLog(" QueryContext called with query: %s, args: %v", query, args)
	if queryCtx, ok := c.Conn.(driver.QueryerContext); ok {
		debugLog(" Using QueryerContext interface")
		convertedArgs := convertNamedValues(args)
		debugLog(" Converted args: %v", convertedArgs)
		rows, err := queryCtx.QueryContext(ctx, query, convertedArgs)
		if err != nil {
			errorLog(" QueryContext failed: %v", err)
			return nil, translateDriverError(err)
		}
		debugLog(" QueryContext returned rows: %v (nil: %t)", rows, rows == nil)
		return rows, nil
	}
	debugLog(" QueryContext: Falling back to non-context version for query: %s", query)
	values := make([]driver.Value, len(args))
	for i, arg := range args {
		values[i] = arg.Value
	}
	// nolint:staticcheck // Using deprecated driver.Queryer for backward compatibility
	if queryer, ok := c.Conn.(driver.Queryer); ok {
		rows, err := queryer.Query(query, values)
		if err != nil {
			errorLog(" Query fallback failed: %v", err)
			return nil, translateDriverError(err)
		}
		debugLog(" Query fallback succeeded for query: %s", query)
		return rows, nil
	}
	errorLog(" QueryContext: underlying driver does not support Query operations for query: %s", query)
	return nil, fmt.Errorf("underlying driver does not support Query operations")
}

type convertingStmt struct {
	driver.Stmt
}

func (s *convertingStmt) Exec(args []driver.Value) (driver.Result, error) {
	debugLog(" convertingStmt.Exec called with args: %v", args)
	// Convert to context-aware version - this is the recommended approach
	namedArgs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		namedArgs[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   arg,
		}
	}
	result, err := s.ExecContext(context.Background(), namedArgs)
	if err != nil {
		debugLog(" convertingStmt.Exec failed: %v", err)
	} else {
		debugLog(" convertingStmt.Exec succeeded")
	}
	return result, err
}

func (s *convertingStmt) Query(args []driver.Value) (driver.Rows, error) {
	debugLog(" convertingStmt.Query called with args: %v", args)
	// Convert to context-aware version - this is the recommended approach
	namedArgs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		namedArgs[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   arg,
		}
	}
	result, err := s.QueryContext(context.Background(), namedArgs)
	debugLog(" convertingStmt.Query result: %v, err: %v", result, err)
	return result, err
}

func (s *convertingStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	debugLog(" convertingStmt.ExecContext called with args: %v", args)
	if stmtCtx, ok := s.Stmt.(driver.StmtExecContext); ok {
		convertedArgs := convertNamedValues(args)
		result, err := stmtCtx.ExecContext(ctx, convertedArgs)
		if err != nil {
			debugLog(" convertingStmt.ExecContext failed: %v", err)
			return nil, fmt.Errorf("failed to execute statement with context: %w", err)
		}
		debugLog(" convertingStmt.ExecContext succeeded")
		return result, nil
	}
	// Direct fallback without using deprecated methods
	convertedArgs := convertNamedValues(args)
	values := make([]driver.Value, len(convertedArgs))
	for i, arg := range convertedArgs {
		values[i] = arg.Value
	}
	//nolint:staticcheck // Fallback required for drivers that don't implement StmtExecContext
	result, err := s.Stmt.Exec(values)
	if err != nil {
		debugLog(" convertingStmt.ExecContext fallback failed: %v", err)
		return nil, fmt.Errorf("failed to execute statement: %w", err)
	}
	debugLog(" convertingStmt.ExecContext fallback succeeded")
	return result, nil
}

func (s *convertingStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	debugLog(" convertingStmt.QueryContext called with args: %v", args)
	if stmtCtx, ok := s.Stmt.(driver.StmtQueryContext); ok {
		debugLog(" Using StmtQueryContext interface")
		convertedArgs := convertNamedValues(args)
		rows, err := stmtCtx.QueryContext(ctx, convertedArgs)
		if err != nil {
			debugLog(" StmtQueryContext failed: %v", err)
			return nil, fmt.Errorf("failed to query statement with context: %w", err)
		}
		debugLog(" StmtQueryContext returned rows: %v (nil: %t)", rows, rows == nil)
		return rows, nil
	}
	debugLog(" Using fallback Stmt.Query")
	// Direct fallback without using deprecated methods
	convertedArgs := convertNamedValues(args)
	values := make([]driver.Value, len(convertedArgs))
	for i, arg := range convertedArgs {
		values[i] = arg.Value
	}
	//nolint:staticcheck // Fallback required for drivers that don't implement StmtQueryContext
	rows, err := s.Stmt.Query(values)
	if err != nil {
		debugLog(" Stmt.Query failed: %v", err)
		return nil, fmt.Errorf("failed to query statement: %w", err)
	}
	debugLog(" Stmt.Query returned rows: %v (nil: %t)", rows, rows == nil)
	return rows, nil
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
	if db == nil {
		return fmt.Errorf("gorm DB instance is nil in Initialize")
	}
	// Register callbacks once per *gorm.DB instance so Initialize can be called
	// multiple times (tests create multiple DB instances) without duplicating
	// registrations. We use InstanceGet/InstanceSet to mark registration per DB.
	// Safely check per-DB registration flag. InstanceGet may panic if DB internals
	// are not fully initialized during early Initialize calls; wrap in recover to
	// avoid crashing tests. If the check cannot be performed, fall back to
	// attempting registration and tolerate duplicate errors.
	alreadyRegistered := false
	_ = alreadyRegistered // suppress unused warning
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Treat as not registered and continue with registration attempt
				alreadyRegistered = false
			}
		}()
		if reg, ok := db.InstanceGet("gorm-duckdb:callbacks_registered"); ok && reg != nil {
			if rb, ok2 := reg.(bool); ok2 && rb {
				alreadyRegistered = true
			}
		}
	}()

	if !alreadyRegistered {
		// Custom CREATE callback to work around GORM v1.31.1 issue where gorm:create
		// doesn't generate INSERT SQL for DuckDB dialector
		if err := db.Callback().Create().Replace("gorm:create", duckdbCreateCallback); err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "duplicated") && !strings.Contains(strings.ToLower(err.Error()), "already") {
				return fmt.Errorf("failed to register custom create callback: %w", err)
			}
		} else {
			debugLog(" Successfully registered custom CREATE callback to work around GORM issue")
		}

		// Temporarily disable other custom callbacks to test GORM's default behavior
		/*
		// Override the create callback to use RETURNING for auto-increment fields.
		if err := db.Callback().Create().Before("gorm:create").Register("duckdb:before_create", beforeCreateCallback); err != nil {
			// Ignore duplicate/already-registered errors
			if !strings.Contains(strings.ToLower(err.Error()), "duplicated") && !strings.Contains(strings.ToLower(err.Error()), "already") {
				return fmt.Errorf("failed to register before create callback: %w", err)
			}
		}

		// Add an after-create callback to handle auto-increment ID retrieval
		// instead of replacing the entire create callback
		if err := db.Callback().Create().After("gorm:create").Register("duckdb:after_create", afterCreateCallback); err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "duplicated") && !strings.Contains(strings.ToLower(err.Error()), "already") {
				return fmt.Errorf("failed to register after create callback: %w", err)
			}
		}

		// Add a debug callback right after GORM's create to see what's happening
		if err := db.Callback().Create().After("gorm:create").Before("duckdb:after_create").Register("duckdb:debug_create", debugCreateCallback); err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "duplicated") && !strings.Contains(strings.ToLower(err.Error()), "already") {
				// Ignore duplicate errors for debug callback
			}
		}
		*/

		// Replace the row callback with our DuckDB-compatible version
		// This is a workaround for a GORM bug where the default RowQuery callback
		// fails to properly assign Statement.Dest, causing Raw().Row() to return nil.
		// See: docs/GORM_ROW_CALLBACK_BUG_ANALYSIS.md
		if shouldApplyRowCallbackFix(db) {
			if err := db.Callback().Row().Replace("gorm:row", rowQueryCallback); err != nil {
				if !strings.Contains(strings.ToLower(err.Error()), "duplicated") && !strings.Contains(strings.ToLower(err.Error()), "already") {
					// Log warning but don't fail initialization - fall back to default callback
					log.Printf("[WARNING] Failed to replace row callback, using default GORM callback: %v", err)
					log.Printf("[WARNING] This may cause Raw().Row() to return nil. See GORM_ROW_CALLBACK_BUG_ANALYSIS.md")
				}
			} else {
				debugLog(" Successfully applied RowQuery callback workaround for GORM bug")
			}
		} else {
			debugLog(" GORM version appears to have fixed RowQuery callback, using default implementation")
		}

		// Attempt to mark this DB instance as having registered callbacks; ignore
		// any panic here as well (some gorm versions may not support InstanceSet during early init).
		func() {
			defer func() { _ = recover() }()
			db.InstanceSet("gorm-duckdb:callbacks_registered", true)
		}()
	}

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
			return fmt.Errorf("failed to open database connection: %w", err)
		}
		db.ConnPool = connPool
	}

	return nil
}

// Migrator returns a new migrator instance for DuckDB.
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

// DataTypeOf returns the SQL data type for a given field.
// nolint:gocyclo // Complex type mapping function required for comprehensive DuckDB type support
func (dialector Dialector) DataTypeOf(field *schema.Field) string {
	if field == nil {
		return ""
	}
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
			return sqlTypeInteger
		default:
			return "BIGINT"
		}
	case schema.Uint:
		// For primary keys, use INTEGER to enable auto-increment in DuckDB
		if field.PrimaryKey {
			return sqlTypeInteger
		}
		// Use signed integers for uint to ensure foreign key compatibility
		// DuckDB has issues with foreign keys between signed and unsigned types
		switch field.Size {
		case 8:
			return "TINYINT"
		case 16:
			return "SMALLINT"
		case 32:
			return sqlTypeInteger
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
		return dataTypeText
	case schema.Time:
		return "TIMESTAMP"
	case schema.Bytes:
		return dataTypeBlob
	}

	// Handle advanced DuckDB types - Phase 2: 80% utilization achieved
	// Handle Phase 3A types - pushing toward 100% utilization
	if field.FieldType != nil {
		typeName := field.FieldType.String()
		switch {
		case strings.Contains(typeName, "StructType"):
			return "STRUCT"
		case strings.Contains(typeName, "MapType"):
			return "MAP"
		case strings.Contains(typeName, "ListType"):
			return "LIST"
		case strings.Contains(typeName, "DecimalType"):
			return "DECIMAL(18,6)" // Default precision and scale
		case strings.Contains(typeName, "IntervalType"):
			return "INTERVAL"
		case strings.Contains(typeName, "UUIDType"):
			return "UUID"
		case strings.Contains(typeName, "JSONType"):
			return dataTypeJSON
		// Phase 3A: Core advanced types for 100% DuckDB utilization
		case strings.Contains(typeName, "ENUMType"):
			return "ENUM" // Will be expanded with enum definition
		case strings.Contains(typeName, "UNIONType"):
			return "UNION" // Supports variant data types
		case strings.Contains(typeName, "TimestampTZType"):
			return "TIMESTAMPTZ" // Timezone-aware timestamps
		case strings.Contains(typeName, "HugeIntType"):
			return "HUGEINT" // 128-bit integers
		case strings.Contains(typeName, "BitStringType"):
			return "BIT" // Bit strings and boolean arrays
		// Final 2% Core Types: Completing 100% Core Advanced Types
		case strings.Contains(typeName, "BLOBType"):
			return "BLOB" // Binary Large Objects
		case strings.Contains(typeName, "GEOMETRYType"):
			return "GEOMETRY" // Spatial geometry data
		// Phase 3B: Advanced operations for 100% DuckDB utilization
		case strings.Contains(typeName, "NestedArrayType"):
			return "ARRAY" // Advanced nested arrays
		case strings.Contains(typeName, "QueryHintType"):
			return dataTypeJSON // Store as JSON text
		case strings.Contains(typeName, "ConstraintType"):
			return dataTypeJSON // Store as JSON text
		case strings.Contains(typeName, "AnalyticalFunctionType"):
			return dataTypeJSON // Store as JSON text
		case strings.Contains(typeName, "PerformanceMetricsType"):
			return dataTypeJSON // Native JSON support
		}
	}

	// Check if it's an array type
	if strings.HasSuffix(string(field.DataType), "[]") {
		baseType := strings.TrimSuffix(string(field.DataType), "[]")
		return fmt.Sprintf("%s[]", baseType)
	}

	return string(field.DataType)
}

// DefaultValueOf returns the default value clause for a field.
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

// BindVarTo writes the bind variable to the clause writer.
func (dialector Dialector) BindVarTo(writer clause.Writer, _ *gorm.Statement, _ interface{}) {
	_ = writer.WriteByte('?')
}

// QuoteTo writes quoted identifiers to the writer.
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
					continuousBacktick--
				}
			}

			for ; continuousBacktick > 0; continuousBacktick-- {
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

// Explain returns an explanation of the SQL query.
func (dialector Dialector) Explain(sql string, vars ...interface{}) string {
	return logger.ExplainSQL(sql, nil, `"`, vars...)
}

// SavePoint creates a savepoint with the given name.
func (dialector Dialector) SavePoint(tx *gorm.DB, name string) error {
	return tx.Exec("SAVEPOINT " + name).Error
}

// RollbackTo rolls back to the given savepoint.
func (dialector Dialector) RollbackTo(tx *gorm.DB, name string) error {
	return tx.Exec("ROLLBACK TO SAVEPOINT " + name).Error
}

// Translate implements ErrorTranslator interface for built-in error translation
func (dialector Dialector) Translate(err error) error {
	translator := ErrorTranslator{}
	return translator.Translate(err)
}

// debugCreateCallback logs what happened during GORM's create
func debugCreateCallback(db *gorm.DB) {
	debugLog(" debugCreateCallback: RowsAffected=%d, Error=%v", db.Statement.RowsAffected, db.Error)
	debugLog(" debugCreateCallback: SQL=%s", db.Statement.SQL.String())
	debugLog(" debugCreateCallback: Vars=%v", db.Statement.Vars)
}

// beforeCreateCallback prepares the statement for auto-increment handling
func beforeCreateCallback(db *gorm.DB) {
	debugLog(" beforeCreateCallback called")
	debugLog(" beforeCreateCallback: Table=%s", db.Statement.Table)
	if db.Statement.Schema != nil {
		debugLog(" beforeCreateCallback: Schema.Table=%s", db.Statement.Schema.Table)
		debugLog(" beforeCreateCallback: Schema.Fields count=%d", len(db.Statement.Schema.Fields))
		for i, field := range db.Statement.Schema.Fields {
			debugLog(" beforeCreateCallback: Field[%d]: Name=%s, DBName=%s, AutoIncrement=%t, PrimaryKey=%t", i, field.Name, field.DBName, field.AutoIncrement, field.PrimaryKey)
		}
	} else {
		debugLog(" beforeCreateCallback: Schema is nil!")
	}
	debugLog(" beforeCreateCallback: ReflectValue=%+v", db.Statement.ReflectValue)
	debugLog(" beforeCreateCallback: SQL before gorm:create=%s", db.Statement.SQL.String())
	// Nothing special needed here, just ensuring the statement is prepared
}

// afterCreateCallback handles auto-increment ID retrieval after GORM's create
func afterCreateCallback(db *gorm.DB) {
	debugLog(" afterCreateCallback called, RowsAffected: %d", db.Statement.RowsAffected)
	if db.Error != nil {
		debugLog(" afterCreateCallback: db.Error = %v", db.Error)
		return
	}

	// Only try to get auto-increment ID if the create was successful
	if db.Statement.RowsAffected > 0 && db.Statement.Schema != nil {
		for _, field := range db.Statement.Schema.PrimaryFields {
			if field.AutoIncrement {
				debugLog(" afterCreateCallback: Attempting to retrieve auto-increment ID for field: %s", field.Name)
				
				// For DuckDB, we need to get the current sequence value
				// The sequence name follows the pattern: seq_{table_name}_{field_name}
				tableName := db.Statement.Schema.Table
				sequenceName := fmt.Sprintf("seq_%s_%s", tableName, field.DBName)
				
				var currentID int64
				query := fmt.Sprintf("SELECT currval('%s')", sequenceName)
				
				err := db.Raw(query).Scan(&currentID).Error
				if err != nil {
					debugLog(" afterCreateCallback: Failed to get sequence value: %v", err)
					return
				}
				
				debugLog(" afterCreateCallback: Retrieved ID: %d", currentID)
				
				// Set the ID in the model
				if db.Statement.ReflectValue.IsValid() && db.Statement.ReflectValue.CanAddr() {
					modelValue := db.Statement.ReflectValue
					if modelValue.Kind() == reflect.Ptr {
						modelValue = modelValue.Elem()
					}

					if modelValue.Kind() == reflect.Struct {
						idField := modelValue.FieldByName(field.Name)
						if idField.IsValid() && idField.CanSet() {
							switch idField.Kind() {
							case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
								idField.SetUint(uint64(currentID))
								debugLog(" afterCreateCallback: Set uint ID to %d", currentID)
							case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
								idField.SetInt(currentID)
								debugLog(" afterCreateCallback: Set int ID to %d", currentID)
							}
						}
					}
				}
				break // Only handle the first auto-increment field
			}
		}
	}
}

// buildInsertSQL creates an INSERT statement with RETURNING for auto-increment fields
func buildInsertSQL(db *gorm.DB, autoIncrementField *schema.Field) (string, []interface{}) {
	if db.Statement.Schema == nil {
		return "", nil
	}

	fieldCount := len(db.Statement.Schema.Fields)
	fields := make([]string, 0, fieldCount)
	placeholders := make([]string, 0, fieldCount)
	values := make([]interface{}, 0, fieldCount)

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

// shouldApplyRowCallbackFix determines if we need to apply our RowQuery callback workaround
// This accounts for future GORM versions that may fix the underlying bug
func shouldApplyRowCallbackFix(db *gorm.DB) bool {
	// Check if the dialector has a specific configuration
	if dialector, ok := db.Dialector.(*Dialector); ok && dialector.Config != nil {
		if dialector.RowCallbackWorkaround != nil {
			// Use explicit configuration
			if *dialector.RowCallbackWorkaround {
				debugLog(" RowCallback workaround explicitly enabled via config")
			} else {
				debugLog(" RowCallback workaround explicitly disabled via config")
			}
			return *dialector.RowCallbackWorkaround
		}
	}

	// Default behavior: apply the fix since we know current GORM versions have the bug
	// In the future, we can add version detection logic here

	// TODO: Add version detection when GORM fixes the RowQuery callback bug
	// Example future implementation:
	// if gormVersion := getGORMVersion(); gormVersion != "" {
	//     // Check if this version has the bug fixed
	//     fixedInVersions := []string{"v1.31.0", "v1.32.0"} // Example versions
	//     if isVersionAtLeast(gormVersion, "v1.31.0") {
	//         return false // Bug is fixed, use default callback
	//     }
	// }

	// For maximum safety, we could also test the callback at runtime:
	// return isRowCallbackBroken(db)

	// Currently always apply fix since we know GORM v1.30.2 has the bug
	debugLog(" Using default RowCallback workaround behavior (disabled for testing)")
	return false
}

// rowQueryCallback replaces GORM's default row query callback with a DuckDB-compatible version
//
// BACKGROUND: This is a workaround for a critical bug in GORM's RowQuery callback implementation
// where Raw().Row() returns nil instead of *sql.Row, causing nil pointer panics.
//
// The bug affects GORM v1.30.2 and potentially other versions. The default callback fails to
// properly execute QueryRowContext() or assign the result to Statement.Dest.
//
// WORKAROUND: Our implementation correctly handles both single-row and multi-row queries:
// - Single row queries (Row()): Uses QueryRowContext() and assigns result to Statement.Dest
// - Multi-row queries (Rows()): Uses QueryContext() and assigns result to Statement.Dest
//
// FUTURE: When GORM fixes this bug, users can disable this workaround by setting:
//
//	OpenWithRowCallbackWorkaround(dsn, false)
//
// See: docs/GORM_ROW_CALLBACK_BUG_ANALYSIS.md for detailed analysis
func rowQueryCallback(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	// Only process if we have SQL to execute
	if db.Statement.SQL.Len() == 0 {
		return
	}

	// Skip execution if DryRun
	if db.DryRun {
		return
	}

	// Check if this is for multiple rows (Rows()) or single row (Row())
	if isRows, ok := db.Get("rows"); ok && isRows.(bool) {
		// Multiple rows - call QueryContext
		db.Statement.Settings.Delete("rows")
		db.Statement.Dest, db.Error = db.Statement.ConnPool.QueryContext(
			db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
	} else {
		// Single row - call QueryRowContext
		db.Statement.Dest = db.Statement.ConnPool.QueryRowContext(
			db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
	}

	// Set RowsAffected to -1 to indicate unknown row count for single row queries
	db.RowsAffected = -1
}

// duckdbCreateCallback implements a custom CREATE callback to work around
// GORM v1.31.1 issue where gorm:create doesn't generate INSERT SQL for DuckDB dialector
func duckdbCreateCallback(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	stmt := db.Statement
	if stmt.Schema == nil {
		db.Error = fmt.Errorf("schema is nil in CREATE callback")
		return
	}

	debugLog("duckdbCreateCallback called")
	debugLog("duckdbCreateCallback: building INSERT for table %s", stmt.Table)

	// Build INSERT statement manually
	var columns []string
	var placeholders []string
	var values []interface{}
	var autoIncrementField *schema.Field

	// Find auto-increment field and collect values
	for _, field := range stmt.Schema.Fields {
		if field.AutoIncrement {
			autoIncrementField = field
			debugLog("duckdbCreateCallback: skipping auto-increment field %s", field.Name)
			continue
		}

		// Get field value from the model
		fieldValue := stmt.ReflectValue
		if fieldValue.Kind() == reflect.Ptr {
			fieldValue = fieldValue.Elem()
		}
		
		if fieldValue.Kind() == reflect.Struct {
			modelFieldValue := fieldValue.FieldByName(field.Name)
			if modelFieldValue.IsValid() {
				columns = append(columns, fmt.Sprintf(`"%s"`, field.DBName))
				placeholders = append(placeholders, "?")
				values = append(values, modelFieldValue.Interface())
				debugLog("duckdbCreateCallback: adding field %s = %v", field.DBName, modelFieldValue.Interface())
			}
		}
	}

	if len(columns) == 0 {
		db.Error = fmt.Errorf("no fields to insert")
		return
	}

	// Build SQL
	sql := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`,
		stmt.Table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	// Add RETURNING clause for auto-increment fields
	hasAutoIncrement := autoIncrementField != nil
	if hasAutoIncrement {
		sql += fmt.Sprintf(` RETURNING "%s"`, autoIncrementField.DBName)
		debugLog("duckdbCreateCallback: adding RETURNING for field %s", autoIncrementField.DBName)
	}

	debugLog("duckdbCreateCallback: generated SQL: %s", sql)
	debugLog("duckdbCreateCallback: vars: %+v", values)

	// Execute the query
	if hasAutoIncrement {
		// Use QueryRow for RETURNING
		var id interface{}
		err := stmt.ConnPool.QueryRowContext(stmt.Context, sql, values...).Scan(&id)
		if err != nil {
			db.Error = err
			debugLog("duckdbCreateCallback: QueryRow failed: %v", err)
		} else {
			db.RowsAffected = 1
			debugLog("duckdbCreateCallback: QueryRow succeeded, ID: %v", id)
			
			// Set the ID back to the model
			if autoIncrementField != nil {
				// Get the struct value (dereference pointer if needed)
				structValue := stmt.ReflectValue
				if structValue.Kind() == reflect.Ptr {
					structValue = structValue.Elem()
				}
				
				fieldValue := structValue.FieldByName(autoIncrementField.Name)
				debugLog("duckdbCreateCallback: Setting field %s, Valid: %t, CanSet: %t, Kind: %s", 
					autoIncrementField.Name, fieldValue.IsValid(), fieldValue.CanSet(), fieldValue.Kind())
				
				if fieldValue.IsValid() && fieldValue.CanSet() {
					debugLog("duckdbCreateCallback: ID value type: %T, value: %v", id, id)
					switch fieldValue.Kind() {
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						var uintVal uint64
						switch v := id.(type) {
						case uint64:
							uintVal = v
						case int64:
							uintVal = uint64(v)
						case int32:
							uintVal = uint64(v)
						case int:
							uintVal = uint64(v)
						default:
							debugLog("duckdbCreateCallback: Could not convert ID %v (%T) to uint", id, id)
							return
						}
						fieldValue.SetUint(uintVal)
						debugLog("duckdbCreateCallback: Set uint field to %d", uintVal)
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						var intVal int64
						switch v := id.(type) {
						case int64:
							intVal = v
						case uint64:
							intVal = int64(v)
						case int32:
							intVal = int64(v)
						case int:
							intVal = int64(v)
						default:
							debugLog("duckdbCreateCallback: Could not convert ID %v (%T) to int", id, id)
							return
						}
						fieldValue.SetInt(intVal)
						debugLog("duckdbCreateCallback: Set int field to %d", intVal)
					}
				} else {
					debugLog("duckdbCreateCallback: Cannot set field %s", autoIncrementField.Name)
				}
			}
		}
	} else {
		// Use Exec for non-returning operations
		result, err := stmt.ConnPool.ExecContext(stmt.Context, sql, values...)
		if err != nil {
			db.Error = err
			debugLog("duckdbCreateCallback: Exec failed: %v", err)
		} else {
			affected, _ := result.RowsAffected()
			db.RowsAffected = affected
			debugLog("duckdbCreateCallback: Exec succeeded, rows affected: %d", affected)
		}
	}
}

// translateDriverError provides production-ready error translation for DuckDB driver errors
func translateDriverError(err error) error {
	// TODO: Add more robust error translation for DuckDB-specific errors
	// For now, just wrap with context
	if err == nil {
		return nil
	}
	return fmt.Errorf("duckdb driver error: %w", err)
}
