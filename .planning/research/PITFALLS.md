# Pitfalls Research

**Domain:** Data-deletion sidecar / rule engine for bookmark cleanup (Karakeep API)
**Researched:** 2026-03-18
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: Irreversible Hard Deletes With No Safety Net

**What goes wrong:**
Karakeep's `DELETE /v1/bookmarks/{id}` is a permanent hard delete -- the row is removed from the database and assets are cleaned up. There is no trash, no soft-delete, no undo endpoint. A bug in rule evaluation, a misconfigured YAML file, or a logic error in condition matching can permanently destroy bookmarks the user intended to keep.

**Why it happens:**
Developers treat delete as "just another API call" and trust that the rule logic is correct. In practice, rule engines have subtle edge cases (see Pitfall 3), and YAML configs are hand-edited by users who make typos. The combination of "user-authored rules" + "irreversible action" is inherently dangerous.

**How to avoid:**
1. **Mandatory dry-run on first run.** If the config has never been seen before (hash check), force a dry-run and log what *would* happen before executing.
2. **Archive-before-delete as the default pattern.** Rules should archive first; only a separate rule (with its own retention period) should delete archived items. Never allow a rule to jump straight from "active" to "deleted" without an explicit user opt-in.
3. **Per-run deletion caps.** If a single run would delete more than N bookmarks (configurable, default 50), halt and log a warning. This catches catastrophic misconfigurations ("oops, I matched everything").
4. **Structured run reports.** Every run produces a structured log (JSON or plaintext) showing exactly what was archived/deleted and why. Users can audit after the fact.

**Warning signs:**
- A rule matches far more bookmarks than expected in dry-run output.
- User reports "all my bookmarks are gone" after first real run.
- No dry-run mode exists or it is not the default for new configs.

**Phase to address:**
Phase 1 (MVP). Dry-run mode, archive-before-delete, and deletion caps must ship in the very first usable version. These are not optional safety features -- they are table stakes for a deletion tool.

---

### Pitfall 2: Mutating Data During Cursor Pagination

**What goes wrong:**
Karaclean must paginate through all bookmarks (default page size: 20, max: 100) using cursor-based pagination keyed on `id` + `createdAt`. If Karaclean deletes or archives bookmarks *while still paginating*, the cursor can become invalid or skip items. Specifically: if a bookmark that appears on a future page gets archived (changing its filter match), the cursor may skip over items that shifted position, or return items that were already processed.

**Why it happens:**
The natural implementation is "fetch page, process page, fetch next page." This interleaves reads and writes, which is a classic pagination race condition. Cursor-based pagination is safer than offset-based, but mutating the dataset that the cursor traverses still causes problems when filter criteria change (e.g., `archived=false` filter -- archiving items changes what subsequent pages return).

**How to avoid:**
1. **Two-phase execution: collect then act.** Phase 1: paginate through ALL bookmarks and collect IDs + metadata into an in-memory list. Phase 2: apply rules to the collected list. Phase 3: execute mutations (archive/delete) against the collected IDs. Never mutate during pagination.
2. **Reasonable memory bounds.** Even with 10,000 bookmarks, storing IDs + minimal metadata (id, createdAt, source, tags, archived, favourited) is well under 10MB. This is fine for a sidecar.
3. **Max page size.** Always request `limit=100` (the API maximum) to minimize the number of round-trips.

**Warning signs:**
- Dry-run reports a different count than the actual run processes.
- Some bookmarks are "missed" and survive multiple cleanup cycles.
- Intermittent off-by-one errors in deletion counts.

**Phase to address:**
Phase 1 (MVP). The collect-then-act pattern must be the architecture from day one. Retrofitting it later means rewriting the core loop.

---

### Pitfall 3: Rule Evaluation Ambiguity When Bookmarks Match Multiple Rules

**What goes wrong:**
A bookmark matches Rule A ("archive RSS items older than 7 days") and Rule B ("never delete favourited items"). What happens? Or: Rule A says "delete" and Rule B says "archive." Which wins? Without a clear conflict resolution strategy, the tool either produces unpredictable results or silently applies the wrong action.

