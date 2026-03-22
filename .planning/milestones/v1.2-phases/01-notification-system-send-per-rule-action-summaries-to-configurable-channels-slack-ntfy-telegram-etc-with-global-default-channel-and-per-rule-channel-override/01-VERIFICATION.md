---
phase: 01-notification-system-send-per-rule-action-summaries-to-configurable-channels-slack-ntfy-telegram-etc-with-global-default-channel-and-per-rule-channel-override
verified: 2026-03-20T12:00:00Z
status: human_needed
score: 16/16 must-haves verified (automated); 1 item requires human
re_verification: false
human_verification:
  - test: "Send a real notification via ShoutrrrNotifier to a live ntfy or Slack endpoint"
    expected: "A notification message appears in the configured channel with the correct [karaclean] <rule-name> header and accurate counts"
    why_human: "ShoutrrrNotifier.Send requires a live external service. No unit test exists for it (by design — plan explicitly excluded network tests). The implementation compiles and follows the shoutrrr.CreateSender pattern, but actual delivery can only be confirmed end-to-end."
  - test: "Run karaclean with dryRun: true and a notifications block configured"
    expected: "Notification message in the channel starts with '[DRY-RUN] [karaclean] <rule-name>'"
    why_human: "Dry-run prefix is verified in unit tests for FormatNotification, but real delivery of the prefixed message to an external channel requires human observation."
---

# Phase 01: Notification System Verification Report

**Phase Goal:** Add per-rule notification dispatch via Shoutrrr to configurable channels (Slack, ntfy, Telegram, etc.) with global default and per-rule override, best-effort delivery, and specific message format with counts
**Verified:** 2026-03-20T12:00:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | YAML config with notifications block parses into Config.Notifications struct | VERIFIED | `config.go` lines 11-28: `Notifications *Notifications` field, YAML tag present. `TestLoad_ValidNotifications` in `config_test.go` asserts 2 channels, default, and per-rule notify field. Full test suite passes. |
| 2 | Rule with notify field parses into Rule.Notify *string | VERIFIED | `config.go` line 37: `Notify *string \`yaml:"notify"\``. `TestLoad_ValidNotifications` asserts `Rules[0].Notify == "slack-team"` and `Rules[1].Notify == nil`. |
| 3 | Invalid Shoutrrr URL rejected at config validation time | VERIFIED | `validate.go` lines 205-210: `shoutrrr.CreateSender(ch.URL)` called; error produces ValidationError. `TestValidateNotifications/channel with invalid shoutrrr URL` uses `badscheme://foo` and expects "invalid shoutrrr URL". Suite passes. |
| 4 | Rule referencing undefined channel rejected at config validation time | VERIFIED | `validate.go` lines 224-232: per-rule notify reference checked against `n.Channels`. `TestValidateNotifications/rule notify references undefined channel` asserts "rules[0].notify: references undefined channel". |
| 5 | Default referencing undefined channel rejected at config validation time | VERIFIED | `validate.go` lines 214-221. `TestValidateNotifications/default references undefined channel` covers this case. |
| 6 | Config without notifications block loads without error (opt-in) | VERIFIED | `validate.go` lines 183-194: `if n == nil { ... }` returns only errors if rules have notify set. `TestLoad_NoNotifications` loads `valid_full.yaml` (no notifications block) and asserts `cfg.Notifications == nil` with no error. |
| 7 | Rule.Notify referencing a channel when Notifications is nil is rejected | VERIFIED | `validate.go` lines 183-194: orphan notify check. `TestValidateNotifications/rule notify set but notifications nil` asserts "rules[0].notify: references channel". |
| 8 | FormatNotification produces the exact message format (conditional lines) | VERIFIED | `notify.go` lines 66-92: conditional line rendering. `TestFormatNotification` has 7 table-driven cases covering dry-run prefix, size-conditional deleted line, archived/errors/deleted omission. Suite passes. |
| 9 | Dry-run messages have [DRY-RUN] prefix | VERIFIED | `notify.go` lines 69-71. `TestFormatNotification/dry run prefix` asserts exact string match. |
| 10 | HasActivity returns true only when deleted>0 OR archived>0 OR errors>0; excepted-only is false | VERIFIED | `notify.go` lines 25-27. `TestHasActivity` has 5 cases covering deleted, archived, errors, excepted-only (false), and all-zeros (false). |
| 11 | Notifier interface enables testable notification dispatch | VERIFIED | `notify.go` lines 32-34: `type Notifier interface { Send(url, message, title string) error }`. `mockNotifier` in `run_test.go` implements it and is used across 8 notification integration tests. |
| 12 | ShoutrrrNotifier sends via CreateSender with title param | VERIFIED (partial — compile-time only) | `notify.go` lines 40-54: `shoutrrr.CreateSender(url)` called, `params.SetTitle(title)`, `sender.Send(message, &params)` called. Implementation correct but not network-tested. See human verification. |
| 13 | ResolveChannelURL returns correct URL for rule override, default fallback, and empty cases | VERIFIED | `notify.go` lines 106-122. `TestResolveChannelURL` has 4 cases: override, default fallback, empty default, nil notifications. Suite passes. |
| 14 | Run() accumulates per-rule summaries and dispatches notifications after all bookmarks | VERIFIED | `run.go` lines 52-114: `ruleSummaries` initialized, incremented in bookmark loop (Excepted/Errors/Deleted/Archived/TotalBytes), dispatched after loop. `TestRunNotification_ActiveRule` confirms 1 call with correct URL and message. |
| 15 | Rules with no activity and rules with no channel configured are silent | VERIFIED | `run.go` lines 100-106: `HasActivity()` guard and `channelURL == ""` guard. `TestRunNotification_Silent_NoActivity` and `TestRunNotification_Silent_NoChannel` confirm 0 calls. |
| 16 | Notification delivery failure is logged but does not abort run or return error | VERIFIED | `run.go` lines 110-112: `log.Printf("notification failed for rule %q: %v", ...)` — no error propagation. `TestRunNotification_FailureNonFatal` passes `mockNotifier{err: errors.New("send failed")}` and asserts `Run()` returns nil error and correct summary. |

