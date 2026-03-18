# Phase 2: API Client and Authentication - Research

**Researched:** 2026-03-18
**Domain:** Go HTTP client generation from OpenAPI spec, bearer token authentication
**Confidence:** HIGH

## Summary

Phase 2 builds the communication layer between karaclean and the Karakeep API. The user has locked in oapi-codegen as the code generation tool, reading the OpenAPI spec at `karakeep-upstream/packages/open-api/karakeep-openapi-spec.json`. The spec is OpenAPI 3.0, which oapi-codegen v2 handles natively. The generated client goes into `internal/karakeep/`, and a `KarakeepAPI` interface is defined in `internal/engine/` for mock-based testing.

The Karakeep API uses bearer token auth, cursor-based pagination (via `cursor`/`limit` query params and `nextCursor` in responses), and returns `401 Unauthorized` as plain text on auth failure. The `GET /users/me` endpoint is the designated auth check. The `GET /bookmarks` endpoint returns `PaginatedBookmarks` with a `Bookmark` schema containing all fields needed by downstream phases (id, createdAt, archived, favourited, source, tags, note).

**Primary recommendation:** Use oapi-codegen v2.6.0 to generate types + client code from the existing OpenAPI spec, wrap with a thin `KarakeepAPI` interface in the engine package, and use `net/http/httptest` for testing the client against fake server responses.

<user_constraints>

## User Constraints (from CONTEXT.md)

### Locked Decisions
- Use oapi-codegen to generate the Karakeep HTTP client from the OpenAPI spec at `karakeep-upstream/packages/open-api/karakeep-openapi-spec.json`
- Generate client interface + concrete HTTP implementation (not typed functions only) -- enables mock-based testing by passing the interface to the engine
- Generated code lives in `internal/karakeep/`
- Karakeep base URL: `KARAKEEP_URL` environment variable (not a YAML config field)
- API bearer token: `KARAKEEP_API_KEY` environment variable (not a YAML config field -- secrets out of config files)
- Both are required; if either is missing at startup, fail fast with a clear error naming the missing variable
- Use `GET /users/me` to validate the token on startup -- semantically correct auth check
- On invalid/expired token (HTTP 401): exit with message `"authentication failed: invalid API token (check KARAKEEP_API_KEY)"`
- On network/connection error: fail immediately, no retry -- Docker Compose restart policies handle orchestration
- Auth check happens after config load and env var validation, before any rule execution

### Claude's Discretion
- oapi-codegen configuration details (generator config file format, which sub-packages to use)
- KarakeepAPI interface method set -- expose only what Phase 2 needs (ListBookmarks + auth check) vs full spec surface
- Exact env var reading/validation code structure in main.go

### Deferred Ideas (OUT OF SCOPE)
- None

</user_constraints>

<phase_requirements>

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| CONF-03 | Container validates API token against Karakeep on startup before executing any rules | `GET /users/me` endpoint returns 200 on valid token, 401 on invalid; auth check integrates into main.go after config load; oapi-codegen generates typed client with bearer auth support via `securityprovider.NewSecurityProviderBearerToken` |

</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| oapi-codegen | v2.6.0 | Generate Go client + types from OpenAPI 3.0 spec | User-locked decision; most popular Go OpenAPI codegen tool |
| oapi-codegen/runtime | latest (pulled as transitive dep) | Runtime helpers for generated client (parameter binding, types) | Required by generated code |
| securityprovider | (part of oapi-codegen/v2) | Bearer token auth via `WithRequestEditorFn` | Built-in auth support, no custom code needed |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| net/http | stdlib | HTTP transport for generated client | Always -- generated client wraps stdlib http.Client |
| net/http/httptest | stdlib | Test server for client tests | All client tests -- avoids real HTTP calls |
| context | stdlib | Request context propagation | All API calls accept context.Context |
| encoding/json | stdlib | JSON serialization (used by generated code) | Transitive -- generated code handles this |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| oapi-codegen | Hand-rolled HTTP client | N/A -- user locked oapi-codegen |
| oapi-codegen | ogen (alternative Go OpenAPI codegen) | N/A -- user locked oapi-codegen |
| httptest | Real Karakeep instance in tests | Too heavy for unit tests; httptest is fast, deterministic |

**Installation:**
```bash
# Add oapi-codegen as a Go tool dependency (Go 1.24+)
cd /var/home/lm/git/karaclean
go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0
```

