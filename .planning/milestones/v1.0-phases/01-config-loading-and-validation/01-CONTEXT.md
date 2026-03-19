# Phase 1: Config Loading and Validation - Context

**Gathered:** 2026-03-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Define the complete v1 YAML config schema, parse it into typed Go structs, and validate it with clear, actionable error messages at startup. This phase establishes the config format users will write for all subsequent phases. Rule execution is out of scope — later phases add that on top of this schema.

</domain>

<decisions>
## Implementation Decisions

### Secrets and connection details
- API token and server URL are NOT in the YAML config file — they are passed as env vars
- Env var names: `KARACLEAN_API_URL` and `KARACLEAN_API_TOKEN`
- Namespaced to the tool to avoid collision with Karakeep's own env vars in a compose stack

### Top-level config structure
- YAML file contains: `timezone`, `schedule`, `rules`
- No connection details in the file — those are env vars (see above)

### Rule anatomy
- Each rule has: `name` (optional label for logs), `conditions`, `unless`, `action`
- Example:
  ```yaml
  rules:
    - name: old-rss-cleanup
      conditions:
        olderThan: 30
        source: rss
      unless:
        favourited: true
      action: archive
  ```

### Condition semantics
- Multiple conditions within a rule combine with AND — all must match for the rule to fire
- Multiple unless exceptions combine with OR — any exception match protects the bookmark
- This is intentional asymmetry: be specific about what you target, be protective about what you spare
- AND/OR logical combinators within a single rule are deferred to v2 (RULE-02)
- OR across rules is achieved by writing multiple rules

### Schema completeness
- Phase 1 defines the FULL v1 schema, even though execution comes in later phases
- All v1 condition fields: `olderThan` (days, int), `source` (enum), `archived` (bool), `favourited` (bool), `hasTag` (string), `lacksTag` (string)
- All v1 exception fields (under `unless`): `favourited` (bool), `hasTag` (string), `hasNote` (bool), `archived` (bool)
- Both v1 actions: `archive`, `delete` (string enum)
- Valid `source` values: `rss`, `web`, `api`, `mobile`, `extension`, `cli`, `import`
- Unknown fields in the YAML cause a validation error (strict parsing — CONF-02)

### Config file discovery
- Precedence: `--config` flag > `KARACLEAN_CONFIG` env var > `/config/karaclean.yaml` (default)
- Compose usage: mount config at default path, no flag needed
- Override: pass env var or flag for non-default paths

### Error presentation
- All validation errors are collected and reported at once — no fix-one-at-a-time
- Errors include field paths matching YAML structure: `rules[0].action`, `rules[1].conditions.source`
- Example output:
  ```
  config validation failed:
    - rules[0].action: invalid value "remove" (must be archive or delete)
    - rules[1].conditions.source: invalid value "feed" (must be rss, web, api, mobile, extension, cli, import)
    - rules[2]: missing required field "action"
  ```

### Claude's Discretion
- YAML parsing library choice (gopkg.in/yaml.v3 is standard; go.yaml.in/yaml/v4 if stable by dev time — STATE.md notes this)
- Whether `name` is required or optional on rules
- Exact Go struct layout and package organization within the `internal/config/` package
- Whether to use a custom validator or implement validation inline

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` — CONF-01, CONF-02 define config parsing and strict validation requirements; traceability table shows Phase 1 scope
- `.planning/PROJECT.md` — Tech stack constraints (Go, YAML, Docker), key decisions table, out-of-scope items

### Karakeep API (for field definitions)
- `karakeep-upstream/` — Karakeep source submodule; contains OpenAPI spec for reference on bookmark field names and source enum values (do not modify)

No external ADRs or design docs yet — requirements are fully captured in decisions above.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- None — greenfield project. No Go code exists yet.

### Established Patterns
- None yet — this phase establishes the foundation. Subsequent phases will follow patterns set here.

### Integration Points
- `cmd/karaclean/` — main entry point (to be created); reads config on startup
- `internal/config/` — config types and validation logic (to be created)
- Phase 2 will consume the parsed config struct to initialize the API client

</code_context>

<deferred>
## Deferred Ideas

- AND/OR logical combinators within a single rule — planned as v2 RULE-02
- `--validate` flag (validate config without running) — planned as v2 TOOL-01

</deferred>

---

*Phase: 01-config-loading-and-validation*
*Context gathered: 2026-03-18*
