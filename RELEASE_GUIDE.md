# Release Guide

> Guidelines for creating releases of the Signals library

This document describes the release process for maintainers and contributors with write access.

---

## Prerequisites

Before creating a release, ensure:

- ✅ All tests pass locally and in CI
- ✅ Code coverage meets requirements (>70%)
- ✅ golangci-lint reports 0 issues
- ✅ Documentation is up to date
- ✅ CHANGELOG.md includes all changes

---

## Safety First

**Always create a backup before destructive git operations:**

```bash
# Create a git bundle backup
git bundle create ../signals-backup-$(date +%Y%m%d).bundle --all
```

**Restore from backup if needed:**
```bash
git clone ../signals-backup-YYYYMMDD.bundle signals-restore
```

---

## 🎯 Git Flow Strategy

### Branch Structure

```
main        - Production-ready code ONLY (protected, green CI always)
  ↑
release/*   - Release candidates (RC)
  ↑
develop     - Active development (default branch for PRs)
  ↑
feature/*   - Feature branches
```

### Branch Rules

#### `main` Branch
- ✅ **ALWAYS** production-ready
- ✅ **ALWAYS** green CI (all tests passing)
- ✅ **ONLY** accepts merges from `release/*` branches
- ❌ **NEVER** commit directly to main
- ❌ **NEVER** push without green CI
- ❌ **NEVER** force push
- 🏷️ **Tags created ONLY after CI passes**

#### `develop` Branch
- Default branch for development
- Accepts feature branches
- May contain work-in-progress code
- Should pass tests, but can have warnings
- **Current default branch**

#### `release/*` Branches
- Format: `release/v0.1.0`, `release/v0.2.0`, `release/v1.0.0`
- Created from `develop`
- Only bug fixes and documentation updates allowed
- No new features
- Merges to both `main` and `develop`

#### `feature/*` Branches
- Format: `feature/batch-updates`, `feature/resource-tracking`
- Created from `develop`
- Merged back to `develop` with `--no-ff`

---

## 🔧 Pre-Release Validation Script

### Location
`scripts/pre-release-check.sh` (and `.bat` for Windows)

### Purpose
Runs **all quality checks locally** before creating a release, matching CI requirements exactly.

### When to Use

#### 1. Before Every Commit (Recommended)
```bash
# Quick validation before committing
bash scripts/pre-release-check.sh

# If script passes (green/yellow), safe to commit:
git add .
git commit -m "..."
git push
```

#### 2. Before Creating Release Branch (Mandatory)
```bash
# MUST pass before starting release process
bash scripts/pre-release-check.sh

# Only proceed if output shows:
# ✅ "All checks passed! Ready for release."
```

#### 3. Before Merging to Main (Mandatory)
```bash
# Final validation on release branch
git checkout release/v0.2.0
bash scripts/pre-release-check.sh

# If errors found, fix them before merging to main
```

#### 4. After Major Changes (Recommended)
- After refactoring
- After dependency updates
- After documentation updates
- After fixing bugs

### What the Script Validates

1. **Go version**: 1.25+ required
2. **Code formatting**: `gofmt -l .` must be clean
3. **Static analysis**: `go vet ./...` must pass
4. **Build**: `go build ./...` must succeed
5. **go.mod**: `go mod verify` and `go mod tidy` check
6. **Tests**: All tests passing (51 tests for v0.1.0-beta)
7. **Coverage**: >70% required
8. **Race detector**: Clean (if GCC available)
9. **golangci-lint**: 0 issues required
10. **TODO/FIXME**: Check for pending work
11. **Documentation**: All critical files present
12. **Examples**: Verify example code runs

### Exit Codes

- **0 (green)**: All checks passed, ready for release
- **0 (yellow)**: Checks passed with warnings (review recommended)
- **1 (red)**: Checks failed with errors (must fix before release)

### Warnings vs Errors

**Warnings (yellow)** - Non-blocking, but review recommended:
- Uncommitted changes detected
- GCC not found (Windows) - race detector unavailable
- Test coverage slightly below target
- Some TODO comments

