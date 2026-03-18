# Phase 8: Scheduler and Deployment - Context

**Gathered:** 2026-03-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Wrap the existing single-run path in a cron-driven daemon. The application runs as a long-lived process, executing rules on a user-defined schedule with timezone support, graceful signal handling, and Docker packaging. Single-run logic is already complete (Phase 7) — this phase adds the scheduler loop, signal handling, and deployment artifacts.

</domain>

<decisions>
## Implementation Decisions

### Graceful Shutdown
- Complete the current in-progress run before exiting — do not abort mid-run on SIGTERM/SIGINT
- Log a shutdown message (e.g., `received signal, shutting down`) before exiting so operators can distinguish an intentional stop from a crash
- Replace `context.Background()` (Phase 7 placeholder) with a signal-aware cancellable context; pass it through to the scheduler loop

### Run-on-start Behavior
- Execute rules immediately at startup before waiting for the first cron tick
- Prevents the operator from having to wait hours after a container restart for the first cleanup pass

### Cron Expression Format
- Standard 5-field cron format: `minute hour day month weekday` (e.g., `0 3 * * *`)
- No seconds field — not needed for a bookmark GC tool that runs daily or hourly
- Use `robfig/cron` v3 with `cron.WithParser(cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow))` to enforce 5-field parsing

### Cron Validation
- Validate the cron expression at config-load time (startup) — fail fast with a clear error
- Missing or empty `schedule` field is a validation error: exit with `"schedule is required"`
- Validation lives alongside existing `config.Validate()` so all config errors surface together

### Timezone
- Roadmap already specifies: default to UTC with a startup warning if `timezone` is unset
- Validate the timezone string at startup (e.g., `time.LoadLocation(cfg.Timezone)`) — exit with error if invalid
- Pass the resolved `*time.Location` to the cron scheduler

### Claude's Discretion
- Exact log format for "next run at" startup message (if any)
- Whether overlapping runs are possible (cron library handles this — skip if previous run is still in progress)
- Exact exit code on signal-driven shutdown (0 is fine)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §SCHED-01, SCHED-02, SCHED-03 — cron schedule, timezone default/warning, daemon lifecycle

### Existing Code
- `cmd/karaclean/main.go` — current single-run entry point; Phase 8 wraps this with a scheduler loop
- `internal/config/config.go` — `Config` struct already has `Timezone string` and `Schedule string` fields; validation must be extended to check these
- `internal/engine/run.go` — `engine.Run()` is the unit of work the scheduler calls on each tick

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `engine.Run(ctx, client, rules, dryRun)` — the work unit; scheduler calls this each tick and at startup
- `config.Config.Schedule` / `config.Config.Timezone` — fields already parsed from YAML, just unused
- `resolveDryRun()` in `main.go` — already handles flag/env/config precedence; no changes needed
- `requireEnv()` in `main.go` — already used for KARAKEEP_URL and KARAKEEP_API_KEY

### Established Patterns
- `log.Printf` for operational output (phases 6–7)
- `fmt.Fprintf(os.Stderr, ...) + os.Exit(1)` for startup failures
- `config.Validate()` returns `[]ValidationError` — extend with schedule/timezone checks

### Integration Points
- `main.go` Step 5 ("single run") becomes the daemon loop; Steps 0–4 remain unchanged
- Signal handling wraps the new scheduler loop (after auth check, before first run)

</code_context>

<specifics>
## Specific Ideas

- No specific UI/UX references — standard Go patterns apply
- Docker image: scratch base, static binary (CGO_ENABLED=0 GOOS=linux go build) — already in roadmap
- docker-compose.yml should show Karaclean as a sidecar alongside a Karakeep service with env vars and volume mount for config

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 08-scheduler-and-deployment*
*Context gathered: 2026-03-18*
