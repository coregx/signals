# Signals

> **Type-safe reactive state management for Go, inspired by Angular Signals**

[![Release](https://img.shields.io/github/v/release/coregx/signals?style=flat-square&logo=github&color=blue)](https://github.com/coregx/signals/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/coregx/signals?style=flat-square&logo=go)](https://go.dev)
[![Go Report Card](https://goreportcard.com/badge/github.com/coregx/signals?style=flat-square)](https://goreportcard.com/report/github.com/coregx/signals)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=flat-square&logo=go)](https://pkg.go.dev/github.com/coregx/signals)
[![CI](https://img.shields.io/github/actions/workflow/status/coregx/signals/test.yml?branch=main&style=flat-square&logo=github&label=tests)](https://github.com/coregx/signals/actions)
[![codecov](https://codecov.io/gh/coregx/signals/graph/badge.svg)](https://codecov.io/gh/coregx/signals)
[![License](https://img.shields.io/github/license/coregx/signals?style=flat-square&color=blue)](LICENSE)
[![Stars](https://img.shields.io/github/stars/coregx/signals?style=flat-square&logo=github)](https://github.com/coregx/signals/stargazers)

A modern, production-grade reactive programming library for Go 1.25+ that brings Angular's powerful signals pattern to the Go ecosystem with full type safety, zero allocations in hot paths, and comprehensive concurrency support.

---

## Features

- **Pure Go** - No dependencies, works everywhere Go works
- **Type-Safe** - Full generic support with Go 1.25+ type parameters
- **Thread-Safe** - Built-in synchronization for concurrent access
- **Zero Allocations** - Hot paths designed for zero heap allocations
- **Angular-Inspired** - API design inspired by Angular Signals
- **Fine-Grained Reactivity** - Only re-compute what changed
- **Glitch-Free** - Atomic updates prevent intermediate states
- **Lazy Evaluation** - Computed values calculate only when needed
- **Effect Cleanup** - Automatic resource management with cleanup callbacks
- **Production Ready** - 55 tests, 94.5% coverage, 28 benchmarks

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
    count := signals.New(0)

    // Create computed value with explicit dependencies
    doubled := signals.Computed(func() int {
        return count.Get() * 2
    }, count.AsReadonly())

    // Create effect that runs immediately and on dependency changes
    eff := signals.Effect(func() {
        fmt.Printf("Count: %d, Doubled: %d\n", count.Get(), doubled.Get())
    }, count.AsReadonly(), doubled)
    defer eff.Stop()
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
firstName := signals.New("John")
lastName := signals.New("Doe")

fullName := signals.Computed(func() string {
    return firstName.Get() + " " + lastName.Get()
}, firstName.AsReadonly(), lastName.AsReadonly())

// Effect with cleanup
eff := signals.EffectWithCleanup(func() func() {
    fmt.Println("Full name:", fullName.Get())

    // Return cleanup function (runs before next execution and on Stop)
    return func() {
        fmt.Println("Effect cleaned up")
    }
}, fullName)

firstName.Set("Jane")  // Cleanup runs, then effect re-runs
lastName.Set("Smith")   // Cleanup runs, then effect re-runs

eff.Stop()  // Final cleanup runs, effect stops
```

[More examples →](cmd/example/)

---

## Documentation

### Reference
- **[API Reference](https://pkg.go.dev/github.com/coregx/signals)** - Complete API documentation
- **[Examples](cmd/example/)** - Working code examples

### Advanced
- **[Architecture Overview](docs/dev/ARCHITECTURE.md)** - How it works internally
- **[Implementation Guide](docs/dev/IMPLEMENTATION_GUIDE.md)** - Development guide
- **[Angular Signals Analysis](docs/dev/ANGULAR_SIGNALS_ANALYSIS.md)** - Comparison with Angular

---

## Current Status

**Version**: v0.1.0 (Stable - Production-ready!)

**Production Readiness: Core functionality complete and stable!**

**[See detailed roadmap →](ROADMAP.md)**

### Fully Implemented

#### Phase 1: Core Signal[T]
- Signal creation and basic operations
- Thread-safe read/write with RWMutex
- Subscription system with automatic unsubscribe
- Context-based auto-cancellation
- Panic recovery with custom handlers
- Read-only view support
- Custom equality functions
- Comprehensive test coverage (19 tests)
- Zero allocations in hot paths (verified by benchmarks)

#### Phase 2: Computed[T]
- Lazy evaluation with automatic caching
- Explicit dependency tracking and invalidation
- Fine-grained reactivity
- Glitch-free execution
- Thread-safe recomputation
- Comprehensive test coverage (17 tests)
- Optimized performance (minimal allocations)

#### Phase 3: Effect
- Explicit dependency tracking
- Cleanup function support (runs before re-execution and on Stop)
- Panic recovery with custom handlers
- Immediate execution on creation (Angular pattern)
- Comprehensive test coverage (16 tests)
- Concurrent effect management

### Test Coverage

| Package | Tests | Coverage | Benchmarks |
|---------|-------|----------|------------|
| signals | 55    | 94.5%    | 28         |

**Key Metrics**:
- 19 Signal tests
- 17 Computed tests
- 16 Effect tests
- 3 Internal tests (type erasure, reflection fallback)
- Zero allocations in signal read/write hot paths
- Race detector clean (all tests pass with `-race`)

### Performance Characteristics

```
BenchmarkSignal_Get            46738254     27.72 ns/op    0 B/op    0 allocs/op
BenchmarkSignal_Set            20975720     52.42 ns/op    0 B/op    0 allocs/op
BenchmarkComputed_Get_Clean    79329402     19.84 ns/op    0 B/op    0 allocs/op
BenchmarkEffect_Execute         8391343    139.6  ns/op    0 B/op    0 allocs/op
```

*Zero allocations in hot paths ensure minimal GC pressure*

### Remaining Work

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

See [ROADMAP.md](ROADMAP.md) for detailed timeline.

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
| Angular-Inspired | Yes | No | No |
| Fine-Grained Reactivity | Yes | Observable-based | Stream-based |
| Dependencies | Zero | Multiple | Multiple |
| Learning Curve | Low (if you know Angular) | Medium | Medium |

---

## Angular Signals Compatibility

This library is designed to be conceptually compatible with Angular Signals:

| Angular Signals | Go Signals | Status |
|----------------|------------|--------|
| `signal(T)` | `New[T](value)` | Complete |
| `computed(() => T)` | `Computed[T](fn, deps...)` | Complete |
| `effect(() => {})` | `Effect(fn, deps...)` | Complete |
| `signal.set(value)` | `signal.Set(value)` | Complete |
| `signal()` | `signal.Get()` | Complete |
| `signal.update(fn)` | `signal.Update(fn)` | Complete |
| `signal.asReadonly()` | `signal.AsReadonly()` | Complete |
| Automatic tracking | Explicit dependencies | Adapted for Go |
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

**Status**: Stable - Production-ready!
**Version**: v0.1.0

---

*Built with care for the Go community*
