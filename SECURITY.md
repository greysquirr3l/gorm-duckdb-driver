# Security Policy

## Supported Versions

We actively maintain security updates for the following versions:

| Version | Supported          | Go Version | Status | Notes |
| ------- | ------------------ | ---------- | ------ | ----- |
| 0.6.x   | :white_check_mark: | 1.24+      | Active | Native Array Support |
| 0.5.x   | :white_check_mark: | 1.24+      | Active | 100% GORM Compliance |
| 0.4.x   | :warning: Limited  | 1.24+      | Legacy | Advanced DuckDB Types |
| 0.3.x   | :warning: Limited  | 1.24+      | Legacy | Extension Management |
| 0.2.x   | :x:                | 1.24+      | EOL    | Basic Functionality |
| 0.1.x   | :x:                | 1.21+      | EOL    | Deprecated |
| < 0.1   | :x:                | N/A        | EOL    | Unsupported |

## Reporting a Vulnerability

### :rotating_light: Critical Security Issues

For **critical security vulnerabilities** that could lead to:

- SQL injection attacks
- Data exposure
- Authentication bypass
- Remote code execution

**DO NOT** open a public GitHub issue.

### :mailbox: Private Disclosure Process

1. **Email**: Send details to `s0ma@protonmail.me`
2. **Subject**: `[SECURITY] GORM DuckDB Driver - [Brief Description]`
3. **Include**:
   - Detailed vulnerability description
   - Steps to reproduce (with code examples)
   - Affected versions
   - Potential impact assessment
   - Suggested mitigation (if known)
   - Your contact information for follow-up

### :clock1: Response Timeline

| Action | Timeframe |
| ------ | --------- |
| Initial acknowledgment | 24 hours |
| Preliminary assessment | 72 hours |
| Status update | Weekly |
| Fix development | 2-4 weeks |
| Security advisory | Upon fix release |

### :trophy: Recognition

Security researchers who responsibly disclose vulnerabilities will be:

- Credited in release notes (if desired)
- Listed in our security acknowledgments
- Notified of fix releases

## Security Best Practices

### :shield: Database Connection Security

#### Connection String Protection

```go
// ❌ BAD: Hardcoded credentials
dsn := "duckdb://user:password@host/db"

// ✅ GOOD: Environment variables
dsn := fmt.Sprintf("duckdb://%s:%s@%s/%s", 
    os.Getenv("DB_USER"),
    os.Getenv("DB_PASS"), 
    os.Getenv("DB_HOST"),
    os.Getenv("DB_NAME"))

// ✅ BETTER: Use configuration struct (v0.5.2+)
config := duckdb.Config{
    DSN: os.Getenv("DUCKDB_DSN"),
    DefaultStringSize: 256,
}
db, err := gorm.Open(duckdb.New(config), &gorm.Config{})
```

#### Memory Database Security

```go
// ❌ BAD: Shared memory database in production
db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{})

// ✅ GOOD: Temporary file with proper cleanup
tmpFile, err := os.CreateTemp("", "app_*.db")
if err != nil {
    return err
}
defer os.Remove(tmpFile.Name()) // Clean up

db, err := gorm.Open(duckdb.Open(tmpFile.Name()), &gorm.Config{})

// ✅ BEST: Production configuration with extensions (v0.5.2+)
db, err := gorm.Open(duckdb.OpenWithExtensions("production.db", &duckdb.ExtensionConfig{
    AutoInstall:       true,
    PreloadExtensions: []string{"json", "parquet"}, // Only trusted extensions
    Timeout:           30 * time.Second,
}), &gorm.Config{})
```

### :lock: Input Validation & SQL Injection Prevention

#### Safe Query Patterns

```go
// ✅ GOOD: Use GORM's built-in parameterization
var users []User
db.Where("name = ? AND age > ?", userInput, ageLimit).Find(&users)

// ✅ GOOD: Named parameters
db.Where("name = @name AND age > @age", 
    sql.Named("name", userInput), 
    sql.Named("age", ageLimit)).Find(&users)

// ❌ BAD: String concatenation
query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", userInput)
db.Raw(query).Scan(&users)
```

#### Array Input Validation

```go
// ✅ GOOD: Validate array inputs (v0.5.2+ Advanced Types)
func validateStringArray(arr duckdb.StringArray) error {
    if len(arr) > 1000 { // Prevent large arrays
        return errors.New("array too large")
    }
    
    for _, item := range arr {
        if len(item) > 255 {
            return errors.New("string too long")
        }
        if strings.Contains(item, "'") || strings.Contains(item, ";") {
            return errors.New("invalid characters")
        }
        // Additional validation for SQL injection patterns
        if strings.Contains(strings.ToLower(item), "drop ") ||
           strings.Contains(strings.ToLower(item), "delete ") ||
           strings.Contains(strings.ToLower(item), "update ") {
            return errors.New("potentially dangerous SQL keywords")
        }
    }
    return nil
}

// ✅ GOOD: Validate advanced DuckDB types (v0.5.2+)
func validateAdvancedTypes(data interface{}) error {
    switch v := data.(type) {
    case duckdb.JSONType:
        // Validate JSON structure and size
        if len(v.Data) > 1024*1024 { // 1MB limit
            return errors.New("JSON payload too large")
        }
    case duckdb.DecimalType:
        // Validate decimal precision
        if v.Precision > 38 || v.Scale > 38 {
            return errors.New("decimal precision/scale out of range")
        }
    case duckdb.UUIDType:
        // Validate UUID format
        if !isValidUUID(string(v)) {
            return errors.New("invalid UUID format")
        }
    }
    return nil
}
```