**Errors (red)** - Blocking, must fix:
- Code not formatted
- go vet failures
- Build failures
- Test failures
- golangci-lint issues (must be 0)
- Coverage significantly below 70%
- Missing documentation files

---

## 📋 Version Naming

### Semantic Versioning

Format: `MAJOR.MINOR.PATCH[-PRERELEASE]`

Examples:
- `v0.1.0-beta` - Current version (core complete)
- `v0.1.0` - Stable core (bug fixes only)
- `v0.2.0` - Documentation + Examples
- `v0.3.0` - Advanced features
- `v1.0.0-rc.1` - Release candidate 1
- `v1.0.0` - First stable release
- `v1.1.0` - Minor feature update
- `v1.1.1` - Patch/bugfix

### Version Increment Rules

**MAJOR** (1.0.0 → 2.0.0):
- Breaking API changes
- Major architectural changes
- Requires migration guide
- **NOTE**: For Go, MAJOR v2+ requires new module path (e.g., `/v2`)

**MINOR** (0.1.0 → 0.2.0):
- New features (backward compatible)
- New reactive primitives
- Performance improvements
- API additions (no breaking changes)

**PATCH** (0.1.0 → 0.1.1):
- Bug fixes
- Performance improvements
- Documentation updates
- Security patches

**PRERELEASE**:
- `-alpha` - Early testing, unstable API
- `-beta` - Feature complete for milestone, testing phase
- `-rc.N` - Release candidate (N = 1, 2, 3...)

### Signals Library Versioning Strategy

**Current Path**: `v0.x.x` until `v1.0.0`

- `v0.1.0-beta`: Core complete (current)
- `v0.1.0`: Stable core (bug fixes)
- `v0.2.0`: Documentation + Examples
- `v0.3.0`: Advanced features (batching, resource tracking)
- `v1.0.0-rc.1`: Release candidate (API stable)
- `v1.0.0`: First stable release

**Rationale**: Avoid `v2.0.0` approach (requires new import path). Use `v0.x.x` progression until feature-complete, then `v1.0.0` stable.

---

## ✅ Pre-Release Checklist

**CRITICAL**: Complete ALL items before creating release branch!

### 1. Automated Quality Checks

**Run our pre-release validation script**:

```bash
# ONE COMMAND runs ALL checks (matches CI exactly)
bash scripts/pre-release-check.sh
```

This script validates:
- ✅ Go version (1.25+)
- ✅ Code formatting (gofmt)
- ✅ Static analysis (go vet)
- ✅ All tests passing (51 tests)
- ✅ Race detector
- ✅ Coverage >70%
- ✅ golangci-lint (0 issues required)
- ✅ go.mod integrity
- ✅ No TODO/FIXME comments
- ✅ All documentation present
- ✅ Examples run successfully

**Manual checks** (if script not available):

```bash
# Format code
go fmt ./...

# Verify formatting
if [ -n "$(gofmt -l .)" ]; then
  echo "ERROR: Code not formatted"
  gofmt -l .
  exit 1
fi

# Static analysis
go vet ./...

# Linting (strict)
golangci-lint run --config .golangci.yml --timeout=5m ./...
# Must show: "0 issues."

# All tests
go test ./...
# All must PASS

# Coverage check
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1
# Minimum: >70%

# Race detector (if GCC available)
go test -race ./...
```

### 2. Dependencies

```bash
# Verify modules
go mod verify

# Tidy and check diff
go mod tidy
git diff go.mod go.sum
# Should show NO changes

# Check dependencies
go list -m all
# Signals has ZERO production dependencies (✅ pure Go)
```

### 3. Documentation

- [ ] README.md updated with latest features
- [ ] CHANGELOG.md entry created for this version
- [ ] All public APIs have godoc comments
- [ ] Examples are up-to-date and tested
- [ ] Migration guide (if breaking changes)
- [ ] ROADMAP.md updated with progress
- [ ] Known limitations documented

### 4. GitHub Actions

- [ ] `.github/workflows/test.yml` exists
- [ ] CI passes on latest `develop` commit
- [ ] Coverage badge updated (if changed)

### 5. Project-Specific Checks

