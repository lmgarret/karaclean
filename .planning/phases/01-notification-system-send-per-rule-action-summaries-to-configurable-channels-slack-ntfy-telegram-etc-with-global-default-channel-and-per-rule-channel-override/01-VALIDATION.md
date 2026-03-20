---
phase: 1
slug: notification-system-send-per-rule-action-summaries-to-configurable-channels-slack-ntfy-telegram-etc-with-global-default-channel-and-per-rule-channel-override
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-20
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | None needed (go test) |
| **Quick run command** | `go test ./internal/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 1-01-01 | 01 | 0 | NOTIF-CFG | unit | `go test ./internal/config/ -run TestNotif` | ❌ W0 | ⬜ pending |
| 1-01-02 | 01 | 0 | NOTIF-VAL | unit | `go test ./internal/config/ -run TestValidateNotif` | ❌ W0 | ⬜ pending |
| 1-01-03 | 01 | 0 | NOTIF-FMT | unit | `go test ./internal/engine/ -run TestFormatNotif` | ❌ W0 | ⬜ pending |
| 1-01-04 | 01 | 0 | NOTIF-SEND | unit | `go test ./internal/engine/ -run TestSendNotif` | ❌ W0 | ⬜ pending |
| 1-02-01 | 02 | 1 | NOTIF-RUN | unit | `go test ./internal/engine/ -run TestRunNotif` | ❌ W0 | ⬜ pending |
| 1-02-02 | 02 | 1 | NOTIF-SILENT | unit | `go test ./internal/engine/ -run TestRunSilent` | ❌ W0 | ⬜ pending |
| 1-02-03 | 02 | 1 | NOTIF-FAIL | unit | `go test ./internal/engine/ -run TestNotifFail` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/engine/notify.go` — new file: RuleSummary, FormatNotification, SendNotification stubs
- [ ] `internal/engine/notify_test.go` — test stubs for NOTIF-FMT, NOTIF-SEND
- [ ] `go get github.com/nicholas-fedor/shoutrrr@v0.14.0` — add Shoutrrr dependency

*Existing infrastructure (config_test.go, validate_test.go, run_test.go) extended in place.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Slack message appears in channel | NOTIF-SEND | Requires live Slack webhook | Set `notifications.channels.test.url: slack://...`, run karaclean with active rule, verify message in Slack |
| ntfy notification delivered | NOTIF-SEND | Requires live ntfy endpoint | Set ntfy URL, run with deletable bookmarks, check ntfy subscription |
| Dry-run prefix shown | NOTIF-FMT | Visual confirmation of `[DRY-RUN]` prefix in message | Run with `dryRun: true`, observe notification content |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