**Why it happens:**
Simple rule engines evaluate rules independently and apply the first match, the last match, or all matches -- and none of these are obviously correct for a deletion tool. The "correct" behavior depends on the user's intent, which is hard to infer.

**How to avoid:**
1. **Exception-first evaluation.** Process exception/protection conditions before action conditions. If any rule says "protect this bookmark" (favourited, has highlights, has notes, is in a list), it is protected regardless of what other rules say. Protection always wins over deletion.
2. **Most-conservative-action-wins.** When multiple action rules match, apply the least destructive action: keep > archive > delete. A bookmark that matches both an "archive" rule and a "delete" rule gets archived, not deleted.
3. **Explicit rule ordering in config.** Rules are evaluated top-to-bottom in the YAML. Document this clearly. But ordering should not affect safety -- the protection-first and conservative-action principles override ordering.
4. **Dry-run output must show which rules matched each bookmark** and which action was selected, so users can debug conflicts.

**Warning signs:**
- Users report "I favourited this but it got deleted anyway."
- Dry-run output shows a bookmark matching multiple rules with no indication of which one "won."
- The rule engine has no concept of rule priority or conflict resolution.

**Phase to address:**
Phase 1 (MVP). The conflict resolution strategy must be designed before the first rule is evaluated. It is extremely hard to change later without breaking existing user configs.

---

### Pitfall 4: Silent YAML Misconfiguration Leading to No-Op or Over-Deletion

**What goes wrong:**
Go's YAML unmarshalling silently ignores misspelled keys and assigns zero values. A user writes `favourited: true` but misspells it as `favorited: true` (American spelling). The field is silently ignored, the condition is never checked, and the rule matches everything instead of just favourited items. Depending on the rule action, this either does nothing (if the misspelled field was the action trigger) or deletes everything (if the misspelled field was a protection condition).

**Why it happens:**
Go's `encoding/yaml` and `gopkg.in/yaml.v3` both silently drop unknown keys by default. This is a well-known footgun in the Go ecosystem. Combined with the fact that YAML has no schema, users get zero feedback when their config is wrong.

**How to avoid:**
1. **Strict unmarshalling with unknown field detection.** Use `mapstructure` with `ErrorUnused: true`, or use `yaml.Decoder` with `KnownFields(true)` (available in `gopkg.in/yaml.v3`). Reject configs with any unrecognized keys.
2. **Validate after parsing.** After unmarshalling, run explicit validation: every rule must have at least one condition and exactly one action. Conditions must reference valid fields. Age durations must parse. Cron expressions must parse.
3. **Provide a `validate` subcommand.** Users can run `karaclean validate --config /path/to/config.yaml` to check their config without running any rules.
4. **Fail loudly on startup.** If config validation fails, the container must exit with a non-zero status and a clear error message. Never silently run with a partially-parsed config.

**Warning signs:**
- Users report "nothing happened" after a run (rules matched nothing due to misspelled conditions).
- Users report "everything got deleted" (protection conditions were silently dropped).
- No config validation exists or it only checks syntax, not semantics.

**Phase to address:**
Phase 1 (MVP). Config validation is a safety-critical feature for a deletion tool. Ship it alongside the rule engine.

---

### Pitfall 5: Cron Timezone and DST Edge Cases

**What goes wrong:**
A user configures `schedule: "0 3 * * *"` expecting it to run at 3am local time. The Docker container runs in UTC. The job runs at 3am UTC, which is 10pm the previous day in US Pacific. Or worse: during a DST transition, a job scheduled at 2:30am either runs twice or never runs at all.

**Why it happens:**
Docker containers default to UTC. Users think in local time. The `robfig/cron` library (the standard Go cron library) defaults to the machine's local timezone, which in Docker is UTC. DST transitions cause the 2am-3am hour to be skipped or repeated, and cron libraries handle this inconsistently.

