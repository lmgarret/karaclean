# Phase 1: Notification System - Research

**Researched:** 2026-03-20
**Domain:** Go notification dispatch via Shoutrrr, config schema extension, per-rule summary accumulation
**Confidence:** HIGH

## Summary

This phase adds a notification system that sends per-rule action summaries to configurable channels after each run. The user has locked all major decisions: Shoutrrr library for multi-provider notifications via URL format, named channels in config, global default + per-rule override, best-effort delivery, and a specific message format.

The core technical work involves: (1) extending Config/Rule structs with a `Notifications` block and `Notify` field, (2) accumulating per-rule summaries during Run(), (3) formatting and dispatching notifications via Shoutrrr after each rule completes, and (4) validating channel references and Shoutrrr URLs at config load time.

**Primary recommendation:** Use `shoutrrr.Send(url, message)` for each notification (one URL per channel, no need for ServiceRouter). Accumulate per-rule summaries in a `map[string]*RuleSummary` inside Run(), dispatch after the bookmark loop completes for each rule.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Named channels block at top level: `notifications.channels.<name>.url`
- Each channel has a single `url:` field in Shoutrrr URL format
- Global default referenced by name: `notifications.default: <channel-name>`
- Per-rule override: `notify: <channel-name>` string field on Rule (one channel only)
- No channel + no default = silent (notifications are opt-in, no error, no warning)
- No multi-channel per rule at launch
- Per-rule message format with counts only (no bookmark list)
- `[karaclean] <rule-name>` title, size shown only when TotalBytes > 0, errors shown only when > 0
- Dry-run prefix: `[DRY-RUN] [karaclean] <rule-name>`
- Only notify when something happened: deleted > 0 OR archived > 0 OR errors > 0
- Rules with zero actions do not send notifications
- Notification delivery failure is non-fatal: log error, continue
- No retry on notification failure at launch
- Use Shoutrrr (github.com/nicholas-fedor/shoutrrr)

### Claude's Discretion
- Per-rule summary accumulation mechanism (how Run() tracks per-rule counts)
- Message title field usage (Shoutrrr supports optional title for services that accept it)
- Exact Shoutrrr API usage (Send vs router approach)

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope.
</user_constraints>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/nicholas-fedor/shoutrrr | v0.14.0 | Multi-provider notification dispatch via URL | 25+ providers, URL-based config, actively maintained fork of containrrr/shoutrrr |

### Supporting
No additional libraries needed. Shoutrrr handles all provider-specific logic.

### Alternatives Considered
None -- Shoutrrr is a locked decision.

**Installation:**
```bash
go get github.com/nicholas-fedor/shoutrrr@v0.14.0
```

**Version verification:** v0.14.0 published Mar 10, 2026 per pkg.go.dev.

## Architecture Patterns

### Recommended Project Structure
```
internal/
  config/
    config.go          # Add Notifications struct, Notify *string to Rule
    validate.go        # Add channel reference + Shoutrrr URL validation
    config_test.go     # Config parsing tests for new fields
    validate_test.go   # Validation tests for new rules
  engine/
    run.go             # Add per-rule summary accumulation + notification dispatch
    notify.go          # NEW: RuleSummary type, FormatNotification(), SendNotification()
    notify_test.go     # NEW: Tests for formatting and send logic
    run_test.go        # Extended with notification dispatch tests
```

### Pattern 1: Per-Rule Summary Accumulation

**What:** Track per-rule action counts during Run() using a map keyed by rule index/name.
**When to use:** During the bookmark evaluation loop in Run().

**Recommended approach:** Use a `RuleSummary` struct (similar to `RunSummary` but per-rule) and accumulate counts as bookmarks are processed. After the bookmark loop, iterate rules and dispatch notifications for rules that had activity.

