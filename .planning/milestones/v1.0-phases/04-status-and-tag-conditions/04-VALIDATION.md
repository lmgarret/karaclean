---
phase: 4
slug: status-and-tag-conditions
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-18
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | N/A (Go convention) |
| **Quick run command** | `go test ./internal/engine/ ./internal/config/ -v` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/engine/ ./internal/config/ -v`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 4-01-01 | 01 | 0 | COND-03..06 | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | ❌ W0 | ⬜ pending |
| 4-01-02 | 01 | 1 | COND-03 | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | ✅ (extend) | ⬜ pending |
| 4-01-03 | 01 | 1 | COND-04 | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | ✅ (extend) | ⬜ pending |
| 4-01-04 | 01 | 1 | COND-05 | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | ✅ (extend) | ⬜ pending |
| 4-01-05 | 01 | 1 | COND-06 | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | ✅ (extend) | ⬜ pending |
| 4-01-06 | 01 | 1 | COND-03..06 | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | ✅ (extend) | ⬜ pending |
| 4-02-01 | 02 | 1 | COND-05..06 | unit | `go test ./internal/config/ -run TestValidate -v` | ✅ (extend) | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/engine/matcher_test.go` — add `boolPtr(b bool) *bool` helper (needed for archived/favourited test cases; already exists in config package but not engine package)

*All existing test infrastructure covers the remaining requirements.*

---

## Manual-Only Verifications

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