**Version verification:** oapi-codegen v2.6.0 confirmed via `go install` download on 2026-03-18. Runtime package version is pulled transitively.

## Architecture Patterns

### Recommended Project Structure
```
internal/
  karakeep/
    generate.go          # //go:generate directive + oapi-codegen config reference
    oapi-codegen.yaml    # codegen config: types + client
    client.gen.go        # generated types + client (DO NOT EDIT)
  engine/
    api.go               # KarakeepAPI interface definition
cmd/
  karaclean/
    main.go              # startup: config load -> env var read -> client init -> auth check
```

### Pattern 1: oapi-codegen Configuration (Two-in-One)
**What:** Single YAML config file generating both types and client into one package
**When to use:** When the spec is not enormous and you only need one consumer package
**Example:**
```yaml
# internal/karakeep/oapi-codegen.yaml
package: karakeep
output: client.gen.go
generate:
  models: true
  client: true
```

Then in `generate.go`:
```go
package karakeep

//go:generate go tool oapi-codegen --config=oapi-codegen.yaml ../../karakeep-upstream/packages/open-api/karakeep-openapi-spec.json
```

### Pattern 2: Bearer Token Client Construction
**What:** Create oapi-codegen client with bearer auth using built-in security provider
**When to use:** Every time the client is instantiated
**Example:**
```go
import (
    "github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
    "github.com/lm/karaclean/internal/karakeep"
)

bearerAuth, err := securityprovider.NewSecurityProviderBearerToken(apiKey)
if err != nil {
    return fmt.Errorf("creating auth provider: %w", err)
}

client, err := karakeep.NewClientWithResponses(
    baseURL+"/api/v1",
    karakeep.WithRequestEditorFn(bearerAuth.Intercept),
)
```

### Pattern 3: KarakeepAPI Interface (Narrow Surface)
**What:** Define a minimal interface in the engine package with only the methods Phase 2 needs
**When to use:** Interface lives in `internal/engine/api.go`; downstream phases extend it as needed
**Example:**
```go
package engine

import "context"

// KarakeepAPI defines the subset of Karakeep API operations used by the engine.
// New methods are added as phases require them.
type KarakeepAPI interface {
    // CheckAuth validates the API token by calling GET /users/me.
    // Returns nil on success, error on 401 or network failure.
    CheckAuth(ctx context.Context) error

    // ListBookmarks retrieves all bookmarks using cursor-based pagination.
    // Returns the complete list across all pages.
    ListBookmarks(ctx context.Context) ([]Bookmark, error)
}
```

**Recommendation:** Start narrow (only CheckAuth + ListBookmarks). Phases 6+ will add UpdateBookmark/DeleteBookmark. Growing the interface is cheaper than shrinking it.

### Pattern 4: Thin Wrapper Over Generated Client
**What:** A struct in `internal/karakeep/` that implements the engine's `KarakeepAPI` interface by delegating to the generated `ClientWithResponses`
**When to use:** Bridge between generated code and the engine interface
**Example:**
```go
package karakeep

import (
    "context"
    "fmt"
    "net/http"

    "github.com/lm/karaclean/internal/engine"
)

// Client wraps the oapi-codegen generated client and implements engine.KarakeepAPI.
type Client struct {
    inner *ClientWithResponses
}

func (c *Client) CheckAuth(ctx context.Context) error {
    resp, err := c.inner.GetCurrentUserWithResponse(ctx)
    if err != nil {
        return fmt.Errorf("auth check: %w", err)
    }
    if resp.StatusCode() == http.StatusUnauthorized {
        return fmt.Errorf("authentication failed: invalid API token (check KARAKEEP_API_KEY)")
    }
    if resp.StatusCode() != http.StatusOK {
        return fmt.Errorf("auth check: unexpected status %d", resp.StatusCode())
    }
    return nil
}
```

