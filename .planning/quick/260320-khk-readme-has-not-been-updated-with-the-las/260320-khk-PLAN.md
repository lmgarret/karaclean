---
phase: quick
plan: 260320-khk
type: execute
wave: 1
depends_on: []
files_modified: [README.md, karaclean.example.yaml]
autonomous: true
requirements: [readme-update-per-rule-dryrun, readme-update-bookmark-size, readme-update-notifications]

must_haves:
  truths:
    - "README documents per-rule dryRun override with semantics and example"
    - "README documents notification system with config reference and example"
    - "README observability section mentions bookmark size in logs and total_size in summary"
    - "karaclean.example.yaml includes notifications section and per-rule dryRun example"
  artifacts:
    - path: "README.md"
      provides: "Complete documentation for all current features"
    - path: "karaclean.example.yaml"
      provides: "Fully commented example config with all features"
  key_links: []
---

<objective>
Update README.md and karaclean.example.yaml to document three features implemented since the last README update:

1. **Per-rule dryRun override** (quick task 260319-uni) -- Rule-level `dryRun` *bool that overrides global dryRun setting
2. **Bookmark size in logs** (quick task 260320-emk) -- Human-readable size in per-action log lines and total_size in run summary
3. **Notification system** (Phase 01) -- Shoutrrr-based notifications with channels, global default, per-rule override via `notify` field

Purpose: Users cannot discover or use these features without documentation.
Output: Updated README.md and karaclean.example.yaml
</objective>

<execution_context>
@/var/home/lm/git/karaclean/.claude/get-shit-done/workflows/execute-plan.md
@/var/home/lm/git/karaclean/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@README.md
@karaclean.example.yaml
@internal/config/config.go
@internal/config/testdata/valid_notifications.yaml
</context>

<tasks>

<task type="auto">
  <name>Task 1: Update README.md with per-rule dryRun, bookmark size, and notifications</name>
  <files>README.md</files>
  <action>
Update the existing README.md with the following additions. Preserve all existing content and writing style (concise, technical, no fluff).

**1. Per-rule dryRun override -- add to Rule Fields table and Dry-Run Mode section:**

In the "Rule Fields" table, add a new row:
| `dryRun` | bool | No | Override global dry-run for this rule. `true` forces dry-run, `false` forces live mode, omitted inherits global setting. |

In the "Dry-Run Mode" section, add a subsection after the precedence list:

### Per-Rule Override

Individual rules can override the global dry-run setting with a per-rule `dryRun` field:

```yaml
rules:
  - name: safe-archive
    conditions:
      olderThan: "30d"
    action: archive
    dryRun: true  # Always dry-run, regardless of global setting

  - name: aggressive-delete
    conditions:
      olderThan: "90d"
    action: delete
    # Inherits global dryRun setting (no override)
```

Resolution: per-rule `dryRun` (if set) takes precedence over global `dryRun`. The CLI flag and env var set the global value; per-rule overrides apply on top.

**2. Bookmark size in logs -- update Observability section:**

Update the "Each run" bullet to mention size:
- **Each run:** Logs a summary line: `run complete: archived=N deleted=M no_match=K excepted=J errors=E total_size=X.X MB`

Add a new bullet after "Each run":
- **Bookmark size:** Action log lines include human-readable file size (e.g., `size=1.2 MB`) when the bookmark has associated content. The run summary includes `total_size` showing total bytes processed (including dry-run actions).

**3. Notification system -- add new major section after "Observability":**

## Notifications

