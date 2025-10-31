#!/usr/bin/env bash
# Pre-Release Validation Script for Signals Library
# This script runs all quality checks before creating a release
# Ensures production-ready quality standards

set -e  # Exit on first error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Header
echo ""
echo "================================================"
echo "  Signals Library - Pre-Release Check"
echo "  Production-Ready Quality Validation"
echo "================================================"
echo ""

# Track overall status
ERRORS=0
WARNINGS=0

# 1. Check Go version
log_info "Checking Go version..."
GO_VERSION=$(go version | awk '{print $3}')
REQUIRED_VERSION="go1.23"
if [[ "$GO_VERSION" < "$REQUIRED_VERSION" ]]; then
    log_error "Go version $REQUIRED_VERSION+ required, found $GO_VERSION"
    ERRORS=$((ERRORS + 1))
else
    log_success "Go version: $GO_VERSION"
fi
echo ""

# 2. Check git status
log_info "Checking git status..."
if git diff-index --quiet HEAD --; then
    log_success "Working directory is clean"
else
    log_warning "Uncommitted changes detected"
    git status --short
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 3. Code formatting check
log_info "Checking code formatting (gofmt -l .)..."
UNFORMATTED=$(gofmt -l . 2>/dev/null || true)
if [ -n "$UNFORMATTED" ]; then
    # Filter out vendor directories if any
    UNFORMATTED=$(echo "$UNFORMATTED" | grep -v "^vendor/" || true)
fi

if [ -n "$UNFORMATTED" ]; then
    log_error "The following files need formatting:"
    echo "$UNFORMATTED"
    echo ""
    log_info "Run: go fmt ./..."
    ERRORS=$((ERRORS + 1))
else
    log_success "All files are properly formatted"
fi
echo ""

# 4. Go vet
log_info "Running go vet..."
if go vet ./... 2>&1; then
    log_success "go vet passed"
else
    log_error "go vet failed"
    ERRORS=$((ERRORS + 1))
fi
echo ""

# 5. Build all packages
log_info "Building all packages..."
if go build ./... 2>&1; then
    log_success "Build successful"
else
    log_error "Build failed"
    ERRORS=$((ERRORS + 1))
fi
echo ""

# 6. go.mod validation
log_info "Validating go.mod..."
go mod verify
if [ $? -eq 0 ]; then
    log_success "go.mod verified"
else
    log_error "go.mod verification failed"
    ERRORS=$((ERRORS + 1))
fi

# Check if go.mod needs tidying
go mod tidy
if git diff --quiet go.mod go.sum 2>/dev/null; then
    log_success "go.mod is tidy"
else
    log_warning "go.mod needs tidying (run 'go mod tidy')"
    git diff go.mod go.sum 2>/dev/null || true
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 6.5. Verify golangci-lint configuration
log_info "Verifying golangci-lint configuration..."
if command -v golangci-lint &> /dev/null; then
    if golangci-lint config verify 2>&1; then
        log_success "golangci-lint config is valid"
    else
        log_error "golangci-lint config is invalid"
        ERRORS=$((ERRORS + 1))
    fi
else
    log_warning "golangci-lint not installed - skipping config verification"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 7. Run tests with race detector (CRITICAL for signals library!)
USE_WSL=0
WSL_DISTRO=""

# Helper function to find WSL distro with Go installed
find_wsl_distro() {
    if ! command -v wsl &> /dev/null; then
        return 1
    fi

    # Try common distros first
    for distro in "Gentoo" "Ubuntu" "Debian" "Alpine"; do
        if wsl -d "$distro" bash -c "command -v go &> /dev/null" 2>/dev/null; then
            echo "$distro"
            return 0
        fi
    done

    return 1
}

if command -v gcc &> /dev/null || command -v clang &> /dev/null; then
    log_info "Running tests with race detector..."
    RACE_FLAG="-race"
    TEST_CMD="go test -race ./... 2>&1"
