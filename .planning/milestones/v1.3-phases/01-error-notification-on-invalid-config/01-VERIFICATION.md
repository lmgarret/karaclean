---
phase: 01-error-notification-on-invalid-config
verified: 2026-03-22T18:09:05Z
status: passed
score: 4/4 must-haves verified
re_verification: false
---

# Phase 1: Error Notification on Invalid Config — Verification Report

**Phase Goal:** Send error notification to default channel when config validation fails at startup, toggleable via `notifyOnError` field, with lenient fallback for YAML syntax errors.
**Verified:** 2026-03-22T18:09:05Z
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | When `notifyOnError` is true and config validation fails, a notification is sent to the default channel before returning the error | VERIFIED | `Load()` calls `SendConfigError(cfg.Notifications, notifier[0], validationErr)` on validation failure (config.go:127); `TestLoad_NotifyOnError/invalid_config_with_notifyOnError_triggers_notification` passes |
| 2 | When YAML decode fails (syntax error / unknown field), a lenient partial parse extracts the notifications section and sends error notification if `notifyOnError` is true | VERIFIED | `lenientNotify()` function in config.go:138–152; decoder without `KnownFields(true)` decodes into `notificationsOnly`; `TestLoad_LenientFallback` passes |
| 3 | When `notifyOnError` is nil/false or notifications section is absent, no error notification is attempted | VERIFIED | `SendConfigError` early-returns on nil, `*n.NotifyOnError == false`, empty `Default`, or missing channel (config.go:159–171); 4 no-op subtests in `TestSendConfigError` pass |
| 4 | If the error notification send itself fails, the failure is logged and the original config error is still returned | VERIFIED | `log.Printf("WARNING: failed to send config error notification: %v", err)` at config.go:176; `Load()` always returns original error regardless of send result; `TestSendConfigError/send_failure_is_logged_not_propagated` passes |

**Score:** 4/4 truths verified

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/config/config.go` | `NotifyOnError` field, `ConfigErrorNotifier` interface, `SendConfigError`, two-pass `Load`, `lenientNotify`, `notificationsOnly` | VERIFIED | All elements present; `NotifyOnError *bool` at line 21, `ConfigErrorNotifier` at line 13, `SendConfigError` at line 158, `Load` variadic at line 105, `lenientNotify` at line 138, `notificationsOnly` at line 96 |
| `internal/config/config_test.go` | `TestLoad_NotifyOnError`, `TestSendConfigError`, `TestLoad_LenientFallback` | VERIFIED | All three test functions present and passing (11 subtests total) |
| `internal/config/testdata/valid_notify_on_error.yaml` | Valid config with `notifyOnError: true` | VERIFIED | File exists, contains `notifyOnError: true` under `notifications` block |
| `internal/config/testdata/invalid_with_notify_on_error.yaml` | Invalid config (missing `action`) with `notifyOnError: true` | VERIFIED | File exists, rule has no `action` field, `notifyOnError: true` present |
| `internal/config/testdata/syntax_error_with_notifications.yaml` | Config that fails strict decode (unknown field) with `notifyOnError: true` | VERIFIED | File exists, contains `unknownField:` triggering `KnownFields(true)` rejection; lenient parse succeeds |
| `cmd/karaclean/main.go` | Passes `&engine.ShoutrrrNotifier{}` to `config.Load` | VERIFIED | `config.Load(path, &engine.ShoutrrrNotifier{})` at line 69 |
| `README.md` | `notifyOnError` in config table, `### Error Notifications` subsection, lenient parse mention | VERIFIED | Table row at line 392; `### Error Notifications` at line 423; lenient parse mentioned at line 427 |
| `karaclean.example.yaml` | `notifyOnError: true` inside notifications block with comment | VERIFIED | Line 29: `notifyOnError: true  # Send a notification when config validation fails at startup` |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `internal/config/config.go` | `engine.Notifier` contract | `ConfigErrorNotifier` interface mirrors `Send(url, message, title string) error` | VERIFIED | Interface defined at config.go:13–15; `engine.ShoutrrrNotifier` satisfies it structurally (Go duck typing); `main.go` passes `&engine.ShoutrrrNotifier{}` as `ConfigErrorNotifier` to `Load` |
| `internal/config/config.go` | Default channel URL resolution | `n.Channels[n.Default]` lookup inside `SendConfigError` | VERIFIED | config.go:168–171 performs map lookup; uses channel URL directly (no import of `engine.ResolveChannelURL` needed — correct, since that would cause import cycle); wiring verified |

