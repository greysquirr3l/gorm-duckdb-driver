---
name: Bug Report
about: Create a detailed bug report to help us improve the DuckDB GORM driver
title: '[BUG] Brief description of the issue'
labels: 'bug'
assignees: ''
---

# DuckDB GORM Driver Bug Report

<!-- 
   INSTRUCTIONS: Replace all [bracketed placeholders] with actual content.
   Remove any sections that don't apply to your specific issue.
   This template covers various bug report scenarios based on actual driver issues.
-->

**Report Date:** <!-- YYYY-MM-DD -->  
**Driver Version:** `github.com/greysquirr3l/gorm-duckdb-driver vX.X.X`  
**Issue Type:** <!-- SQL Syntax Error | Application Crash | Performance Issue | Feature Request | Compatibility Issue -->  
**Severity:** <!-- Critical/Blocker | High | Medium | Low -->  

<!-- Add status banner for resolved issues -->
<!-- **🎉 STATUS: RESOLVED IN [vX.X.X]** ✅   -->
<!-- **Resolution Date:** [YYYY-MM-DD]       -->
<!-- **Fixed By:** [Brief description]       -->

## Summary

<!-- Provide a clear, concise summary of the issue. What is the problem and what component is affected? -->

## Environment

### Software Versions

- **Go Version:** <!-- e.g., 1.21.0 -->
- **GORM Version:** <!-- e.g., v1.25.12 -->
- **DuckDB Driver:** `github.com/greysquirr3l/gorm-duckdb-driver vX.X.X`
- **DuckDB Bindings:** <!-- e.g., github.com/duckdb/duckdb-go-bindings vX.X.X -->
- **Operating System:** <!-- macOS | Linux | Windows (version if relevant) -->

### Dependencies

```go
// Include relevant go.mod entries or replace directives
require (
    gorm.io/gorm v1.25.12
    gorm.io/driver/duckdb v0.2.0
)

// If using custom fork
replace gorm.io/driver/duckdb => github.com/greysquirr3l/gorm-duckdb-driver v0.2.0
```

### Project Context (Optional)

- **Project Name:** <!-- Your project name -->
- **Use Case:** <!-- Brief description of what you're building -->
- **DuckDB Features Used:** <!-- e.g., Arrays, Extensions, JSON, etc. -->

## Error Details

### Primary Error Message

```text
<!-- Paste the exact error message or output here -->
```

### Stack Trace (if applicable)

```plaintext
<!-- Paste full stack trace if available, especially for crashes -->
```

### Error Location

```go
// File: path/to/file.go:line
// Paste the relevant code that triggers the error
```

## Root Cause Analysis

<!-- Explain what you believe is causing the issue -->

### Technical Details

1. **Issue Description:**
   - <!-- Detailed explanation of what's happening -->
   - <!-- Why this is problematic -->

2. **Driver Behavior:**
   - **Current (Incorrect):** <!-- What the driver currently does -->
   - **Expected (Correct):** <!-- What the driver should do -->

3. **Missing Implementation/Fixes Needed:**
   - <!-- List specific functionality that's missing -->
   - <!-- Areas of the driver that need improvement -->

## Reproduction Steps

### Minimal Reproduction Case

```go
package main

import (
    "gorm.io/gorm"
    duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
)

type TestModel struct {
    // Define your test model here
}

func main() {
    db, err := gorm.Open(duckdb.Open("test.db"), &gorm.Config{})
    if err != nil {
        panic(err)
    }
    
    // Add code that reproduces the issue
    err = db.AutoMigrate(&TestModel{})
    if err != nil {
        panic(err) // FAILS HERE
    }
}
```

### Full Reproduction Steps

1. **Setup:**

   ```bash
   # Commands to set up the environment
   ```

2. **Run:**

   ```bash
   # Commands to run and reproduce the issue
   ```

3. **Observe:**
   - <!-- What you should see happen -->
   - <!-- What actually happens instead -->

## Expected Behavior

<!-- Describe what should happen instead of the error -->

### Working Process Should Be

1. **Step 1:** ✅ <!-- What should work -->
2. **Step 2:** ✅ <!-- What should work -->
3. **Step 3:** ✅ <!-- What should work -->

### Required Driver Features (if applicable)

- **Feature 1:** <!-- Description of what the driver needs -->
- **Feature 2:** <!-- Description of what the driver needs -->

## Models/Code Involved

### Example Model

```go
type YourModel struct {
    // Include relevant model definitions that trigger the issue
}
```

### Query/Operation

```go
// Include the specific GORM operations that fail
```

## Impact Assessment

### Severity: **[CRITICAL/HIGH/MEDIUM/LOW]**

- **Application Startup:** <!-- ✅ Works | ❌ Fails | ⚠️ Partial -->
- **Database Operations:** <!-- ✅ Works | ❌ Fails | ⚠️ Partial -->
- **Production Deployment:** <!-- ✅ Possible | ❌ Blocked | ⚠️ Risky -->
- **Development Testing:** <!-- ✅ Works | ❌ Blocked | ⚠️ Limited -->

### Business Impact

- <!-- Describe the real-world impact on your application/business -->
- <!-- List any features that are completely blocked -->
- <!-- Note any performance implications -->

## Workaround Attempts

### 1. **[Workaround Name]** <!-- ✅ Works | ❌ Fails | ⚠️ Partial -->

```go
// Code for attempted workaround
```

**Result:** <!-- Description of outcome and limitations -->

### 2. **[Another Workaround]** <!-- ✅ Works | ❌ Fails | ⚠️ Partial -->

```go
// Code for another attempted workaround
```