else
    # Try to find WSL distro with Go
    WSL_DISTRO=$(find_wsl_distro)
    if [ -n "$WSL_DISTRO" ]; then
        log_info "GCC not found locally, but WSL2 ($WSL_DISTRO) detected!"
        log_info "Running tests with race detector via WSL2 $WSL_DISTRO..."
        USE_WSL=1
        RACE_FLAG="-race"

        # Convert Windows path to WSL path (D:\projects\signals -> /mnt/d/projects/signals)
        # pwd in MSYS returns /d/projects/signals, need to convert to /mnt/d/projects/signals
        CURRENT_DIR=$(pwd)
        if [[ "$CURRENT_DIR" =~ ^/([a-z])/ ]]; then
            # Already in /d/... format (MSYS), convert to /mnt/d/...
            WSL_PATH="/mnt${CURRENT_DIR}"
        else
            # Windows format D:\... convert to /mnt/d/...
            DRIVE_LETTER=$(echo "$CURRENT_DIR" | cut -d: -f1 | tr '[:upper:]' '[:lower:]')
            PATH_WITHOUT_DRIVE=${CURRENT_DIR#*:}
            WSL_PATH="/mnt/$DRIVE_LETTER${PATH_WITHOUT_DRIVE//\\//}"
        fi

        TEST_CMD="wsl -d \"$WSL_DISTRO\" bash -c \"cd \\\"$WSL_PATH\\\" && go test -race ./... 2>&1\""
    else
        log_warning "GCC not found, running tests WITHOUT race detector"
        log_info "Install GCC (mingw-w64) or setup WSL2 with Go for race detection"
        log_info "  Windows: https://www.mingw-w64.org/"
        log_info "  WSL2: https://docs.microsoft.com/en-us/windows/wsl/install"
        WARNINGS=$((WARNINGS + 1))
        RACE_FLAG=""
        TEST_CMD="go test ./... 2>&1"
    fi
fi

if [ $USE_WSL -eq 1 ]; then
    TEST_OUTPUT=$(eval "$TEST_CMD")
else
    TEST_OUTPUT=$(eval "$TEST_CMD")
fi

# Check if race detector failed to build (known issue with some Go versions)
if echo "$TEST_OUTPUT" | grep -q "hole in findfunctab\|build failed.*race"; then
    log_warning "Race detector build failed (known Go runtime issue)"
    log_info "Falling back to tests without race detector..."

    if [ $USE_WSL -eq 1 ]; then
        TEST_CMD="wsl -d \"$WSL_DISTRO\" bash -c \"cd \\\"$WSL_PATH\\\" && go test ./... 2>&1\""
    else
        TEST_CMD="go test ./... 2>&1"
    fi

    TEST_OUTPUT=$(eval "$TEST_CMD")
    RACE_FLAG=""
    WARNINGS=$((WARNINGS + 1))
fi

if echo "$TEST_OUTPUT" | grep -q "FAIL"; then
    log_error "Tests failed or race conditions detected"
    echo "$TEST_OUTPUT"
    echo ""
    if [ -n "$RACE_FLAG" ]; then
        log_error "Race conditions are CRITICAL failures for signals library!"
    fi
    ERRORS=$((ERRORS + 1))
elif echo "$TEST_OUTPUT" | grep -q "PASS\|ok"; then
    if [ $USE_WSL -eq 1 ] && [ -n "$RACE_FLAG" ]; then
        log_success "All tests passed with race detector (via WSL2 $WSL_DISTRO)"
    elif [ -n "$RACE_FLAG" ]; then
        log_success "All tests passed with race detector (0 races)"
    else
        log_success "All tests passed (race detector not available)"
    fi
else
    log_error "Unexpected test output"
    echo "$TEST_OUTPUT"
    ERRORS=$((ERRORS + 1))
fi
echo ""

# 8. Stress test race detector (signals-specific)
if [ -n "$RACE_FLAG" ]; then
    log_info "Running race detector stress test (10 iterations)..."
    RACE_STRESS_PASS=true
    for i in {1..10}; do
        if [ $USE_WSL -eq 1 ]; then
            if ! wsl -d "$WSL_DISTRO" bash -c "cd \"$WSL_PATH\" && go test -race -count=1 ./... 2>&1" &> /dev/null; then
                RACE_STRESS_PASS=false
                break
            fi
        else
            if ! go test -race -count=1 ./... &> /dev/null; then
                RACE_STRESS_PASS=false
                break
            fi
        fi
    done

    if [ "$RACE_STRESS_PASS" = true ]; then
        log_success "Race detector stress test passed (10/10)"
    else
        log_error "Race detector stress test failed at iteration $i"
        ERRORS=$((ERRORS + 1))
    fi
else
    log_warning "Race detector not available, skipping stress test"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 9. Test coverage check (CRITICAL: 90%+ requirement)
log_info "Checking test coverage..."
COVERAGE_FILE=$(mktemp)
go test -cover -coverprofile="$COVERAGE_FILE" ./... > /dev/null 2>&1
COVERAGE=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | grep "total:" | awk '{print $3}' | sed 's/%//')
rm -f "$COVERAGE_FILE"

