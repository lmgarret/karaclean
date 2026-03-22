# Phase 1: Notification System - Context

**Gathered:** 2026-03-20
**Status:** Ready for planning

<domain>
## Phase Boundary

After each run, send per-rule action summaries to configurable notification channels.
A global default channel applies to all rules; individual rules can override with a
specific channel. Notification delivery is best-effort and non-fatal to the run.

</domain>

<decisions>
## Implementation Decisions

### Config schema
- Named channels block at top level: `notifications.channels.<name>.url`
- Each channel has a single `url:` field in Shoutrrr URL format
- Global default referenced by name: `notifications.default: <channel-name>`
- Per-rule override: `notify: <channel-name>` string field on `Rule` (one channel only)
- No channel + no default = silent (notifications are opt-in, no error, no warning)
- No multi-channel per rule at launch

Example config shape:
```yaml
notifications:
  channels:
    my-ntfy:
      url: ntfy://ntfy.sh/karaclean-alerts
    slack-team:
      url: slack://TOKEN-A/TOKEN-B/TOKEN-C
  default: my-ntfy

rules:
  - name: old-rss
    notify: slack-team   # override
    ...
  - name: web-junk
    # uses default: my-ntfy
    ...
  - name: no-notify-rule
    # no notify + no default → silent
    ...
```

### Notification content
- Per-rule message format (counts only, no bookmark list):
  ```
  [karaclean] <rule-name>
  deleted: N (X.X MB)   ← size only shown when TotalBytes > 0
  archived: N
  excepted: N
  errors: N             ← only shown when errors > 0
  ```
- Dry-run prefix: `[DRY-RUN] [karaclean] <rule-name>` (dry-run notifications are sent)
- Only notify when something happened: notify if deleted > 0 OR archived > 0 OR errors > 0
- Rules with zero actions (all no_match or all excepted-only) do not send notifications

### Trigger conditions
- Errors in a rule's actions (API failures) are included in the notification (errors: N)
- Notification delivery failure is non-fatal: log the error, continue — notification is
  observability, not critical path
- No retry on notification failure at launch

### Notification library
- Use Shoutrrr (github.com/nicholas-fedor/shoutrrr) — Go library, 20+ providers via URL
- Supported services include: ntfy, Slack, Telegram, Discord, Gotify, Matrix, Pushover,
  Teams, Google Chat, Rocketchat, Mattermost, Pushbullet, Signal, and more
- Single `url:` field per channel covers all providers — no per-provider config structs needed

### Claude's Discretion
- Per-rule summary accumulation mechanism (how Run() tracks per-rule counts)
- Message title field usage (Shoutrrr supports optional title for services that accept it)
- Exact Shoutrrr API usage (Send vs router approach)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Existing config and engine
- `internal/config/config.go` — Current Config and Rule structs; Notify field must be added to Rule
- `internal/engine/run.go` — Run() orchestrator; per-rule summary tracking must be added
- `internal/engine/actions.go` — ActionResult, HumanSize(); per-rule summaries use these

### Notification library
- Shoutrrr GitHub: https://github.com/nicholas-fedor/shoutrrr
- Supported services overview: all services use URL format, see library docs for each provider

No internal specs — requirements fully captured in decisions above.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `HumanSize(bytes int64) string` in `internal/engine/actions.go` — already formats bytes as human-readable; reuse for the `deleted: N (X MB)` line
- `ActionResult.Size int64` — per-bookmark size already tracked; accumulate into per-rule TotalBytes
- `RunSummary` in `internal/engine/run.go` — global summary pattern; per-rule summary follows same shape

### Established Patterns
- `*bool` pointer types for optional config fields (distinguish nil from false) — `notify` is a `*string` on Rule to distinguish "no channel" from empty string
- `log.Printf` for all logging — notification errors should follow same pattern
- `KnownFields(true)` YAML parsing — new `notifications:` block and `notify:` field must be added to config structs to avoid unknown-field errors at startup

### Integration Points
- `Run()` in `engine/run.go` — currently returns a single `RunSummary`; needs to accumulate a per-rule summary map and dispatch notifications after each rule completes its bookmarks
- `Config` struct in `config/config.go` — `Notifications` field added at top level; `Notify *string` added to `Rule`
- `config.Validate()` in `config/validate.go` — validate that `notify` references a defined channel name; validate Shoutrrr URLs at startup

</code_context>

<specifics>
## Specific Ideas

- Config example from discussion: `url: ntfy://ntfy.sh/karaclean-alerts` for ntfy, `url: slack://TOKEN-A/TOKEN-B/TOKEN-C` for Slack
- Message format shown during discussion:
  ```
  [karaclean] old-rss
  deleted: 12 (4.2 MB)
  archived: 3
  excepted: 1
  ```

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 01-notification-system*
*Context gathered: 2026-03-20*
