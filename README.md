# Karaclean

A Docker sidecar that automatically cleans up Karakeep bookmarks based on declarative YAML rules.

## What is Karaclean?

[Karakeep](https://github.com/karakeep-app/karakeep) is a self-hosted bookmark manager. Over time, bookmarks accumulate -- RSS feeds import dozens a day, browser extensions capture pages you never revisit, and your collection grows into a sprawling backlog. Cleaning it up manually is tedious and easy to forget.

Karaclean solves this by letting you define declarative YAML rules that describe which bookmarks to archive or delete, and when. It runs as a Docker sidecar alongside your Karakeep instance on a cron schedule, evaluating every bookmark against your rules each cycle.

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
| `action` | string | Yes | `archive` or `delete` |
| `dryRun` | bool | No | Override global dry-run for this rule. `true` forces dry-run, `false` forces live mode, omitted inherits global setting. |
| `notify` | string | No | Channel name for this rule's notification (overrides `default`). Must reference a channel defined in `notifications.channels`. |

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

The Docker image is published to `ghcr.io/lmgarret/karaclean` with two tag strategies:

| Tag | Description |
|-----|-------------|
| `latest` | Always points to the most recent build from the `main` branch |
| `<sha>` | Short Git commit SHA (e.g., `abc1234`) for pinning to a specific build |

```bash
docker pull ghcr.io/lmgarret/karaclean:latest
docker pull ghcr.io/lmgarret/karaclean:<sha>
```

**Recommendation:** Pin to a SHA tag in production to avoid unexpected changes from new builds.

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
- **Each run:** Logs a summary line: `run complete: archived=N deleted=M no_match=K excepted=J errors=E total_size=X.X MB`
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

## Building from Source

```bash
go build -o karaclean ./cmd/karaclean
./karaclean --config karaclean.yaml
```

Requires Go 1.26 or later.

## Config Validation

Karaclean validates your config file thoroughly at startup, before any rules execute:

- **Unknown fields are rejected.** A typo like `olderThen` instead of `olderThan` produces a clear error immediately, rather than being silently ignored.
- **Missing required fields** (`conditions`, `action`, `schedule`) produce descriptive error messages.
- **Invalid enum values** for `source` (must be one of `rss`, `web`, `api`, `mobile`, `extension`, `cli`, `import`) and `action` (must be `archive` or `delete`) are caught.
- **Invalid duration formats** in `olderThan` are validated (must match `Nh`, `Nd`, `Nw`, `Nmo`, or `Ny`).
- **Invalid cron expressions** in `schedule` are caught.
- **Invalid timezone names** in `timezone` are caught.
- **Empty tag values** in `hasTag` and `lacksTag` are rejected.
- **Empty list names** in `inList` are rejected.
- **List name validation** at startup: all list names referenced in `inList` conditions and exceptions are checked against the Karakeep API. If any configured list name doesn't exist, Karaclean reports all missing names and exits.
- **All errors are collected and reported together**, not one at a time, so you can fix everything in a single pass.

## Built With

This project was built using [get-shit-done](https://github.com/lmignot/get-shit-done), an AI coding workflow for Claude Code. All phases -- from config parsing through CI -- were planned and executed with GSD's structured approach.
