# Security Policy

## Supported Versions

We actively maintain security updates for the following versions:

| Version | Supported          | Go Version | Status |
| ------- | ------------------ | ---------- | ------ |
| 0.2.x   | :white_check_mark: | 1.24+      | Active |
| 0.1.x   | :warning: Limited  | 1.21+      | Legacy |
| < 0.1   | :x:                | N/A        | Unsupported |

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
```

#### Memory Database Security

```go
// ❌ BAD: Shared memory database
db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{})

// ✅ GOOD: Temporary file with proper cleanup
tmpFile, err := os.CreateTemp("", "app_*.db")
if err != nil {
    return err
}
defer os.Remove(tmpFile.Name()) // Clean up

db, err := gorm.Open(duckdb.Open(tmpFile.Name()), &gorm.Config{})
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
// ✅ GOOD: Validate array inputs
func validateStringArray(arr duckdb.StringArray) error {
    for _, item := range arr {
        if len(item) > 255 {
            return errors.New("string too long")
        }
        if strings.Contains(item, "'") || strings.Contains(item, ";") {
            return errors.New("invalid characters")
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

// ✅ GOOD: Validate extension paths
allowedExtensions := map[string]bool{
    "json": true,
    "parquet": true,
}

func loadExtension(db *gorm.DB, ext string) error {
    if !allowedExtensions[ext] {
        return fmt.Errorf("extension %s not allowed", ext)
    }
    return db.Exec(fmt.Sprintf("LOAD %s", ext)).Error
}
```

### :computer: Memory & Resource Management

#### Connection Pool Security

```go
// ✅ GOOD: Configure connection limits
db, err := gorm.Open(duckdb.Open(dsn), &gorm.Config{})
if err != nil {
    return err
}

sqlDB, err := db.DB()
if err != nil {
    return err
}

// Prevent resource exhaustion
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
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

2. **Extension Loading**: Dynamic extension loading
   - Disable if not needed: `SET enable_external_access=false`
   - Validate extension sources

3. **Memory Usage**: Large datasets can cause OOM
   - Monitor memory consumption
   - Implement query timeouts

### :gear: Configuration Hardening

```sql
-- Disable dangerous features in production
SET enable_external_access = false;
SET enable_object_cache = false;
SET enable_http_metadata_cache = false;
```

## Dependency Security

### :package: Regular Updates

- Monitor Go security advisories: https://pkg.go.dev/vuln/
- Update DuckDB bindings regularly
- Use `go mod tidy` and `go mod vendor` for reproducible builds

### :mag: Vulnerability Scanning

```bash
# Check for known vulnerabilities
go list -json -deps ./... | nancy sleuth

# Use govulncheck
govulncheck ./...
```

## Compliance & Auditing

### :memo: Security Logging

```go
// Log security-relevant events
func auditQuery(query string, user string) {
    log.Printf("AUDIT: User %s executed query: %s", user, 
        sanitizeForLogging(query))
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
- [Go Security Best Practices](https://go.dev/doc/security/)
- [GORM Security Guide](https://gorm.io/docs/security.html)

**Last Updated**: August 2025
