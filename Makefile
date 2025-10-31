# Signals Library Makefile
# Type-safe reactive state management for Go

PROJECT = signals
VERSION ?= v0.1.0-alpha

# Default target
.DEFAULT_GOAL := test

# Build library (check compilation)
build:
	@echo "Building $(PROJECT) library..."
	go build ./...

# Run all tests
test:
	@echo "Running tests..."
	go test -v -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"
	@echo ""
	@echo "Opening coverage report in browser..."
	@if command -v xdg-open > /dev/null; then \
		xdg-open coverage.html; \
	elif command -v open > /dev/null; then \
		open coverage.html; \
	elif command -v start > /dev/null; then \
		start coverage.html; \
	else \
		echo "Please open coverage.html manually"; \
	fi

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	go test -v -race -coverprofile=coverage.out ./...

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Run benchmarks with memory profiling
benchmark-mem:
	@echo "Running benchmarks with memory profiling..."
	go test -bench=. -benchmem -memprofile=mem.out ./...
	@echo "Memory profile: mem.out"
	@echo "View with: go tool pprof -http=:8080 mem.out"

# Run benchmarks with CPU profiling
benchmark-cpu:
	@echo "Running benchmarks with CPU profiling..."
	go test -bench=. -benchmem -cpuprofile=cpu.out ./...
	@echo "CPU profile: cpu.out"
	@echo "View with: go tool pprof -http=:8080 cpu.out"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run --timeout=5m

# Run linter and save results
lint-report:
	@echo "Running linter with report..."
	golangci-lint run --timeout=5m > lint_results.log 2>&1 || true
	@echo "Linter results saved to lint_results.log"
	@cat lint_results.log

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -w -s .
	go mod tidy

# Check code formatting (CI-friendly, no changes)
fmt-check:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "ERROR: The following files are not formatted:"; \
		gofmt -l .; \
		echo ""; \
		echo "Run 'make fmt' to fix formatting issues."; \
		exit 1; \
	fi
	@echo "All files are properly formatted ✓"

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f coverage.out coverage.html
	rm -f lint_results.log
	rm -f mem.out cpu.out
	rm -f *.test
	go clean -testcache

# Build example
example:
	@echo "Building example..."
	go build -o /dev/null ./cmd/example

# Run example
run-example:
	@echo "Running example..."
	go run ./cmd/example

# Development workflow
dev: fmt lint test
	@echo "Development checks complete!"

# CI/CD checks (includes formatting check)
ci: fmt-check vet lint test-race
	@echo "CI checks passed!"

# Pre-commit checks
pre-commit: fmt vet lint test
	@echo "Pre-commit checks complete!"
	@echo ""
	@echo "Summary:"
	@echo "  ✓ Code formatted"
	@echo "  ✓ Vet passed"
	@echo "  ✓ Linter passed"
	@echo "  ✓ Tests passed"
	@echo ""
	@echo "Ready to commit!"

# Pre-release checks (comprehensive)
pre-release: clean fmt-check vet lint test-race benchmark
	@echo "Pre-release checks complete!"
	@echo ""
	@echo "Summary:"
	@echo "  ✓ Code formatted"
	@echo "  ✓ Vet passed"
	@echo "  ✓ Linter passed"
	@echo "  ✓ Tests passed (with race detector)"
	@echo "  ✓ Benchmarks completed"
	@echo ""
	@echo "Ready for release!"

# Install golangci-lint (if not installed)
install-lint:
	@echo "Installing golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
	else \
		echo "golangci-lint is already installed"; \
	fi

# Show coverage percentage
coverage:
	@echo "Calculating coverage..."
	@go test -coverprofile=coverage.out ./... > /dev/null 2>&1
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

# Security scan (if gosec is installed)
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy
	@echo "Dependencies updated!"

# Verify dependencies
deps-verify:
	@echo "Verifying dependencies..."
	go mod verify
	@echo "Dependencies verified!"

# Help
help:
	@echo "Signals Library Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build           - Build library (check compilation)"
	@echo "  make test            - Run tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make test-race       - Run tests with race detector"
	@echo "  make benchmark       - Run benchmarks"
	@echo "  make benchmark-mem   - Run benchmarks with memory profiling"
	@echo "  make benchmark-cpu   - Run benchmarks with CPU profiling"
	@echo "  make lint            - Run linter"
	@echo "  make lint-report     - Run linter and save to file"
	@echo "  make fmt             - Format code"
	@echo "  make fmt-check       - Check code formatting (CI)"
	@echo "  make vet             - Run go vet"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make example         - Build example"
	@echo "  make run-example     - Run example"
	@echo "  make dev             - Full development workflow"
	@echo "  make ci              - CI/CD checks"
	@echo "  make pre-commit      - Pre-commit checks"
	@echo "  make pre-release     - Pre-release checks (comprehensive)"
	@echo "  make install-lint    - Install golangci-lint"
	@echo "  make coverage        - Show coverage percentage"
	@echo "  make security        - Run security scan (requires gosec)"
	@echo "  make deps-update     - Update dependencies"
	@echo "  make deps-verify     - Verify dependencies"
	@echo ""
	@echo "Version: $(VERSION)"
	@echo "Go Version: 1.25+"

.PHONY: build test test-coverage test-race benchmark benchmark-mem benchmark-cpu \
	lint lint-report fmt fmt-check vet clean example run-example \
	dev ci pre-commit pre-release install-lint coverage security \
	deps-update deps-verify help
