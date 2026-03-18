---
plan: 02-01
phase: 02-api-client-and-authentication
status: complete
wave: 1
---

# Plan 02-01 Summary — API Client Generation

## What Was Done

Generated the Karakeep API client from the OpenAPI spec, defined engine-layer types and interface, and created the thin wrapper implementing `engine.KarakeepAPI`.

## Artifacts Created

| File | Purpose |
|------|---------|
| `internal/karakeep/oapi-codegen.yaml` | codegen config: package karakeep, generates models+client |
| `internal/karakeep/generate.go` | `//go:generate` directive pointing at OpenAPI spec |
| `internal/karakeep/client.gen.go` | Generated types and HTTP client (DO NOT EDIT) |
| `internal/karakeep/client.go` | Thin wrapper (`KarakeepClient`) implementing `engine.KarakeepAPI` |
| `internal/engine/api.go` | `KarakeepAPI` interface: `CheckAuth` + `ListBookmarks` |
| `internal/engine/bookmark.go` | Domain `Bookmark` type for rule evaluation |

## Key Decisions

- **Wrapper named `KarakeepClient`** (not `Client`): oapi-codegen already generates a `Client` type and `NewClient` func in the same package — name collision required renaming.
- **`CreatedAt` parsed via `time.RFC3339`**: generated type uses `string`, wrapper converts to `time.Time`.
- **`Source` is `*BookmarkSource`** in generated code — wrapper dereferences and converts to `string`.
- **`Note` is `*string`** — wrapper converts nil to empty string.
- **Empty bookmark slice**: initialized to `[]engine.Bookmark{}` (not nil) to match test expectations.

## Verification

- `go generate ./internal/karakeep/` — OK
- `go build ./...` — OK
- `go test ./internal/karakeep/... ./internal/engine/... -v -count=1` — all PASS (tests went straight GREEN)
