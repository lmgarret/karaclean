---
phase: 1
slug: config-loading-and-validation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-18
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — Wave 0 creates `internal/config/config_test.go` |
| **Quick run command** | `go test ./internal/config/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/config/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 1-01-01 | 01 | 0 | CONF-01 | unit | `go test ./internal/config/... -run TestLoad` | ❌ W0 | ⬜ pending |
| 1-01-02 | 01 | 1 | CONF-01 | unit | `go test ./internal/config/... -run TestLoad` | ✅ | ⬜ pending |
| 1-01-03 | 01 | 1 | CONF-02 | unit | `go test ./internal/config/... -run TestKnownFields` | ✅ | ⬜ pending |
| 1-01-04 | 01 | 1 | CONF-01 | unit | `go test ./internal/config/... -run TestValidate` | ✅ | ⬜ pending |
| 1-02-01 | 02 | 2 | CONF-01, CONF-02 | integration | `go test ./...` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/config/config_test.go` — test stubs for CONF-01 and CONF-02 (load, known fields, validate)
- [ ] `go.mod` and `go.sum` — module initialized with `go.yaml.in/yaml/v3` dependency

*Note: No existing test infrastructure — Wave 0 bootstraps go module and test file stubs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Error message clarity for unknown fields | CONF-02 | Message quality is subjective | Run `go run ./cmd/karaclean/ --config testdata/unknown_field.yaml` and verify the error is human-readable |
| Error message clarity for semantic errors | CONF-01 | Message quality is subjective | Run `go run ./cmd/karaclean/ --config testdata/invalid_enum.yaml` and verify all errors are reported at once |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
