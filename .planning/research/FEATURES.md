# Feature Research

**Domain:** Bookmark garbage collection / content retention automation (sidecar for Karakeep)
**Researched:** 2026-03-18
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete or unsafe.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Age-based conditions (`olderThan`) | Every retention tool (FreshRSS, Miniflux, email rules) uses age as primary filter. Foundational. | LOW | Compare `createdAt` against duration. Use Go `time.Duration` or days integer. |
| Source-based conditions (`source`) | RSS is the high-volume source. Users need to target RSS bookmarks specifically without touching web/extension saves. | LOW | Enum match against `source` field: rss, web, api, mobile, extension, cli, import, singlefile. |
| Archived/favourited status conditions | Archive-then-delete is the core workflow per PROJECT.md. Must filter on these. | LOW | Direct boolean fields on bookmark object. API also supports server-side filtering. |
| Tag-based conditions (`hasTag`, `lacksTag`) | Tags are Karakeep's primary organizational primitive. Users tag things "keep" or "ephemeral". | LOW | Tags array is embedded in bookmark list response. String matching. |
| Exception conditions (unless clauses) | Email filters, FreshRSS, Miniflux all protect favourites/starred from deletion. Without exceptions, users will lose data they care about. | MEDIUM | Implement as nested conditions with NOT semantics. E.g., `unless: [favourited: true]`. |
| Archive action | Non-destructive first step. Mirrors Karakeep's native archive. | LOW | `PATCH /v1/bookmarks/{id}` with `{ "archived": true }`. |
| Delete action | The actual garbage collection. Permanent. | LOW | `DELETE /v1/bookmarks/{id}`. Returns 204. |
| Two-phase archive-then-delete | Core value proposition per PROJECT.md. Archive first, delete archived items after grace period. | MEDIUM | Two separate rules: one archives unarchived items matching criteria, one deletes archived items older than retention period. Could also be a single rule with `phases`. |
| Dry-run mode | Standard safety feature for any destructive automation (Terraform plan, Ansible --check, git clean -n, GCP cleanup policies). Without this, users won't trust the tool. | LOW | Skip actual API calls, log what would happen. Flag in config or CLI arg. |
| Cron scheduling | Users expect set-and-forget. Manual runs are for testing only. | LOW | Go cron library (robfig/cron). Schedule string in YAML config. |
| YAML config file | Declared in PROJECT.md as the interface. Power users expect declarative config. | MEDIUM | Define schema, validate on startup, fail fast on bad config. |
| Structured logging | Users running Docker sidecars rely on `docker logs` for observability. Must know what happened. | LOW | JSON-structured logs with timestamp, rule name, action, bookmark ID, result. |

### Differentiators (Competitive Advantage)

