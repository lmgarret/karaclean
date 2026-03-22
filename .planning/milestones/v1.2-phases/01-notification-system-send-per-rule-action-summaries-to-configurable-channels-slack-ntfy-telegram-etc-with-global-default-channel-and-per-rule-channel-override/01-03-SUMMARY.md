---
phase: 01-notification-system-send-per-rule-action-summaries-to-configurable-channels-slack-ntfy-telegram-etc-with-global-default-channel-and-per-rule-channel-override
plan: 03
subsystem: engine
tags: [notifications, run-integration, shoutrrr, per-rule-dispatch]

requires:
  - phase: 01-01
    provides: "Config types (Notifications, NotificationChannel), Rule.Notify field"
  - phase: 01-02
    provides: "RuleSummary, FormatNotification, Notifier interface, ResolveChannelURL"
provides:
  - "Run() with per-rule summary accumulation and notification dispatch"
  - "Backward-compatible nil notifier/notifications handling"
  - "main.go wiring ShoutrrrNotifier and cfg.Notifications to Run()"
affects: []

tech-stack:
  added: []
  patterns: [per-rule-summary-accumulation-in-bookmark-loop, notification-dispatch-after-evaluation]

key-files:
  created: []
  modified: [internal/engine/run.go, internal/engine/run_test.go, cmd/karaclean/main.go]

key-decisions:
  - "main.go creates notifier only when cfg.Notifications is non-nil (nil = no Shoutrrr overhead)"
  - "Run() signature extended with trailing params for backward compatibility with nil, nil"

patterns-established:
  - "Per-rule RuleSummary accumulation keyed by rule index in bookmark loop"
  - "Non-fatal notification dispatch: log.Printf on failure, no error propagation"

requirements-completed: [NOTIF-RUN, NOTIF-SILENT, NOTIF-FAIL]

duration: 2min
completed: 2026-03-20
---

# Phase 01 Plan 03: Run Integration Summary

**Per-rule notification dispatch wired into Run() with summary accumulation, channel resolution, and non-fatal error handling**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-20T11:25:17Z
- **Completed:** 2026-03-20T11:27:35Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Run() accumulates per-rule RuleSummary counters during bookmark evaluation
- Notification dispatch after bookmark loop: HasActivity check, channel resolution, formatted message with title
- Backward compatible: nil notifier and nil notifications handled gracefully (no panics)
- Notification failure logged but non-fatal (Run() still returns successful RunSummary)
- main.go creates ShoutrrrNotifier only when notifications configured
- 8 new notification integration tests covering all dispatch scenarios

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: Failing notification tests** - `cf3dbe1` (test)
2. **Task 1 GREEN + Task 2: Run() implementation and main.go wiring** - `5df19e5` (feat)

_Note: main.go changes (Task 2) were required for compilation during Task 1 GREEN phase, so both were committed together._

## Files Created/Modified
- `internal/engine/run.go` - Run() signature extended, per-rule RuleSummary accumulation, notification dispatch loop
- `internal/engine/run_test.go` - mockNotifier, testNotifications helper, 8 new TestRunNotification_* tests, existing tests updated with nil params
- `cmd/karaclean/main.go` - ShoutrrrNotifier creation, both Run() call sites updated with notification params

## Decisions Made
- main.go creates notifier only when cfg.Notifications is non-nil -- avoids Shoutrrr overhead when notifications not configured
- Run() signature uses trailing params (notifications, notifier) for backward compatibility with nil, nil in existing tests
- Per-rule RuleSummary uses index-based access (ruleIdx) matching rule loop iteration order

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Notification system is fully wired end-to-end: config parsing, validation, formatting, dispatch, and Run() integration
- Phase 01 (notification-system) is complete
- Users can add `notifications:` section to config YAML to enable per-rule notifications

---
*Phase: 01-notification-system*
*Completed: 2026-03-20*