```go
// RuleSummary records per-rule action counts for notification.
type RuleSummary struct {
    RuleName   string
    Archived   int
    Deleted    int
    Excepted   int
    Errors     int
    TotalBytes int64
}

// HasActivity returns true if this rule should trigger a notification.
func (s RuleSummary) HasActivity() bool {
    return s.Deleted > 0 || s.Archived > 0 || s.Errors > 0
}
```

The accumulation happens inside Run() by maintaining a `map[int]*RuleSummary` (keyed by rule index in the rules slice). Each bookmark's outcome increments the appropriate rule's summary.

### Pattern 2: Notification Dispatch (after bookmark loop)

**What:** After all bookmarks are processed, iterate rule summaries and send notifications.
**When to use:** At the end of Run(), before returning RunSummary.

```go
// After the bookmark loop in Run():
for i, rule := range rules {
    rs := ruleSummaries[i]
    if rs == nil || !rs.HasActivity() {
        continue
    }
    channelURL := resolveChannel(cfg.Notifications, rule.Notify)
    if channelURL == "" {
        continue  // no channel configured, silent
    }
    msg := FormatNotification(rs, dryRun)
    if err := SendNotification(channelURL, msg, rs.RuleName, dryRun); err != nil {
        log.Printf("notification failed for rule %q: %v", rule.Name, err)
    }
}
```

### Pattern 3: Shoutrrr Direct Send (recommended over ServiceRouter)

**What:** Use the package-level `shoutrrr.Send()` function per notification.
**Why:** Each notification goes to exactly one URL. No need for a multi-URL router. Simpler, no state to manage.

```go
import "github.com/nicholas-fedor/shoutrrr"

func SendNotification(url, message, title string, dryRun bool) error {
    return shoutrrr.Send(url, message)
}
```

**Note on title:** Shoutrrr v0.14.0 package-level `Send(rawURL, message)` returns `error` (not `[]error`). It does NOT accept params. To use title, you would need `CreateSender` + `ServiceRouter.Send(message, params)`. Since the message format already includes the rule name in the first line, the title is optional/nice-to-have. If title is desired, use:

```go
sender, err := shoutrrr.CreateSender(url)
if err != nil {
    return err
}
params := types.Params{}
params.SetTitle(title)
errs := sender.Send(message, &params)
// errs is []error, one per service
```

**Recommendation:** Use `CreateSender` + `ServiceRouter.Send()` with title params. The title improves UX on services that support it (Slack, ntfy, Telegram, Discord all show titles prominently). Cache senders by channel name to avoid re-creating per notification.

### Pattern 4: Channel Resolution

**What:** Resolve a rule's effective notification channel URL.
**When to use:** Before sending each notification.

```go
func ResolveChannelURL(channels map[string]NotificationChannel, ruleNotify *string, defaultChannel string) string {
    name := defaultChannel
    if ruleNotify != nil {
        name = *ruleNotify
    }
    if name == "" {
        return ""
    }
    ch, ok := channels[name]
    if !ok {
        return ""  // should not happen if validation passed
    }
    return ch.URL
}
```

### Anti-Patterns to Avoid
- **Creating a global ServiceRouter with all channel URLs:** Each rule sends to exactly one channel. A multi-URL router would send to ALL channels, which is wrong.
- **Sending notifications inside the bookmark loop:** Notifications should be sent per-rule after all bookmarks are processed, not per-bookmark.
- **Making notification failure fatal:** The context doc is explicit -- log and continue, never abort the run.
- **Accumulating summaries by rule name (string key):** Use rule index. Two rules could theoretically have the same name (though validation could prevent this).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Multi-provider notification dispatch | Custom HTTP clients per service | Shoutrrr | 25+ providers with auth, retries, URL parsing built in |
| Shoutrrr URL validation | Regex or string parsing | Shoutrrr's own URL parsing at config validation time | Library knows valid URL formats per service |
| Human-readable byte formatting | New formatter | Existing `HumanSize()` in `actions.go` | Already tested and used throughout |

