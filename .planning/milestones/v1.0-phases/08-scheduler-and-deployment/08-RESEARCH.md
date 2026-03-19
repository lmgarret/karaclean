# Phase 8: Scheduler and Deployment - Research

**Researched:** 2026-03-18
**Domain:** Go cron scheduling, signal handling, Docker packaging
**Confidence:** HIGH

## Summary

Phase 8 wraps the existing single-run path (Phase 7) in a cron-driven daemon with timezone support, graceful shutdown, and Docker packaging. The `robfig/cron` v3 library is the standard Go cron scheduler and provides all required features: custom 5-field parsing, timezone via `WithLocation`, overlap prevention via `SkipIfStillRunning`, and graceful stop via `Stop()` which returns a context that completes when running jobs finish.

The existing codebase already has `Schedule` and `Timezone` fields in `config.Config`, `engine.Run()` as the work unit, and `context.Background()` placeholders in `main.go` ready to be replaced with signal-aware contexts. The Docker image uses a standard multi-stage build pattern with `scratch` base, requiring CA certificates and timezone data to be copied from the builder stage.

**Primary recommendation:** Use `robfig/cron/v3` with `cron.WithParser` (5-field), `cron.WithLocation`, and `cron.WithChain(cron.SkipIfStillRunning)`. Signal handling via `signal.NotifyContext`. Multi-stage Dockerfile with `golang:1.26-alpine` builder and `scratch` final image.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Graceful shutdown: complete in-progress run before exiting on SIGTERM/SIGINT; log shutdown message; replace `context.Background()` with signal-aware cancellable context
- Run-on-start: execute rules immediately at startup before first cron tick
- Cron format: standard 5-field (`minute hour day month weekday`); use `robfig/cron` v3 with `cron.WithParser(cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow))` to enforce 5-field parsing
- Cron validation: validate at config-load time in `config.Validate()`; missing/empty `schedule` is a validation error
- Timezone: default UTC with startup warning if unset; validate with `time.LoadLocation()`; pass resolved `*time.Location` to cron scheduler
- Docker: scratch base, static binary (`CGO_ENABLED=0 GOOS=linux go build`)
- docker-compose.yml: sidecar alongside Karakeep with env vars and volume mount for config

### Claude's Discretion
- Exact log format for "next run at" startup message
- Whether to use `SkipIfStillRunning` for overlap prevention
- Exact exit code on signal-driven shutdown (0 is fine)

### Deferred Ideas (OUT OF SCOPE)
None.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| SCHED-01 | User defines a cron schedule expression in YAML config | `robfig/cron` v3 parser validates 5-field expressions; `config.Validate()` extended with schedule check |
| SCHED-02 | User defines explicit timezone in config (defaults to UTC with startup warning if unset) | `time.LoadLocation()` validates timezone strings; `cron.WithLocation()` passes to scheduler; `log.Printf` warning at startup |
| SCHED-03 | Container runs as a daemon executing rules on the defined schedule | `cron.Start()` runs scheduler goroutine; `signal.NotifyContext` for SIGTERM/SIGINT; `cron.Stop()` for graceful shutdown; run-on-start before cron loop |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/robfig/cron/v3` | v3.0.1 | Cron scheduler | De facto Go cron library; 13k+ GitHub stars; stable since 2020; built-in timezone, job wrappers, graceful stop |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `os/signal` | stdlib | Signal handling | Capture SIGTERM/SIGINT for graceful shutdown |
| `context` | stdlib | Cancellation propagation | Signal-aware context passed to `engine.Run()` |
| `time` | stdlib | Timezone loading | `time.LoadLocation()` for timezone validation and cron location |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `robfig/cron` | `go-co-op/gocron` | gocron is higher-level but heavier; robfig/cron is simpler, lighter, and sufficient |
| `scratch` base | `gcr.io/distroless/static` | Distroless includes CA certs and tzdata automatically but is ~2MB larger; scratch is smaller but requires manual COPY |

**Installation:**
```bash
go get github.com/robfig/cron/v3@v3.0.1
```

**Version verification:** v3.0.1 is the latest release (published 2020-01-04, stable, no further releases). Verified via `proxy.golang.org`.

## Architecture Patterns

### Recommended Project Structure
```
cmd/karaclean/main.go          # Daemon entry point (scheduler loop + signal handling)
internal/config/config.go       # Config struct (Schedule, Timezone fields already exist)
internal/config/validate.go     # Extended with schedule + timezone validation
internal/engine/run.go          # engine.Run() unchanged — work unit called by scheduler
Dockerfile                      # Multi-stage build (NEW)
docker-compose.yml              # Sidecar deployment example (NEW)
karaclean.example.yaml          # Full example config documenting all features (NEW)
```

### Pattern 1: Signal-Aware Daemon Loop
**What:** Use `signal.NotifyContext` to create a cancellable context, then block on `<-ctx.Done()` after starting the cron scheduler.
**When to use:** Any long-lived Go process that needs graceful SIGTERM/SIGINT handling.
**Example:**
```go
// Source: Go stdlib os/signal docs
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
defer stop()

