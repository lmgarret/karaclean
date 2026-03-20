# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

---

## Milestone: v1.0 — MVP

**Shipped:** 2026-03-19
**Phases:** 10 | **Plans:** 20 | **Sessions:** ~3

### What Was Built

- Declarative YAML config loader with strict unknown-field rejection and collected semantic validation errors
- Karakeep API client via oapi-codegen with paginated `ListBookmarks`, auth-on-startup gate, and mockable interface
- Full condition engine: `olderThan`, `source`, `archived`, `favourited`, `hasTag`, `lacksTag`
- Exception evaluation: `unless favourited/hasTag/hasNote/archived/notArchived` with OR semantics
- Collect-then-act run orchestrator (`engine.Run()`) with structured `RunSummary` log
- Archive and delete actions with multi-source dry-run precedence (flag → env → config)
- Production Docker sidecar: cron daemon, timezone support, graceful SIGTERM, scratch image with embedded tzdata
- Extensive README with full config reference, CLI docs, and Docker usage
- GitHub Actions CI: test (`-race`), golangci-lint v2, conditional ghcr.io push on main

### What Worked

- **Phase boundary clarity** — each phase delivered a single testable capability that the next phase built on. No phase needed to backtrack to fix a prior phase's design.
- **Context + Research → Planning pipeline** — discuss-phase captured locked decisions upfront; research caught golangci-lint v2 breaking changes and GHCR permission patterns before they became runtime surprises.
- **Collect-then-act pattern** — designing engine.Run() with pagination-first/mutate-second made the entire rule evaluation safe by construction; no rework needed.
- **oapi-codegen for Karakeep client** — generating from the OpenAPI spec eliminated all HTTP boilerplate. The only friction was a naming collision (`Client` → `KarakeepClient`) caught immediately in tests.
- **Pointer types for optional config** — `*int`, `*string`, `*bool` everywhere made nil-vs-zero-value unambiguous across config, validation, and engine; no special-casing needed.

### What Was Inefficient

- **Phase 02 VERIFICATION.md skipped** — the verifier was not run after executing Phase 02, leaving CONF-03 with a stale `[ ]` in REQUIREMENTS.md and no VERIFICATION.md. Caught only at milestone audit. Cost: one extra session to run verifier retroactively.
- **ROADMAP.md progress table drift** — the progress table fell out of sync with actual completion during execution (some phases showed as not started when they were complete). Minor, but required a cleanup pass.
- **Nyquist status flags not updated** — 6 VALIDATION.md files shipped with `nyquist_compliant: false` because executors didn't update the flags post-task. The coverage was real; the flags were stale. Consider making flag updates part of the executor's done criteria.
- **go.mod indirect annotations** — 4 direct dependencies were left annotated `// indirect`. `go mod tidy` was not run as part of any phase. Low severity but cosmetically incorrect.

### Patterns Established

- **`_ "time/tzdata"` in main.go for scratch images** — embeds the timezone database into the binary; essential for any Go binary in a scratch container that uses `time.LoadLocation`.
- **`KarakeepClient` vs generated `Client`** — when using oapi-codegen, always check for name collisions in the generated package before naming your wrapper.
- **Engine package has zero karakeep imports** — `KarakeepAPI` interface lives in `engine/`, enabling mock-based testing without any HTTP dependency in tests.
- **Duration parser in `internal/duration/`** — utility packages shared across config and engine belong in `internal/`, not in either package, to avoid import cycles.
- **Table-driven tests with named subtests** — all condition/exception/validation tests use this pattern; consistent and easy to extend.
- **golangci-lint v2 config format** — `version: "2"` at top level, `linters.settings` (not `linters-settings`), `gosimple` removed (merged into `staticcheck`). Required for v2.x compatibility.

### Key Lessons

1. **Run the verifier after every phase execution, not just at milestone audit.** One missed verifier invocation for Phase 02 created a tracking gap that surfaced only at audit time. The code was correct — the paper trail was missing.
2. **Nyquist flag updates should be part of plan execution, not a separate pass.** The `nyquist_compliant` and `wave_0_complete` flags in VALIDATION.md are only useful if kept current. Consider adding them to the executor's done criteria.
3. **Research pays off for greenfield CI.** The phase-researcher caught the golangci-lint v2 config format change and the GHCR `packages: write` permission requirement before planning. Both would have been hard runtime failures.
4. **`go mod tidy` should run as part of the final phase** (or as a task in the CI phase). Indirect annotation drift is cosmetic but should be caught before shipping.
5. **Strict YAML parsing (`KnownFields: true`) is a better default than lenient for config files.** It prevented silent misconfiguration from the start and required no special handling in later phases.

