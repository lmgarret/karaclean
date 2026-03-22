---
phase: 1
slug: list-based-bookmark-exclusion
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-22
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | none — uses `go test` |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test -race -count=1 ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test -race -count=1 ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | D-01 | unit | `go test ./internal/config/ -run TestLoad_InList` | ❌ W0 | ⬜ pending |
| 01-01-02 | 01 | 1 | D-08 | unit | `go test ./internal/config/ -run TestStringOrSlice` | ❌ W0 | ⬜ pending |
| 01-01-03 | 01 | 1 | D-11 | unit | `go test ./internal/config/ -run TestValidate_InList` | ❌ W0 | ⬜ pending |
| 01-02-01 | 02 | 1 | API | unit | `go test ./internal/karakeep/ -run TestListLists` | ❌ W0 | ⬜ pending |
| 01-02-02 | 02 | 1 | API | unit | `go test ./internal/karakeep/ -run TestGetListBookmarks` | ❌ W0 | ⬜ pending |
| 01-02-03 | 02 | 1 | D-13 | unit | `go test ./cmd/karaclean/ -run TestValidateListNames` | ❌ W0 | ⬜ pending |
| 01-03-01 | 03 | 2 | D-09 | unit | `go test ./internal/engine/ -run TestMatchesConditions_InList` | ❌ W0 | ⬜ pending |
| 01-03-02 | 03 | 2 | D-10 | unit | `go test ./internal/engine/ -run TestMatchesExceptions_InList` | ❌ W0 | ⬜ pending |
| 01-03-03 | 03 | 2 | D-04 | unit | `go test ./internal/engine/ -run TestPreloadListSets` | ❌ W0 | ⬜ pending |
| 01-03-04 | 03 | 2 | D-05 | unit | `go test ./internal/engine/ -run TestPreloadListSets_NoLists` | ❌ W0 | ⬜ pending |
| 01-03-05 | 03 | 2 | D-03 | unit | `go test ./cmd/karaclean/ -run TestValidateListNames_Missing` | ❌ W0 | ⬜ pending |
| 01-03-06 | 03 | 2 | E2E | integration | `go test ./internal/engine/ -run TestRun_InList` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/config/config_test.go` — add StringOrSlice and inList loading tests
- [ ] `internal/config/validate_test.go` — add inList validation tests
- [ ] `internal/engine/matcher_test.go` — add inList condition/exception tests (update mockAPI)
- [ ] `internal/engine/api_test.go` — update mockAPI with new interface methods
- [ ] `internal/engine/run_test.go` — add preloadListSets and Run with inList tests
- [ ] `internal/config/testdata/valid_inlist_string.yaml` — test fixture
- [ ] `internal/config/testdata/valid_inlist_list.yaml` — test fixture

---

## Manual-Only Verifications

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
