---
phase: 8
slug: scheduler-and-deployment
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-18
---

# Phase 8 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing Go test infrastructure |
| **Quick run command** | `go test ./internal/scheduler/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/scheduler/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 8-01-01 | 01 | 1 | SCHED-01 | unit | `go test ./internal/scheduler/... -run TestScheduler` | ❌ W0 | ⬜ pending |
| 8-01-02 | 01 | 1 | SCHED-02 | unit | `go test ./internal/scheduler/... -run TestTimezone` | ❌ W0 | ⬜ pending |
| 8-01-03 | 01 | 1 | SCHED-03 | unit | `go test ./internal/scheduler/... -run TestGracefulShutdown` | ❌ W0 | ⬜ pending |
| 8-02-01 | 02 | 2 | SCHED-01 | integration | `go build -o /dev/null ./cmd/karaclean` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/scheduler/scheduler_test.go` — stubs for SCHED-01, SCHED-02, SCHED-03

*Existing go test infrastructure covers the framework; only test file stubs need to be created.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Docker image builds and runs as sidecar | SCHED-01 | Requires Docker daemon | `docker build -t karaclean . && docker run --rm karaclean --help` |
| Graceful shutdown on SIGTERM in container | SCHED-03 | Requires running container | `docker run -d karaclean && docker stop <id>` — verify clean exit |
| docker-compose.yml starts alongside Karakeep | SCHED-01 | Requires full compose stack | `docker-compose up` in examples/ — verify both containers start |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