**Score:** 16/16 truths verified (15 fully automated, 1 verified at compile-time with human delivery check pending)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/config/config.go` | Notifications, NotificationChannel structs; Notify field on Rule | VERIFIED | Lines 11-38: both types defined with yaml tags, fields present on Config and Rule |
| `internal/config/validate.go` | validateNotifications function | VERIFIED | Lines 181-236: full implementation with shoutrrr.CreateSender, channel ref checks, orphan notify check |
| `internal/config/config_test.go` | Config parsing tests for notification fields | VERIFIED | Lines 321-369: `TestLoad_ValidNotifications` and `TestLoad_NoNotifications` — substantive, both pass |
| `internal/config/validate_test.go` | Validation tests for notification rules | VERIFIED | Lines 415-530: `TestValidateNotifications` with 8 table-driven test cases — substantive, all pass |
| `internal/config/testdata/valid_notifications.yaml` | Test fixture with notifications block | VERIFIED | File exists: 2 channels (my-ntfy, slack-team), default set, rules[0] has notify |
| `internal/engine/notify.go` | RuleSummary, Notifier interface, ShoutrrrNotifier, FormatNotification, ResolveChannelURL | VERIFIED | 122-line file: all 5 exports present and substantive |
| `internal/engine/notify_test.go` | Tests for formatting, HasActivity, channel resolution | VERIFIED | 234-line file: `TestFormatNotification` (7 cases), `TestHasActivity` (5 cases), `TestResolveChannelURL` (4 cases), `TestFormatNotificationTitle` (2 cases) |
| `internal/engine/run.go` | Updated Run() with per-rule summary accumulation and notification dispatch | VERIFIED | Lines 41-117: new signature, `ruleSummaries` init and accumulation, notification dispatch loop |
| `internal/engine/run_test.go` | Tests for notification dispatch integration | VERIFIED | Lines 283-509: `mockNotifier`, `testNotifications`, 8 `TestRunNotification_*` tests |
| `cmd/karaclean/main.go` | Updated callers passing ShoutrrrNotifier and cfg.Notifications | VERIFIED | Lines 73-77: notifier creation; lines 92, 107: both Run() call sites updated |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `internal/config/validate.go` | `shoutrrr.CreateSender` | URL validation at startup | WIRED | Line 9: `"github.com/nicholas-fedor/shoutrrr"` imported; line 205: `shoutrrr.CreateSender(ch.URL)` called in `validateNotifications` |
| `internal/config/config.go` | `internal/config/validate.go` | `Validate()` calls `validateNotifications` | WIRED | `validate.go` line 49: `errs = append(errs, validateNotifications(c.Notifications, c.Rules)...)` |
| `internal/engine/notify.go` | `shoutrrr.CreateSender` | `ShoutrrrNotifier.Send` | WIRED | Line 41: `sender, err := shoutrrr.CreateSender(url)` in `Send()` method |
| `internal/engine/notify.go` | `internal/engine/actions.go` | `HumanSize` call in `FormatNotification` | WIRED | Line 76: `HumanSize(rs.TotalBytes)` — same package, `HumanSize` exported from `actions.go` |
| `internal/engine/run.go` | `internal/engine/notify.go` | `FormatNotification`, `ResolveChannelURL`, `notifier.Send` | WIRED | Lines 103: `ResolveChannelURL`; 108: `FormatNotification`; 109: `FormatNotificationTitle`; 110: `notifier.Send` — all called in dispatch loop |
| `internal/engine/run.go` | `internal/config/config.go` | `*config.Notifications` parameter | WIRED | Line 41: `notifications *config.Notifications` in Run() signature |
| `cmd/karaclean/main.go` | `internal/engine/run.go` | Calls Run() with ShoutrrrNotifier and cfg.Notifications | WIRED | Lines 92, 107: both `engine.Run(ctx, client, cfg.Rules, dryRun, cfg.Notifications, notifier)` calls present |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| NOTIF-CFG | 01-01 | Config parsing with notifications block | SATISFIED | `Notifications` and `NotificationChannel` types in config.go; `Notify *string` on Rule; `TestLoad_ValidNotifications` passes |
| NOTIF-VAL | 01-01 | Validate channel refs and Shoutrrr URLs | SATISFIED | `validateNotifications` in validate.go with Shoutrrr URL validation and channel ref checks; 8 test cases pass |
| NOTIF-FMT | 01-02 | FormatNotification output correctness | SATISFIED | `FormatNotification` in notify.go with conditional lines; 7 test cases covering exact output format pass |
| NOTIF-SEND | 01-02 | SendNotification calls Shoutrrr correctly | PARTIALLY SATISFIED | `ShoutrrrNotifier.Send` implemented with `shoutrrr.CreateSender` + title params; compiles and follows correct API. No `TestSendNotif` unit test (requires network — documented exclusion). Live delivery requires human verification. |
| NOTIF-RUN | 01-03 | Run() accumulates per-rule summaries and dispatches | SATISFIED | `ruleSummaries` accumulation in bookmark loop; dispatch loop after bookmarks; 8 integration tests pass |
| NOTIF-SILENT | 01-03 | No notification when no activity or no channel | SATISFIED | `HasActivity()` guard and empty `channelURL` guard in Run(); `TestRunNotification_Silent_NoActivity` and `TestRunNotification_Silent_NoChannel` pass |
| NOTIF-FAIL | 01-03 | Notification failure is non-fatal | SATISFIED | `log.Printf` on Send error, no error propagation; `TestRunNotification_FailureNonFatal` passes |

**Note on NOTIF-SEND:** The plan explicitly excluded network tests for `ShoutrrrNotifier.Send`. The VALIDATION.md pre-documents this as a live-endpoint human test. The implementation is correct at the code level (uses `shoutrrr.CreateSender`, sets title params, iterates errors). This does not constitute a gap — it is an acknowledged human verification item.

### Anti-Patterns Found

None. Scanned all 5 modified/created source files for TODO, FIXME, placeholder comments, empty implementations, and console-log-only stubs. Zero findings.

### Human Verification Required

#### 1. Live Notification Delivery

**Test:** Configure a `notifications:` block in karaclean.yaml pointing to a real ntfy topic URL (e.g., `ntfy://ntfy.sh/my-test-topic`). Run karaclean against a Karakeep instance with at least one bookmark matching a rule. Subscribe to the ntfy topic before running.

