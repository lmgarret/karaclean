# Roadmap: Karaclean

## Overview

Karaclean delivers a Docker sidecar that automatically archives and deletes Karakeep bookmarks based on user-defined YAML rules. The roadmap builds from config parsing and API integration through a layered rule engine (conditions, exceptions, actions), then adds the run orchestrator with collect-then-act safety, and finishes with cron scheduling and Docker packaging. Each phase delivers a testable, coherent capability that the next phase builds on.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Config Loading and Validation** - YAML config parsing with strict unknown-field rejection and semantic validation
- [ ] **Phase 2: API Client and Authentication** - Karakeep HTTP client with typed responses and startup auth verification
- [ ] **Phase 3: Age and Source Conditions** - Matcher foundation with time-based and source-based bookmark filtering
- [ ] **Phase 4: Status and Tag Conditions** - Archived, favourited, and tag-based condition matching
- [ ] **Phase 5: Exception Evaluation** - Protection clauses that prevent rules from acting on bookmarks the user cares about
- [ ] **Phase 6: Actions and Dry-Run** - Archive and delete mutations with dry-run mode that blocks all changes
- [ ] **Phase 7: Run Orchestrator and Observability** - Collect-then-act execution loop with structured run summaries
- [ ] **Phase 8: Scheduler and Deployment** - Cron-based daemon with timezone support, Docker image, and compose example

## Phase Details

### Phase 1: Config Loading and Validation
**Goal**: Users can write a YAML config file and get immediate, clear feedback if it contains errors
**Depends on**: Nothing (first phase)
**Requirements**: CONF-01, CONF-02
**Success Criteria** (what must be TRUE):
  1. User can write a YAML config file with rules containing conditions, exceptions, and actions, and the application parses it into typed Go structs
  2. User receives a clear startup error if the YAML contains a misspelled or unknown field (strict parsing rejects silent misconfiguration)
  3. User receives a clear startup error if the config is semantically invalid (e.g., missing required fields, invalid enum values for source or action)
  4. Go module is initialized with project structure (`cmd/karaclean/`, `internal/config/`) and the config types compile
**Plans**: TBD

Plans:
- [ ] 01-01: TBD
- [ ] 01-02: TBD

### Phase 2: API Client and Authentication
**Goal**: The application can communicate with Karakeep and confirms a valid connection before doing any work
**Depends on**: Phase 1
**Requirements**: CONF-03
**Success Criteria** (what must be TRUE):
  1. Application validates the API bearer token against Karakeep on startup and exits with a clear error if authentication fails
  2. API client can list bookmarks with cursor-based pagination and return typed Go structs with all relevant fields (id, createdAt, archived, favourited, source, tags, note)
  3. API client interface (`KarakeepAPI`) is defined in the engine package, enabling mock-based testing without real HTTP calls
**Plans**: TBD

Plans:
- [ ] 02-01: TBD
- [ ] 02-02: TBD

### Phase 3: Age and Source Conditions
**Goal**: Rules can identify bookmarks based on how old they are and where they came from
**Depends on**: Phase 2
**Requirements**: COND-01, COND-02
**Success Criteria** (what must be TRUE):
  1. A rule with `olderThan: 30d` matches only bookmarks created more than 30 days ago
  2. A rule with `source: rss` matches only bookmarks ingested from RSS feeds (and similarly for web, api, mobile, extension, cli, import)
  3. Conditions compose correctly -- a rule with both `olderThan` and `source` matches only bookmarks satisfying both
  4. Matcher functions are pure (no I/O) and have comprehensive unit tests
**Plans**: TBD

Plans:
- [ ] 03-01: TBD
- [ ] 03-02: TBD

### Phase 4: Status and Tag Conditions
**Goal**: Rules can filter bookmarks by their archived/favourited status and tag presence
**Depends on**: Phase 3
**Requirements**: COND-03, COND-04, COND-05, COND-06
**Success Criteria** (what must be TRUE):
  1. A rule with `archived: true` matches only archived bookmarks; `archived: false` matches only non-archived bookmarks
  2. A rule with `favourited: true` matches only favourited bookmarks; `favourited: false` matches only non-favourited bookmarks
  3. A rule with `hasTag: read-later` matches only bookmarks that carry the specified tag
  4. A rule with `lacksTag: keep` matches only bookmarks that do not carry the specified tag
  5. All condition types (age, source, status, tags) can be combined in a single rule and all must match for the rule to apply
