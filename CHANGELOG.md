# Changelog

All notable changes to the Signals library will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned for v0.2.0
- Resource tracking and lifecycle management
- Advanced batching strategies
- Performance monitoring and debugging tools
- Additional utility functions
- User guides and tutorials
- Migration guides
- Best practices documentation

## [0.1.0-beta] - 2025-10-31

### 🎉 Initial Beta Release

**Production-ready core functionality!** All three reactive primitives implemented with comprehensive testing, documentation, and CI/CD.

This beta release marks the completion of core functionality. The library is ready for early adopters and production use, with full user guides coming in v1.0.0.

### Added - Phase 1: Core Signal[T]

**Signal Creation and Operations**
- `NewSignal[T](value)` - Create reactive signal with type safety
- `Signal.Get()` - Read current value (thread-safe, zero allocations)
- `Signal.Set(value)` - Update value (thread-safe, notifies subscribers)
- `Signal.Update(fn)` - Update via function (atomic read-modify-write)
- `Signal.Subscribe(fn)` - Subscribe to changes with automatic cleanup
- `Signal.Unsubscribe(id)` - Manual unsubscription
- `Signal.AsReadonly()` - Create read-only view

**Thread Safety**
- Full RWMutex protection for concurrent access
- Safe concurrent reads and writes
- Race detector clean (verified in CI)

**Panic Recovery**
- Built-in panic recovery in user callbacks
- Custom panic handlers via `WithPanicHandler` option
- Safe error propagation

**Testing**
- 24 comprehensive tests covering:
  - Basic operations (Get, Set, Update)
  - Subscriptions and notifications
  - Concurrent access patterns
  - Panic recovery
  - Memory leak prevention
  - Read-only views
- Zero allocations verified in hot paths

### Added - Phase 2: Computed[T]

**Computed Values**
- `NewComputed[T](fn)` - Create computed value with automatic dependency tracking
- Lazy evaluation with intelligent caching
- Automatic invalidation when dependencies change
- Fine-grained reactivity (only recompute what changed)
- Glitch-free execution (atomic updates)

**Dependency Management**
- Automatic dependency tracking during computation
- Circular dependency detection with clear error messages
- Efficient dependency graph management

**Thread Safety**
- Thread-safe recomputation
- Concurrent access protection
- No race conditions

**Testing**
- 13 comprehensive tests covering:
  - Basic computed values
  - Dependency tracking
  - Lazy evaluation
  - Invalidation and caching
  - Circular dependency detection
  - Multiple dependencies
  - Thread safety
- Performance benchmarks for caching efficiency

### Added - Phase 3: Effect

**Effect System**
- `NewEffect(fn)` - Create reactive effect with dependency tracking
- Automatic re-execution when dependencies change
- Effect batching (multiple updates trigger single execution)
- Immediate vs deferred execution modes

**Lifecycle Management**
- `Effect.Cleanup()` - Manual cleanup
- `WithCleanup(fn)` - Cleanup callback option
- Context-based cancellation support
- Proper resource disposal

**Error Handling**
- Panic recovery in effect functions
- Custom panic handlers
- Safe concurrent effect management

**Testing**
- 14 comprehensive tests covering:
  - Basic effect execution
  - Dependency tracking
  - Cleanup functions
  - Batching behavior
  - Panic recovery
  - Concurrent effects
  - Context cancellation
- Benchmarks for effect performance

### Performance

**Zero Allocations in Hot Paths**
- `Signal.Get()`: 0.51 ns/op, 0 B/op, 0 allocs/op
- `Signal.Set()`: 28.6 ns/op, 0 B/op, 0 allocs/op
- `Computed.Get()`: 54.4 ns/op, 0 B/op, 0 allocs/op
- `Effect.Run()`: 204 ns/op, 0 B/op, 0 allocs/op

**Benchmarks**
- 12 benchmarks covering all hot paths
- Memory allocation tracking
- Performance regression detection

### Documentation

**Developer Documentation**
- Architecture overview
- Implementation guide
- Angular Signals analysis
- Roadmap
- Senior review notes

**Repository Structure**
- Professional README with badges
- MIT License
- Contributing guidelines
- Code of Conduct
- Security policy
- CI/CD with GitHub Actions
- Cross-platform testing (Linux, macOS, Windows)

**Code Quality**
- 51 tests total
- 67.9% code coverage
- golangci-lint with 30+ linters
- Zero linter issues
- Race detector clean

### Project Statistics

- **Total Tests**: 51
- **Coverage**: 67.9%
- **Benchmarks**: 12
- **Supported Platforms**: Linux, macOS, Windows
- **Go Version**: 1.25+
- **Dependencies**: Zero (pure Go standard library)

## Version History

### Version Numbering

- **v0.1.0-alpha**: Initial alpha release (current)
- **v0.2.0-alpha**: Documentation and guides (planned)
- **v0.3.0-beta**: Beta testing phase (planned)
- **v1.0.0**: Stable release (planned Q2 2025)

### Semantic Versioning

Starting from v1.0.0, this project will follow strict semantic versioning:

- **Major (X.0.0)**: Breaking API changes
- **Minor (0.X.0)**: New features, backward compatible
- **Patch (0.0.X)**: Bug fixes, backward compatible

### Pre-1.0 Versioning

Before v1.0.0, minor version bumps may include breaking changes. We will clearly document all breaking changes in this changelog.

---

## Links

- [GitHub Repository](https://github.com/coregx/signals)
- [API Documentation](https://pkg.go.dev/github.com/coregx/signals)
- [Issue Tracker](https://github.com/coregx/signals/issues)
- [Contributing Guide](CONTRIBUTING.md)

---

**Maintained by**: Andy Goryachev
**License**: MIT
**Last Updated**: 2025-10-31