**Expected:** A message appears in the ntfy topic with the format:
```
[karaclean] <rule-name>
deleted: N
archived: N  (only if > 0)
```

**Why human:** `ShoutrrrNotifier.Send` requires a live external service. Unit tests use `mockNotifier`. The implementation compiles and the shoutrrr API is called correctly, but actual delivery cannot be confirmed programmatically.

#### 2. Dry-Run Prefix in Delivered Message

**Test:** Same setup as above but with `dryRun: true` in the config (or `--dry-run` flag). Observe the notification received in the channel.

**Expected:** Message header reads `[DRY-RUN] [karaclean] <rule-name>`.

**Why human:** The `[DRY-RUN]` prefix is verified in `TestFormatNotification` and `TestRunNotification_DryRun` unit tests, but confirming it survives round-trip through Shoutrrr to an actual service requires human observation.

### Gaps Summary

No gaps. All 16 observable truths verified. All 7 requirement IDs covered. All 10 artifacts are substantive and wired. All key links confirmed in source. Full test suite (`go test ./...`) passes cleanly with 0 failures. Binary builds successfully (`go build ./cmd/karaclean/`).

The two human verification items relate to NOTIF-SEND live delivery, which was pre-acknowledged in the VALIDATION.md as requiring a live endpoint. This is not a gap in implementation — it is a verification limitation.

---

_Verified: 2026-03-20T12:00:00Z_
_Verifier: Claude (gsd-verifier)_