**Plans**: TBD

Plans:
- [ ] 04-01: TBD
- [ ] 04-02: TBD

### Phase 5: Exception Evaluation
**Goal**: Users can protect bookmarks they care about from being affected by cleanup rules
**Depends on**: Phase 4
**Requirements**: EXCP-01, EXCP-02, EXCP-03, EXCP-04
**Success Criteria** (what must be TRUE):
  1. A rule with `unless: favourited` skips any bookmark that is starred, even if all other conditions match
  2. A rule with `unless: hasTag: important` skips any bookmark carrying the specified tag
  3. A rule with `unless: hasNote` skips any bookmark where the user has added a personal note
  4. A rule with `unless: archived` or `unless: notArchived` skips bookmarks based on their archive status
  5. Multiple exception clauses on a single rule are evaluated with OR semantics (any exception triggers a skip)
**Plans**: TBD

Plans:
- [ ] 05-01: TBD
- [ ] 05-02: TBD

### Phase 6: Actions and Dry-Run
**Goal**: Rules can archive or delete bookmarks, and dry-run mode lets users preview changes safely
**Depends on**: Phase 5
**Requirements**: ACTN-01, ACTN-02, ACTN-03
**Success Criteria** (what must be TRUE):
  1. A rule with `action: archive` sets `archived: true` on matched bookmarks via the Karakeep PATCH API
  2. A rule with `action: delete` permanently removes matched bookmarks via the Karakeep DELETE API
  3. When dry-run mode is enabled, no mutations (archive or delete) are executed against the API; all intended actions are logged instead
  4. Dry-run output clearly shows what each bookmark's fate would be (archive vs delete) and why (which rule matched)
**Plans**: TBD

Plans:
- [ ] 06-01: TBD
- [ ] 06-02: TBD

### Phase 7: Run Orchestrator and Observability
**Goal**: A complete rule evaluation run executes safely with collect-then-act ordering and produces a summary report
**Depends on**: Phase 6
**Requirements**: OBS-01
**Success Criteria** (what must be TRUE):
  1. A single run paginates all bookmarks into memory first, evaluates all rules, then executes mutations as a separate phase (collect-then-act pattern prevents pagination race conditions)
  2. Each run produces a structured log summary showing counts: archived N, deleted M, skipped K, errors E
  3. The application can be invoked for a single run (not just as a daemon) for testing and manual use
**Plans**: TBD

Plans:
- [ ] 07-01: TBD
- [ ] 07-02: TBD

### Phase 8: Scheduler and Deployment
**Goal**: Karaclean runs as a production Docker sidecar on a user-defined cron schedule
**Depends on**: Phase 7
**Requirements**: SCHED-01, SCHED-02, SCHED-03
**Success Criteria** (what must be TRUE):
  1. User defines a cron expression in the YAML config and the application executes rules on that schedule
  2. User defines an explicit timezone in config; if omitted, the application defaults to UTC and logs a startup warning
  3. The container runs as a long-lived daemon, executing rules on schedule and shutting down gracefully on SIGTERM/SIGINT
  4. A working Dockerfile produces a minimal image (scratch base, static binary) and a docker-compose.yml example shows sidecar deployment alongside Karakeep
  5. An example YAML config file documents all available conditions, exceptions, and actions
**Plans**: TBD

Plans:
- [ ] 08-01: TBD
- [ ] 08-02: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7 -> 8

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Config Loading and Validation | 0/0 | Not started | - |
| 2. API Client and Authentication | 0/0 | Not started | - |
| 3. Age and Source Conditions | 0/0 | Not started | - |
| 4. Status and Tag Conditions | 0/0 | Not started | - |
| 5. Exception Evaluation | 0/0 | Not started | - |
| 6. Actions and Dry-Run | 0/0 | Not started | - |
| 7. Run Orchestrator and Observability | 0/0 | Not started | - |
| 8. Scheduler and Deployment | 0/0 | Not started | - |