**Key insight:** Shoutrrr's URL format encodes all provider-specific config (tokens, channels, endpoints). The entire notification subsystem reduces to: format message string, call `Send(url, message)`.

## Common Pitfalls

### Pitfall 1: KnownFields(true) Rejects New YAML Fields
**What goes wrong:** Adding `notifications:` to YAML without updating Config struct causes parse failure at startup.
**Why it happens:** The project uses `decoder.KnownFields(true)` which rejects any YAML key not mapped to a struct field.
**How to avoid:** Add ALL new fields to structs BEFORE testing with YAML. This includes `Notifications` on Config and `Notify` on Rule.
**Warning signs:** "unknown field" errors when loading config.

### Pitfall 2: Shoutrrr URL Validation at Startup vs. Send Time
**What goes wrong:** Invalid Shoutrrr URLs pass config validation but fail at notification time. User doesn't know until a run completes.
**Why it happens:** String-only validation (non-empty) doesn't catch malformed URLs.
**How to avoid:** Use `shoutrrr.CreateSender(url)` during config validation. If it returns an error, the URL is invalid. This gives immediate feedback at startup.
**Warning signs:** Notifications silently failing in logs.

### Pitfall 3: Nil Pointer on Rule.Notify
**What goes wrong:** Accessing `*rule.Notify` when Notify is nil causes panic.
**Why it happens:** `Notify *string` uses pointer type per project convention for optional fields.
**How to avoid:** Always nil-check before dereferencing: `if rule.Notify != nil { name = *rule.Notify }`.
**Warning signs:** Panic in Run() when rules don't have notify set.

### Pitfall 4: Run() Signature Change Breaking Existing Tests
**What goes wrong:** Run() needs access to notification config (channels map, default). Changing its signature breaks all existing callers and tests.
**Why it happens:** Current Run() takes `(ctx, api, rules, dryRun)`. Notifications need channel config too.
**How to avoid:** Pass a NotificationConfig struct (or the full Notifications config) as an additional parameter. Update all test call sites. Alternatively, pass a Notifier interface that Run() calls, with a no-op implementation for tests.
**Warning signs:** Compilation errors in run_test.go after changing Run() signature.

### Pitfall 5: Notification Sent for Excepted-Only Rules
**What goes wrong:** A rule that only has excepted bookmarks (no deletes, archives, or errors) sends a notification.
**Why it happens:** `Excepted > 0` is not part of the HasActivity() check per user decision.
**How to avoid:** HasActivity() checks only `deleted > 0 || archived > 0 || errors > 0`. Excepted-only rules are silent.
**Warning signs:** Noisy notifications for rules that didn't do anything meaningful.

## Code Examples

### Config Struct Extension
```go
// In config/config.go

type Config struct {
    Timezone      string         `yaml:"timezone"`
    Schedule      string         `yaml:"schedule"`
    DryRun        bool           `yaml:"dryRun"`
    Notifications *Notifications `yaml:"notifications"`
    Rules         []Rule         `yaml:"rules"`
}

type Rule struct {
    Name       string      `yaml:"name"`
    Conditions *Conditions `yaml:"conditions"`
    Unless     *Exceptions `yaml:"unless"`
    Action     string      `yaml:"action"`
    DryRun     *bool       `yaml:"dryRun"`
    Notify     *string     `yaml:"notify"`
}

type Notifications struct {
    Channels map[string]NotificationChannel `yaml:"channels"`
    Default  string                         `yaml:"default"`
}

type NotificationChannel struct {
    URL string `yaml:"url"`
}
```