### :file_folder: File System Security

#### Database File Permissions

```go
// ✅ GOOD: Restrict file permissions
dbFile := "app.db"
if err := os.Chmod(dbFile, 0600); err != nil { // Owner read/write only
    return fmt.Errorf("failed to set db permissions: %w", err)
}
```

#### Extension Loading Security

```go
// ❌ BAD: Loading arbitrary extensions
db.Exec("LOAD '/path/to/unknown/extension.so'")

// ✅ GOOD: Validate extension paths (v0.5.2+ Extension Management)
allowedExtensions := map[string]bool{
    "json":         true,
    "parquet":      true,
    "autocomplete": true,
    "fts":          true,
    "httpfs":       false, // Disable network access in production
    "spatial":      true,
}

func loadExtension(db *gorm.DB, ext string) error {
    allowed, exists := allowedExtensions[ext]
    if !exists || !allowed {
        return fmt.Errorf("extension %s not allowed", ext)
    }
    
    // Use the extension manager for secure loading (v0.5.2+)
    manager, err := duckdb.GetExtensionManager(db)
    if err != nil {
        return fmt.Errorf("failed to get extension manager: %w", err)
    }
    
    return manager.LoadExtension(ext)
}

// ✅ BEST: Use extension helpers (v0.5.2+)
func setupSecureExtensions(db *gorm.DB) error {
    manager, err := duckdb.GetExtensionManager(db)
    if err != nil {
        return err
    }
    
    helper := duckdb.NewExtensionHelper(manager)
    
    // Load only required extension groups
    if err := helper.EnableAnalytics(); err != nil { // json, parquet, fts
        return err
    }
    
    // Skip cloud access in production
    // if err := helper.EnableCloudAccess(); err != nil {
    //     return err
    // }
    
    return nil
}
```

### :computer: Memory & Resource Management

#### Connection Pool Security

```go
// ✅ GOOD: Configure connection limits (v0.5.2+ Production Config)
db, err := gorm.Open(duckdb.OpenWithExtensions("production.db", &duckdb.ExtensionConfig{
    AutoInstall:       true,
    PreloadExtensions: []string{"json", "parquet"}, // Only trusted extensions
    Timeout:           30 * time.Second,
}), &gorm.Config{
    Logger: logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold:             time.Second,
            LogLevel:                  logger.Warn, // Avoid leaking sensitive data
            IgnoreRecordNotFoundError: true,
            Colorful:                  false, // Production setting
        },
    ),
})

if err != nil {
    return err
}

sqlDB, err := db.DB()
if err != nil {
    return err
}

// DuckDB-optimized security settings
sqlDB.SetMaxOpenConns(25)                   // Prevent connection exhaustion
sqlDB.SetMaxIdleConns(5)                    // Reduce attack surface
sqlDB.SetConnMaxLifetime(5 * time.Minute)   // Force connection refresh
sqlDB.SetConnMaxIdleTime(1 * time.Minute)   // Quick cleanup of idle connections
```

### :globe_with_meridians: Network Security

#### TLS Configuration (for network-enabled builds)

```go
// ✅ GOOD: Enforce TLS for network connections
config := &tls.Config{
    MinVersion: tls.VersionTLS12,
    CipherSuites: []uint16{
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
    },
}
```

## Known Security Considerations

### :warning: DuckDB-Specific Risks

1. **File Access**: DuckDB can read arbitrary files via SQL
   - Validate all file paths in queries
   - Use allowlists for permitted directories
   - Disable external access: `SET enable_external_access=false`

2. **Extension Loading**: Dynamic extension loading (Enhanced in v0.5.2)
   - Use the secure extension manager: `duckdb.GetExtensionManager()`
   - Validate extension sources and use allowlists
   - Prefer extension helpers: `duckdb.NewExtensionHelper()`
   - Disable if not needed: `SET enable_external_access=false`

3. **Memory Usage**: Large datasets can cause OOM
   - Monitor memory consumption with performance metrics types (v0.5.2+)
   - Implement query timeouts and resource limits
   - Use streaming queries for large datasets

4. **Advanced Type Vulnerabilities** (New in v0.5.2):
   - **JSONType**: Validate JSON structure and size limits
   - **BLOBType**: Scan for malicious binary content
   - **UNIONType**: Validate all variant types
   - **StructType/MapType**: Prevent deeply nested structures (DoS)

