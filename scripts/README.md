# Scripts - Automation Tools

Scripts for automating development and release tasks.

---

## 📋 Available Scripts

### pre-release-check.sh / .bat

**Purpose**: Comprehensive quality validation before creating a release

**Runs**:
- ✅ Go version check (1.23+)
- ✅ Git status validation
- ✅ Code formatting (gofmt)
- ✅ Go vet
- ✅ Build verification
- ✅ go.mod validation
- ✅ **Tests with race detector** (CRITICAL!)
- ✅ **Race detector stress test** (10 iterations)
- ✅ **Coverage check** (90%+ required)
- ✅ Benchmarks (performance targets)
- ✅ golangci-lint (if available)
- ✅ TODO/FIXME check
- ✅ Documentation validation
- ✅ Global state detection
- ✅ API consistency checks
- ✅ STATUS.md freshness check

**Usage**:

```bash
# Linux/macOS/Git Bash
./scripts/pre-release-check.sh

# Windows
scripts\pre-release-check.bat
```

**Exit Codes**:
- `0` - All checks passed (or warnings only)
- `1` - Critical errors detected

**When to Run**:
- Before creating release branch
- Before merging to main
- After major refactoring
- Weekly as sanity check

---

## 🎯 Quality Standards

### Required for Release

| Check | Requirement | Severity |
|-------|-------------|----------|
| Race conditions | 0 races | 🔴 CRITICAL |
| Test coverage | ≥ 90% | 🔴 CRITICAL |
| Go vet | 0 issues | 🔴 CRITICAL |
| Build | Success | 🔴 CRITICAL |
| Code format | gofmt clean | 🔴 CRITICAL |

### Performance Targets

| Metric | Target | Severity |
|--------|--------|----------|
| Signal.Get() | < 15ns/op | 🟡 WARNING |
| Signal.Set() | < 200ns/op | 🟡 WARNING |
| Subscribe | < 100ns/op | 🟡 WARNING |
| Computed (clean) | < 15ns/op | 🟡 WARNING |

---

## 🔧 Requirements

### Minimum (Linux/macOS/Windows with Git Bash)

- Go 1.23+
- Git
- GCC or Clang (for race detector)

### Recommended

- golangci-lint (for additional linting)

### Installing GCC

**Linux**:
```bash
sudo apt-get install gcc
```

**macOS**:
```bash
xcode-select --install
```

**Windows**:
- Install [Git for Windows](https://git-scm.com/download/win) (includes Git Bash)
- Install [mingw-w64](https://www.mingw-w64.org/) for GCC

---

## 📊 Example Output

```
================================================
  Signals Library - Pre-Release Check
  Production-Ready Quality Validation
================================================

[INFO] Checking Go version...
[SUCCESS] Go version: go1.23.1

[INFO] Checking git status...
[SUCCESS] Working directory is clean

[INFO] Checking code formatting (gofmt -l .)...
[SUCCESS] All files are properly formatted

[INFO] Running go vet...
[SUCCESS] go vet passed

[INFO] Building all packages...
[SUCCESS] Build successful

[INFO] Running tests with race detector...
[SUCCESS] All tests passed with race detector (0 races)

[INFO] Running race detector stress test (10 iterations)...
[SUCCESS] Race detector stress test passed (10/10)

[INFO] Checking test coverage...
  • Total coverage: 92.5%
[SUCCESS] Coverage meets requirement (≥90%)

[INFO] Running benchmarks...
[SUCCESS] Benchmarks completed
  • Signal.Get(): 12.3ns/op
[SUCCESS] Signal.Get() meets target (<15ns/op)
  • Signal.Set(): 185ns/op
[SUCCESS] Signal.Set() meets target (<200ns/op)

================================================
  Summary
================================================

[SUCCESS] ✅ All checks passed! Ready for release.

Next steps for release:
  1. Update .claude/STATUS.md with release milestone
  2. Create release branch: git checkout -b release/vX.Y.Z
  3. Update version in go.mod (if applicable)
  4. Create CHANGELOG.md entry
  5. Commit: git commit -m 'chore: prepare vX.Y.Z release'
  ...
```

---

## 🚨 Common Issues

### Race Detector Not Available

**Error**:
```
[ERROR] GCC/Clang not found - race detector REQUIRED
```

**Fix**: Install GCC/Clang (see Requirements above)

### Coverage Below 90%

**Error**:
```
[ERROR] Coverage below 90% (85.2%)
Signals library requires 90%+ coverage for production
```

**Fix**: Add more unit tests, especially for:
- Edge cases
- Error paths
- Concurrent scenarios

### Uncommitted Changes

**Warning**:
```
[WARNING] Uncommitted changes detected
```

**Fix**: Commit or stash changes before release

### go.mod Needs Tidying

**Warning**:
```
[WARNING] go.mod needs tidying (run 'go mod tidy')
```

**Fix**:
```bash
go mod tidy
git add go.mod go.sum
git commit -m "chore: tidy go.mod"
```

---

## 📝 Notes

- **Windows users**: The `.bat` file automatically calls the `.sh` script via Git Bash
- **CI/CD**: This script should match CI pipeline checks exactly
- **Race detector**: Signals library REQUIRES race detector validation (thread-safety critical)
- **Coverage**: 90%+ is non-negotiable for production-ready reactive library

---

*Last Updated: 2025-10-31*
