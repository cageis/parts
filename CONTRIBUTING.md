# Contributing to Parts

Thank you for your interest in contributing to Parts! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites
- Go 1.19 or later
- Git

### Development Setup
```bash
# Clone the repository
git clone https://github.com/yourusername/parts.git
cd parts

# Install dependencies
go mod download

# Build the project
make build

# Run tests
make test

# Run all quality checks
make check
```

## Development Workflow

### 1. Fork and Clone
1. Fork the repository on GitHub
2. Clone your fork locally
3. Add the upstream repository as a remote

### 2. Create a Branch
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b bugfix/issue-description
```

### 3. Make Changes
- Write clean, readable code
- Follow existing code patterns and style
- Add tests for new functionality
- Update documentation as needed

### 4. Test Your Changes
```bash
# Run all tests
go test ./...

# Run linting
make check

# Test with examples
make examples

# Test specific functionality
bin/parts --dry-run examples/ssh-config/config examples/ssh-config/partials "#"
```

### 5. Commit and Push
```bash
git add .
git commit -m "feat: add new comment style support"
git push origin feature/your-feature-name
```

### 6. Create a Pull Request
- Use the PR template
- Provide clear description
- Link related issues
- Ensure CI passes

## Code Style Guidelines

### Go Code Style
- Follow standard Go conventions
- Use `gofmt` to format code
- Use meaningful variable names
- Add comments for exported functions
- Keep functions focused and small

### Testing
- Write tests for all new functionality
- Use table-driven tests where appropriate
- Test both success and failure cases
- Include edge cases
- Maintain good test coverage

### Commit Messages
We follow conventional commits format:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `refactor:` for code refactoring
- `test:` for adding tests
- `chore:` for maintenance tasks

Examples:
```
feat: add support for TOML comment style
fix: handle empty partials directory gracefully
docs: update README with new usage examples
```

## Areas for Contribution

### High Priority
- New comment style support
- Performance improvements
- Bug fixes
- Documentation improvements

### Medium Priority
- New file type support
- CLI enhancements
- Example improvements
- Test coverage improvements

### Good First Issues
Look for issues labeled `good-first-issue` which are suitable for newcomers.

## Adding New Comment Styles

To add support for a new comment style:

1. **Add to comments.go:**
   ```go
   // Add to commentStyles map
   "newstyle": {Start: "//", End: ""},
   
   // Add file extensions to extensionToStyle map
   ".newext": "newstyle",
   ```

2. **Add tests:**
   ```go
   // Add test case to commentStyles_test.go
   {"New style", "newstyle", "// PARTIALS>>>>>", "// PARTIALS<<<<<"},
   ```

3. **Update documentation:**
   - Add examples to README.md
   - Update help text if needed
   - Add example in examples/ directory

4. **Test thoroughly:**
   ```bash
   bin/parts --dry-run test.newext ./partials "newstyle"
   bin/parts --dry-run test.newext ./partials "auto"
   ```

## Testing Guidelines

### Unit Tests
- Test all public functions
- Use table-driven tests for multiple test cases
- Mock external dependencies when needed
- Test error conditions

### Integration Tests
- Test complete workflows
- Test with real files
- Test different comment styles
- Test both build and remove operations

### Example Tests
- All examples must work
- Examples are tested in CI
- Add new examples for new features

## Documentation Guidelines

### README Updates
- Keep examples current
- Update feature lists
- Maintain installation instructions
- Update usage patterns

### Code Documentation
- Document all exported functions
- Explain complex logic
- Provide usage examples in comments
- Keep documentation up to date

## Release Process

Releases are handled by maintainers:

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create GitHub release
4. Automated workflows handle building and publishing

## Getting Help

- Check existing issues and discussions
- Create an issue for bugs or feature requests
- Join discussions in GitHub Issues
- Ask questions in pull request reviews

## Code of Conduct

Be respectful and inclusive. We want this project to be welcoming to all contributors.

## Questions?

If you have questions about contributing, please:
1. Check this document
2. Look at existing issues and PRs
3. Create a new issue with your question

Thank you for contributing to Parts! 🎉