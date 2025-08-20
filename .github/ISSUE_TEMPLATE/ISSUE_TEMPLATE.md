# Bug Report Template

Use this template when reporting bugs in the DuckDB GORM driver. Replace the placeholders with your specific information.

---

## Basic Information

**Date:** [Today's date]  
**Project:** [Your project name]  
**Driver Version:** `github.com/greysquirr3l/gorm-duckdb-driver v[X.X.X]`  
**Severity:** [Blocker/Critical/Major/Minor]  

## Issue Summary

[Provide a clear, concise description of the bug]

## Environment

- **Go Version:** [e.g., 1.21.5]
- **GORM Version:** [e.g., v1.25.12]
- **DuckDB Driver Version:** [Your version]
- **Operating System:** [e.g., macOS 14.2]

## Error Details

### Stack Trace

```plaintext
[Paste complete stack trace here]
```

### Error Location

```go
// File: path/to/file.go:line
[Include relevant code where error occurs]
```

## Model Definition

```go
type YourModel struct {
    // Include complete model that causes the issue
}
```

## Reproduction Steps

1. **Setup:**

   ```bash
   # Commands to set up the issue
   ```

2. **Run:**

   ```bash
   # Commands to reproduce the bug
   ```

3. **Minimal Example:**

   ```go
   package main
   
   import (
       "gorm.io/gorm"
       duckdb "gorm.io/driver/duckdb"
   )
   
   func main() {
       // Minimal code that reproduces the issue
   }
   ```

## Expected vs Actual Behavior

**Expected:** [What should happen]  
**Actual:** [What actually happens]

## Impact Assessment

- **Application Startup:** ✅/❌
- **Database Operations:** ✅/❌  
- **Production Ready:** ✅/❌

## Workarounds Attempted

1. **[Approach 1]:** [Result]
2. **[Approach 2]:** [Result]

## Additional Context

[Any other relevant information, configuration, or context]

---

**Reporter:** [Your name]  
**Priority:** [High/Medium/Low]  
**Willing to Help:** [Yes/No]