// Start cron scheduler
c.Start()

// Run immediately at startup
runOnce(ctx, client, rules, dryRun)

// Block until signal received
<-ctx.Done()
log.Println("received signal, shutting down")

// Stop cron and wait for in-progress job
stopCtx := c.Stop()
<-stopCtx.Done()
```

### Pattern 2: Cron with 5-Field Parser and Timezone
**What:** Configure `robfig/cron` with strict 5-field parsing and explicit timezone.
**When to use:** When you want standard cron format without seconds field.
**Example:**
```go
// Source: robfig/cron v3 pkg.go.dev docs
loc, err := time.LoadLocation(cfg.Timezone) // already validated

parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
c := cron.New(
    cron.WithParser(parser),
    cron.WithLocation(loc),
    cron.WithChain(
        cron.SkipIfStillRunning(cron.DefaultLogger),
    ),
)

c.AddFunc(cfg.Schedule, func() {
    runOnce(ctx, client, rules, dryRun)
})
```

### Pattern 3: Config Validation Extension
**What:** Add schedule and timezone validation to the existing `Validate()` method.
**When to use:** Fail-fast at startup, consistent with existing validation pattern.
**Example:**
```go
// Validate schedule field
if c.Schedule == "" {
    errs = append(errs, ValidationError{Field: "schedule", Message: "schedule is required"})
} else {
    parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
    if _, err := parser.Parse(c.Schedule); err != nil {
        errs = append(errs, ValidationError{Field: "schedule", Message: fmt.Sprintf("invalid cron expression: %v", err)})
    }
}

