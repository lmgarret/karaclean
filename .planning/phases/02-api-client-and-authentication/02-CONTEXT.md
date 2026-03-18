# Phase 2: API Client and Authentication - Context

**Gathered:** 2026-03-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Build the HTTP client that communicates with Karakeep and validate the API bearer token on startup before executing any rules. This phase delivers the communication layer only — no rule evaluation, no actions. Outputs: typed Go client in `internal/karakeep/`, `KarakeepAPI` interface in `internal/engine/`, and startup auth check integrated into `cmd/karaclean/main.go`.

</domain>

<decisions>
## Implementation Decisions

### Code generation strategy
- Use oapi-codegen to generate the Karakeep HTTP client from the OpenAPI spec at `karakeep-upstream/packages/open-api/karakeep-openapi-spec.json`
- Generate client interface + concrete HTTP implementation (not typed functions only) — enables mock-based testing by passing the interface to the engine
- Generated code lives in `internal/karakeep/`

### URL and token sourcing
- Karakeep base URL: `KARAKEEP_URL` environment variable (not a YAML config field)
- API bearer token: `KARAKEEP_API_KEY` environment variable (not a YAML config field — secrets out of config files)
- Both are required; if either is missing at startup, fail fast with a clear error naming the missing variable (same fail-fast pattern as config validation in Phase 1)

### Auth failure behavior
- Use `GET /users/me` to validate the token on startup — semantically correct auth check, not a side-effect of data fetching
- On invalid/expired token (HTTP 401): exit with message `"authentication failed: invalid API token (check KARAKEEP_API_KEY)"`
- On network/connection error: fail immediately, no retry — Docker Compose restart policies handle orchestration
- Auth check happens after config load and env var validation, before any rule execution

### Claude's Discretion
- oapi-codegen configuration details (generator config file format, which sub-packages to use)
- KarakeepAPI interface method set — expose only what Phase 2 needs (ListBookmarks + auth check) vs full spec surface
- Exact env var reading/validation code structure in main.go

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Karakeep API spec
- `karakeep-upstream/packages/open-api/karakeep-openapi-spec.json` — Full OpenAPI 3.0 spec used as codegen input; defines bookmark fields, pagination, auth scheme, and `/users/me` endpoint

### Project requirements and roadmap
- `.planning/REQUIREMENTS.md` §CONF-03 — "Container validates API token against Karakeep on startup before executing any rules"
- `.planning/ROADMAP.md` §Phase 2 — Success criteria: KarakeepAPI interface in engine package, cursor-based pagination, typed Go structs

### Existing codebase
- `internal/config/config.go` — Established patterns: pointer types for optional fields, fail-fast error returns, fmt.Errorf wrapping
- `cmd/karaclean/main.go` — Startup entry point where env var reading and auth check integrate after config load

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/config.Load()` + `config.ResolvePath()`: pattern for reading config — same fail-fast approach applies to env var reading for KARAKEEP_URL / KARAKEEP_API_KEY
- `ValidationErrors` type from Phase 1: established pattern for structured error types — follow same approach for auth errors

### Established Patterns
- Fail fast on missing/invalid configuration — exit with descriptive error, no silent degradation
- `fmt.Errorf("context: %w", err)` error wrapping throughout
- Pointer types for optional fields (relevant for generated struct handling)
- `go.yaml.in/yaml/v3` already in go.mod — oapi-codegen adds its own dependencies (net/http, oapi-codegen runtime)

### Integration Points
- `cmd/karaclean/main.go` — currently just loads config and prints success; Phase 2 adds: (1) read KARAKEEP_URL + KARAKEEP_API_KEY, (2) construct API client, (3) call auth check, (4) exit on failure
- `internal/engine/` — new package; receives `KarakeepAPI` interface; downstream phases (rule evaluation, actions) import this interface

</code_context>

<specifics>
## Specific Ideas

- No specific references — open to standard oapi-codegen patterns and idiomatic Go HTTP client conventions

</specifics>

<deferred>
## Deferred Ideas

- None — discussion stayed within phase scope

</deferred>

---

*Phase: 02-api-client-and-authentication*
*Context gathered: 2026-03-18*