**Signals Library Requirements**:
- [ ] All reactive primitives working (Signal, Computed, Effect)
- [ ] Test coverage >70%
- [ ] Zero allocations in hot paths (verified by benchmarks)
- [ ] Thread-safe (race detector clean)
- [ ] Angular Signals compatibility maintained
- [ ] All examples run without errors
- [ ] Benchmark results documented
- [ ] No regressions in existing features

---

## 🚀 Release Process

### Step 1: Create Release Branch

```bash
# Ensure you're on develop and up-to-date
git checkout develop
git pull origin develop

# Verify develop is clean
git status
# Should show: "nothing to commit, working tree clean"

# Run ALL pre-release checks (CRITICAL!)
bash scripts/pre-release-check.sh
# Script must exit with: "All checks passed! Ready for release."
# If errors: FIX THEM before proceeding!

# Create release branch (example: v0.2.0)
git checkout -b release/v0.2.0

# Update version in files
# - README.md (version badges)
# - CHANGELOG.md (add version section)
# - ROADMAP.md (update status)

git add .
git commit -m "chore: prepare v0.2.0 release"
git push origin release/v0.2.0
```

### Step 2: Wait for CI (CRITICAL!)

```bash
# Go to GitHub Actions and WAIT for green CI
# URL: https://github.com/coregx/signals/actions
```

**⏸️ STOP HERE! Do NOT proceed until CI is GREEN!**

✅ **All checks must pass:**
- Unit tests (Linux, macOS, Windows)
- Linting (golangci-lint)
- Code formatting (gofmt)
- Coverage check (>70%)
- Race detector (if available)

❌ **If CI fails:**
1. Fix issues in `release/v0.2.0` branch
2. Commit fixes
3. Push and wait for CI again
4. Repeat until GREEN

### Step 3: Merge to Main (After Green CI)

```bash
# ONLY after CI is green!
git checkout main
git pull origin main

# Merge release branch (--no-ff ensures merge commit)
git merge --no-ff release/v0.2.0 -m "Release v0.2.0

Complete v0.2.0 implementation:
- Comprehensive documentation and guides
- 10+ working examples
- API documentation with usage examples
- Best practices guide
- Migration guide from other reactive libraries
- Test coverage >75%

All features working:
- Signal[T] - Reactive state management
- Computed[T] - Derived state with memoization
- Effect - Side effects with cleanup
- Thread-safe, zero allocations in hot paths

Quality metrics:
- 51+ tests passing
- golangci-lint: 0 issues
- Production-ready for general use"

# Push to main
git push origin main
```

### Step 4: Wait for CI on Main

```bash
# Go to GitHub Actions and verify main branch CI
# https://github.com/coregx/signals/actions

# WAIT for green CI on main branch!
```

**⏸️ STOP! Do NOT create tag until main CI is GREEN!**

### Step 5: Create Tag (After Green CI on Main)

```bash
# ONLY after main CI is green!

# Create annotated tag
git tag -a v0.2.0 -m "Release v0.2.0

Signals Library v0.2.0 - Documentation & Examples

Features:
- Complete user documentation (guides, tutorials, API reference)
- 10+ working examples (basic, advanced, integration)
- Best practices guide and patterns
- Migration guide from other reactive libraries
- Troubleshooting guide and FAQ

Core Functionality (from v0.1.0):
- Signal[T] - Type-safe reactive state
- Computed[T] - Derived state with lazy evaluation
- Effect - Side effects with automatic cleanup
- Thread-safe with sync.RWMutex and atomic operations
- Zero allocations in hot paths
- Angular Signals compatible API

Performance:
- Signal.Get: 28.8 ns/op, 0 allocs
- Signal.Set: 52.8 ns/op, 0 allocs
- Computed.Get (cached): 20.65 ns/op, 0 allocs
- Effect.Execute: 128.8 ns/op, 0 allocs

Quality:
- 51+ unit tests passing
- Coverage >75%
- golangci-lint compliant (0 issues)
- Race detector clean
- Zero production dependencies (pure Go)

API Stability:
- Read/write API stable and production-ready
- No breaking changes from v0.1.0
- Advanced features coming in v0.3.0

See CHANGELOG.md for complete details."

# Push tag
git push origin v0.2.0
```

