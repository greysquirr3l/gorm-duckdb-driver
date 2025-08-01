# Security Policy

## Supported Versions

We currently support the following versions of the GORM DuckDB driver:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | ✅ Yes             |
| < 1.0   | ❌ No              |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please follow these steps:

### 1. Do NOT Open a Public Issue

Please do not report security vulnerabilities through public GitHub issues.

### 2. Report Privately

Send an email to [security@example.com] with the following information:

- **Subject**: Security Vulnerability Report - GORM DuckDB Driver
- **Description**: Detailed description of the vulnerability
- **Steps to Reproduce**: Clear steps to reproduce the issue
- **Impact**: Potential impact and attack scenarios
- **Environment**: Affected versions and configurations
- **Proof of Concept**: If applicable, include PoC code (but please be responsible)

### 3. Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution**: Security fixes will be prioritized and released as soon as possible

### 4. Disclosure Process

1. **Private Discussion**: We'll work with you to understand and validate the issue
2. **Fix Development**: We'll develop and test a fix
3. **Coordinated Disclosure**: We'll coordinate the public disclosure timeline with you
4. **Public Release**: Security advisory and patched version will be released

### 5. Recognition

We appreciate responsible disclosure and will acknowledge your contribution in:

- Security advisory (if desired)
- CHANGELOG.md
- Hall of Fame (if we create one)

## Security Best Practices

When using the GORM DuckDB driver:

### Database Security

- Always validate and sanitize user input before database operations
- Use parameterized queries to prevent SQL injection
- Implement proper access controls and authentication
- Regularly update to the latest version

### Connection Security

- Use secure connection strings
- Avoid hardcoding credentials in source code
- Use environment variables or secure configuration management
- Enable SSL/TLS when possible

### Data Protection

- Encrypt sensitive data at rest
- Implement proper backup security
- Follow data retention policies
- Consider data masking for non-production environments

### Monitoring and Logging

- Monitor database access patterns
- Log security-relevant events
- Implement alerting for suspicious activities
- Regular security audits

## Dependencies

This project depends on:

- [GORM](https://github.com/go-gorm/gorm) - Go ORM library
- [DuckDB](https://github.com/duckdb/duckdb) - In-memory analytical database
- Various Go standard library packages

We regularly monitor our dependencies for security vulnerabilities using:

- GitHub Dependabot
- Go vulnerability database
- Automated security scanning

## Security Tools

We use the following security tools in our CI/CD pipeline:

- **CodeQL**: Static code analysis for vulnerability detection
- **Gosec**: Go security checker
- **Nancy**: Vulnerability scanner for Go dependencies
- **Trivy**: Comprehensive vulnerability scanner

## Contact

For security-related questions or concerns, contact:

- Security Email: [s0ma@protonmail.com]
- Primary Maintainer: [@greysquirr3l](https://github.com/greysquirr3l)

## Updates

This security policy will be updated as needed. Check back regularly for the latest version.

**Last Updated**: [2025-07-31]
**Version**: 1.0