// Validate timezone field (empty is valid — defaults to UTC with warning)
if c.Timezone != "" {
    if _, err := time.LoadLocation(c.Timezone); err != nil {
        errs = append(errs, ValidationError{Field: "timezone", Message: fmt.Sprintf("invalid timezone: %v", err)})
    }
}
```

### Anti-Patterns to Avoid
- **Using `time.Ticker` instead of cron library:** Tickers drift over time and don't handle cron expressions. Use `robfig/cron`.
- **Calling `os.Exit()` in signal handler:** Prevents graceful completion of in-progress run. Use context cancellation instead.
- **Ignoring `cron.Stop()` return context:** The returned context signals when running jobs complete. Not waiting on it risks killing mid-run jobs.
- **Putting cron dependency in config package:** Config validation needs `cron.NewParser` to validate expressions, which adds `robfig/cron` as a dependency of the config package. This is acceptable since it's validation-only and the alternative (regex validation) would be incomplete.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Cron expression parsing | Custom regex or field parser | `cron.NewParser().Parse()` | Cron parsing has edge cases (ranges, steps, month/dow names); library handles all of them |
| Cron scheduling loop | `time.Timer` + next-tick calculation | `robfig/cron` scheduler | DST transitions, missed ticks, and scheduling edge cases are non-trivial |
| Overlap prevention | Manual mutex/flag around job | `cron.SkipIfStillRunning` | Built-in, tested, handles edge cases around panics |
| Timezone validation | String matching against known zones | `time.LoadLocation()` | Relies on system tzdata; covers all IANA zones including aliases |
| Signal handling | Manual `os.Signal` channel + select | `signal.NotifyContext` | stdlib since Go 1.16; cleaner, composable with existing context patterns |

**Key insight:** The cron library and Go stdlib provide everything needed. No custom scheduling logic is necessary.

## Common Pitfalls

### Pitfall 1: Missing tzdata in scratch image
**What goes wrong:** `time.LoadLocation("America/New_York")` fails at runtime because scratch has no timezone database.
**Why it happens:** Go's `time` package looks for tzdata in the filesystem. Scratch has no files.
**How to avoid:** Copy `/usr/share/zoneinfo` from the builder stage, OR embed tzdata with `import _ "time/tzdata"` (adds ~450KB to binary but is simpler and more reliable).
**Warning signs:** Works in development (host has tzdata), fails in container.
**Recommendation:** Use `import _ "time/tzdata"` -- it is simpler and eliminates the filesystem dependency entirely. Go has included this since 1.15.

### Pitfall 2: Missing CA certificates in scratch image
**What goes wrong:** HTTPS calls to Karakeep API fail with `x509: certificate signed by unknown authority`.
**Why it happens:** Scratch has no CA certificate bundle.
**How to avoid:** Copy `/etc/ssl/certs/ca-certificates.crt` from the builder stage.
**Warning signs:** Works locally, fails in container when calling Karakeep API.

### Pitfall 3: cron.Stop() does not cancel running jobs
**What goes wrong:** `cron.Stop()` only stops scheduling new runs. It does NOT cancel a currently-running job's context.
**Why it happens:** robfig/cron does not inject contexts into job functions. The job function must check its own context.
**How to avoid:** Pass the signal-aware context (`ctx` from `signal.NotifyContext`) into the job closure. When signal fires, `ctx` is cancelled, and `engine.Run` receives the cancelled context. `cron.Stop()` then waits for the job to finish via its returned context.
**Warning signs:** Container takes a long time to shut down because the running job ignores the signal.

### Pitfall 4: Run-on-start race with cron scheduler
**What goes wrong:** If run-on-start is executed in a goroutine and the cron scheduler fires immediately, two runs could overlap.
**Why it happens:** Cron tick might fire while the startup run is still in progress.
**How to avoid:** Use `cron.SkipIfStillRunning` wrapper, which will skip the cron-triggered run if the startup run is still executing. Or run the startup execution synchronously before calling `c.Start()`.
**Recommendation:** Run startup execution synchronously before `c.Start()` -- simplest approach, no overlap possible.

### Pitfall 5: Validation importing cron library
**What goes wrong:** Adding `robfig/cron` as a dependency of `internal/config` feels like tight coupling.
**Why it happens:** Config validation needs to parse the cron expression to verify it.
**How to avoid:** This is acceptable. The alternative (not validating the cron expression in config) would mean the error surfaces later at scheduler setup time, breaking the established fail-fast-at-config-load pattern. The `config` package already imports `internal/duration` for similar validation purposes.

## Code Examples

### Complete Daemon Main Function (Skeleton)
```go
// Source: robfig/cron v3 docs + Go stdlib os/signal docs
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/robfig/cron/v3"
    _ "time/tzdata" // embed timezone database for scratch images
)

func main() {
    // Steps 0-4 unchanged from Phase 7...

    // Resolve timezone
    loc := time.UTC
    if cfg.Timezone != "" {
        loc, _ = time.LoadLocation(cfg.Timezone) // already validated
    } else {
        log.Println("WARNING: timezone not set, defaulting to UTC")
    }

    // Create signal-aware context
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
    defer stop()

    // Run immediately at startup
    summary, err := engine.Run(ctx, client, cfg.Rules, dryRun)
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
    log.Printf("run complete: %s", summary)

    // Set up cron scheduler
    parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
    c := cron.New(
        cron.WithParser(parser),
        cron.WithLocation(loc),
        cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)),
    )
    c.AddFunc(cfg.Schedule, func() {
        summary, err := engine.Run(ctx, client, cfg.Rules, dryRun)
        if err != nil {
            log.Printf("error: %v", err)
            return
        }
        log.Printf("run complete: %s", summary)
    })

    // Log next run time
    c.Start()
    entries := c.Entries()
    if len(entries) > 0 {
        log.Printf("next run at %s", entries[0].Next.Format(time.RFC3339))
    }

    // Block until signal
    <-ctx.Done()
    log.Println("received signal, shutting down")

    // Graceful stop -- wait for in-progress job
    stopCtx := c.Stop()
    <-stopCtx.Done()
}
```

### Dockerfile (Multi-Stage Scratch)
```dockerfile
# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o karaclean ./cmd/karaclean

