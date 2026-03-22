# Phase 1: Error notification on invalid config - Context

**Gathered:** 2026-03-22
**Status:** Ready for planning

<domain>
## Phase Boundary

When config validation fails at startup, send an error notification via the configured notification channel before exiting. This behavior is toggleable via a config field. If notifications are not configured or the toggle is off, behavior is unchanged (log error, exit).

</domain>

<decisions>
## Implementation Decisions

### Config toggle placement
- **D-01:** Add `notifyOnError` as a `*bool` field under the `notifications` section (not top-level). Rationale: it's a notification behavior — if there's no `notifications` block, this field doesn't make sense. Nil means false (opt-in, consistent with notifications being opt-in).

### Partial config parsing strategy
- **D-02:** Two-pass approach in `config.Load`: first decode YAML into struct (may succeed even if validation fails), then validate. If decode succeeds but validation fails AND `notifications` is non-nil AND `notifyOnError` is true, attempt to send notification before returning error.
- **D-03:** If YAML decode itself fails (syntax error, unknown fields), attempt a lenient partial parse of just the notifications section to extract channel URLs. Use a separate minimal struct for this fallback parse.

### Error notification channel
- **D-04:** Use the `notifications.default` channel for error notifications. No per-error channel routing (unlike per-rule routing). If no default is set, skip notification silently.

### Notification format
- **D-05:** Title: `[ERROR] [karaclean] config validation failed`. Body: the validation error messages (same format as stderr output). Consistent with existing `FormatNotificationTitle` pattern (`[TAG] [karaclean] description`).

### Scope boundary
- **D-06:** This phase covers config validation errors only (startup). Runtime errors during cron runs are NOT in scope (they already log and continue).

### Claude's Discretion
- Exact implementation of the two-pass parse (whether to use a separate function or modify Load)
- Error handling if the notification send itself fails (log and continue to exit is fine)
- Whether to validate the notifications section separately before attempting to send

</decisions>

<specifics>
## Specific Ideas

- User explicitly wants this toggleable — not always-on when notifications are configured
- Current `exitf` pattern in main.go should still work — notification send happens before exit

</specifics>

<canonical_refs>
## Canonical References

No external specs — requirements are fully captured in decisions above

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `engine.Notifier` interface and `ShoutrrrNotifier`: can reuse for error notification dispatch
- `engine.ResolveChannelURL`: resolves default channel URL from notifications config
- `engine.FormatNotificationTitle`: pattern for title formatting (`[TAG] [karaclean] name`)

### Established Patterns
- Config fields use pointer types for optional values (nil = not set)
- `config.Load` returns `(*Config, error)` — validation errors wrapped in `ValidationErrors`
- `main.go` calls `exitf` on config load failure — notification must happen before this
- Notification dispatch is best-effort (log failures, don't make them fatal)

### Integration Points
- `config.Config` struct — add `NotifyOnError` to `Notifications` struct
- `config.Load` — add notification dispatch on validation failure
- `main.go:loadConfig` — may need adjustment if notification dispatch moves here
- `config.Validate` — notifications validation must still work (circular: validating the section we use to report errors)

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-error-notification-on-invalid-config*
*Context gathered: 2026-03-22*