### Message Formatting
```go
// In engine/notify.go

func FormatNotification(rs *RuleSummary, dryRun bool) string {
    var b strings.Builder

    if dryRun {
        fmt.Fprintf(&b, "[DRY-RUN] ")
    }
    fmt.Fprintf(&b, "[karaclean] %s\n", rs.RuleName)

    if rs.TotalBytes > 0 {
        fmt.Fprintf(&b, "deleted: %d (%s)\n", rs.Deleted, HumanSize(rs.TotalBytes))
    } else if rs.Deleted > 0 {
        fmt.Fprintf(&b, "deleted: %d\n", rs.Deleted)
    }
    if rs.Archived > 0 {
        fmt.Fprintf(&b, "archived: %d\n", rs.Archived)
    }
    if rs.Excepted > 0 {
        fmt.Fprintf(&b, "excepted: %d\n", rs.Excepted)
    }
    if rs.Errors > 0 {
        fmt.Fprintf(&b, "errors: %d\n", rs.Errors)
    }

    return strings.TrimRight(b.String(), "\n")
}
```

### Validation: Channel References
```go
// In config/validate.go

func validateNotifications(n *Notifications, rules []Rule) []ValidationError {
    var errs []ValidationError
    if n == nil {
        return nil  // no notifications configured, that's fine
    }

    // Validate each channel URL via Shoutrrr
    for name, ch := range n.Channels {
        if ch.URL == "" {
            errs = append(errs, ValidationError{
                Field:   fmt.Sprintf("notifications.channels.%s.url", name),
                Message: "url is required",
            })
            continue
        }
        if _, err := shoutrrr.CreateSender(ch.URL); err != nil {
            errs = append(errs, ValidationError{
                Field:   fmt.Sprintf("notifications.channels.%s.url", name),
                Message: fmt.Sprintf("invalid shoutrrr URL: %v", err),
            })
        }
    }

    // Validate default references a defined channel
    if n.Default != "" {
        if _, ok := n.Channels[n.Default]; !ok {
            errs = append(errs, ValidationError{
                Field:   "notifications.default",
                Message: fmt.Sprintf("references undefined channel %q", n.Default),
            })
        }
    }

    // Validate per-rule notify references
    for i, rule := range rules {
        if rule.Notify != nil {
            if _, ok := n.Channels[*rule.Notify]; !ok {
                errs = append(errs, ValidationError{
                    Field:   fmt.Sprintf("rules[%d].notify", i),
                    Message: fmt.Sprintf("references undefined channel %q", *rule.Notify),
                })
            }
        }
    }

    return errs
}
```

### Shoutrrr Send with Title
```go
// In engine/notify.go
import (
    "github.com/nicholas-fedor/shoutrrr"
    "github.com/nicholas-fedor/shoutrrr/pkg/types"
)

func SendNotification(url, message, title string) error {
    sender, err := shoutrrr.CreateSender(url)
    if err != nil {
        return fmt.Errorf("creating sender: %w", err)
    }
    params := types.Params{}
    params.SetTitle(title)
    errs := sender.Send(message, &params)
    for _, e := range errs {
        if e != nil {
            return e  // return first error
        }
    }
    return nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| containrrr/shoutrrr | nicholas-fedor/shoutrrr | 2024 | Active maintenance, v0.14.0, more services |
| `Send()` returns `[]error` | `Send()` returns `error` (pkg-level) | v0.14.0 | Simpler error handling at package level |
| ServiceRouter.Send with `map[string]string` | ServiceRouter.Send with `*types.Params` | v0.14.0 | Use `types.Params` type, not raw map |

**Deprecated/outdated:**
- `containrrr/shoutrrr`: Original, now unmaintained. Use `nicholas-fedor/shoutrrr`.
- Package-level `Send()` no longer accepts params. Use `CreateSender` + `ServiceRouter.Send` for title support.

## Open Questions

1. **Shoutrrr URL validation side effects at startup**
   - What we know: `CreateSender(url)` parses and validates the URL format
   - What's unclear: Does it attempt network connections during creation? If so, config validation could fail when network is unavailable at startup.
   - Recommendation: Test this empirically. If it does connect, fall back to regex-based URL scheme validation (check for known service prefixes like `slack://`, `ntfy://`, `telegram://`).

