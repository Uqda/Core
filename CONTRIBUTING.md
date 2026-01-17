# Contributing to Uqda Network

Thank you for your interest in contributing to Uqda Network! This document provides guidelines and instructions for contributing.

---

## Code of Conduct

- **Be respectful** - Treat everyone with respect
- **Be constructive** - Provide helpful feedback
- **Be patient** - Maintainers are volunteers
- **Be collaborative** - Work together toward common goals

---

## How to Contribute

### Reporting Bugs

**Before reporting a bug:**

1. Check if the issue already exists in [GitHub Issues](https://github.com/Uqda/Core/issues)
2. Verify you're using the latest version
3. Try to reproduce the issue consistently

**When reporting a bug, include:**

- **Description** - Clear description of the problem
- **Steps to Reproduce** - Detailed steps to reproduce
- **Expected Behavior** - What should happen?
- **Actual Behavior** - What actually happens?
- **Environment**:
  - OS and version
  - Uqda version
  - Architecture (x86_64, ARM64, etc.)
- **Logs** - Relevant log output (if applicable)
- **Screenshots** - If applicable

### Suggesting Features

**Before suggesting a feature:**

1. Check if the feature was already requested
2. Consider if it aligns with Uqda's goals
3. Think about implementation complexity

**When suggesting a feature, include:**

- **Use Case** - What problem does it solve?
- **Proposed Solution** - How should it work?
- **Alternatives** - Other solutions you've considered
- **Additional Context** - Any other relevant information

### Code Contributions

#### Getting Started

1. **Fork the repository**
2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/Core.git
   cd Core
   ```
3. **Create a branch**:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

#### Development Setup

**Prerequisites:**
- Go 1.22 or later
- Git
- Make (optional, for build scripts)

**Build:**
```bash
./build
```

**Run tests:**
```bash
go test ./...
```

#### Code Style

- Follow **Go standard formatting** (`gofmt`)
- Follow **Go naming conventions**
- Write **clear, self-documenting code**
- Add **comments** for complex logic
- Keep functions **focused and small**

**Linting:**
```bash
# Run linter (if configured)
golangci-lint run
```

#### Commit Messages

Use clear, descriptive commit messages:

```
Short summary (50 chars or less)

More detailed explanation if needed. Wrap to 72 characters.
Explain what and why, not how.

- Bullet points are okay
- Use present tense ("Add feature" not "Added feature")
- Reference issues: Fixes #123
```

**Examples:**
```
Fix handshake timeout compatibility issue

Reduced handshake timeout from 3s to 5s to maintain
compatibility with older Uqda versions while still
improving performance over original 6s timeout.

Fixes #42
```

```
Add DNS caching to reduce lookup latency

Implements DNS cache with 5-minute TTL for successful
lookups and 30-second TTL for failed lookups.

Closes #38
```

#### Pull Request Process

1. **Update documentation** - If you change functionality, update docs
2. **Add tests** - If applicable, add tests for new features
3. **Update CHANGELOG.md** - Document your changes
4. **Ensure tests pass** - Run `go test ./...`
5. **Create pull request**:
   - Use clear title and description
   - Reference related issues
   - Explain what changed and why
   - Include screenshots (if UI changes)

**PR Template:**
```markdown
## Description
Brief description of changes

## Related Issues
Fixes #123
Related to #456

## Testing
- [ ] Tests pass locally
- [ ] Manual testing completed
- [ ] Documentation updated

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
```

---

## Documentation Contributions

### Types of Documentation

- **README.md** - Project overview and quick start
- **docs/WHITEPAPER.md** - Technical documentation
- **docs/FAQ.md** - Frequently asked questions
- **docs/EXECUTIVE_SUMMARY.md** - One-page overview
- **docs/** - Additional documentation
- **Code comments** - Inline documentation

### Documentation Guidelines

- **Be clear and concise**
- **Use examples** when helpful
- **Keep up to date** with code changes
- **Fix typos and grammar**
- **Improve clarity** of existing docs

---

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./src/core

# Run tests with coverage
go test -cover ./...
```

### Writing Tests

- **Test edge cases** - Not just happy path
- **Test error conditions** - What happens when things go wrong?
- **Keep tests focused** - One test, one concern
- **Use table-driven tests** - For multiple scenarios
- **Mock external dependencies** - When appropriate

---

## Project Structure

```
Core/
├── cmd/
│   ├── uqda/          # Main daemon
│   └── uqdactl/       # Control tool
├── src/
│   ├── admin/         # Admin socket interface
│   ├── core/          # Core networking logic
│   └── ...
├── docs/              # Documentation
├── contrib/           # Contrib files (systemd, etc.)
└── ...
```

---

## Areas for Contribution

### High Priority

- **Performance improvements** - Reduce latency, optimize memory
- **Documentation** - Improve clarity, add examples
- **Testing** - Increase test coverage
- **Platform support** - Improve Windows/macOS support

### Medium Priority

- **Monitoring tools** - Better diagnostics
- **Configuration management** - Easier peer management
- **Error handling** - Better error messages
- **Logging** - Structured logging improvements

### Low Priority

- **UI tools** - Graphical interfaces
- **Integration** - Package manager support
- **Examples** - Sample applications

---

## Review Process

1. **Automated checks** - CI runs tests and linting
2. **Maintainer review** - At least one maintainer reviews
3. **Feedback** - Maintainers may request changes
4. **Approval** - Once approved, PR is merged

**Review criteria:**
- Code quality and style
- Test coverage
- Documentation updates
- Alignment with project goals

---

## Questions?

- **GitHub Discussions** - [Ask questions](https://github.com/Uqda/Core/discussions)
- **GitHub Issues** - [Report problems](https://github.com/Uqda/Core/issues)
- **Email** - uqda@proton.me

---

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (LGPLv3).

---

**Thank you for contributing to Uqda Network!** 🎉