### Cost Observations

- Model mix: ~70% opus (researcher, planner, executor), ~30% sonnet (plan checker, verifier, integration checker)
- Sessions: ~3 main sessions across 2 days
- Notable: Research-first approach on Phase 10 (CI) saved at least one iteration — golangci-lint v2 migration would have caused immediate CI failures without the upfront research.

---

## Milestone: v1.1 — Notifications

**Shipped:** 2026-03-20
**Phases:** 1 | **Plans:** 3 | **Quick Tasks:** 2 (260320-lfo, 260320-ls1)

### What Was Built

- Shoutrrr-backed notification system: per-rule channel dispatch after bookmark evaluation
- `Notifications` config block with named channels (Shoutrrr URLs), global default, and per-rule `notify` override
- `RuleSummary` accumulator with `HasActivity()` gate — silent when nothing happened
- `Notifier` interface with `ShoutrrrNotifier` implementation (testable without live services)
- `ResolveChannelURL`: rule override → global default → nil (silent) resolution chain
- `FormatNotification` / `FormatNotificationTitle` producing `Summary:\ndeleted: N | archived: N` messages
- HTTP 204 fix: `DeleteBookmark` now correctly accepts 204 No Content as success

### What Worked

- **Notifier interface pattern** — decoupling `ShoutrrrNotifier` behind an interface meant all dispatch logic was unit-tested without needing a live Shoutrrr endpoint. Zero friction.
- **`*Notifications` nil pointer for opt-in** — no feature flag, no boolean toggle. Nil = feature absent. Consistent with the pointer-types-for-optional-config pattern from v1.0.
- **Quick tasks for post-ship fixes** — both bugs (HTTP 204, title duplication) were discovered in production use and fixed immediately as quick tasks without disrupting the milestone workflow.
- **Shoutrrr URL validation at config load** — fail-fast before any rules run. No "worked fine until first notification" failures.

### What Was Inefficient

- **HTTP 204 bug shipped with the milestone** — the Karakeep API returns 204 for DELETE (not 200), but the client only accepted 200. This was a testable property that the plan didn't cover. The quick task fix was fast, but the defect shouldn't have shipped.
- **Title duplication in notification body** — initial message format duplicated the rule name in both the title and body. Noticed immediately in first real use; fixed in a quick task. A review of the actual rendered output before shipping would have caught it.

### Patterns Established

- **Post-milestone quick tasks are the right tool for polish** — minor UX/format issues discovered during real use are correctly handled as quick tasks, not phase work. They're small, reversible, and don't need research or planning overhead.
- **Human-verification items (live Shoutrrr delivery)** — the VERIFICATION.md `human_needed` status is the right signal. Don't block milestone completion on items that require live external services; note them and ship.

### Key Lessons

1. **Test the actual rendered output of format functions end-to-end in tests.** The notification body format had title duplication that unit tests didn't expose because they tested the format function in isolation without considering what the title field already contained.
2. **HTTP response code coverage should be part of HTTP client plan templates.** DELETE → 204, POST → 201 — common patterns that should be in the plan's `verify` criteria by default, not discovered post-ship.
3. **Quick tasks are high-leverage post-ship polish.** Two defects shipped, two quick tasks fixed them within the same session. No ceremony, no ceremony overhead.

### Cost Observations

- Model mix: ~80% opus (planner, executor), ~20% sonnet (verifier)
- Sessions: 1 session
- Notable: Short milestone (1 phase, 3 plans) — fast to execute. Quick task cycle time was under 5 minutes each.

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Sessions | Phases | Key Change |
|-----------|----------|--------|------------|
| v1.0 | ~3 | 10 | Initial project — full greenfield build |
| v1.1 | 1 | 1 | Additive feature milestone — quick cycle, 2 post-ship quick fixes |

### Cumulative Quality

| Milestone | Go Tests | Verification Score | LOC |
|-----------|----------|--------------------|-----|
| v1.0 | All pass (race-clean) | 99/100 must-haves | ~11,168 |
| v1.1 | All pass (race-clean) | 16/16 must-haves (1 human item) | ~12,372 |

### Top Lessons (Verified Across Milestones)

1. Run the verifier after every phase — don't save it for the milestone audit.
2. Research-first pays back on phases involving external tooling with breaking changes.
3. HTTP response code coverage (204 for DELETE, 201 for POST) should be in plan verify criteria by default.
4. Quick tasks are the right tool for post-ship format/polish fixes — fast cycle, no overhead.
