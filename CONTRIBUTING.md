# Contributing to GORM DuckDB Driver

Thank you for your interest in contributing to the GORM DuckDB driver! This project aims to provide first-class DuckDB support for the GORM ecosystem.

## Development Setup

### Prerequisites

- Go 1.23 or higher
- Git

### Local Development

```bash
git clone https://github.com/greysquirr3l/gorm-duckdb-driver.git
cd gorm-duckdb-driver
go mod download
go test -v
```

### Running Tests

```bash
# Run all tests
go test -v

# Run specific test
go test -v -run TestConnection

# Run with coverage
go test -v -cover
```

## Contributing Guidelines

### Code Style

- Follow Go best practices and conventions
- Use `go fmt` for formatting
- Ensure code passes `go vet`
- Add appropriate comments for public APIs
- **Follow GORM Style Guidelines**: Use `primaryKey` (camelCase) not `primarykey` in model tags
- **Use Context**: Include context and timeout controls in examples and code
- **Error Handling**: Utilize DuckDB-specific error translation functions
- **Performance**: Implement DuckDB-optimized patterns (batch sizes, field selection)

### Testing

- All new features must include tests
- Maintain or improve test coverage
- Tests should be deterministic and fast

### Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass: `go test -v`
6. Commit with descriptive messages
7. Push to your fork and create a PR

### Issues

- Check existing issues before creating new ones
- Provide clear reproduction steps for bugs
- Include Go version, GORM version, and DuckDB version

## Roadmap to Official GORM Integration

### Current Status

- ✅ Core implementation complete
- ✅ Comprehensive test suite
- ✅ Documentation and examples
- 🔄 Community testing and feedback
- ⏳ Performance optimization
- ⏳ Additional features based on community needs
- ⏳ Submission to go-gorm organization

### How You Can Help

1. **Test the driver** with your applications
2. **Report issues** or edge cases
3. **Contribute features** like advanced data types
4. **Improve documentation** and examples
5. **Performance testing** and optimization
6. **Community feedback** and adoption

## Architecture Notes

### Key Components

- `dialector.go`: Main GORM dialector implementation
- `migrator.go`: Schema migration and introspection
- Connection wrapper for DuckDB-specific handling
- Auto-increment via DuckDB sequences

### Design Decisions

- Uses `information_schema` for introspection (no schema filtering needed)
- Implements connection wrapper for `*time.Time` → `time.Time` conversion
- Follows GORM patterns established by MySQL/PostgreSQL drivers
- Comprehensive error handling and translation

## Getting Help

- GitHub Issues: Bug reports and feature requests
- Discussions: General questions and community chat
- GORM Documentation: [https://gorm.io/docs/]

## License

MIT License - see LICENSE file for details.
