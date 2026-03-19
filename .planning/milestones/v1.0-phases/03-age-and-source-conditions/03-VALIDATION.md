---
phase: 3
slug: age-and-source-conditions
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-18
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` (Go 1.26) |
| **Config file** | none — Go convention, `go test` auto-discovers |
| **Quick run command** | `go test ./internal/engine/ ./internal/config/ ./internal/duration/` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~3 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/engine/ ./internal/config/ ./internal/duration/`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~3 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | COND-01 | unit | `go test ./internal/duration/ -run TestParse -v` | ❌ W0 | ⬜ pending |
| 03-01-02 | 01 | 1 | COND-01 | unit | `go test ./internal/config/ -run TestValidate -v` | ✅ (needs update) | ⬜ pending |
| 03-02-01 | 02 | 2 | COND-01, COND-02 | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | ❌ W0 | ⬜ pending |
| 03-02-02 | 02 | 2 | COND-01+02 | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/duration/duration.go` — new package for `Parse` function
- [ ] `internal/duration/duration_test.go` — covers COND-01 duration parsing (all units, edge cases, boundary)
- [ ] `internal/engine/matcher.go` — new file for `MatchesConditions`
- [ ] `internal/engine/matcher_test.go` — covers COND-01, COND-02, AND composition, boundary cases
- [ ] Update `internal/config/validate_test.go` — existing tests broken by `*int` → `*string` change
- [ ] Update `internal/config/config_test.go` — existing tests reference `intPtr(30)`
- [ ] Update `internal/config/testdata/valid_full.yaml` — `olderThan: 30` → `olderThan: "30d"`
- [ ] Update `internal/config/testdata/valid_minimal.yaml` — same

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
