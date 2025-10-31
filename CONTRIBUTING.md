# Contributing to Signals

Thank you for considering contributing to the Signals library! This document outlines the development workflow and guidelines.

## Ways to Contribute

We welcome contributions in many forms:

- **Bug reports** - Help us identify and fix issues
- **Feature requests** - Suggest new functionality
- **Documentation** - Improve guides, examples, and API docs
- **Code contributions** - Submit bug fixes or new features
- **Testing** - Add test cases, improve coverage
- **Performance** - Optimize hot paths, reduce allocations
- **Examples** - Create usage examples and tutorials

## Development Workflow

### Prerequisites

- Go 1.25 or later
- golangci-lint (install with `make install-lint`)
- Git
- Basic understanding of reactive programming (helpful but not required)

### Setting Up Development Environment

```bash
# Clone repository
git clone https://github.com/coregx/signals.git
cd signals

# Verify Go version
go version  # Should be 1.25 or later

# Install development tools
make install-lint

# Run tests to verify setup
make test
```

### Development Cycle

1. **Create an issue** (for non-trivial changes)
   - Describe the problem or feature
   - Discuss approach with maintainers
   - Get feedback before coding

2. **Fork and branch**
   ```bash
   # Fork repository on GitHub

   # Clone your fork
   git clone https://github.com/YOUR_USERNAME/signals.git
   cd signals

   # Add upstream remote
   git remote add upstream https://github.com/coregx/signals.git

   # Create feature branch
   git checkout -b feature/my-feature
   ```

3. **Make changes**
   ```bash
   # Write code following style guidelines
   # Add tests for new functionality
   # Update documentation
   ```

4. **Test thoroughly**
   ```bash
   # Run all checks
   make pre-commit

   # Or run individual checks:
   make test          # Unit tests
   make test-race     # Race detector
   make benchmark     # Performance tests
   make lint          # Code quality
   make fmt           # Format code
   ```

5. **Commit changes**
   ```bash
   # Stage changes
   git add .

   # Commit with conventional message
   git commit -m "feat: add new feature X"
   ```

6. **Push and create PR**
   ```bash
   # Push to your fork
   git push origin feature/my-feature

   # Create pull request on GitHub
   # Fill out PR template with details
   ```

## Commit Message Guidelines

Follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

- **feat**: New feature (e.g., `feat: add batch update support`)
- **fix**: Bug fix (e.g., `fix: resolve race condition in Effect`)
- **docs**: Documentation changes (e.g., `docs: update README examples`)
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code refactoring (no behavior change)
- **test**: Adding or updating tests
- **chore**: Maintenance tasks (build, dependencies, etc.)
- **perf**: Performance improvements

### Examples

```bash
feat: add support for custom equality comparers
fix: correct race condition in computed value updates
docs: add examples for effect cleanup
refactor: simplify signal subscription logic
test: add benchmark for concurrent signal updates
perf: optimize computed value caching
```

## Code Quality Standards

### Before Committing

Run the pre-commit checks to ensure code quality:

```bash
make pre-commit
```

This runs:
1. Code formatting (`go fmt`)
2. Linting (`golangci-lint`)
3. Unit tests
4. Race detector

### Pull Request Requirements

Your PR must meet these requirements:

- [ ] Code is formatted (`make fmt` or `go fmt ./...`)
- [ ] Linter passes with zero issues (`make lint`)
- [ ] All tests pass (`make test`)
- [ ] Race detector clean (`make test-race`)
- [ ] New code has tests (target: 70%+ coverage for new code)
- [ ] Documentation updated (if applicable)
- [ ] Benchmarks added for performance-critical code
- [ ] Commit messages follow conventions
- [ ] No breaking changes without discussion
- [ ] No sensitive data committed

### Code Coverage

- **Minimum**: 70% overall coverage
- **Target**: 90%+ for business logic
- **Critical paths**: 100% coverage for signal core, computed, effect

Check coverage with:
```bash
make test-coverage
# Opens coverage.html in browser
```

## Coding Standards

### General Principles

Follow Go best practices and idioms:

- **SOLID, DRY, KISS, YAGNI**
- **Clarity over cleverness**
- **Error handling is mandatory**
- **Thread-safety by default**
- **Zero allocations in hot paths**

### Naming Conventions