**How to avoid:**
1. **Require explicit timezone in config.** The YAML config should have a `timezone` field (e.g., `timezone: America/New_York`). Pass this to `robfig/cron` via `cron.WithLocation()`.
2. **Default to UTC and document it.** If no timezone is specified, default to UTC and log a message saying so. Do not silently use the container's local timezone.
3. **Warn about DST-sensitive schedules.** If the configured timezone observes DST and the schedule falls in the 1am-3am window, log a warning at startup.
4. **Use `robfig/cron` v3** which has proper timezone support via `CRON_TZ=` prefix syntax.

**Warning signs:**
- Jobs run at unexpected times after DST transitions.
- Users in non-UTC timezones report the schedule is "off by N hours."
- No timezone configuration exists in the YAML schema.

**Phase to address:**
Phase 1 (MVP). The cron scheduler is a core component. Timezone handling must be correct from the start.

---

### Pitfall 6: N+1 API Calls for Enrichment Data

**What goes wrong:**
The `GET /v1/bookmarks` endpoint returns core bookmark fields, but highlights and list membership require separate API calls per bookmark. For a collection of 5,000 bookmarks, checking highlights requires 5,000 additional API calls. This is extremely slow and may trigger rate limiting or timeouts.

**Why it happens:**
The Karakeep API does not support bulk queries for highlights or list membership. The natural implementation of "for each bookmark, check if it has highlights" creates an N+1 query pattern against the API.

**How to avoid:**
1. **Minimize enrichment calls.** Only fetch highlights/list data for bookmarks that actually need it (i.e., bookmarks that match all other conditions and would be acted upon). Evaluate cheap conditions (age, source, archived, favourited, tags) first, then enrich only the remaining candidates.
2. **Add concurrency control with rate limiting.** Use a semaphore (e.g., 5 concurrent requests) and add delays between batches. Respect any rate-limit headers from the API.
3. **Cache enrichment data within a run.** If multiple rules need the same enrichment data, fetch it once per bookmark per run.
4. **Consider a "fast mode" that skips enrichment-dependent rules** when the collection is very large, with a warning.

**Warning signs:**
- Cleanup runs take 30+ minutes for large collections.
- Karakeep API returns 429 (rate limit) or 503 (overload) errors during runs.
- CPU/memory usage on the Karakeep server spikes during cleanup runs.

**Phase to address:**
Phase 2 (after MVP). The MVP can ship with basic enrichment, but optimization and rate limiting should come early. This is especially important for the "count per feed" rule condition, which requires aggregating across many bookmarks.

---

### Pitfall 7: API Token Expiry and Auth Failures Mid-Run

**What goes wrong:**
Karaclean authenticates via a Bearer token. If the token expires or is revoked mid-run (e.g., user regenerates their API key in Karakeep), the sidecar silently fails to delete some bookmarks, or worse, crashes partway through a batch leaving the collection in an inconsistent state (some items deleted, some not).

**Why it happens:**
Long-running cleanup cycles (especially with enrichment) can take minutes. Token validity is rarely checked proactively. Error handling for 401 responses is often an afterthought.

**How to avoid:**
1. **Check auth on startup.** Make a lightweight API call (e.g., `GET /v1/bookmarks?limit=1`) before starting the rule evaluation loop. If it returns 401, exit immediately with a clear error.
2. **Handle 401 mid-run gracefully.** If any API call returns 401, stop the entire run immediately. Do not continue with partial results. Log clearly that auth failed.
3. **Treat auth failures as non-retriable.** Do not retry 401 errors with backoff -- the token is invalid and retrying will not fix it.
4. **Document token lifecycle.** Make it clear in docs that if the user regenerates their API key, they must update the sidecar's config/env.

**Warning signs:**
- Partial cleanup runs (some bookmarks deleted, others not) without clear error logs.
- Silent failures in the sidecar logs.
- No auth check at startup.