### Anti-Patterns to Avoid
- **Editing generated code directly:** Never modify `client.gen.go`. All customization goes in wrapper types or the oapi-codegen config. Generated files should have `// DO NOT EDIT` header.
- **Exposing generated types in the engine interface:** The engine package should define its own `Bookmark` type (or re-export from karakeep). This decouples the engine from codegen details and makes the interface testable without importing generated code.
- **Retrying on network errors:** The user explicitly decided no retry logic -- Docker Compose restart policies handle this. Do not add exponential backoff or retry loops.
- **Putting secrets in config YAML:** `KARAKEEP_URL` and `KARAKEEP_API_KEY` come from environment variables only. Never add them to the Config struct.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| HTTP client for Karakeep API | Manual http.NewRequest + json.Decode for each endpoint | oapi-codegen generated client | OpenAPI spec has 20+ endpoints; hand-rolling is error-prone and won't track spec changes |
| Bearer token header injection | Custom `RoundTripper` or per-request header setting | `securityprovider.NewSecurityProviderBearerToken` | Already built, tested, composable via `WithRequestEditorFn` |
| Cursor-based pagination loop | Manual `for` loop with raw HTTP calls | Wrapper method on the thin client that calls generated `ListBookmarksWithResponse` in a loop | The generated client handles serialization; you only need the loop logic |
| OpenAPI type definitions | Manual Go structs matching the API JSON | oapi-codegen generated models | Spec has nullable fields, oneOf content types, nested tags -- codegen handles all edge cases |

**Key insight:** oapi-codegen generates both the types AND the HTTP client. The only hand-written code is: (1) the wrapper implementing `KarakeepAPI`, (2) the pagination loop, (3) the startup auth check orchestration in `main.go`.

## Common Pitfalls

### Pitfall 1: Forgetting /api/v1 Base Path
**What goes wrong:** Client hits wrong URLs, gets 404s
**Why it happens:** The OpenAPI spec defines server URL as `{address}/api/v1`. oapi-codegen's `NewClientWithResponses` takes a server URL -- if you pass just the host without `/api/v1`, all paths will be wrong.
**How to avoid:** Always append `/api/v1` to the `KARAKEEP_URL` value when constructing the client. Or validate/normalize at env var read time.
**Warning signs:** 404 responses on all API calls.

### Pitfall 2: oneOf Content Field in Bookmark
**What goes wrong:** The `content` field in the Bookmark schema uses `oneOf` (link | text | asset). oapi-codegen may generate this as an interface or a union type that requires type assertions.
**Why it happens:** Go has no sum types; oapi-codegen represents oneOf via embedded discriminator or merged struct.
**How to avoid:** After running codegen, inspect the generated `Bookmark` struct to see how `content` is represented. For Phase 2, we only need metadata fields (id, createdAt, archived, favourited, source, tags, note) -- the `content` field can be ignored or left as raw JSON.
**Warning signs:** Compilation errors or panics when deserializing bookmark responses.

### Pitfall 3: 401 Response is text/plain, Not JSON
**What goes wrong:** Trying to parse 401 response body as JSON fails
**Why it happens:** The OpenAPI spec defines 401 responses as `text/plain` with body `"Unauthorized"`. The generated `ClientWithResponses` will have a nil parsed body for 401s.
**How to avoid:** Check `resp.StatusCode()` first, not the parsed body. The generated client exposes `StatusCode()` on all response types.
**Warning signs:** nil pointer dereference when accessing response body on 401.

### Pitfall 4: Source Enum Mismatch -- singlefile
**What goes wrong:** The OpenAPI spec includes `singlefile` as a valid source value, but Phase 1's config validation only accepts: `rss, web, api, mobile, extension, cli, import`.
**Why it happens:** The config validation was based on an earlier understanding. The OpenAPI spec has 8 source values, not 7.
**How to avoid:** This is a KNOWN GAP to address -- either update Phase 1 validation to include `singlefile`, or document it as a known limitation. The generated types will include `singlefile` regardless.
**Warning signs:** Bookmarks with `source: "singlefile"` would not match any config rule's source condition.

### Pitfall 5: Pagination -- nextCursor null vs empty string
**What goes wrong:** Pagination loop never terminates or skips the last page
**Why it happens:** The spec says `nextCursor` is `nullable: true` -- it's null (not empty string) when there are no more pages. In Go, oapi-codegen generates this as `*string`. Check for nil, not empty string.
**How to avoid:** Loop condition: `cursor != nil` (or check the pointer). Do not check `cursor != ""`.
**Warning signs:** Infinite loop fetching bookmarks, or missing the last page of results.

### Pitfall 6: Generated Code Import Path
**What goes wrong:** Build fails with import errors
**Why it happens:** oapi-codegen moved from `github.com/deepmap/oapi-codegen` to `github.com/oapi-codegen/oapi-codegen/v2`. Old tutorials reference the wrong path.
**How to avoid:** Use only `github.com/oapi-codegen/oapi-codegen/v2` imports. The runtime package is `github.com/oapi-codegen/runtime`.
**Warning signs:** Go module resolution errors, "module not found" during `go get`.

## Code Examples