**Result:** <!-- Description of outcome and limitations -->

## Proposed Solutions

### 1. **[Solution Name]**

**Priority:** <!-- High | Medium | Low -->  
**Effort:** <!-- High | Medium | Low -->  

```go
// Proposed code changes or implementation approach
```

**Description:** <!-- Explain the proposed solution -->

### 2. **[Alternative Solution]**

**Priority:** <!-- High | Medium | Low -->  
**Effort:** <!-- High | Medium | Low -->  

<!-- Describe alternative approach -->

## Driver Requirements/Implementation Needed

### Must Implement

1. **[Method/Feature Name]:**

   ```go
   // Signature or code example of what needs to be implemented
   ```

2. **[Another Requirement]:**
   - <!-- Specific behavior needed -->
   - <!-- Implementation details -->

### Should Implement (Optional)

- <!-- Nice-to-have features -->
- <!-- Future enhancements -->

## Testing Strategy

### Unit Tests Required

```go
func Test[TestName](t *testing.T) {
    // Test case that should pass once issue is fixed
}
```

### Integration Tests

- <!-- Description of integration tests needed -->
- <!-- Real-world scenarios to validate -->

## Risk Assessment

### Development Risks

- **Timeline Impact:** <!-- High | Medium | Low --> - <!-- explanation -->
- **Technical Debt:** <!-- High | Medium | Low --> - <!-- explanation -->
- **Maintenance Burden:** <!-- High | Medium | Low --> - <!-- explanation -->

### Production Risks

- **Data Loss:** <!-- High | Medium | Low --> - <!-- explanation -->
- **Performance:** <!-- High | Medium | Low --> - <!-- explanation -->
- **Scalability:** <!-- High | Medium | Low --> - <!-- explanation -->

## Additional Context

<!-- Any additional information that might be helpful for debugging or understanding the issue -->

### Configuration

```go
// Include any relevant configuration
db, err := gorm.Open(duckdb.Open("database.db"), &gorm.Config{
    // Your config
})
```

### Extensions Used

<!-- List any DuckDB extensions that are loaded -->
- <!-- Extension 1 -->
- <!-- Extension 2 -->

### Performance Context (if applicable)

- **Data Size:** <!-- Amount of data involved -->
- **Query Complexity:** <!-- Simple/Complex queries -->
- **Concurrent Connections:** <!-- Number of connections -->

### Related Issues

- <!-- Link to related GitHub issues -->
- <!-- Reference to similar problems -->

### Files Demonstrating the Issue

- **File 1:** <!-- path/description -->
- **File 2:** <!-- path/description -->

## Driver Enhancement Impact (for feature requests)

<!-- If this is a feature request, explain how it would improve the driver -->

1. **Benefit 1:** <!-- Description -->
2. **Benefit 2:** <!-- Description -->
3. **Benefit 3:** <!-- Description -->

## Recommendations

### Immediate Actions (timeframe)

1. **Action 1** - <!-- effort estimate -->
2. **Action 2** - <!-- effort estimate -->

### Short-term Actions (timeframe)

1. **Action 1** - <!-- effort estimate -->
2. **Action 2** - <!-- effort estimate -->

### Long-term Actions (timeframe)

1. **Action 1** - <!-- effort estimate -->
2. **Action 2** - <!-- effort estimate -->

## Conclusion

<!-- Summarize the issue and its priority -->

**Next Steps:**

1. <!-- Step 1 (time estimate) -->
2. <!-- Step 2 (time estimate) -->
3. <!-- Step 3 (time estimate) -->

**Total Estimated Resolution Time:** <!-- X hours/days -->

## Checklist

- [ ] I have searched existing issues for similar problems
- [ ] I have provided a minimal reproduction case
- [ ] I have included all relevant version information
- [ ] I have described the expected vs actual behavior
- [ ] I have included stack traces and error messages (if applicable)
- [ ] I have attempted basic troubleshooting steps
- [ ] I have provided business impact context

---

**Reporter:** <!-- Your name/team -->  
**Contact:** <!-- Your contact information -->  
**Last Updated:** <!-- YYYY-MM-DD -->  
**Status:** <!-- Open | In Progress | Resolved | Closed -->  
**Willing to Contribute Fix:** <!-- Yes/No/Maybe -->

<!-- For resolved issues, add resolution section -->
<!--
## Resolution Summary

**Fixed in:** [version]  
**Root Cause:** [Brief explanation]  
**Solution:** [Brief explanation of fix]  

### What Was Fixed

- ✅ [Fix 1]
- ✅ [Fix 2]
- ✅ [Fix 3]

### Verification

- ✅ [How fix was verified]
- ✅ [Test cases that now pass]

**Upgrade Command:** `go get github.com/greysquirr3l/gorm-duckdb-driver@[version]`
-->

---

## Template Usage Notes

**Issue Types:**

- **SQL Syntax Error:** Wrong SQL generation by driver
- **Application Crash:** Panics, nil pointer dereferences, segfaults
- **Performance Issue:** Slow queries, memory leaks, inefficient operations
- **Feature Request:** Missing GORM/DuckDB functionality
- **Compatibility Issue:** Version conflicts, interface problems

**Severity Levels:**

- **Critical/Blocker:** Prevents application startup or core functionality
- **High:** Major feature broken, significant workaround needed
- **Medium:** Feature issue with reasonable workaround available
- **Low:** Minor issue, enhancement, or edge case

**Tips:**

- Always include minimal reproduction code
- Be specific about versions and environment
- Explain the business impact, not just technical details
- Propose solutions when possible
- Update status as issue progresses
