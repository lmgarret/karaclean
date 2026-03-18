---
phase: 7
slug: run-orchestrator-and-observability
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-18
---

# Phase 7 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) |
| **Config file** | none — Go convention |
| **Quick run command** | `go test ./internal/engine/ -run TestRun -v` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/engine/ -run TestRun -v`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 2 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 7-01-01 | 01 | 0 | OBS-01 | unit | `go test ./internal/engine/ -run TestRun -v` | ❌ W0 | ⬜ pending |
| 7-01-02 | 01 | 1 | OBS-01 | unit | `go test ./internal/engine/ -run TestRun -v` | ❌ W0 | ⬜ pending |
| 7-02-01 | 02 | 2 | OBS-01 | build | `go build ./cmd/karaclean/` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/engine/run_test.go` — table-driven tests for Run() covering all OBS-01 behaviors
  - no bookmarks → zero summary
  - no rules → all NoMatch
  - first-match-wins (stops after first matching rule)
  - excepted counter (conditions match but unless protects)
  - errors counter (per-bookmark failure: log-and-continue)
  - dry-run passthrough to ExecuteAction
  - ListBookmarks failure → Run() returns error

*mockAPI already exists in `internal/engine/api_test.go` (same `engine_test` package — directly reusable).*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Summary printed to log output at end of run | OBS-01 | Requires live invocation to observe log output | Run `go run ./cmd/karaclean/ --dry-run` with a populated config; verify log line contains `archived=` `deleted=` `no_match=` `excepted=` `errors=` |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 2s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