Karaclean can send per-rule action summaries to notification channels (Slack, ntfy, Telegram, and any service supported by [Shoutrrr](https://github.com/nicholas-fedor/shoutrrr)). Notifications are opt-in -- omit the `notifications` section to disable them entirely.

### Configuration

Add a `notifications` block to your config file:

```yaml
notifications:
  channels:
    my-ntfy:
      url: "ntfy://ntfy.sh/karaclean-alerts"
    slack-ops:
      url: "slack://hook:TOKEN-A/TOKEN-B/TOKEN-C@webhook"
  default: my-ntfy
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `channels` | map | Yes | Named notification channels, each with a Shoutrrr URL |
| `default` | string | No | Channel name to use when a rule has no `notify` override |

Channel URLs follow the [Shoutrrr URL format](https://github.com/nicholas-fedor/shoutrrr/blob/main/docs/services/overview.md). URLs are validated at startup -- an invalid URL prevents Karaclean from starting.

### Per-Rule Channel Override

Each rule can specify which channel receives its notifications using the `notify` field:

```yaml
rules:
  - name: archive-old-rss
    conditions:
      olderThan: "30d"
      source: rss
    action: archive
    notify: slack-ops  # Send this rule's summary to slack-ops instead of default
```

Channel resolution order:
1. Rule's `notify` field (if set, must reference a defined channel)
2. Global `default` channel (if set)
3. Silent -- no notification sent for this rule

### Message Format

Each notification includes the rule name, action counts, and is sent only when the rule produced activity (archived, deleted, or errored bookmarks). Rules that only matched excepted bookmarks are silent.

### Notification Failures

Notification delivery is best-effort. If a notification fails to send, the error is logged but does not affect the run outcome -- bookmarks are still processed normally.

**4. Update the top-level fields table** to add `notifications`:

| `notifications` | object | No | -- | Notification channels for per-rule action summaries (see [Notifications](#notifications)) |

**5. Update the Rule Fields table** to add `notify`:

| `notify` | string | No | Channel name for this rule's notification (overrides `default`). Must reference a channel defined in `notifications.channels`. |

  </action>
  <verify>
    <automated>grep -c "Notifications\|Per-Rule Channel Override\|dryRun.*bool\|total_size\|Shoutrrr" README.md | xargs test 5 -le</automated>
  </verify>
  <done>README.md documents all three features: per-rule dryRun (in Rule Fields table + Dry-Run section), bookmark size (in Observability), and notifications (new major section with config reference, per-rule override, message format, and failure behavior)</done>
</task>

<task type="auto">
  <name>Task 2: Update karaclean.example.yaml with notifications and per-rule dryRun</name>
  <files>karaclean.example.yaml</files>
  <action>
Update karaclean.example.yaml to include the new features with full comments. Add to the existing file:

**1. Add a notifications section** after the `dryRun` field and before `rules`:

```yaml
# Notifications: opt-in per-rule action summaries via Shoutrrr.
# Omit this entire section to disable notifications.
# Channel URLs use Shoutrrr format: https://github.com/nicholas-fedor/shoutrrr
notifications:
  channels:
    my-ntfy:
      url: "ntfy://ntfy.sh/karaclean-alerts"
    # slack-ops:
    #   url: "slack://hook:TOKEN-A/TOKEN-B/TOKEN-C@webhook"
    # telegram:
    #   url: "telegram://TOKEN@telegram?channels=CHAT_ID"
  default: my-ntfy   # Channel used when a rule has no "notify" override
```

**2. Add per-rule dryRun and notify examples** to the existing rules. Add `dryRun: true` to the delete-ancient-archived rule as a safety example. Add `notify: my-ntfy` to the first rule to show per-rule override. Specifically:

On the archive-old-rss rule, add a comment showing notify is optional:
```yaml
    # notify: slack-ops  # Optional: send this rule's summary to a specific channel
```

On the delete-ancient-archived rule, add dryRun override:
```yaml
    dryRun: true           # Always dry-run this dangerous rule, regardless of global setting
```

**3. Add comments in the summary block at the bottom** documenting the new rule-level fields:
```yaml
# Per-rule overrides:
#   dryRun: true/false  - override global dry-run for this rule (omit to inherit global)
#   notify: "channel"   - send this rule's summary to a specific notification channel
```

Keep existing style: comments above or inline, concise explanations.
  </action>
  <verify>
    <automated>grep -c "notifications\|notify\|dryRun: true" karaclean.example.yaml | xargs test 3 -le</automated>
  </verify>
  <done>karaclean.example.yaml includes notifications section with commented-out Slack/Telegram examples, per-rule dryRun override on the delete rule, and notify field documentation</done>
</task>

</tasks>

<verification>
- README.md contains "Notifications" section with Shoutrrr reference
- README.md Rule Fields table has `dryRun` and `notify` rows
- README.md Observability section mentions `total_size` and bookmark size
- README.md Dry-Run section has "Per-Rule Override" subsection
- karaclean.example.yaml has `notifications:` block with channels
- karaclean.example.yaml has per-rule `dryRun: true` example
- All existing README content preserved (no regressions)
</verification>

<success_criteria>
A user reading README.md can discover and configure all three features (per-rule dryRun, bookmark size logging, notifications) without consulting source code. The example YAML demonstrates all new fields.
</success_criteria>

<output>
After completion, create `.planning/quick/260320-khk-readme-has-not-been-updated-with-the-las/260320-khk-SUMMARY.md`
</output>