if [ -n "$COVERAGE" ]; then
    echo "  • Total coverage: ${COVERAGE}%"
    if awk -v cov="$COVERAGE" 'BEGIN {exit !(cov >= 90.0)}'; then
        log_success "Coverage meets requirement (≥90%)"
    else
        log_error "Coverage below 90% (${COVERAGE}%)"
        log_error "Signals library requires 90%+ coverage for production"
        ERRORS=$((ERRORS + 1))
    fi
else
    log_error "Could not determine coverage"
    ERRORS=$((ERRORS + 1))
fi
echo ""

# 10. Benchmarks validation
log_info "Running benchmarks..."
BENCH_FILE=$(mktemp)
if go test -bench=. -benchmem ./... > "$BENCH_FILE" 2>&1; then
    log_success "Benchmarks completed"

    # Check Signal.Get performance (target: < 15ns/op)
    SIGNAL_GET=$(grep "BenchmarkSignal.*Get" "$BENCH_FILE" 2>/dev/null | awk '{print $3}' | head -1 | sed 's/ns\/op//')
    if [ -n "$SIGNAL_GET" ]; then
        echo "  • Signal.Get(): ${SIGNAL_GET}ns/op"
        if awk -v perf="$SIGNAL_GET" 'BEGIN {exit !(perf < 15)}'; then
            log_success "Signal.Get() meets target (<15ns/op)"
        else
            log_warning "Signal.Get() slower than target (${SIGNAL_GET}ns/op vs 15ns/op)"
            WARNINGS=$((WARNINGS + 1))
        fi
    fi

    # Check Signal.Set performance (target: < 200ns/op)
    SIGNAL_SET=$(grep "BenchmarkSignal.*Set" "$BENCH_FILE" 2>/dev/null | awk '{print $3}' | head -1 | sed 's/ns\/op//')
    if [ -n "$SIGNAL_SET" ]; then
        echo "  • Signal.Set(): ${SIGNAL_SET}ns/op"
        if awk -v perf="$SIGNAL_SET" 'BEGIN {exit !(perf < 200)}'; then
            log_success "Signal.Set() meets target (<200ns/op)"
        else
            log_warning "Signal.Set() slower than target (${SIGNAL_SET}ns/op vs 200ns/op)"
            WARNINGS=$((WARNINGS + 1))
        fi
    fi
else
    log_warning "Benchmarks not yet implemented or failed"
    WARNINGS=$((WARNINGS + 1))
fi
rm -f "$BENCH_FILE"
echo ""

# 11. golangci-lint (if available)
log_info "Running golangci-lint..."
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run --timeout=5m ./... 2>&1 | tail -5 | grep -q "0 issues"; then
        log_success "golangci-lint passed with 0 issues"
    else
        log_warning "Linter found issues"
        golangci-lint run --timeout=5m ./... 2>&1 | tail -10
        WARNINGS=$((WARNINGS + 1))
    fi
else
    log_warning "golangci-lint not installed (optional but recommended)"
    log_info "Install: https://golangci-lint.run/welcome/install/"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 12. Check for critical TODO/FIXME comments in production code
log_info "Checking for TODO/FIXME comments in production code..."
TODO_COUNT=$(grep -r "TODO\|FIXME" --include="*.go" --exclude-dir=vendor --exclude="*_test.go" . 2>/dev/null | wc -l)
if [ "$TODO_COUNT" -gt 0 ]; then
    log_warning "Found $TODO_COUNT TODO/FIXME comments in production code"
    grep -r "TODO\|FIXME" --include="*.go" --exclude-dir=vendor --exclude="*_test.go" . 2>/dev/null | head -5
    log_info "Consider resolving before release"
    WARNINGS=$((WARNINGS + 1))
else
    log_success "No TODO/FIXME comments in production code"
fi
echo ""

# 13. Check documentation (signals-specific)
log_info "Checking documentation..."
DOCS_MISSING=0
REQUIRED_DOCS=".claude/QUICKREF.md .claude/STATUS.md .claude/CLAUDE.md docs/dev/IMPLEMENTATION_GUIDE.md docs/dev/ARCHITECTURE.md"

for doc in $REQUIRED_DOCS; do
    if [ ! -f "$doc" ]; then
        log_error "Missing critical doc: $doc"
        DOCS_MISSING=1
        ERRORS=$((ERRORS + 1))
    fi
done

# Check if public docs are ready (for v1.0+)
if [ -f "README.md" ] && [ -f "API.md" ] && [ -f "GUIDE.md" ]; then
    log_success "Public documentation complete (README, API, GUIDE)"