### Step 6: Merge Back to Develop

```bash
# Keep develop in sync
git checkout develop
git merge --no-ff release/v0.2.0 -m "Merge release v0.2.0 back to develop"
git push origin develop

# Delete release branch (optional, after confirming release is good)
git branch -d release/v0.2.0
git push origin --delete release/v0.2.0
```

### Step 7: Create GitHub Release

1. Go to: https://github.com/coregx/signals/releases/new
2. Select tag: `v0.2.0`
3. Release title: `v0.2.0 - Documentation & Examples`
4. Description: Copy from CHANGELOG.md
5. Check "Set as a pre-release" (for beta releases only)
6. Click "Publish release"

---

## 🔥 Hotfix Process

For critical bugs in production (`main` branch):

```bash
# Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b hotfix/v0.1.1

# Fix the bug
# ... make changes ...

# Test thoroughly
go test ./...
go test -race ./... # if GCC available
golangci-lint run --config .golangci.yml ./...

# Commit
git add .
git commit -m "fix: critical race condition in Signal.Update()"

# Push and wait for CI
git push origin hotfix/v0.1.1

# WAIT FOR GREEN CI!

# Merge to main
git checkout main
git merge --no-ff hotfix/v0.1.1 -m "Hotfix v0.1.1"
git push origin main

# WAIT FOR GREEN CI ON MAIN!

# Create tag
git tag -a v0.1.1 -m "Hotfix v0.1.1 - Fix critical race condition in Signal.Update()"
git push origin v0.1.1

# Merge back to develop
git checkout develop
git merge --no-ff hotfix/v0.1.1 -m "Merge hotfix v0.1.1"
git push origin develop

# Delete hotfix branch
git branch -d hotfix/v0.1.1
git push origin --delete hotfix/v0.1.1
```

---

## 📊 CI Requirements

### Must Pass Before Release

All GitHub Actions workflows must be GREEN:

1. **Unit Tests** (3 platforms)
   - Linux (ubuntu-latest)
   - macOS (macos-latest)
   - Windows (windows-latest)
   - Go versions: 1.23, 1.24, 1.25

2. **Code Quality**
   - go vet (no errors)
   - golangci-lint (34+ linters, 0 issues required)
   - gofmt (all files formatted)

3. **Coverage**
   - Overall: ≥70%
   - Core primitives: ≥90% (Signal, Computed, Effect)

4. **Race Detection**
   - go test -race ./... (no data races)

5. **Benchmarks**
   - All benchmarks run successfully
   - Zero allocations in hot paths maintained

---

## 🚫 NEVER Do This

❌ **NEVER commit directly to main**
```bash
# WRONG!
git checkout main
git commit -m "quick fix"  # ❌ NO!
```

❌ **NEVER push to main without green CI**
```bash
# WRONG!
git push origin main  # ❌ WAIT for CI first!
```

❌ **NEVER create tags before CI passes**
```bash
# WRONG!
git tag v0.2.0  # ❌ WAIT for green CI on main!
git push origin v0.2.0
```

❌ **NEVER force push to main or develop**
```bash
# WRONG!
git push -f origin main  # ❌ NEVER!
```

❌ **NEVER skip lint or format checks**
```bash
# WRONG!
git commit -m "skip CI" --no-verify  # ❌ NO!
```

❌ **NEVER push without running lint locally**
```bash
# WRONG WORKFLOW:
git commit -m "feat: something"
git push  # ❌ Run lint FIRST!

# CORRECT WORKFLOW:
golangci-lint run --config .golangci.yml ./...  # ✅ Check FIRST
go fmt ./...                                      # ✅ Format FIRST
go test ./...                                     # ✅ Test FIRST
git commit -m "feat: something"
git push
```

---

## ✅ Always Do This

✅ **ALWAYS run checks before commit**
```bash
# Recommended: Use our pre-release script
bash scripts/pre-release-check.sh

# Or manual workflow:
go fmt ./...
golangci-lint run --config .golangci.yml ./...
go test ./...
git add .
git commit -m "..."
git push
```

