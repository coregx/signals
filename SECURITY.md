# Security Policy

## Supported Versions

Signals library is currently in beta. We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.0-beta | :white_check_mark: |
| < 0.1.0 | :x:                |

Future stable releases (v1.0+) will follow semantic versioning with long-term support.

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability in the Signals library, please report it responsibly.

### How to Report

**DO NOT** open a public GitHub issue for security vulnerabilities.

Instead, please report security issues by:

1. **Private Security Advisory** (preferred):
   https://github.com/coregx/signals/security/advisories/new

2. **Email** to maintainers:
   Create a private GitHub issue or contact via discussions

### What to Include

Please include the following information in your report:

- **Description** of the vulnerability
- **Steps to reproduce** the issue (include minimal code example)
- **Affected versions** (which versions are impacted)
- **Potential impact** (data races, memory corruption, DoS, etc.)
- **Suggested fix** (if you have one)
- **Your contact information** (for follow-up questions)

### Response Timeline

- **Initial Response**: Within 48-72 hours
- **Triage & Assessment**: Within 1 week
- **Fix & Disclosure**: Coordinated with reporter

We aim to:
1. Acknowledge receipt within 72 hours
2. Provide an initial assessment within 1 week
3. Work with you on a coordinated disclosure timeline
4. Credit you in the security advisory (unless you prefer to remain anonymous)

## Security Considerations for Reactive State

Reactive state management in a concurrent environment introduces specific security considerations.

### 1. Thread-Safety Guarantees

**Risk**: Concurrent access to shared state can cause data races and undefined behavior.

**Mitigation in Library**:
- All public APIs are thread-safe by design
- RWMutex protection on all signal operations
- Atomic state updates in computed values
- Effect execution serialization

**User Responsibilities**:
```go
// ✅ GOOD - Library handles thread-safety
signal := signals.NewSignal(0)
go signal.Set(1)  // Safe
go signal.Get()   // Safe

// ❌ BAD - Don't bypass library APIs
type MySignal struct {
    *signals.Signal[int]
}
// Direct field access not protected
```

### 2. Panic Recovery

**Risk**: User-provided functions (computed, effects, updaters) can panic and crash the application.

**Mitigation**:
- Built-in panic recovery in all user callbacks
- Custom panic handlers via `WithPanicHandler` option
- Panics logged and converted to errors where possible

**Best Practices**:
```go
// ✅ GOOD - Panic is caught and logged
computed := signals.NewComputed(func() int {
    panic("oops")  // Caught by library
})

// ✅ BETTER - Handle errors explicitly
computed := signals.NewComputed(func() int {
    if value < 0 {
        log.Printf("Invalid value: %d", value)
        return 0
    }
    return value
})
```

### 3. Memory Leaks

**Risk**: Unmanaged subscriptions and effects can leak memory.

**Mitigation**:
- Effects provide cleanup functions
- Subscriptions can be unsubscribed
- Context-based cancellation for long-running effects

**User Best Practices**:
```go
// ❌ BAD - Effect leaks if not cleaned up
effect := signals.NewEffect(func() {
    // Long-running work
})

// ✅ GOOD - Clean up when done
effect := signals.NewEffect(func() {
    // Work
})
defer effect.Cleanup()

// ✅ BETTER - Context-based lifecycle
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

effect := signals.NewEffect(func() {
    select {
    case <-ctx.Done():
        return
    default:
        // Work
    }
})
```

### 4. Resource Exhaustion

**Risk**: Circular dependencies or unbounded effect chains can cause infinite loops.

**Mitigation**:
- Circular dependency detection in computed values
- Maximum recursion depth limits
- Effect batching prevents cascading updates

**Detection**:
```go
// ❌ BAD - Circular dependency detected and panics
a := signals.NewSignal(0)
b := signals.NewComputed(func() int {
    return a.Get() + 1
})
c := signals.NewComputed(func() int {
    return b.Get() + 1
})
// If 'a' depends on 'c', library detects and panics

// ✅ GOOD - Acyclic dependency graph
a := signals.NewSignal(0)
b := signals.NewComputed(func() int {
    return a.Get() + 1
})
c := signals.NewComputed(func() int {
    return b.Get() + 1
})
```

### 5. Type Safety

**Risk**: Type assertions and interface conversions can panic if misused.

**Mitigation**:
- Full generic support (Go 1.25+)
- Compile-time type checking
- No runtime type assertions in hot paths