**Phase to address:**
Phase 1 (MVP). Auth validation and error handling are foundational.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Load all bookmarks into memory at once | Simple implementation | OOM for very large collections (50k+) | MVP -- most users have <10k bookmarks. Add streaming later if needed. |
| No persistent state between runs | No database needed, simpler sidecar | Cannot track "already processed" bookmarks, re-evaluates everything each run | Always acceptable for this use case -- rules are idempotent by design. |
| Single-threaded rule evaluation | No concurrency bugs | Slow for very large collections | MVP -- optimize only if users report speed issues. |
| Hardcoded API version (`/api/v1`) | No version negotiation needed | Breaks if Karakeep changes API version | Acceptable until Karakeep ships v2. Track upstream releases. |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Karakeep `GET /v1/bookmarks` | Assuming all bookmark data is in the list response | Highlights and list membership require separate `GET /v1/bookmarks/{id}/highlights` and `GET /v1/lists` calls. Plan for this from the start. |
| Karakeep cursor pagination | Using offset-based mental model with cursor-based API | Cursor is `{id, createdAt}` based. Always drain all pages before mutating. Never assume page count is stable. |
| Karakeep `PATCH /v1/bookmarks/{id}` | Sending the full bookmark object | Only send the fields you want to change (e.g., `{"archived": true}`). Sending extra fields may overwrite user data. |
| Docker Compose networking | Hardcoding `localhost` for Karakeep URL | Use Docker service names (e.g., `http://karakeep:3000`). Document that the sidecar must be on the same Docker network. |
| YAML config file mounting | Expecting hot-reload of config changes | Config is read at startup. Document that container restart is required after config changes. (Hot-reload is a Phase 2+ feature.) |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Fetching highlights for every bookmark | Run takes 30+ minutes, Karakeep becomes slow | Lazy enrichment: only fetch highlights for bookmarks that pass all other conditions | >500 bookmarks needing enrichment |
| Requesting page size of 20 (the default) | 500 API calls to paginate 10,000 bookmarks | Always use `limit=100` (API max) | >1,000 bookmarks total |
| No request concurrency control | Karakeep API overwhelmed, 429/503 errors | Semaphore with max 5 concurrent requests, exponential backoff on errors | >2,000 bookmarks with enrichment |
| Re-evaluating all rules against all bookmarks | O(rules * bookmarks) per run | Index bookmarks by relevant fields in memory, short-circuit on first disqualifying condition | >10,000 bookmarks with >10 rules |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Logging the API bearer token | Token exposed in container logs, accessible to anyone with Docker log access | Never log the token value. Log only that auth succeeded/failed. Mask the token in any debug output. |
| Storing the token in the YAML config file | Config file is often committed to version control | Accept the token via environment variable (`KARAKEEP_API_TOKEN`), not in the YAML file. Support both, but recommend env var in docs. |
| No TLS verification for Karakeep connection | MITM attack could intercept token and bookmark data | Default to verifying TLS. Allow `--insecure` flag for local/dev setups only, with a warning. |
| Running the container as root | Container escape could compromise host | Use a non-root user in the Dockerfile. The sidecar only needs network access and config file read access. |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| No feedback on what was cleaned up | User has no idea if Karaclean is working or doing nothing | Emit a structured summary after each run: "Archived: 12, Deleted: 3, Skipped: 847, Errors: 0" |
| Dry-run output is too verbose (lists every bookmark) | User cannot find the signal in the noise | Group dry-run output by rule: "Rule 'clean-old-rss': would archive 45 bookmarks, would delete 12 bookmarks" |
| Error messages reference internal Go types | User sees `*yaml.TypeError` and has no idea what to fix | Wrap all config errors with user-friendly messages: "Line 12: unknown field 'favorited' -- did you mean 'favourited'?" |
| No indication of next scheduled run | User doesn't know when cleanup will happen next | Log the next scheduled run time at startup and after each run completes |
| Silent no-op when API is unreachable | User thinks cleanup is happening, but Karakeep is down | Log a clear error when Karakeep is unreachable. Exit with non-zero status if the first health check fails. |

## "Looks Done But Isn't" Checklist

