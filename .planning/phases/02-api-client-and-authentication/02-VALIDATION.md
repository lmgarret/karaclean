---
phase: 2
slug: api-client-and-authentication
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-18
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib), Go 1.26.1 |
| **Config file** | None needed — `go test` works out of the box |
| **Quick run command** | `go test ./internal/karakeep/... ./internal/engine/... -v -count=1` |
| **Full suite command** | `go test ./... -v -count=1` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/karakeep/... ./internal/engine/... -v -count=1`
- **After every plan wave:** Run `go test ./... -v -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 0 | CONF-03c | unit | `go test ./internal/karakeep/... -run TestCheckAuth_Success -v` | ❌ W0 | ⬜ pending |
| 02-01-02 | 01 | 0 | CONF-03d | unit | `go test ./internal/karakeep/... -run TestCheckAuth_Unauthorized -v` | ❌ W0 | ⬜ pending |
| 02-01-03 | 01 | 0 | CONF-03e | unit | `go test ./internal/karakeep/... -run TestCheckAuth_NetworkError -v` | ❌ W0 | ⬜ pending |
| 02-01-04 | 01 | 0 | CONF-03f | unit | `go test ./internal/karakeep/... -run TestListBookmarks_SinglePage -v` | ❌ W0 | ⬜ pending |
| 02-01-05 | 01 | 0 | CONF-03g | unit | `go test ./internal/karakeep/... -run TestListBookmarks_Pagination -v` | ❌ W0 | ⬜ pending |
| 02-01-06 | 01 | 0 | CONF-03h | unit | `go test ./internal/karakeep/... -run TestListBookmarks_Empty -v` | ❌ W0 | ⬜ pending |
| 02-02-01 | 02 | 0 | CONF-03i | unit | `go test ./internal/engine/... -run TestMockAPI -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/karakeep/client_test.go` — stubs for CONF-03c through CONF-03h (httptest-based)
- [ ] `internal/engine/api_test.go` — stubs for CONF-03i (mock implements interface)
- [ ] Generated code from `go generate` must exist before tests can compile

*Wave 0 creates test stubs before implementation tasks begin.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Auth check exit message clarity | CONF-03 | Requires running binary against live/mock Karakeep | Run `KARAKEEP_URL=http://... KARAKEEP_API_KEY=invalid ./karaclean`; verify exit message is human-readable |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
