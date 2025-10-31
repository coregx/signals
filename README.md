# Signals

> **Type-safe reactive state management for Go, inspired by Angular Signals**

[![Go Version](https://img.shields.io/github/go-mod/go-version/coregx/signals?style=flat&logo=go)](https://go.dev)
[![Go Report Card](https://goreportcard.com/badge/github.com/coregx/signals)](https://goreportcard.com/report/github.com/coregx/signals)
[![CI](https://github.com/coregx/signals/actions/workflows/test.yml/badge.svg)](https://github.com/coregx/signals/actions)
[![codecov](https://codecov.io/gh/coregx/signals/branch/main/graph/badge.svg)](https://codecov.io/gh/coregx/signals)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![GoDoc](https://pkg.go.dev/badge/github.com/coregx/signals)](https://pkg.go.dev/github.com/coregx/signals)

A modern, production-grade reactive programming library for Go 1.25+ that brings Angular's powerful signals pattern to the Go ecosystem with full type safety, zero allocations in hot paths, and comprehensive concurrency support.

---

## Features

- **Pure Go** - No dependencies, works everywhere Go works
- **Type-Safe** - Full generic support with Go 1.25+ type parameters
- **Thread-Safe** - Built-in synchronization for concurrent access
- **Zero Allocations** - Hot paths designed for zero heap allocations
- **Angular-Compatible** - API design inspired by Angular Signals
- **Fine-Grained Reactivity** - Only re-compute what changed
- **Glitch-Free** - Atomic updates prevent intermediate states
- **Lazy Evaluation** - Computed values calculate only when needed
- **Effect Batching** - Multiple updates trigger single effect execution
- **Production Ready** - 51 tests, 67.9% coverage, comprehensive benchmarks

---

## Quick Start

### Installation

```bash
go get github.com/coregx/signals
```

Requires Go 1.25 or later.

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/coregx/signals"
)

func main() {
    // Create a reactive signal
    count := signals.NewSignal(0)

    // Create computed value that auto-updates
    doubled := signals.NewComputed(func() int {
        return count.Get() * 2
    })

    // Create effect that runs when dependencies change
    signals.NewEffect(func() {
        fmt.Printf("Count: %d, Doubled: %d\n", count.Get(), doubled.Get())
    })
    // Output: Count: 0, Doubled: 0

    // Update signal - effect automatically re-runs
    count.Set(5)
    // Output: Count: 5, Doubled: 10

    count.Set(10)
    // Output: Count: 10, Doubled: 20
}
```

### Advanced Example

```go
// Multiple dependencies
firstName := signals.NewSignal("John")
lastName := signals.NewSignal("Doe")

fullName := signals.NewComputed(func() string {
    return firstName.Get() + " " + lastName.Get()
})

// Effect with cleanup
effect := signals.NewEffect(func() {
    fmt.Println("Full name:", fullName.Get())
}, signals.WithCleanup(func() {
    fmt.Println("Effect cleaned up")
}))

firstName.Set("Jane")  // Effect re-runs
lastName.Set("Smith")  // Effect re-runs

effect.Cleanup()  // Manual cleanup when done
```

[More examples →](cmd/example/)

---

## Documentation

### Getting Started
- **[Installation Guide](docs/guides/INSTALLATION.md)** - Install and verify the library *(coming soon)*
- **[Quick Start Guide](docs/guides/QUICKSTART.md)** - Get started in 5 minutes *(coming soon)*
- **[Core Concepts](docs/guides/CONCEPTS.md)** - Understand signals, computed, and effects *(coming soon)*

### Reference
- **[API Reference](https://pkg.go.dev/github.com/coregx/signals)** - Complete API documentation
- **[Examples](cmd/example/)** - Working code examples
- **[Troubleshooting](docs/guides/TROUBLESHOOTING.md)** - Common issues and solutions *(coming soon)*

### Advanced
- **[Architecture Overview](docs/dev/ARCHITECTURE.md)** - How it works internally
- **[Implementation Guide](docs/dev/IMPLEMENTATION_GUIDE.md)** - Development guide
- **[Angular Signals Analysis](docs/dev/ANGULAR_SIGNALS_ANALYSIS.md)** - Comparison with Angular

---

## Current Status

**Version**: v0.1.0-beta (Core features complete - production-ready for early adopters)

**Production Readiness: Core functionality complete! Documentation and guides coming soon!**

📋 **[See detailed roadmap →](ROADMAP.md)**

### Fully Implemented (67% Complete)

#### Phase 1: Core Signal[T]
- Signal creation and basic operations
- Thread-safe read/write with RWMutex
- Subscription system with automatic unsubscribe
- Update notifications with batching
- Panic recovery with custom handlers
- Read-only view support
- Comprehensive test coverage (24 tests)
- Zero allocations in hot paths (verified by benchmarks)

#### Phase 2: Computed[T]
- Lazy evaluation with automatic caching
- Dependency tracking and invalidation
- Fine-grained reactivity
- Glitch-free execution
- Circular dependency detection
- Thread-safe recomputation
- Comprehensive test coverage (13 tests)
- Optimized performance (minimal allocations)

#### Phase 3: Effect
- Automatic dependency tracking
- Effect scheduling and batching
- Cleanup function support
- Context-based cancellation
- Panic recovery
- Immediate vs deferred execution
- Comprehensive test coverage (14 tests)
- Concurrent effect management

### Test Coverage

| Package | Tests | Coverage | Benchmarks |
|---------|-------|----------|------------|
| signals | 51    | 67.9%    | 12         |

**Key Metrics**:
- 24 Signal tests
- 13 Computed tests
- 14 Effect tests
- Zero allocations in signal read/write hot paths
- Race detector clean (all tests pass with `-race`)

### Performance Characteristics

```
Benchmark_Signal_Get          1000000000    0.51 ns/op    0 B/op    0 allocs/op
Benchmark_Signal_Set          41869632     28.6 ns/op     0 B/op    0 allocs/op
Benchmark_Computed_Get        22285714     54.4 ns/op     0 B/op    0 allocs/op
Benchmark_Effect_Run          5865354     204 ns/op       0 B/op    0 allocs/op
```

*Zero allocations in hot paths ensure minimal GC pressure*

### Remaining Work (33% - Documentation & Guides)

#### Phase 4: Documentation (In Progress)
- User guides and tutorials
- API documentation examples
- Migration guides
- Best practices guide

#### Phase 5: Advanced Features (Planned v0.2.0)
- Resource tracking and lifecycle management
- Advanced batching strategies
- Performance monitoring and debugging tools
- Additional utility functions

See [ROADMAP.md](docs/dev/ROADMAP.md) for detailed timeline.

---

## Development

### Requirements
- Go 1.25 or later
- golangci-lint (for linting)
- No external runtime dependencies

### Building

```bash
# Clone repository
git clone https://github.com/coregx/signals.git
cd signals

# Run tests
make test

# Run tests with race detector
make test-race

# Run benchmarks
make benchmark

# Run linter
make lint
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detector
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem ./...
```

### Code Quality

This project maintains high code quality standards:

```bash
# Format code
make fmt

# Run linter (zero issues required)
make lint

# Run all pre-commit checks
make pre-commit
```

---

## Contributing

Contributions are welcome! This is an early-stage project and we'd love your help.

**Before contributing**:
1. Read [CONTRIBUTING.md](CONTRIBUTING.md) - Development workflow and guidelines
2. Check [open issues](https://github.com/coregx/signals/issues)
3. Review the [Architecture Overview](docs/dev/ARCHITECTURE.md)

**Ways to contribute**:
- Report bugs
- Suggest features
- Improve documentation
- Submit pull requests
- Star the project

---

## Comparison with Other Libraries

| Feature | Signals | RxGo | Reactor |
|---------|---------|------|---------|
| Type-Safe Generics | Yes (Go 1.25+) | Limited | No |
| Zero Allocations | Yes (hot paths) | No | No |
| Thread-Safe | Yes (built-in) | Yes | Partial |
| Angular-Compatible | Yes | No | No |
| Fine-Grained Reactivity | Yes | Observable-based | Stream-based |
| Dependencies | Zero | Multiple | Multiple |
| Learning Curve | Low (if you know Angular) | Medium | Medium |

---

## Angular Signals Compatibility

This library is designed to be conceptually compatible with Angular Signals:

| Angular Signals | Go Signals | Status |
|----------------|------------|--------|
| `signal(T)` | `NewSignal[T](value)` | Complete |
| `computed(() => T)` | `NewComputed[T](fn)` | Complete |
| `effect(() => {})` | `NewEffect(fn)` | Complete |
| `signal.set(value)` | `signal.Set(value)` | Complete |
| `signal()` | `signal.Get()` | Complete |
| `signal.update(fn)` | `signal.Update(fn)` | Complete |
| `signal.asReadonly()` | `signal.AsReadonly()` | Complete |
| Automatic tracking | Automatic tracking | Complete |
| Glitch-free | Glitch-free | Complete |
| Lazy computed | Lazy computed | Complete |

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- The Angular team for the signals design pattern
- The Go team for generics and type parameters
- All contributors to this project

---

## Support

- [API Documentation](https://pkg.go.dev/github.com/coregx/signals)
- [Issue Tracker](https://github.com/coregx/signals/issues)
- [Discussions](https://github.com/coregx/signals/discussions)

---

**Status**: Beta - Core complete, production-ready for early adopters
**Version**: v0.1.0-beta
**Last Updated**: 2025-10-31

---

*Built with care for the Go community*
