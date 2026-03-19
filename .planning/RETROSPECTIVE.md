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

## Cross-Milestone Trends

### Process Evolution

| Milestone | Sessions | Phases | Key Change |
|-----------|----------|--------|------------|
| v1.0 | ~3 | 10 | Initial project — full greenfield build |

### Cumulative Quality

| Milestone | Go Tests | Verification Score | LOC |
|-----------|----------|--------------------|-----|
| v1.0 | All pass (race-clean) | 99/100 must-haves | ~11,168 |

### Top Lessons (Verified Across Milestones)

1. Run the verifier after every phase — don't save it for the milestone audit.
2. Research-first pays back on phases involving external tooling with breaking changes.