**Note on `ResolveChannelURL`:** The PLAN listed `engine.ResolveChannelURL` as a key link. In the actual implementation, `SendConfigError` resolves the channel URL directly (`n.Channels[n.Default]`) rather than calling `engine.ResolveChannelURL`, which is the correct approach — it avoids the import cycle (`config -> engine -> config`). The key link's intent (resolving default channel URL) is fully satisfied.

---

### Requirements Coverage

No separate `REQUIREMENTS.md` exists in this project. Requirements are tracked by ID in `ROADMAP.md`. Individual requirement descriptions are encoded in the PLAN acceptance criteria.

| Requirement | Source Plan | Description (from PLAN acceptance criteria) | Status | Evidence |
|-------------|-------------|----------------------------------------------|--------|----------|
| ERRNOTIF-01 | 01-01-PLAN.md | `NotifyOnError *bool` field on `Notifications` struct with `yaml:"notifyOnError"` | SATISFIED | `NotifyOnError *bool \`yaml:"notifyOnError"\`` at config.go:21 |
| ERRNOTIF-02 | 01-01-PLAN.md | `SendConfigError` dispatches to default channel on validation failure when `notifyOnError` is true | SATISFIED | `SendConfigError` at config.go:158; dispatch path tested and passing |
| ERRNOTIF-03 | 01-01-PLAN.md | Lenient fallback parse for YAML decode errors extracts notifications and dispatches if possible | SATISFIED | `lenientNotify` + `notificationsOnly` at config.go:96–152; `TestLoad_LenientFallback` passes |
| ERRNOTIF-04 | 01-01-PLAN.md | Notification failure is best-effort: logged but not propagated; original error always returned | SATISFIED | `log.Printf` at config.go:176; Load always returns original `validationErr`; test verifies no panic or error propagation |

All 4 requirement IDs declared in `PLAN.md` frontmatter accounted for. No orphaned IDs.

---

### Anti-Patterns Found

Scanned files: `internal/config/config.go`, `internal/config/config_test.go`, `cmd/karaclean/main.go`, `README.md`, `karaclean.example.yaml`.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | No anti-patterns found | — | — |

No TODOs, FIXMEs, placeholder implementations, empty returns, or stub handlers found in phase-modified files.

---

### Human Verification Required

None. All behaviors are verifiable programmatically via unit tests. The notification integration (sending via Shoutrrr to a real ntfy/Slack/Telegram endpoint) is covered by `engine.ShoutrrrNotifier` which was implemented and tested in a prior phase.

---

### Test Suite Results

- `go test -race -run "TestLoad_NotifyOnError|TestSendConfigError|TestLoad_LenientFallback" ./internal/config/` — **PASS** (11 subtests)
- `go test -race ./...` — **PASS** (all packages)
- `golangci-lint run ./...` — **0 issues**

---

### Summary

Phase 1 goal is fully achieved. The `notifyOnError` feature is implemented end-to-end:

1. The `NotifyOnError *bool` field is on `Notifications` with opt-in semantics (nil = false).
2. `Load()` uses a two-pass approach: YAML decode then validate; on validation failure it calls `SendConfigError` if a notifier is provided.
3. `lenientNotify()` handles the YAML decode failure path by re-parsing without `KnownFields(true)` to extract just the notifications section.
4. `SendConfigError` is a no-op for all disabled states (nil notifications, false/nil flag, no default, missing channel) and best-effort on send failure.
5. `main.go` wires `&engine.ShoutrrrNotifier{}` into `config.Load` for production use.
6. All acceptance criteria from both tasks are met. Documentation is complete.

---

_Verified: 2026-03-22T18:09:05Z_
_Verifier: Claude (gsd-verifier)_
