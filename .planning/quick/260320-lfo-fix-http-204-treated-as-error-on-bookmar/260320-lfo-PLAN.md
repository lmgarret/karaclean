---
phase: quick
plan: 260320-lfo
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/karakeep/client.go
  - internal/karakeep/client_test.go
autonomous: true
requirements: ["fix-http-204-delete"]

must_haves:
  truths:
    - "DeleteBookmark succeeds when API returns HTTP 204 No Content"
    - "DeleteBookmark still succeeds when API returns HTTP 200 OK"
    - "DeleteBookmark still errors on non-success status codes (e.g. 500)"
  artifacts:
    - path: "internal/karakeep/client.go"
      provides: "DeleteBookmark accepting 200 and 204"
      contains: "http.StatusNoContent"
    - path: "internal/karakeep/client_test.go"
      provides: "Test coverage for 204 success path"
      contains: "StatusNoContent"
  key_links:
    - from: "internal/karakeep/client.go"
      to: "internal/engine/api.go"
      via: "KarakeepAPI interface"
      pattern: "DeleteBookmark"
---

<objective>
Fix HTTP 204 No Content being treated as an error in DeleteBookmark.

Purpose: The Karakeep API returns 204 No Content on successful DELETE, but the client only accepts 200 OK, causing every successful deletion to be logged as an error.
Output: Patched client.go accepting both 200 and 204, updated tests confirming both paths.
</objective>

<execution_context>
@/var/home/lm/git/karaclean/.claude/get-shit-done/workflows/execute-plan.md
@/var/home/lm/git/karaclean/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/karakeep/client.go (lines 102-112: DeleteBookmark method)
@internal/karakeep/client_test.go (lines 286-325: existing delete tests)

<interfaces>
From internal/karakeep/client.go:
```go
// DeleteBookmark permanently removes the bookmark via DELETE /bookmarks/{id}.
func (c *KarakeepClient) DeleteBookmark(ctx context.Context, id string) error
```

Bug: Line 108 checks `resp.StatusCode() != http.StatusOK` — rejects 204.
Fix: Accept both 200 and 204 as success.
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Fix DeleteBookmark to accept HTTP 204 and update tests</name>
  <files>internal/karakeep/client.go, internal/karakeep/client_test.go</files>
  <behavior>
    - Test: DeleteBookmark succeeds when server returns 204 No Content (new test case)
    - Test: DeleteBookmark succeeds when server returns 200 OK (existing test, keep passing)
    - Test: DeleteBookmark errors when server returns 500 (existing test, keep passing)
  </behavior>
  <action>
1. In `internal/karakeep/client_test.go`, add a new test `TestDeleteBookmark_Success204` that sets up an httptest server returning `http.StatusNoContent` (204) with no body. Verify `DeleteBookmark` returns nil. Run tests -- this MUST FAIL (red) since client.go still only accepts 200.

2. In `internal/karakeep/client.go` line 108, change the status check in `DeleteBookmark` from:
   ```go
   if resp.StatusCode() != http.StatusOK {
   ```
   to:
   ```go
   if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
   ```
   This accepts both 200 OK and 204 No Content as success. All other status codes remain errors.

3. Run all tests to confirm green: existing 200 success test, new 204 success test, and 500 error test all pass.

4. Run `golangci-lint run ./internal/karakeep/...` to confirm no lint issues.
  </action>
  <verify>
    <automated>cd /var/home/lm/git/karaclean && go test ./internal/karakeep/... -run TestDeleteBookmark -v && golangci-lint run ./internal/karakeep/...</automated>
  </verify>
  <done>DeleteBookmark accepts both 200 and 204 as success. Three test cases pass: 200 success, 204 success, 500 error. Lint clean.</done>
</task>

</tasks>

<verification>
- `go test ./internal/karakeep/... -v` -- all tests pass
- `go test ./... -count=1` -- full test suite passes (no regressions)
- `golangci-lint run ./...` -- no lint errors
</verification>

<success_criteria>
- DeleteBookmark returns nil for HTTP 204 responses
- DeleteBookmark returns nil for HTTP 200 responses (no regression)
- DeleteBookmark returns error for non-success status codes
- All existing tests continue to pass
</success_criteria>

<output>
After completion, create `.planning/quick/260320-lfo-fix-http-204-treated-as-error-on-bookmar/260320-lfo-SUMMARY.md`
</output>