5. **Schema Introspection**: Enhanced metadata access (v0.5.2)
   - `ColumnTypes()` exposes detailed schema information
   - `TableType()` reveals table metadata
   - Implement proper access controls for sensitive schema data

### :gear: Configuration Hardening

```sql
-- Disable dangerous features in production (v0.5.2+ recommendations)
SET enable_external_access = false;      -- Prevent file system access
SET enable_object_cache = false;         -- Reduce memory attack surface  
SET enable_http_metadata_cache = false;  -- Disable network metadata caching
SET memory_limit = '2GB';                -- Prevent memory exhaustion
SET threads = 4;                         -- Limit CPU usage
SET max_expression_depth = 100;          -- Prevent deep recursion attacks
```

#### GORM-Level Security Configuration (v0.5.2+)

```go
// ✅ GOOD: Secure GORM configuration
config := &gorm.Config{
    Logger: logger.New(
        log.New(logOutput, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold:             time.Second,
            LogLevel:                  logger.Warn, // Don't log sensitive data
            IgnoreRecordNotFoundError: true,
            Colorful:                  false,
        },
    ),
    NamingStrategy: schema.NamingStrategy{
        SingularTable: false, // Use consistent naming
    },
    DisableForeignKeyConstraintWhenMigrating: false, // Enforce referential integrity
    PrepareStmt:                              true,  // Use prepared statements
}

// Enhanced error handling with security considerations
db.Callback().Query().Before("gorm:query").Register("security_check", func(db *gorm.DB) {
    // Implement query validation, rate limiting, etc.
    if isRateLimited(db.Statement.Context) {
        db.AddError(errors.New("rate limit exceeded"))
        return
    }
})
```

## Dependency Security

### :package: Regular Updates

- Monitor Go security advisories: https://pkg.go.dev/vuln/
- Update DuckDB bindings regularly (current: v2.3.3+)
- Update GORM DuckDB driver regularly (latest: v0.5.2)
- Use `go mod tidy` and `go mod vendor` for reproducible builds
- Monitor GitHub Security Advisories for this repository

### :mag: Vulnerability Scanning

```bash
# Check for known vulnerabilities
go list -json -deps ./... | nancy sleuth

# Use govulncheck (recommended)
govulncheck ./...

# GORM DuckDB driver specific tests (v0.5.2+)
go test -v -run TestSecurity      # Run security-focused tests
go test -v -run TestCompliance    # Verify GORM compliance (security relevant)
```

## Compliance & Auditing

### :memo: Security Logging

```go
// Log security-relevant events (Enhanced for v0.5.2)
func auditQuery(query string, user string, metadata map[string]interface{}) {
    // Sanitize sensitive data before logging
    sanitized := sanitizeForLogging(query)
    
    log.Printf("AUDIT: User %s executed query: %s, metadata: %v", 
        user, sanitized, metadata)
    
    // Enhanced logging for advanced features (v0.5.2+)
    if strings.Contains(strings.ToLower(query), "information_schema") {
        log.Printf("SCHEMA_ACCESS: User %s accessed schema information", user)
    }
    
    if strings.Contains(strings.ToLower(query), "load ") {
        log.Printf("EXTENSION_LOAD: User %s attempted extension loading", user)
    }
}

// Advanced type access logging (v0.5.2+)
func auditAdvancedTypeAccess(typeUsed string, user string, operation string) {
    log.Printf("ADVANCED_TYPE: User %s used %s for %s operation", 
        user, typeUsed, operation)
}

// GORM interface access logging (v0.5.2+)
func auditGORMOperation(operation string, table string, user string) {
    log.Printf("GORM_AUDIT: User %s performed %s on table %s", 
        user, operation, table)
}
```

### :lock: Data Protection

- **Encryption at Rest**: Encrypt database files using OS-level encryption
- **Data Minimization**: Only collect necessary data
- **Retention Policies**: Implement data retention and deletion policies

## Emergency Response

### :sos: Security Incident Response

1. **Immediate**: Isolate affected systems
2. **Assessment**: Evaluate scope and impact  
3. **Mitigation**: Apply temporary fixes
4. **Communication**: Notify affected users
5. **Recovery**: Implement permanent fixes
6. **Lessons Learned**: Update security measures

### :telephone_receiver: Emergency Contacts

- **Security Team**: `s0ma@protonmail.me`
- **Incident Response**: Create GitHub issue with `[URGENT]` prefix for non-security incidents

---

## Additional Resources

- [OWASP Database Security](https://owasp.org/www-project-database-security/)
- [DuckDB Security Documentation](https://duckdb.org/docs/sql/configuration)
- [Go Security Best Practices](https://go.dev/security/)
- [GORM Security Guide](https://gorm.io/docs/security.html)
- [GORM DuckDB Driver v0.5.2 Documentation](https://github.com/greysquirr3l/gorm-duckdb-driver)
- [GitHub Security Advisory Database](https://github.com/advisories)

**Last Updated**: November 2025 (v0.6.1 - Native Array Support Release)