Verified patterns from official sources and OpenAPI spec analysis:

### oapi-codegen Configuration File
```yaml
# internal/karakeep/oapi-codegen.yaml
# Source: https://pkg.go.dev/github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen#GenerateOptions
package: karakeep
output: client.gen.go
generate:
  models: true
  client: true
```

### go:generate Directive
```go
// internal/karakeep/generate.go
package karakeep

//go:generate go tool oapi-codegen --config=oapi-codegen.yaml ../../karakeep-upstream/packages/open-api/karakeep-openapi-spec.json
```

### Client Construction with Bearer Auth
```go
// Source: https://github.com/oapi-codegen/oapi-codegen (securityprovider example)
import (
    "fmt"
    "github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
    "github.com/lm/karaclean/internal/karakeep"
)

func NewKarakeepClient(baseURL, apiKey string) (*karakeep.Client, error) {
    bearerAuth, err := securityprovider.NewSecurityProviderBearerToken(apiKey)
    if err != nil {
        return nil, fmt.Errorf("creating auth provider: %w", err)
    }

    inner, err := karakeep.NewClientWithResponses(
        baseURL+"/api/v1",
        karakeep.WithRequestEditorFn(bearerAuth.Intercept),
    )
    if err != nil {
        return nil, fmt.Errorf("creating API client: %w", err)
    }

    return &karakeep.Client{Inner: inner}, nil
}
```

### Pagination Loop Pattern
```go
func (c *Client) ListBookmarks(ctx context.Context) ([]Bookmark, error) {
    var all []Bookmark
    var cursor *string

    for {
        resp, err := c.inner.ListBookmarksWithResponse(ctx, &ListBookmarksParams{
            Cursor: cursor,
            Limit:  intPtr(100),
        })
        if err != nil {
            return nil, fmt.Errorf("listing bookmarks: %w", err)
        }
        if resp.StatusCode() != http.StatusOK {
            return nil, fmt.Errorf("listing bookmarks: unexpected status %d", resp.StatusCode())
        }

        all = append(all, resp.JSON200.Bookmarks...)

        if resp.JSON200.NextCursor == nil {
            break
        }
        cursor = resp.JSON200.NextCursor
    }

    return all, nil
}
```

### httptest Pattern for Client Tests
```go
func TestCheckAuth_Success(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/api/v1/users/me" {
            t.Errorf("unexpected path: %s", r.URL.Path)
        }
        if r.Header.Get("Authorization") != "Bearer test-token" {
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte("Unauthorized"))
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]any{"id": "user1", "localUser": true})
    }))
    defer srv.Close()

    client, _ := NewKarakeepClient(srv.URL, "test-token")
    err := client.CheckAuth(context.Background())
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

### Env Var Reading Pattern (follows existing config.ResolvePath style)
```go
func requireEnv(key string) (string, error) {
    val := os.Getenv(key)
    if val == "" {
        return "", fmt.Errorf("required environment variable %s is not set", key)
    }
    return val, nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `github.com/deepmap/oapi-codegen` | `github.com/oapi-codegen/oapi-codegen/v2` | May 2024 | Must use new import path |
| `go install` for tools | `go get -tool` (Go 1.24+) | Go 1.24 (Feb 2025) | Tool deps tracked in go.mod, reproducible builds |
| `//go:generate oapi-codegen` | `//go:generate go tool oapi-codegen` | Go 1.24+ | Uses module-tracked tool, not PATH binary |
| Pointer for optional fields | `omitzero` tag option (Go 1.24+) | Go 1.24 | oapi-codegen v2.5.0+ supports `prefer-skip-optional-pointer-with-omitzero` -- but pointer approach is fine and matches existing codebase |

**Deprecated/outdated:**
- `github.com/deepmap/oapi-codegen`: Old import path, no longer maintained under that org
- `-generate types,client` CLI flags: Superseded by YAML config file approach

## Open Questions

1. **How does oapi-codegen handle the Bookmark `content` oneOf field?**
   - What we know: The spec uses `oneOf` with discriminator on `type` field (link/text/asset). oapi-codegen generates merged structs or interface types for oneOf.
   - What's unclear: Exact generated Go type -- need to run codegen and inspect output.
   - Recommendation: Run `go generate` early in implementation, inspect the `Bookmark` struct, and decide whether to expose `content` in the engine's `Bookmark` type or omit it (Phase 2 does not need it).

2. **Should the engine define its own Bookmark type or reuse the generated one?**
   - What we know: The engine interface should be mockable without importing generated code. A separate type provides decoupling.
   - What's unclear: Whether the mapping overhead is worth it for a small project.
   - Recommendation: Define a slim `engine.Bookmark` with only the fields needed for rule evaluation (id, createdAt, archived, favourited, source, tags, note). The karakeep wrapper maps generated types to engine types. This keeps the engine package independent of codegen.

3. **Rate limiting behavior**
   - What we know: API docs mention rate limiting with 429 responses. STATE.md notes "Karakeep API rate limiting on reads is undocumented."
   - What's unclear: Whether rate limiting applies to self-hosted instances, what the limits are.
   - Recommendation: For Phase 2, do not implement rate limit handling. If 429 is encountered, treat it as an unexpected status error. Revisit if it becomes a problem in practice.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib), Go 1.26.1 |
