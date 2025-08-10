# Contributing to filesystem-exporter

Thank you for your interest in contributing to filesystem-exporter! This document provides guidelines and information for contributors.

## Development Setup

### Prerequisites
- Go 1.21 or later
- Docker (optional, for containerized development)
- Make (optional, for using Makefile targets)

### Local Development

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/your-username/filesystem-exporter.git
   cd filesystem-exporter
   ```

2. **Install dependencies**
   ```bash
   make deps
   # or
   go mod download
   ```

3. **Run tests**
   ```bash
   make test
   # or
   go test ./...
   ```

4. **Build the application**
   ```bash
   make build
   # or
   go build -o filesystem-exporter ./cmd/main.go
   ```

5. **Run the application**
   ```bash
   make run
   # or
   ./filesystem-exporter
   ```

## Code Quality

### Code Style
- Follow Go conventions and best practices
- Use `gofmt` for code formatting
- Follow the project's linting rules (see `.golangci.yml`)

### Running Linters
```bash
make lint
# or
golangci-lint run
```

### Running Tests
```bash
make test
# or
go test -v ./...
```

### Test Coverage
```bash
make test-coverage
# or
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Making Changes

### Branch Strategy
1. Create a feature branch from `main`
2. Make your changes
3. Add tests for new functionality
4. Ensure all tests pass
5. Update documentation if needed
6. Submit a pull request

### Commit Messages
We follow [Conventional Commits](https://www.conventionalcommits.org/) for commit messages:

- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `style:` for formatting changes
- `refactor:` for code refactoring
- `test:` for adding or updating tests
- `chore:` for maintenance tasks

Examples:
```
feat: add support for custom metrics labels
fix: resolve memory leak in directory collector
docs: update README with new configuration options
```

### Pull Request Process

1. **Create a descriptive PR title** that follows conventional commits
2. **Fill out the PR template** completely
3. **Ensure CI passes** - all tests, linting, and security scans must pass
4. **Add tests** for new functionality
5. **Update documentation** if needed
6. **Request review** from maintainers

## Testing

### Unit Tests
- Write unit tests for new functionality
- Ensure existing tests continue to pass
- Aim for good test coverage

### Integration Tests
- Test configuration loading and validation
- Test metrics collection functionality
- Test error handling and edge cases

### Manual Testing
- Test with different configuration scenarios
- Verify metrics output in Prometheus format
- Test Docker container builds and runs

## Documentation

### Code Documentation
- Add comments for complex logic
- Document exported functions and types
- Include examples where helpful

### User Documentation
- Update README.md for user-facing changes
- Add configuration examples
- Update CHANGELOG.md for significant changes

## Release Process

This project uses [Release Please](https://github.com/google-github-actions/release-please-action) for automated releases.

### For Contributors
- Use conventional commit messages
- Merge to `main` branch
- Release Please will automatically create releases

### For Maintainers
- Review and merge PRs
- Monitor CI/CD pipeline
- Review and publish releases

## Getting Help

- **Issues**: Use GitHub issues for bug reports and feature requests
- **Discussions**: Use GitHub Discussions for questions and general discussion
- **Security**: Report security issues privately to maintainers

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## License

By contributing to filesystem-exporter, you agree that your contributions will be licensed under the MIT License.