# Final stage
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/karaclean /karaclean
ENTRYPOINT ["/karaclean"]
```

### docker-compose.yml (Sidecar Example)
```yaml
services:
  karakeep:
    image: ghcr.io/karakeep-app/karakeep:latest
    # ... karakeep config ...

  karaclean:
    image: karaclean:latest
    environment:
      - KARAKEEP_URL=http://karakeep:3000
      - KARAKEEP_API_KEY=${KARAKEEP_API_KEY}
    volumes:
      - ./karaclean.yaml:/config/karaclean.yaml:ro
    depends_on:
      - karakeep
    restart: unless-stopped
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `os.Signal` channel + select | `signal.NotifyContext` | Go 1.16 (2021) | Cleaner context-based signal handling |
| Copy tzdata files in Dockerfile | `import _ "time/tzdata"` | Go 1.15 (2020) | Embedded tzdata eliminates filesystem dependency |
| robfig/cron v2 (6-field default) | robfig/cron v3 (5-field default + WithParser) | 2020 | Configurable parser, functional options, Job wrappers |

**Deprecated/outdated:**
- robfig/cron v1/v2: Use v3 with module path `github.com/robfig/cron/v3`
- `CRON_TZ=` prefix in expressions: Still supported but `WithLocation` is cleaner for global timezone

## Open Questions

1. **Startup run error handling**
   - What we know: Single-run mode (Phase 7) exits on `engine.Run` error
   - What's unclear: Should the daemon also exit on startup run error, or log and continue to cron schedule?
   - Recommendation: Exit on startup error -- if the first run fails (e.g., API unreachable), the container should crash and let Docker restart policy handle retry. This matches the existing fail-fast pattern.

2. **Log format for cron library**
   - What we know: `cron.DefaultLogger` uses Go's default logger; `SkipIfStillRunning` logs when skipping
   - What's unclear: Whether cron library logs are consistent with existing `log.Printf` format
   - Recommendation: Use `cron.DefaultLogger` -- it uses `log.Printf` internally, consistent with existing patterns.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none -- Go convention, `go test` works out of the box |
| Quick run command | `go test ./... -count=1 -short` |
| Full suite command | `go test ./... -count=1 -v` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SCHED-01 | Schedule field validated at config load; invalid cron rejected | unit | `go test ./internal/config/ -run TestValidate -count=1 -v` | Partially (validate_test.go exists, needs schedule cases) |
| SCHED-02 | Timezone validated; empty defaults to UTC with warning | unit | `go test ./internal/config/ -run TestValidate -count=1 -v` | Partially (validate_test.go exists, needs timezone cases) |
| SCHED-03 | Daemon runs on schedule, handles SIGTERM gracefully | integration | `go test ./cmd/karaclean/ -run TestScheduler -count=1 -v -short` | No (new test file needed) |

### Sampling Rate
- **Per task commit:** `go test ./... -count=1 -short`
- **Per wave merge:** `go test ./... -count=1 -v`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/config/validate_test.go` -- add schedule and timezone validation test cases (file exists, needs new cases)
- [ ] `cmd/karaclean/main_test.go` -- add scheduler/signal handling tests (file exists, needs new test functions)
- [ ] `go get github.com/robfig/cron/v3@v3.0.1` -- new dependency

## Sources

### Primary (HIGH confidence)
- [robfig/cron v3 - pkg.go.dev](https://pkg.go.dev/github.com/robfig/cron/v3) - API reference for New, WithParser, WithLocation, WithChain, SkipIfStillRunning, Stop
- [Go stdlib os/signal](https://pkg.go.dev/os/signal) - signal.NotifyContext API
- [Go stdlib time/tzdata](https://pkg.go.dev/time/tzdata) - embedded timezone database
- proxy.golang.org - verified v3.0.1 is latest (published 2020-01-04)

### Secondary (MEDIUM confidence)
- [Docker multi-stage builds for Go](https://docs.docker.com/build/building/multi-stage/) - official Docker docs on multi-stage patterns
- [Go Docker best practices](https://docs.docker.com/guides/golang/build-images/) - official Docker guide for Go

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - robfig/cron v3 is the de facto Go cron library, verified via pkg.go.dev
- Architecture: HIGH - patterns are standard Go idioms (signal.NotifyContext, context propagation, multi-stage Docker)
- Pitfalls: HIGH - tzdata and CA cert issues are well-documented; cron.Stop() behavior verified in official docs

**Research date:** 2026-03-18
**Valid until:** 2026-04-18 (stable domain, robfig/cron v3 unchanged since 2020)