| Config file | None needed -- `go test` works out of the box |
| Quick run command | `go test ./internal/karakeep/... ./internal/engine/... -v -count=1` |
| Full suite command | `go test ./... -v -count=1` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CONF-03a | Env var validation: missing KARAKEEP_URL fails fast | unit | `go test ./cmd/karaclean/... -run TestRequireEnv -v` | Wave 0 |
| CONF-03b | Env var validation: missing KARAKEEP_API_KEY fails fast | unit | `go test ./cmd/karaclean/... -run TestRequireEnv -v` | Wave 0 |
| CONF-03c | Auth check success (200 from /users/me) | unit | `go test ./internal/karakeep/... -run TestCheckAuth_Success -v` | Wave 0 |
| CONF-03d | Auth check failure (401 from /users/me) | unit | `go test ./internal/karakeep/... -run TestCheckAuth_Unauthorized -v` | Wave 0 |
| CONF-03e | Auth check network error | unit | `go test ./internal/karakeep/... -run TestCheckAuth_NetworkError -v` | Wave 0 |
| CONF-03f | ListBookmarks single page | unit | `go test ./internal/karakeep/... -run TestListBookmarks_SinglePage -v` | Wave 0 |
| CONF-03g | ListBookmarks pagination (multi-page) | unit | `go test ./internal/karakeep/... -run TestListBookmarks_Pagination -v` | Wave 0 |
| CONF-03h | ListBookmarks empty result | unit | `go test ./internal/karakeep/... -run TestListBookmarks_Empty -v` | Wave 0 |
| CONF-03i | KarakeepAPI interface is mockable | unit | `go test ./internal/engine/... -run TestMockAPI -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/karakeep/... ./internal/engine/... -v -count=1`
- **Per wave merge:** `go test ./... -v -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/karakeep/client_test.go` -- covers CONF-03c through CONF-03h (httptest-based)
- [ ] `internal/engine/api_test.go` -- covers CONF-03i (mock implements interface)
- [ ] Generated code must exist first (`go generate`) before tests can compile

## Sources

### Primary (HIGH confidence)
- Karakeep OpenAPI spec at `karakeep-upstream/packages/open-api/karakeep-openapi-spec.json` -- Bookmark schema, PaginatedBookmarks, /users/me endpoint, /bookmarks endpoint, auth scheme, pagination params
- Existing codebase: `internal/config/config.go`, `internal/config/validate.go`, `cmd/karaclean/main.go` -- established patterns

### Secondary (MEDIUM confidence)
- [oapi-codegen GitHub repository](https://github.com/oapi-codegen/oapi-codegen) -- v2.6.0 confirmed, configuration format, generate options, securityprovider pattern
- [oapi-codegen GoDoc - GenerateOptions](https://pkg.go.dev/github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen#GenerateOptions) -- client and models boolean fields in YAML config
- [oapi-codegen runtime package](https://pkg.go.dev/github.com/oapi-codegen/runtime) -- runtime dependency for generated code

### Tertiary (LOW confidence)
- oapi-codegen oneOf handling -- based on WebSearch results; exact behavior for this spec's Bookmark content field needs validation by running codegen

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - oapi-codegen v2.6.0 verified via download, user locked the choice
- Architecture: HIGH - patterns derived from official oapi-codegen examples and existing codebase conventions
- Pitfalls: HIGH - derived from direct OpenAPI spec analysis (401 text/plain, nullable cursor, /api/v1 base path)
- oneOf handling: LOW - need to run codegen to see actual output for this specific spec

**Research date:** 2026-03-18
**Valid until:** 2026-04-18 (stable domain, oapi-codegen v2 is mature)