- **Public types/functions**: `PascalCase` (e.g., `NewSignal`, `Get`)
- **Private types/functions**: `camelCase` (e.g., `notifySubscribers`)
- **Interfaces**: Noun or adjective (e.g., `Readable`, `Signal`)
- **Test functions**: `Test*` (e.g., `TestSignal_BasicOperations`)
- **Benchmark functions**: `Benchmark_*` (e.g., `Benchmark_Signal_Get`)

### File Organization

- **Implementation**: `signal.go`, `computed.go`, `effect.go`
- **Tests**: `*_test.go` (same package)
- **Benchmarks**: `*_bench_test.go`
- **Types**: `types.go` (shared types)
- **Options**: `options.go` (functional options)
- **Examples**: `cmd/example/`

### Error Handling

```go
// ✅ GOOD - Always check errors
if err := something(); err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// ❌ BAD - Never ignore errors
something()

// ✅ GOOD - Validate inputs
func NewSignal[T any](value T, opts ...Option[T]) *Signal[T] {
    if value == nil {
        panic("value cannot be nil")
    }
    // ...
}
```

### Thread-Safety

All public APIs must be thread-safe:

```go
// ✅ GOOD - Protected access
type Signal[T any] struct {
    mu    sync.RWMutex
    value T
}

func (s *Signal[T]) Get() T {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.value
}

// ❌ BAD - Unprotected access
func (s *Signal[T]) Get() T {
    return s.value  // Race condition!
}
```

### Performance

Hot paths must have zero allocations:

```go
// ✅ GOOD - Zero allocations
func (s *Signal[T]) Get() T {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.value
}

// ❌ BAD - Allocates on every call
func (s *Signal[T]) Get() T {
    return s.getValue()  // Function call may allocate
}
```

Verify with benchmarks:
```bash
go test -bench=Benchmark_Signal_Get -benchmem
# Should show: 0 B/op    0 allocs/op
```

## Testing Guidelines

### Test Structure

Use table-driven tests for multiple cases:

```go
func TestSignal_Update(t *testing.T) {
    tests := []struct {
        name     string
        initial  int
        updateFn func(int) int
        expected int
    }{
        {"increment", 5, func(v int) int { return v + 1 }, 6},
        {"double", 10, func(v int) int { return v * 2 }, 20},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := NewSignal(tt.initial)
            s.Update(tt.updateFn)
            assert.Equal(t, tt.expected, s.Get())
        })
    }
}
```

### Test Coverage

- **Happy path** - Normal usage
- **Edge cases** - Empty values, zero values, nil
- **Error cases** - Invalid inputs, panics
- **Concurrency** - Race conditions, concurrent access
- **Performance** - Benchmarks for hot paths

### Benchmarks

Add benchmarks for performance-critical code:

```go
func Benchmark_Signal_Get(b *testing.B) {
    s := NewSignal(42)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = s.Get()
    }
}
```

## Documentation Guidelines

### Code Comments

- **Public APIs**: Must have godoc comments
- **Complex logic**: Explain why, not what
- **TODOs**: Use `// TODO(username): description`

```go
// NewSignal creates a new reactive signal with the given initial value.
// Signals are thread-safe and can be read/written from multiple goroutines.
//
// Example:
//
//	count := NewSignal(0)
//	count.Set(5)
//	fmt.Println(count.Get()) // Output: 5
func NewSignal[T any](value T, opts ...Option[T]) *Signal[T] {
    // Implementation
}
```

### Examples

Provide runnable examples:

```go
func ExampleSignal() {
    count := NewSignal(0)
    count.Set(5)
    fmt.Println(count.Get())
    // Output: 5
}
```

## Performance Optimization

### Zero-Allocation Guidelines

1. **Avoid interface boxing** in hot paths
2. **Reuse buffers** with sync.Pool if needed
3. **Minimize allocations** in Get/Set operations
4. **Profile before optimizing** with pprof

### Benchmarking

Always benchmark performance claims:

```bash
# Run benchmarks
make benchmark

# Compare with baseline
go test -bench=. -benchmem -count=5 > new.txt
# (after changes)
benchstat old.txt new.txt
```

## Getting Help

- **Documentation**: Check [docs/](docs/) and [README.md](README.md)
- **Issues**: Search existing issues first
- **Discussions**: Ask questions in GitHub Discussions
- **Architecture**: Review [Architecture Overview](docs/dev/ARCHITECTURE.md)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to Signals!**

Your contributions help make reactive programming in Go better for everyone.