Features that set Karaclean apart. Not expected in a v1 sidecar, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| RSS feed-scoped rules (`rssFeedId`) | Karakeep API supports `rssFeedId` filter on list endpoint. Lets users set different retention per feed (keep tech blogs 90 days, delete news feeds after 7 days). No existing tool does per-feed retention as a sidecar. | MEDIUM | Requires fetching feed list via `GET /v1/feeds` to resolve feed names to IDs. Rules reference feeds by name, Karaclean resolves to ID. |
| Note/highlight presence as protection signals | Users who annotate bookmarks clearly value them. "Has highlights" or "has note" as exception conditions prevents accidental deletion of engaged-with content. | HIGH | Highlights require per-bookmark `GET /v1/bookmarks/{id}/highlights` call. Expensive at scale. Note field is on bookmark object (cheap). Consider highlights as opt-in condition. |
| List membership conditions | Bookmarks in curated lists are intentionally organized. "In list X" or "in any list" as exception conditions. | HIGH | Lists require per-bookmark `GET /v1/bookmarks/{id}/lists` call. Same N+1 cost as highlights. Consider as opt-in. |
| Count-based retention (`keepNewest: N`) | "Keep the 50 newest RSS bookmarks per feed, delete the rest." Common in RSS readers (FreshRSS max articles per feed). | MEDIUM | Requires sorting by createdAt, counting per group, marking excess for action. |
| Bookmark type conditions (`type: link/text/asset`) | Different retention for links vs text notes vs uploaded assets. Text notes are often personal; links from RSS are ephemeral. | LOW | Available on `content.type` field in bookmark response. |
| Run summary / report | After each cron run, emit a summary: "Archived 42, deleted 17, skipped 3 (protected)". Helps users tune rules. | LOW | Aggregate counters during run, log summary at end. |
| Config validation with `--validate` flag | Let users test config syntax without running cleanup. Fast feedback loop. | LOW | Parse YAML, validate schema, print errors or "OK", exit. |
| AND/OR logical combinators for conditions | Email filters support "all conditions" vs "any condition". Complex rules need boolean logic. | MEDIUM | Default to AND (all conditions must match). Support explicit `matchAll` / `matchAny` grouping. |
| Rule ordering and priority | Email filters process rules in order. First-match-wins or all-match semantics. Users need control over which rule "wins" for a bookmark. | MEDIUM | Process rules in YAML order. Consider `stopAfterMatch: true` per rule for first-match semantics. |
| Tag-based actions (add tag before archive/delete) | Tag bookmarks with "karaclean:archived" before archiving. Provides audit trail visible in Karakeep UI. | LOW | `POST /v1/bookmarks/{id}/tags` with tag name. Tag auto-created if missing. |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems for this project.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Web UI for rule management | Visual config is easier than YAML for non-technical users. | Massive scope increase (frontend framework, auth, state management). Violates sidecar simplicity. PROJECT.md explicitly defers this. | Ship YAML-only. Add `--validate` flag. Consider UI as separate future project if demand exists. |
| Reading progress conditions (`readingProgressPercent`) | "Delete unread RSS items after 30 days" is appealing. | Reading progress is NOT exposed via REST API -- it's tRPC-only. Would require reverse-engineering internal endpoints or direct DB access, both of which violate the HTTP API constraint. | Use "has highlights" or "has note" as proxy signals for engagement. Or wait for Karakeep to expose this via REST. |
| Undo / restore deleted bookmarks | Safety net for accidental deletion. | Karakeep DELETE is permanent (204, no trash). Implementing undo would require Karaclean to maintain its own backup storage, which is a different product. | Rely on archive-then-delete two-phase pattern as the safety net. Dry-run mode catches config errors before they cause damage. |
| Multi-user support | Shared Karakeep instances. | One API key = one user context. Multi-user means multiple configs, multiple schedules, auth management. Complexity explosion for v1. | Run one Karaclean container per user. Compose supports multiple sidecar instances trivially. |
| Real-time webhook triggers | React to new bookmarks immediately instead of waiting for cron. | Karakeep has no webhook/event system. Would require polling at high frequency, which is wasteful and no better than cron for garbage collection (batch is fine). | Cron at reasonable intervals (hourly/daily). Garbage collection is not time-sensitive. |
| Content-based rules (regex on title/body) | "Delete bookmarks with 'sponsored' in the title." | Requires fetching full content (`includeContent: true`) for every bookmark on every run. Massive API load. Also fragile -- content changes, encoding issues, false positives. | Use Karakeep's built-in AI tagging to auto-tag content, then filter on tags in Karaclean. Leverage the platform, don't replicate it. |
| Direct database access | Faster than API, can do complex queries. | PROJECT.md explicitly forbids this. Couples to Karakeep internals, breaks on upgrades, bypasses Karakeep's auth and business logic. | HTTP API only. Accept the pagination overhead. |
| Push notifications / alerting | "Email me when bookmarks are deleted." | Out of scope per PROJECT.md. Notification systems are their own infrastructure (SMTP, webhooks, Slack). | Structured logs. Users can use log monitoring tools (Loki, etc.) on Docker logs. |

## Feature Dependencies

```
[YAML Config Parsing]
    |
    +--requires--> [Rule Condition Engine]
    |                  |
    |                  +--requires--> [Karakeep API Client]
    |                  |                  |
    |                  |                  +--provides--> [Archive Action]
    |                  |                  +--provides--> [Delete Action]
    |                  |                  +--provides--> [Tag Action]
    |                  |
    |                  +--enhances--> [Exception Conditions]
    |                  +--enhances--> [AND/OR Combinators]
    |
    +--requires--> [Cron Scheduler]
    |
    +--enhances--> [Dry-Run Mode]
    +--enhances--> [Structured Logging]
    +--enhances--> [Run Summary]

[Feed-Scoped Rules]
    +--requires--> [Karakeep API Client] (feeds endpoint)
    +--requires--> [Rule Condition Engine]

[Highlight/List Conditions]
    +--requires--> [Karakeep API Client] (per-bookmark sub-endpoints)
    +--requires--> [Rule Condition Engine]
    * WARNING: N+1 API calls -- expensive at scale
```

### Dependency Notes