- [ ] **Dry-run mode:** Verify it actually prevents ALL mutations, not just deletes (also must skip archive calls).
- [ ] **Cursor pagination:** Verify the last page is handled correctly (no off-by-one: the API returns `nextCursor: null` when done).
- [ ] **Config validation:** Verify unknown fields are rejected, not just that known fields parse correctly.
- [ ] **Rule exceptions:** Verify that exception conditions (e.g., "unless favourited") work with AND/OR semantics as documented.
- [ ] **Empty collection handling:** Verify the tool handles zero bookmarks gracefully (no division by zero in stats, no misleading "deleted 0" logs).
- [ ] **Large collection handling:** Test with 5,000+ bookmarks to catch pagination and memory issues.
- [ ] **Container restart:** Verify the container restarts cleanly after a crash (no stale state, no PID files, no lock files).
- [ ] **Concurrent runs:** Verify that if the cron schedule triggers faster than a run completes, the tool does not start overlapping runs.

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Accidental mass deletion | HIGH -- data is permanently gone | No recovery from Karaclean. User must restore from Karakeep backup (if they have one) or re-import bookmarks. This is why prevention (dry-run, caps, archive-first) is critical. |
| Config typo causing no-op | LOW | Fix YAML, restart container. No data lost. |
| Config typo causing over-matching | HIGH if action is delete, MEDIUM if archive | For archives: bulk un-archive via API. For deletes: no recovery (see above). |
| Token expiry mid-run | LOW | Update token in env/config, restart container. Partial runs leave data in a safe state (some items archived/deleted, rest untouched). |
| Timezone misconfiguration | LOW | Fix timezone in config, restart. No data risk, just timing. |
| API pagination race condition | MEDIUM | If using collect-then-act pattern: no issue. If not: some bookmarks may be missed or processed twice. Re-run will catch missed items. |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Irreversible hard deletes | Phase 1 (MVP) | Dry-run mode works, archive-before-delete is the default, deletion caps are configurable |
| Pagination mutation race | Phase 1 (MVP) | Integration test: delete during pagination produces correct results |
| Rule conflict resolution | Phase 1 (MVP) | Unit tests: bookmark matching multiple rules gets most-conservative action |
| Silent YAML misconfiguration | Phase 1 (MVP) | Config with unknown fields is rejected; `validate` subcommand exists |
| Cron timezone edge cases | Phase 1 (MVP) | Config requires explicit timezone; DST-sensitive schedules produce warnings |
| N+1 API enrichment | Phase 2 (optimization) | Lazy enrichment implemented; benchmark with 5,000 bookmarks |
| Auth failures mid-run | Phase 1 (MVP) | 401 response stops the run; startup auth check exists |
| Overlapping cron runs | Phase 2 | Mutex or skip-if-running logic prevents concurrent runs |

## Sources

- Karakeep source code: `karakeep-upstream/packages/trpc/models/bookmarks.ts` -- confirmed hard delete, cursor pagination on `{id, createdAt}`, page size default 20 / max 100
- Karakeep source code: `karakeep-upstream/packages/trpc/routers/bookmarks.ts` -- confirmed rate limiting exists on bookmark creation, no rate limiting on reads/deletes
- [Go YAML silent zero-value bug](https://buildsoftwaresystems.com/post/go-config-yaml-safer-mapstructure-fix/) -- mapstructure with `ErrorUnused: true` as mitigation
- [Handling Timezone Issues in Cron Jobs (2025 Guide)](https://dev.to/cronmonitor/handling-timezone-issues-in-cron-jobs-2025-guide-52ii) -- DST skip/double-fire behavior
- [robfig/cron](https://github.com/robfig/cron) -- Go cron library timezone support via `cron.WithLocation()`
- [REST API Pagination and Race Conditions](https://www.samanvayfoundation.org/articles/on-software-architecture/rest-api-pagination-and-race-condition/) -- mutation-during-pagination patterns
- [Slack Engineering: Evolving API Pagination](https://slack.engineering/evolving-api-pagination-at-slack/) -- cursor-based pagination best practices
- [Martin Fowler: Rules Engine](https://martinfowler.com/bliki/RulesEngine.html) -- rule engine complexity and conflict resolution
- [FlexRule: Rule Engine Inference](https://www.flexrule.com/archives/rule-engine-inference/) -- rule priority and evaluation order patterns

---
*Pitfalls research for: Karaclean -- data-deletion sidecar for Karakeep*
*Researched: 2026-03-18*
