# Requirements: Karaclean

**Defined:** 2026-03-18
**Core Value:** Users can define flexible, declarative cleanup rules that keep their Karakeep instance lean without ever touching bookmarks they care about.

## v1 Requirements

### Config (CONF)

- [x] **CONF-01**: User can define rules in a YAML config file mounted into the container
- [x] **CONF-02**: Config validation rejects unknown fields at startup (strict YAML parsing — prevents silent misconfiguration from misspelled keys)
- [ ] **CONF-03**: Container validates API token against Karakeep on startup before executing any rules

### Conditions (COND)

- [x] **COND-01**: Rules can match bookmarks older than N days (`olderThan` condition)
- [x] **COND-02**: Rules can filter by source: `rss`, `web`, `api`, `mobile`, `extension`, `cli`, `import`
- [x] **COND-03**: Rules can match on archived status (`archived: true/false`)
- [x] **COND-04**: Rules can match on favourited status (`favourited: true/false`)
- [x] **COND-05**: Rules can match bookmarks that have a specific tag (`hasTag`)
- [x] **COND-06**: Rules can match bookmarks that lack a specific tag (`lacksTag`)

### Exceptions (EXCP)

- [x] **EXCP-01**: Rules support `unless favourited` — skip bookmark if starred
- [x] **EXCP-02**: Rules support `unless hasTag` — skip bookmark if it has a specific tag
- [x] **EXCP-03**: Rules support `unless hasNote` — skip if user has added a personal note
- [x] **EXCP-04**: Rules support `unless archived` / `unless notArchived` exception clause

### Actions (ACTN)

- [x] **ACTN-01**: Rules can archive bookmarks (`archived: true` via Karakeep PATCH API)
- [x] **ACTN-02**: Rules can permanently delete bookmarks (Karakeep DELETE API)
- [x] **ACTN-03**: Dry-run mode logs all intended actions without executing any mutations

### Scheduling (SCHED)

- [x] **SCHED-01**: User defines a cron schedule expression in YAML config
- [x] **SCHED-02**: User defines explicit timezone in config (defaults to UTC with a startup warning if unset)
- [x] **SCHED-03**: Container runs as a daemon executing rules on the defined schedule

### Observability (OBS)

- [x] **OBS-01**: Each run produces a structured log summary (archived: N, deleted: M, skipped: K, errors: E)

## v2 Requirements

### Safety

- **SAFE-01**: Per-run deletion cap — halt if a single run would delete more than N bookmarks (configurable)

### Rule Engine Enhancements

- **RULE-01**: RSS feed-scoped rules — target specific feeds by ID or name for per-feed retention policies
- **RULE-02**: AND/OR logical combinators — compose complex multi-condition rules
- **RULE-03**: Count-based retention (`keepNewest: N`) — keep only the N most recent bookmarks per feed
- **RULE-04**: Bookmark type conditions (`type: link/text/asset`)
- **RULE-05**: Note presence condition (`hasNote: true`) — promoted to condition (currently exception-only)
- **RULE-06**: Tag-based actions (add/remove tag before archive/delete for audit trail)
- **RULE-07**: Rule priority and `stopAfterMatch` semantics

### Tooling

- **TOOL-01**: `--validate` CLI flag — validate config file without executing rules

## Out of Scope

| Feature | Reason |
|---------|--------|
| Web UI | YAML config first; UI is additive, not v1 |
| Multi-user support | Single API key per container; additive if needed |
| Direct database access | HTTP API only to stay decoupled from Karakeep internals |
| Reading progress conditions | tRPC-only in Karakeep — not accessible via REST API |
| Highlight presence conditions | Requires per-bookmark API call (N+1 cost); defer to v2+ |
| List membership conditions | Same N+1 cost as highlights; defer to v2+ |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| CONF-01 | Phase 1: Config Loading and Validation | Complete |
| CONF-02 | Phase 1: Config Loading and Validation | Complete |
| CONF-03 | Phase 2: API Client and Authentication | Pending |
| COND-01 | Phase 3: Age and Source Conditions | Complete |
| COND-02 | Phase 3: Age and Source Conditions | Complete |
| COND-03 | Phase 4: Status and Tag Conditions | Complete |
| COND-04 | Phase 4: Status and Tag Conditions | Complete |
| COND-05 | Phase 4: Status and Tag Conditions | Complete |
| COND-06 | Phase 4: Status and Tag Conditions | Complete |
| EXCP-01 | Phase 5: Exception Evaluation | Complete |
| EXCP-02 | Phase 5: Exception Evaluation | Complete |
| EXCP-03 | Phase 5: Exception Evaluation | Complete |
| EXCP-04 | Phase 5: Exception Evaluation | Complete |
| ACTN-01 | Phase 6: Actions and Dry-Run | Complete |
| ACTN-02 | Phase 6: Actions and Dry-Run | Complete |
| ACTN-03 | Phase 6: Actions and Dry-Run | Complete |
| OBS-01  | Phase 7: Run Orchestrator and Observability | Complete |
| SCHED-01 | Phase 8: Scheduler and Deployment | Complete |
| SCHED-02 | Phase 8: Scheduler and Deployment | Complete |
| SCHED-03 | Phase 8: Scheduler and Deployment | Complete |

**Coverage:**
- v1 requirements: 20 total
- Mapped to phases: 20
- Unmapped: 0

---
*Requirements defined: 2026-03-18*
*Last updated: 2026-03-18 after roadmap creation*