**Type Safety**:
```go
// ✅ GOOD - Type-safe at compile time
signal := signals.NewSignal[int](0)
signal.Set(42)        // OK
signal.Set("hello")   // Compile error

// ✅ GOOD - Type inference
computed := signals.NewComputed(func() string {
    return fmt.Sprintf("Value: %d", signal.Get())
})
```

## Known Security Considerations

### 1. Concurrent Access

**Status**: Fully mitigated.

**Risk Level**: Low

**Description**: Signals library is designed for concurrent access. All public APIs are protected with RWMutex.

**Mitigation**:
- RWMutex on all signal state
- Read-write lock optimization (multiple readers, single writer)
- Race detector clean (verified in CI)

### 2. User-Provided Functions

**Status**: Partially mitigated.

**Risk Level**: Medium

**Description**: Computed values and effects execute user-provided functions. Malicious or buggy functions can:
- Panic and crash
- Infinite loop
- Excessive memory allocation

**Mitigation**:
- Panic recovery with custom handlers
- Circular dependency detection
- No automatic mitigation for infinite loops (user responsibility)

**User Recommendations**:
- Timeout long-running computations
- Use context for cancellation
- Test computed functions thoroughly

### 3. Memory Management

**Status**: Active monitoring.

**Risk Level**: Low to Medium

**Description**: Subscriptions and effects can leak if not properly cleaned up.

**Mitigation**:
- Cleanup functions for effects
- Unsubscribe for subscriptions
- Documentation of lifecycle management

**Testing**:
- Memory leak tests in test suite
- Benchmark tests verify zero allocations in hot paths

### 4. Dependency Security

Signals library has zero runtime dependencies:

- **No external dependencies** (pure Go standard library)
- **No C dependencies** (no CGo)
- **Minimal attack surface**

**Development Dependencies** (testing only):
- None (uses standard `testing` package)

## Security Testing

### Current Testing

- Unit tests with concurrent access (race detector)
- Panic recovery tests
- Memory leak tests
- Thread-safety verification
- 51 tests, 67.9% coverage

### Continuous Integration

- Race detector on all tests (`go test -race`)
- Cross-platform testing (Linux, macOS, Windows)
- golangci-lint with 30+ linters including security checks

### Planned for v1.0

- Fuzzing for user-provided functions
- Static analysis with gosec
- Formal verification of thread-safety guarantees
- Stress testing with millions of concurrent operations

## Security Best Practices for Users

### 1. Always Clean Up Effects

```go
// Create effect with cleanup
effect := signals.NewEffect(func() {
    // Work
}, signals.WithCleanup(func() {
    // Cleanup resources
}))

// Clean up when done
defer effect.Cleanup()
```

### 2. Use Context for Long-Running Work

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

signals.NewEffect(func() {
    select {
    case <-ctx.Done():
        return
    default:
        // Work
    }
})
```

### 3. Handle Panics in User Code

```go
// Library catches panics, but handle explicitly when possible
signals.NewComputed(func() int {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Panic in computed: %v", r)
        }
    }()
    // Work that might panic
    return value
})
```

### 4. Validate Inputs

```go
signal := signals.NewSignal(0)

// Validate before setting
signal.Update(func(current int) int {
    newValue := current + 1
    if newValue > maxValue {
        return maxValue
    }
    return newValue
})
```

### 5. Use Read-Only Views

```go
// Expose read-only view to untrusted code
signal := signals.NewSignal(42)
readonly := signal.AsReadonly()

// Safe to share - cannot modify
func shareWithUntrusted(r signals.Readable[int]) {
    value := r.Get()  // OK
    // r.Set(100)     // Compile error
}
```

## Security Disclosure History

No security vulnerabilities have been reported or fixed yet (project is in beta).

When vulnerabilities are addressed, they will be listed here with:
- **CVE ID** (if assigned)
- **Affected versions**
- **Fixed in version**
- **Severity** (Critical/High/Medium/Low)
- **Credit** to reporter

## Security Contact

- **GitHub Security Advisory**: https://github.com/coregx/signals/security/advisories/new
- **Public Issues** (for non-sensitive bugs): https://github.com/coregx/signals/issues
- **Discussions**: https://github.com/coregx/signals/discussions

## Bug Bounty Program

Signals library does not currently have a bug bounty program. We rely on responsible disclosure from the security community.

If you report a valid security vulnerability:
- Public credit in security advisory (if desired)
- Acknowledgment in CHANGELOG
- Our gratitude and recognition in README
- Priority review and quick fix

---

**Thank you for helping keep Signals secure!**

*Security is a journey, not a destination. We continuously improve our security posture with each release.*