✅ **ALWAYS wait for green CI before proceeding**
```bash
# Correct workflow:
git push origin release/v0.2.0
# ⏸️ WAIT for green CI
git checkout main
git merge --no-ff release/v0.2.0
git push origin main
# ⏸️ WAIT for green CI on main
git tag -a v0.2.0 -m "..."
git push origin v0.2.0
```

✅ **ALWAYS use annotated tags**
```bash
# Good
git tag -a v0.2.0 -m "Release v0.2.0"

# Bad
git tag v0.2.0  # Lightweight tag
```

✅ **ALWAYS update CHANGELOG.md**
- Document all changes
- Include breaking changes
- Add known limitations
- Reference ROADMAP progress

✅ **ALWAYS test on all platforms locally if possible**
```bash
# At minimum:
go test ./...
go test -race ./... # if GCC available
golangci-lint run --config .golangci.yml ./...
go mod verify
```

✅ **ALWAYS maintain zero production dependencies**
- Pure Go implementation
- Test dependencies are OK
- No external libraries for core functionality

---

## 📝 Release Checklist Template

Copy this for each release:

```markdown
## Release v0.2.0 Checklist

### Pre-Release
- [ ] All tests passing locally (`go test ./...`)
- [ ] Race detector clean (`go test -race ./...`)
- [ ] Code formatted (`go fmt ./...`, `gofmt -l .` = empty)
- [ ] Linter clean (`golangci-lint run ./...` = 0 issues)
- [ ] Dependencies verified (`go mod verify`)
- [ ] CHANGELOG.md updated
- [ ] ROADMAP.md updated with progress
- [ ] README.md updated (if needed)
- [ ] Version bumped in relevant files
- [ ] All documentation examples tested

### Release Branch
- [ ] Created release/v0.2.0 from develop
- [ ] Pushed to GitHub
- [ ] CI GREEN on release branch
- [ ] All checks passed (tests, lint, format, coverage)

### Main Branch
- [ ] Merged release branch to main (`--no-ff`)
- [ ] Pushed to origin
- [ ] CI GREEN on main
- [ ] All checks passed

### Tagging
- [ ] Created annotated tag v0.2.0
- [ ] Tag message includes full changelog
- [ ] Pushed tag to origin
- [ ] GitHub release created

### Cleanup
- [ ] Merged back to develop
- [ ] Deleted release branch
- [ ] Verified pkg.go.dev updated
- [ ] Announced release (if applicable)
```

---

## 🎯 Summary: Golden Rules

1. **main = Production ONLY** - Always green CI, always stable
2. **Wait for CI** - NEVER proceed without green CI
3. **Tags LAST** - Only after main CI is green
4. **No Direct Commits** - Use release branches
5. **Annotated Tags** - Always use `git tag -a`
6. **Full Testing** - Run `golangci-lint` + `go test` before commit
7. **Document Everything** - Update CHANGELOG.md, README.md, ROADMAP.md
8. **Git Flow** - develop → release/* → main → tag
9. **Check Lint ALWAYS** - `golangci-lint run ./...` before every push
10. **Pure Go** - Zero production dependencies (test dependencies OK)

---

## 🔧 Signals-Specific Guidelines

### Before Release

**Angular Signals Compatibility**:
- [ ] API matches Angular Signals concepts
- [ ] Behavior matches Angular's reactive model
- [ ] Documentation references Angular when applicable

**Performance**:
- [ ] Zero allocations in hot paths (verified by benchmarks)
- [ ] Signal.Get < 30ns/op
- [ ] Signal.Set < 60ns/op
- [ ] Computed.Get (cached) < 25ns/op

**Documentation**:
- [ ] All public APIs have godoc comments
- [ ] Examples demonstrate real-world usage
- [ ] Best practices documented
- [ ] Migration guides (if breaking changes)

**Testing**:
- [ ] Test with concurrent access (race detector)
- [ ] Verify memory leaks (unsubscribe tests)
- [ ] Test with different types (generics validation)
- [ ] Benchmark all core operations

---

**Remember**: A release can always wait. A broken production release cannot be undone.

**When in doubt, wait for CI!**

**Always run lint before push!**

---

*Last Updated: 2025-10-31*
*Signals Library v0.1.0-beta Release Process*
