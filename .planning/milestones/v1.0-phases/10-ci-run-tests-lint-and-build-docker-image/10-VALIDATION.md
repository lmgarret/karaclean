---
phase: 10
slug: ci-run-tests-lint-and-build-docker-image
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-19
---

# Phase 10 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing Go test files |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test -race ./...` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test -race ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 10-01-01 | 01 | 1 | golangci-lint config | lint | `golangci-lint run` | ✅ / ❌ W0 | ⬜ pending |
| 10-01-02 | 01 | 1 | CI workflow file | integration | `cat .github/workflows/ci.yml` | ❌ W0 | ⬜ pending |
| 10-01-03 | 01 | 1 | tests pass in CI | unit | `go test -race ./...` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `.github/workflows/ci.yml` — CI workflow (new file)
- [ ] `.golangci.yml` — linter config (new file)
- [ ] Lint violations fixed in existing code before CI enforces them

*Note: No test stubs needed — existing tests cover Go source. Wave 0 is creating new CI config files.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| CI runs on PR | GitHub Actions trigger | Requires GitHub push/PR | Create a test PR and verify CI jobs appear |
| ghcr.io push on main | Docker registry push | Requires GitHub push to main | Merge to main, check ghcr.io packages |
| Concurrency cancel | In-progress run cancelled | Requires two rapid pushes | Push twice quickly, verify first run cancelled |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