elif [[ "$VERSION" == v1.* ]]; then
    log_error "Public documentation required for v1.0+ releases"
    DOCS_MISSING=1
    ERRORS=$((ERRORS + 1))
else
    log_info "Public documentation not yet required (pre-v1.0)"
fi

if [ $DOCS_MISSING -eq 0 ]; then
    log_success "All critical documentation present"
fi
echo ""

# 14. Check for global state (signals anti-pattern)
log_info "Checking for global state (anti-pattern for signals)..."
GLOBAL_VARS=$(grep -r "^var.*=" --include="*.go" --exclude="*_test.go" --exclude-dir=vendor . 2>/dev/null | grep -v "^var.*func" | grep -v "^var.*interface" | wc -l)
if [ "$GLOBAL_VARS" -gt 0 ]; then
    log_warning "Found $GLOBAL_VARS potential global variables (review required)"
    grep -r "^var.*=" --include="*.go" --exclude="*_test.go" --exclude-dir=vendor . 2>/dev/null | grep -v "^var.*func" | grep -v "^var.*interface" | head -5
    log_info "Signals library should avoid global state"
    WARNINGS=$((WARNINGS + 1))
else
    log_success "No global state detected"
fi
echo ""

# 15. Check API consistency (signals-specific)
log_info "Checking API consistency..."
API_ISSUES=0

# Check for ReadonlySignal interface
if ! grep -q "type ReadonlySignal\[" *.go 2>/dev/null; then
    log_error "ReadonlySignal interface not found (required for v1.0)"
    API_ISSUES=1
    ERRORS=$((ERRORS + 1))
fi

# Check for AsReadonly method
if ! grep -q "AsReadonly()" *.go 2>/dev/null; then
    log_warning "AsReadonly() method not found (encapsulation pattern)"
    API_ISSUES=1
    WARNINGS=$((WARNINGS + 1))
fi

# Check for context.Context support
if ! grep -q "context.Context" *.go 2>/dev/null; then
    log_warning "context.Context support not detected"
    API_ISSUES=1
    WARNINGS=$((WARNINGS + 1))
fi

if [ $API_ISSUES -eq 0 ]; then
    log_success "API consistency checks passed"
fi
echo ""

# 16. Check STATUS.md is up-to-date
log_info "Checking STATUS.md..."
if [ -f ".claude/STATUS.md" ]; then
    LAST_UPDATE=$(grep "Last Updated" .claude/STATUS.md | head -1 | awk '{print $4}')
    TODAY=$(date +%Y-%m-%d)

    if [ "$LAST_UPDATE" == "$TODAY" ]; then
        log_success "STATUS.md updated today"
    else
        log_warning "STATUS.md last updated: $LAST_UPDATE (consider updating)"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    log_warning "STATUS.md not found"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# Summary
echo "================================================"
echo "  Summary"
echo "================================================"
echo ""

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    log_success "✅ All checks passed! Ready for release."
    echo ""
    log_info "Next steps for release:"
    echo "  1. Update .claude/STATUS.md with release milestone"
    echo "  2. Create release branch: git checkout -b release/vX.Y.Z"
    echo "  3. Update version in go.mod (if applicable)"
    echo "  4. Create CHANGELOG.md entry"
    echo "  5. Commit: git commit -m 'chore: prepare vX.Y.Z release'"
    echo "  6. Push and create PR: git push origin release/vX.Y.Z"
    echo "  7. After PR merge, create tag: git tag -a vX.Y.Z -m 'Release vX.Y.Z'"
    echo "  8. Push tag: git push origin vX.Y.Z"
    echo "  9. Create GitHub release with notes"
    echo ""
    exit 0
elif [ $ERRORS -eq 0 ]; then
    log_warning "⚠️  Checks completed with $WARNINGS warning(s)"
    echo ""
    log_info "Review warnings above. Consider fixing before release."
    echo ""
    exit 0
else
    log_error "❌ Checks failed with $ERRORS error(s) and $WARNINGS warning(s)"
    echo ""
    log_error "CRITICAL: Fix all errors before creating release"
    echo ""
    log_info "Common fixes:"
    echo "  • Race conditions: Review ARCHITECTURE.md for thread-safety patterns"
    echo "  • Coverage < 90%: Add more unit tests"
    echo "  • Format issues: Run 'go fmt ./...'"
    echo "  • Go vet errors: Fix reported issues"
    echo ""
    exit 1
fi
