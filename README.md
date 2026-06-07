# Karaclean

[![CI](https://github.com/lmgarret/karaclean/actions/workflows/ci.yml/badge.svg)](https://github.com/lmgarret/karaclean/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lmgarret/karaclean)](https://goreportcard.com/report/github.com/lmgarret/karaclean)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A Docker sidecar that automatically cleans up Karakeep bookmarks based on declarative YAML rules.

## What is Karaclean?

[Karakeep](https://github.com/karakeep-app/karakeep) is a self-hosted bookmark manager. Over time, bookmarks accumulate -- RSS feeds import dozens a day, browser extensions capture pages you never revisit, and your collection grows into a sprawling backlog. Cleaning it up manually is tedious and easy to forget.

Karaclean solves this by letting you define declarative YAML rules that describe which bookmarks to archive, tag, favourite or delete, and when. It runs as a Docker sidecar alongside your Karakeep instance on a cron schedule, evaluating every bookmark against your rules each cycle.

Safety is built in. A **dry-run mode** lets you preview exactly what would happen before any mutations execute. **Exception clauses** protect bookmarks you care about -- favourites, tagged items, bookmarks with personal notes, or bookmarks in specific lists. **Strict config validation** rejects unknown fields at startup, so a typo like `olderThen` is caught immediately instead of silently ignored.

Note: this project has been an exploration of AI coding tools for me. Although I do use karaclean with my own Karakeep instance, use it at your own risk!

## Quick Start

1. **Get a Karakeep API key.** In the Karakeep web UI, go to Settings > API Keys and create a new key.

2. **Create a config file** named `karaclean.yaml`:

```yaml
schedule: "0 3 * * *"
rules:
  - name: archive-old-rss
    conditions:
      olderThan: "30d"
      source: rss
    unless:
      favourited: true
    action: archive
```

3. **Create a `docker-compose.yml`** (or add the karaclean service to your existing one):

```yaml
services:
  karaclean:
    image: ghcr.io/lmgarret/karaclean:latest
    environment:
      - KARAKEEP_URL=http://karakeep:3000
      - KARAKEEP_API_KEY=your-api-key-here
    volumes:
      - ./karaclean.yaml:/config/karaclean.yaml:ro
    depends_on:
      - karakeep
    restart: unless-stopped
```

4. **Start the service:**

```bash
docker compose up -d
```

5. **Check the logs:**

```bash
docker compose logs karaclean
```

You should see an initial run summary followed by the next scheduled run time.

## Configuration Reference

Configuration is a single YAML file. See [`karaclean.example.yaml`](karaclean.example.yaml) for a fully commented example.

### Top-level Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `schedule` | string | Yes | -- | Cron expression (5-field: minute hour day month weekday) |
| `timezone` | string | No | UTC (with warning) | IANA timezone name (e.g., `America/New_York`) |
| `dryRun` | bool | No | `false` | Log actions without executing mutations |
| `notifications` | object | No | -- | Notification channels for per-rule action summaries (see [Notifications](#notifications)) |
| `rules` | list | Yes | -- | List of cleanup rules (at least one required) |

### Rule Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Human-readable rule identifier (used in logs) |
| `conditions` | object | Yes | Matching criteria (all must match -- AND semantics) |
| `unless` | object | No | Exception criteria (any match skips -- OR semantics) |
| `action` | string | Yes | What to do with matched bookmarks. One of `archive`, `unarchive`, `delete`, `tag`, `untag`, `favourite`, `unfavourite` (see [Actions](#actions)) |
| `tag` | string | Conditional | Tag name to attach/detach. **Required** for `action: tag` and `action: untag`; must not be set for any other action. |
| `dryRun` | bool | No | Override global dry-run for this rule. `true` forces dry-run, `false` forces live mode, omitted inherits global setting. |
| `notify` | string | No | Channel name for this rule's notification (overrides `default`). Must reference a channel defined in `notifications.channels`. |

### Actions

The `action` field selects what karaclean does with each matched bookmark. Only `delete` is irreversible; the rest are safe, reversible operations against the Karakeep API.

| Action | Effect | Requires `tag` | Reversible |
|--------|--------|:---:|:---:|
| `archive` | Set the bookmark's archived flag to `true` | — | ✅ (`unarchive`) |
| `unarchive` | Set the bookmark's archived flag to `false` | — | ✅ (`archive`) |
| `delete` | Permanently remove the bookmark | — | ❌ |
| `tag` | Attach the named tag (created if it doesn't exist) | ✅ | ✅ (`untag`) |
| `untag` | Detach the named tag | ✅ | ✅ (`tag`) |
| `favourite` | Star the bookmark (favourited = `true`) | — | ✅ (`unfavourite`) |
| `unfavourite` | Unstar the bookmark (favourited = `false`) | — | ✅ (`favourite`) |

> **Testing changes safely.** Before letting any rule run live, preview it with [dry-run mode](#dry-run-mode) — it logs exactly what each rule *would* do without touching a single bookmark. This is the recommended way to validate a new or edited rule, especially anything using `delete`. Dry-run works for every action, including the destructive ones.
>
> **A standing review workflow.** Dry-run is great for vetting a rule, but its output lives in logs. If you'd rather curate inside Karakeep itself, use `tag` as a non-destructive stand-in for `delete`: point a `tag` rule at the conditions you'd eventually delete on (e.g. `tag: delete-candidate`), then browse that tag in Karakeep for a periodic manual pass. Delete the survivors with a later `delete` rule, or clear the tag from keepers with `untag`.

### Conditions (AND semantics)

All specified conditions must match for a bookmark to be selected. Unspecified conditions are ignored.

| Field | Type | Example | Description |
|-------|------|---------|-------------|
| `olderThan` | string | `"30d"` | Match bookmarks created more than N ago. Formats: `Nh` (hours), `Nd` (days), `Nw` (weeks), `Nmo` (months = 30 days), `Ny` (years = 365 days). Uses strictly-greater-than comparison (exact boundary does not match). |
| `source` | string | `"rss"` | Match bookmarks from this source. Valid values: `rss`, `web`, `api`, `mobile`, `extension`, `cli`, `import` |
| `archived` | bool | `true` | `true` = only archived bookmarks, `false` = only non-archived |
| `favourited` | bool | `true` | `true` = only favourited bookmarks, `false` = only non-favourited |
| `hasTag` | string | `"read-later"` | Match bookmarks that have this exact tag (case-sensitive) |
| `lacksTag` | string | `"keep"` | Match bookmarks that do NOT have this tag (case-sensitive) |
| `inList` | string or list | `"Read Later"` | Match bookmarks that belong to any of the specified Karakeep lists (OR semantics, case-sensitive). Accepts a single list name or a list of names. List names are validated at startup. |

### Exceptions (OR semantics)

If any exception matches, the bookmark is protected from the rule's action.

| Field | Type | Example | Description |
|-------|------|---------|-------------|
| `favourited` | bool | `true` | Skip if bookmark matches this star status |
| `hasTag` | string | `"important"` | Skip if bookmark has this tag (case-sensitive) |
| `hasNote` | bool | `true` | Skip if bookmark has a personal note (whitespace-only notes count as empty) |
| `archived` | bool | `true` | Skip if bookmark matches this archive status |
| `inList` | string or list | `"Important"` | Skip if bookmark belongs to any of the specified lists (OR semantics, case-sensitive) |

## Rule Examples

### 1. Archive old RSS after 30 days, protect favourites

```yaml
- name: archive-old-rss
  conditions:
    olderThan: "30d"
    source: rss
  unless:
    favourited: true
  action: archive
```

RSS bookmarks older than 30 days get archived. Any bookmark you have starred is left alone.

### 2. Delete ancient archived bookmarks, protect tagged and noted

```yaml
- name: delete-ancient-archived
  conditions:
    olderThan: "90d"
    archived: true
  unless:
    hasTag: keep-forever
    hasNote: true
  action: delete
```

Archived bookmarks older than 90 days are permanently deleted -- unless they carry the `keep-forever` tag or have a personal note attached.

### 3. Archive stale untagged bookmarks after 60 days

```yaml
- name: archive-stale-untagged
  conditions:
    olderThan: "60d"
    favourited: false
    lacksTag: keep
  unless:
    archived: true
  action: archive
```

Non-favourited bookmarks older than 60 days that lack the `keep` tag are archived. Already-archived bookmarks are skipped.

### 4. Delete all web bookmarks older than 1 year

```yaml
- name: delete-old-web
  conditions:
    olderThan: "1y"
    source: web
  action: delete
```

An aggressive cleanup rule with no exceptions. All web-sourced bookmarks older than 365 days are permanently deleted.

### 5. Archive bookmarks in the "Read Later" list after 14 days

```yaml
- name: archive-read-later
  conditions:
    olderThan: "14d"
    inList: "Read Later"
  unless:
    favourited: true
  action: archive
```

Bookmarks in the "Read Later" list older than 14 days are archived. Favourited items are protected.

### 6. Delete old RSS, but protect bookmarks in curated lists

```yaml
- name: delete-old-rss-except-curated
  conditions:
    olderThan: "60d"
    source: rss
  unless:
    inList:
      - "Best Articles"
      - "Reference"
  action: delete
```

RSS bookmarks older than 60 days are deleted -- unless they've been added to the "Best Articles" or "Reference" lists. The `inList` exception accepts a list of names (OR semantics: membership in any listed list protects the bookmark).

### 7. Archive mobile bookmarks older than 2 weeks unless they have notes

```yaml
- name: archive-mobile-quick
  conditions:
    olderThan: "2w"
    source: mobile
  unless:
    hasNote: true
  action: archive
```

Mobile bookmarks are often quick saves. This rule archives them after 2 weeks, but keeps any that have personal notes.

### 8. Flag deletion candidates with a tag instead of deleting

```yaml
- name: flag-stale-for-review
  conditions:
    olderThan: "90d"
    archived: true
  unless:
    hasTag: keep-forever
    hasNote: true
  action: tag
  tag: delete-candidate
```

Instead of deleting straight away, this tags old archived bookmarks with `delete-candidate`. Browse that tag in Karakeep for a final manual review, then delete the survivors (or remove the tag with an `untag` rule). A safer on-ramp to automated cleanup.

### 9. Unarchive recent items so they resurface for review

```yaml
- name: resurface-recently-archived
  conditions:
    archived: true
    hasTag: revisit
  action: unarchive
```

Bookmarks tagged `revisit` are pulled back out of the archive so they show up in your main list again -- a gentle reminder to take another look.

## Rule Evaluation

Rules are evaluated with the following semantics:

- **First-match-wins:** Rules are evaluated in order (top to bottom). The first rule whose conditions match a bookmark is the one applied. Subsequent rules are not checked for that bookmark.
- **Collect-then-act:** All bookmarks are fetched from the Karakeep API first, then rules are evaluated against the full set. This prevents pagination race conditions.
- **Exceptions override conditions:** If a bookmark matches a rule's conditions but also matches any of its `unless` exceptions, the bookmark is marked as "excepted" and left untouched.
- **Unmatched bookmarks are ignored:** If no rule's conditions match a bookmark, it is left completely untouched.

## Dry-Run Mode

Dry-run mode logs all intended actions without executing any mutations against the Karakeep API. This is the recommended way to test new rules before going live.

There are three ways to enable dry-run, listed in precedence order (highest first):

1. **CLI flag:** `--dry-run` (highest precedence)
2. **Environment variable:** `KARACLEAN_DRY_RUN=true` (or `1`)
3. **Config file:** `dryRun: true` (lowest precedence)

When dry-run is active, the log output shows `DRY-RUN archive: bookmark <id> (rule: <name>)` for each action that would have been taken.

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

**Recommendation:** Always run with dry-run enabled when setting up new rules. Review the logs, then disable dry-run once you are satisfied with the behavior.

## CLI Reference

| Flag | Default | Description |
|------|---------|-------------|
| `--config PATH` | (see resolution below) | Path to YAML config file |
| `--dry-run` | `false` | Enable dry-run mode (no mutations) |
| `--version` | -- | Print version, commit, and build date, then exit |

### Config Path Resolution

The config file path is resolved using the first match:

1. `--config PATH` flag (explicit path)
2. `KARACLEAN_CONFIG` environment variable
3. `/config/karaclean.yaml` (default, matches Docker volume mount convention)

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `KARAKEEP_URL` | Yes | Karakeep instance URL (e.g., `http://karakeep:3000`) |
| `KARAKEEP_API_KEY` | Yes | API key from Karakeep Settings > API Keys |
| `KARACLEAN_CONFIG` | No | Config file path override (default: `/config/karaclean.yaml`) |
| `KARACLEAN_DRY_RUN` | No | Set to `true` or `1` to enable dry-run mode |

## Docker

Karaclean is designed to run as a Docker container.

- **Base image:** Built from `scratch` (minimal attack surface, ~10MB final image)
- **Binary:** Statically compiled Go binary with no external dependencies
- **Volumes:** Mount your config file at `/config/karaclean.yaml` (read-only recommended)
- **Networking:** The container needs network access to your Karakeep API URL
- **Signals:** Responds to `SIGTERM` and `SIGINT` for graceful shutdown (waits for in-progress jobs to complete)
- **Timezone:** Embeds the Go timezone database (`time/tzdata`), so there is no need to mount `/usr/share/zoneinfo`

### Image Tags

Images are published to `ghcr.io/lmgarret/karaclean` and are **multi-arch**
(`linux/amd64` and `linux/arm64`, so they run on x86 servers and ARM boards like a
Raspberry Pi or Apple Silicon alike). Pick a tag based on how much you want to trade
automatic updates for stability:

| Tag | Points to | Use case |
|-----|-----------|----------|
| `latest` | Newest stable release | Track the latest release, hands-off |
| `1` | Latest `1.x` release | Get features and fixes within a major version |
| `1.4` | Latest `1.4.x` patch | Get bug-fix patches within a minor version |
| `1.4.2` | One exact, immutable release | Reproducible, pinned deployments |
| `edge` | Latest `main` build (unstable) | Try unreleased changes |
| `<sha>` | One exact commit | Debug a specific build |

```bash
docker pull ghcr.io/lmgarret/karaclean:latest   # newest release
docker pull ghcr.io/lmgarret/karaclean:1        # newest 1.x
docker pull ghcr.io/lmgarret/karaclean:1.4.2    # exact release
```

**Recommendation:** Pin to a minor (`:1.4`) or exact (`:1.4.2`) tag in production so
updates are deliberate. For the strongest guarantee, pin to an immutable **digest** --
it can never move, even if a tag is re-pushed:

```bash
docker pull ghcr.io/lmgarret/karaclean@sha256:<digest>
```

Avoid `:edge` and `:<sha>` outside of testing -- they track unreleased code.

### Building the Image

```bash
docker build -t karaclean .
```

### Running Directly

```bash
docker run \
  -e KARAKEEP_URL=http://host:3000 \
  -e KARAKEEP_API_KEY=your-api-key \
  -v ./karaclean.yaml:/config/karaclean.yaml:ro \
  karaclean
```

## Docker Compose

The repository includes a [`docker-compose.yml`](docker-compose.yml) that runs Karaclean alongside Karakeep:

```yaml
services:
  karakeep:
    image: ghcr.io/karakeep-app/karakeep:latest
    ports:
      - "3000:3000"
    # Add your Karakeep configuration here

  karaclean:
    build: .
    environment:
      - KARAKEEP_URL=http://karakeep:3000
      - KARAKEEP_API_KEY=${KARAKEEP_API_KEY}
    volumes:
      - ./karaclean.yaml:/config/karaclean.yaml:ro
    depends_on:
      - karakeep
    restart: unless-stopped
```

| Field | Purpose |
|-------|---------|
| `KARAKEEP_URL` | Points to the Karakeep service by its Docker Compose service name |
| `KARAKEEP_API_KEY` | Pulled from your shell environment or `.env` file |
| `volumes` | Mounts your local config as read-only inside the container |
| `depends_on` | Ensures Karakeep starts before Karaclean |
| `restart: unless-stopped` | Keeps Karaclean running across Docker restarts |

## Observability

Karaclean logs structured information at key points:

- **Startup:** Logs authentication check result, dry-run status, timezone (with a warning if defaulting to UTC), and the next scheduled run time.
- **Each run:** Logs a summary line: `run complete: archived=N deleted=M no_match=K excepted=J errors=E total_size=X.X MB`. The `archived`, `deleted` and total counters always appear; counters for the other actions (`unarchived`, `tagged`, `untagged`, `favourited`, `unfavourited`) are added only when non-zero.
- **Bookmark size:** Action log lines include human-readable file size (e.g., `size=1.2 MB`) when the bookmark has associated content. The run summary includes `total_size` showing total bytes processed (including dry-run actions).
- **Immediate first run:** Karaclean executes a run immediately at startup (synchronous, before the cron scheduler begins), so you get feedback right away.
- **Cron schedule:** Subsequent runs follow the configured cron schedule.
- **Overlap protection:** If a run takes longer than the cron interval, the next scheduled run is automatically skipped (`SkipIfStillRunning`).

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
| `notifyOnError` | bool | No | `false` | Send a notification to the default channel when config validation fails at startup |

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

Each notification includes the rule name, action counts, and is sent only when the rule produced activity (any action was taken, or the rule errored). Rules that only matched excepted bookmarks are silent.

### Notification Failures

Notification delivery is best-effort. If a notification fails to send, the error is logged but does not affect the run outcome -- bookmarks are still processed normally.

### Error Notifications

Set `notifyOnError: true` in the `notifications` block to receive a notification when config validation fails at startup. This is useful when running karaclean as an unattended Docker sidecar -- you'll be alerted via your notification channel instead of having to check container logs.

The error notification is sent to the `default` channel. If no default channel is set, the notification is silently skipped. Even if the YAML file has syntax errors, karaclean attempts a lenient parse to extract the notifications section and deliver the error alert.

## Building from Source

```bash
go build -o karaclean ./cmd/karaclean
./karaclean --config karaclean.yaml
```

Requires Go 1.26 or later.

The binary is version-stamped at build time. Check it with:

```bash
karaclean --version
# karaclean 1.4.2 (commit abc1234, built 2026-06-05)
```

## Releases

Releases are cut by pushing a Git tag of the form `vX.Y.Z`. A [GitHub Actions
workflow](.github/workflows/release.yml) then builds the multi-arch image, publishes
the full tag ladder (`X.Y.Z`, `X.Y`, `X`, and `latest`), and creates a
[GitHub Release](https://github.com/lmgarret/karaclean/releases) with a changelog
generated from the commits since the previous tag. See the [Image Tags](#image-tags)
table for how to consume them, and [CONTRIBUTING.md](CONTRIBUTING.md#releasing) for the
maintainer steps.

## Config Validation

Karaclean validates your config file thoroughly at startup, before any rules execute:

- **Unknown fields are rejected.** A typo like `olderThen` instead of `olderThan` produces a clear error immediately, rather than being silently ignored.
- **Missing required fields** (`conditions`, `action`, `schedule`) produce descriptive error messages.
- **Invalid enum values** for `source` (must be one of `rss`, `web`, `api`, `mobile`, `extension`, `cli`, `import`) and `action` (must be one of `archive`, `unarchive`, `delete`, `tag`, `untag`, `favourite`, `unfavourite`) are caught.
- **Tag-field consistency** is enforced: `tag`/`untag` actions require a non-empty `tag`, and any other action that sets `tag` is rejected.
- **Invalid duration formats** in `olderThan` are validated (must match `Nh`, `Nd`, `Nw`, `Nmo`, or `Ny`).
- **Invalid cron expressions** in `schedule` are caught.
- **Invalid timezone names** in `timezone` are caught.
- **Empty tag values** in `hasTag` and `lacksTag` are rejected.
- **Empty list names** in `inList` are rejected.
- **List name validation** at startup: all list names referenced in `inList` conditions and exceptions are checked against the Karakeep API. If any configured list name doesn't exist, Karaclean reports all missing names and exits.
- **All errors are collected and reported together**, not one at a time, so you can fix everything in a single pass.

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for how to set
up the project and the checks to run before opening a pull request. For security
issues, please follow the [security policy](SECURITY.md).

## License

Karaclean is released under the [MIT License](LICENSE).

## Built With

This project was built using [get-shit-done](https://github.com/lmignot/get-shit-done), an AI coding workflow for Claude Code. All phases -- from config parsing through CI -- were planned and executed with GSD's structured approach.