- **Rule Condition Engine requires Karakeep API Client:** Conditions evaluate bookmark properties fetched from the API.
- **Exception Conditions enhance Rule Condition Engine:** Exceptions are just negated conditions applied after primary matching.
- **Dry-Run Mode enhances everything:** Cross-cutting concern. When enabled, actions log intent but skip API mutation calls.
- **Feed-Scoped Rules require feeds endpoint:** Must call `GET /v1/feeds` to resolve feed names to IDs for the `rssFeedId` filter on the bookmarks endpoint.
- **Highlight/List Conditions are expensive:** Each requires a separate API call per bookmark. Must be opt-in per rule to avoid hammering the API. These should be exception conditions ("unless has highlights") rather than primary filters.

## MVP Definition

### Launch With (v1)

Minimum viable product -- what's needed to validate the concept.

- [ ] **YAML config parsing with validation** -- Foundation. Fail fast on bad config.
- [ ] **Age-based condition** (`olderThan: 30d`) -- Primary filter for garbage collection.
- [ ] **Source condition** (`source: rss`) -- Target high-volume RSS bookmarks.
- [ ] **Archived/favourited status conditions** -- Enable two-phase workflow.
- [ ] **Tag-based conditions** (`hasTag`, `lacksTag`) -- Users' primary organizational signal.
- [ ] **Exception conditions** (`unless`) -- Safety. "Delete old RSS, unless favourited or tagged 'keep'."
- [ ] **Archive action** -- Non-destructive first phase.
- [ ] **Delete action** -- Actual cleanup.
- [ ] **Dry-run mode** -- Safety. Must exist before any destructive action ships.
- [ ] **Cron scheduling** -- Set-and-forget operation.
- [ ] **Structured JSON logging** -- Observability via `docker logs`.
- [ ] **Run summary** -- "Archived 42, deleted 17" after each run.

### Add After Validation (v1.x)

Features to add once core is working and users provide feedback.

- [ ] **RSS feed-scoped rules** -- Add when users request per-feed retention policies. Requires feed name resolution.
- [ ] **AND/OR logical combinators** -- Add when users need complex multi-condition rules beyond simple AND.
- [ ] **Count-based retention** (`keepNewest: N`) -- Add when users want to cap bookmarks per feed rather than age-based cleanup.
- [ ] **Bookmark type conditions** (`type: link`) -- Add when users want different policies for links vs text vs assets.
- [ ] **Config validation CLI flag** (`--validate`) -- Add for better DX once config schema stabilizes.
- [ ] **Tag-based actions** (add tag before archive/delete) -- Add for audit trail when users request it.
- [ ] **Rule priority / stop-after-match** -- Add when users report rules conflicting.
- [ ] **Note presence as condition** (`hasNote: true`) -- Cheap to check (field on bookmark object). Add as engagement signal.

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] **Highlight presence conditions** -- Requires per-bookmark API call. Defer until API cost is understood at scale.
- [ ] **List membership conditions** -- Same N+1 cost as highlights. Defer.
- [ ] **Web UI** -- Only if YAML proves too friction-heavy for the target audience.
- [ ] **Reading progress conditions** -- Blocked by Karakeep REST API (tRPC-only). Revisit when/if Karakeep exposes this.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| YAML config + validation | HIGH | MEDIUM | P1 |
| Age condition (`olderThan`) | HIGH | LOW | P1 |
| Source condition | HIGH | LOW | P1 |
| Archived/favourited conditions | HIGH | LOW | P1 |
| Tag conditions | HIGH | LOW | P1 |
| Exception conditions (`unless`) | HIGH | MEDIUM | P1 |
| Archive action | HIGH | LOW | P1 |
| Delete action | HIGH | LOW | P1 |
| Dry-run mode | HIGH | LOW | P1 |
| Cron scheduling | HIGH | LOW | P1 |
| Structured logging | HIGH | LOW | P1 |
| Run summary | MEDIUM | LOW | P1 |
| Feed-scoped rules | HIGH | MEDIUM | P2 |
| AND/OR combinators | MEDIUM | MEDIUM | P2 |
| Count-based retention | MEDIUM | MEDIUM | P2 |
| Bookmark type conditions | MEDIUM | LOW | P2 |
| Config validate flag | MEDIUM | LOW | P2 |
| Tag-based actions | LOW | LOW | P2 |
| Note presence condition | MEDIUM | LOW | P2 |
| Rule priority / stop-after-match | LOW | MEDIUM | P2 |
| Highlight conditions | MEDIUM | HIGH | P3 |
| List membership conditions | MEDIUM | HIGH | P3 |
| Web UI | MEDIUM | HIGH | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Competitor Feature Analysis