2. **Notifier interface vs. direct Shoutrrr calls in Run()**
   - What we know: Run() tests use mockAPI. Adding direct Shoutrrr calls makes Run() harder to test.
   - What's unclear: Best way to make notification testable.
   - Recommendation: Define a `Notifier` interface with `Send(url, message, title string) error`. Production uses Shoutrrr implementation. Tests use a mock. Pass Notifier to Run() or a wrapping function.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | None needed (go test) |
| Quick run command | `go test ./internal/...` |
| Full suite command | `go test ./...` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| NOTIF-CFG | Config parsing with notifications block | unit | `go test ./internal/config/ -run TestNotif -x` | Extend config_test.go |
| NOTIF-VAL | Validate channel refs and Shoutrrr URLs | unit | `go test ./internal/config/ -run TestValidateNotif -x` | Extend validate_test.go |
| NOTIF-FMT | FormatNotification output correctness | unit | `go test ./internal/engine/ -run TestFormatNotif -x` | New notify_test.go |
| NOTIF-SEND | SendNotification calls Shoutrrr correctly | unit | `go test ./internal/engine/ -run TestSendNotif -x` | New notify_test.go |
| NOTIF-RUN | Run() accumulates per-rule summaries and dispatches | unit | `go test ./internal/engine/ -run TestRunNotif -x` | Extend run_test.go |
| NOTIF-SILENT | No notification when no activity or no channel | unit | `go test ./internal/engine/ -run TestRunSilent -x` | Extend run_test.go |
| NOTIF-FAIL | Notification failure is non-fatal | unit | `go test ./internal/engine/ -run TestNotifFail -x` | Extend run_test.go |

### Sampling Rate
- **Per task commit:** `go test ./internal/...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/engine/notify.go` -- new file for RuleSummary, FormatNotification, SendNotification
- [ ] `internal/engine/notify_test.go` -- tests for notification formatting and send
- [ ] `go get github.com/nicholas-fedor/shoutrrr@v0.14.0` -- add dependency

## Sources

### Primary (HIGH confidence)
- [pkg.go.dev/github.com/nicholas-fedor/shoutrrr@v0.14.0](https://pkg.go.dev/github.com/nicholas-fedor/shoutrrr@v0.14.0) -- Send() signature, CreateSender, ServiceRouter.Send, types.Params
- [pkg.go.dev/...shoutrrr@v0.14.0/pkg/router](https://pkg.go.dev/github.com/nicholas-fedor/shoutrrr@v0.14.0/pkg/router) -- ServiceRouter methods (Send, Enqueue, Flush)
- [pkg.go.dev/...shoutrrr@v0.14.0/pkg/types](https://pkg.go.dev/github.com/nicholas-fedor/shoutrrr@v0.14.0/pkg/types) -- Params type, TitleKey, StdLogger
- Existing codebase: config.go, validate.go, run.go, actions.go -- all read directly

### Secondary (MEDIUM confidence)
- [Shoutrrr official docs](https://shoutrrr.nickfedor.com/v0.11.0/usage/go-package/) -- Go package usage examples
- [Shoutrrr GitHub](https://github.com/nicholas-fedor/shoutrrr) -- README, service list
- [Ntfy service docs](https://shoutrrr.nickfedor.com/v0.12.0/services/push/ntfy/) -- ntfy URL format

### Tertiary (LOW confidence)
- CreateSender network side effects at startup -- needs empirical validation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- single library, locked decision, verified version on pkg.go.dev
- Architecture: HIGH -- straightforward extension of existing patterns (Config structs, Run() loop, validation)
- Pitfalls: HIGH -- identified from direct code reading (KnownFields, pointer types, Run() signature)
- Shoutrrr API details: HIGH -- verified on pkg.go.dev for v0.14.0

**Research date:** 2026-03-20
**Valid until:** 2026-04-20 (stable domain, library API unlikely to change within 30 days)