No direct competitor exists for "Karakeep sidecar garbage collector." The closest analogues are retention features in RSS readers and email filter systems.

| Feature | FreshRSS | Miniflux | Email Filters (Gmail/Outlook) | Karaclean Approach |
|---------|----------|----------|-------------------------------|-------------------|
| Age-based cleanup | Global + per-feed max age | Global `CLEANUP_ARCHIVE_READ_DAYS` (60d default) | Age-based rules available | Per-rule `olderThan` with any duration |
| Per-source retention | Per-feed override of global policy | No per-feed policy (requested in issue #770) | Per-sender rules | Per-rule `source` filter + future `rssFeedId` |
| Protect favourites | "Never delete favourites" global toggle | Starred entries never deleted | Star/flag prevents auto-delete | `unless: [favourited: true]` per rule |
| Protect annotated | "Never delete with labels" toggle | N/A | N/A | `unless: [hasNote: true]` or `hasHighlights: true` per rule |
| Count-based limit | Max articles per feed | N/A | N/A | Future `keepNewest: N` per rule |
| Two-phase (archive then delete) | Read -> purge pipeline | Read entries archived, then deleted | Archive -> delete after N days | First-class two-rule pattern |
| Dry-run | No | No | No | `dryRun: true` in config or `--dry-run` CLI flag |
| Scheduling | Built-in (part of app) | Built-in `CLEANUP_FREQUENCY_HOURS` | Continuous (always-on) | Cron expression in YAML |
| Rule exceptions | Implicit (never delete favourites/labels) | Implicit (never delete starred) | Exception conditions in rules | Explicit `unless` clause per rule |

## API Constraints Affecting Features

Critical constraints discovered from Karakeep source code analysis:

| Constraint | Impact | Mitigation |
|------------|--------|------------|
| `readingProgressPercent` is tRPC-only, not in REST API | Cannot use reading progress as a condition | Use note/highlight presence as engagement proxy |
| Highlights require per-bookmark GET (`/bookmarks/{id}/highlights`) | N+1 calls for highlight-based conditions; expensive at scale | Make highlight conditions opt-in; only fetch when rule explicitly uses them |
| List membership requires per-bookmark GET (`/bookmarks/{id}/lists`) | Same N+1 problem as highlights | Same mitigation: opt-in only |
| List endpoint max 100 bookmarks per page (`MAX_NUM_BOOKMARKS_PER_PAGE`) | Must paginate through all bookmarks; high-volume users may have thousands | Efficient cursor-based pagination; apply server-side filters (`archived`, `favourited`, `rssFeedId`) to reduce page count |
| Tags are embedded in bookmark list response | Tag-based conditions are cheap (no extra API calls) | Prefer tag-based conditions over highlight/list-based ones |
| `rssFeedId` filter on list endpoint | Can efficiently fetch only bookmarks from a specific RSS feed | Use for feed-scoped rules to avoid fetching all bookmarks |
| Delete returns 204 (no body, permanent) | No undo possible via API | Archive-then-delete pattern is the safety net |

## Sources

- Karakeep OpenAPI source: `karakeep-upstream/packages/open-api/lib/bookmarks.ts` (HIGH confidence, primary source)
- Karakeep shared types: `karakeep-upstream/packages/shared/types/bookmarks.ts` (HIGH confidence, primary source)
- Karakeep tRPC routers: `karakeep-upstream/packages/trpc/routers/bookmarks.ts` (HIGH confidence, confirms readingProgress is tRPC-only)
- [FreshRSS Configuration docs](https://freshrss.github.io/FreshRSS/en/users/05_Configuration.html) (MEDIUM confidence)
- [Miniflux Configuration Parameters](https://miniflux.app/docs/configuration.html) (MEDIUM confidence)
- [Miniflux per-feed retention policy issue #770](https://github.com/miniflux/v2/issues/770) (MEDIUM confidence)
- [FreshRSS per-feed/category purge policy issue #6601](https://github.com/FreshRSS/FreshRSS/issues/6601) (MEDIUM confidence)
- [Rules Engine Pattern - DevIQ](https://deviq.com/design-patterns/rules-engine-pattern/) (MEDIUM confidence)
- [Dry-Run Engineering - DEV Community](https://dev.to/danieljglover/dry-run-engineering-the-simple-practice-that-prevents-production-disasters-ek0) (MEDIUM confidence)

---
*Feature research for: Bookmark garbage collection sidecar (Karaclean)*
*Researched: 2026-03-18*
