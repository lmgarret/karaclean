---
phase: quick
plan: 260320-emk
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/engine/bookmark.go
  - internal/engine/actions.go
  - internal/engine/actions_test.go
  - internal/engine/run.go
  - internal/engine/run_test.go
  - internal/karakeep/client.go
autonomous: true
requirements: [display-bookmark-size-in-logs, total-size-in-summary]

must_haves:
  truths:
    - "Per-bookmark action logs show human-readable size (e.g. 1.2 MB) when size is known"
    - "Per-bookmark action logs show no size token when size is zero/unknown"
    - "RunSummary includes total bytes processed (archived+deleted) in its String() output"
  artifacts:
    - path: "internal/engine/bookmark.go"
      provides: "Size field on Bookmark struct"
      contains: "Size"
    - path: "internal/engine/actions.go"
      provides: "Size display in bookmarkSummary and ActionResult tracking"
    - path: "internal/engine/run.go"
      provides: "TotalBytes in RunSummary, accumulated during run loop"
    - path: "internal/karakeep/client.go"
      provides: "Size extraction from Karakeep API Content union"
  key_links:
    - from: "internal/karakeep/client.go"
      to: "internal/engine/bookmark.go"
      via: "toEngineBookmark maps Content2.Size to Bookmark.Size"
      pattern: "Size.*float32|AsBookmarkContent2"
    - from: "internal/engine/actions.go"
      to: "internal/engine/run.go"
      via: "ActionResult.Size feeds RunSummary.TotalBytes accumulation"
      pattern: "TotalBytes.*Size|result\\.Size"
---

<objective>
Add bookmark size (in bytes) to per-bookmark action log lines and accumulate total size in the run summary.

Purpose: When karaclean deletes/archives bookmarks, the user wants to see how much storage each bookmark consumed and the total size freed in the summary line.

Output: Updated engine with size-aware logging and summary.
</objective>

<execution_context>
@/var/home/lm/git/karaclean/.claude/get-shit-done/workflows/execute-plan.md
@/var/home/lm/git/karaclean/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/engine/bookmark.go
@internal/engine/actions.go
@internal/engine/actions_test.go
@internal/engine/run.go
@internal/engine/run_test.go
@internal/karakeep/client.go

<interfaces>
<!-- Karakeep API Content union for asset bookmarks (only type with size) -->
From internal/karakeep/client.gen.go:
```go
type BookmarkContent2 struct {
    AssetId   string                    `json:"assetId"`
    AssetType BookmarkContent2AssetType `json:"assetType"`
    Content   *string                   `json:"content,omitempty"`
    FileName  *string                   `json:"fileName,omitempty"`
    Size      *float32                  `json:"size,omitempty"`
    SourceUrl *string                   `json:"sourceUrl,omitempty"`
    Type      BookmarkContent2Type      `json:"type"`
}

// Access via:
func (t Bookmark_Content) AsBookmarkContent2() (BookmarkContent2, error)
```

<!-- Current engine domain types -->
From internal/engine/bookmark.go:
```go
type Bookmark struct {
    ID         string
    CreatedAt  time.Time
    Archived   bool
    Favourited bool
    Source     string
    Tags       []string
    Note       string
}
```

From internal/engine/actions.go:
```go
type ActionResult struct {
    BookmarkID string
    RuleName   string
    Action     string
    DryRun     bool
    Err        error
}

func bookmarkSummary(b Bookmark) string // formats "id=X source=Y tags=Z"
func ExecuteAction(ctx, api, action, bookmark, ruleName, dryRun) ActionResult
```

From internal/engine/run.go:
```go
type RunSummary struct {
    Archived int `json:"archived"`
    Deleted  int `json:"deleted"`
    NoMatch  int `json:"no_match"`
    Excepted int `json:"excepted"`
    Errors   int `json:"errors"`
}
func (s RunSummary) String() string // "archived=2 deleted=1 ..."
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Add Size to Bookmark, ActionResult, and bookmarkSummary log output</name>
  <files>internal/engine/bookmark.go, internal/engine/actions.go, internal/engine/actions_test.go, internal/karakeep/client.go</files>
  <behavior>
    - bookmarkSummary with Size > 0 includes "size=1.2 MB" (human-readable) in output
    - bookmarkSummary with Size == 0 omits size token entirely (no "size=0 B")
    - DRY-RUN log line includes size when present
    - Live action log line includes size when present
    - Error log line includes size when present
    - ActionResult gains Size int64 field, populated from bookmark.Size
    - toEngineBookmark extracts Size from Content2 (asset bookmarks); 0 for link/text/unknown types
  </behavior>
  <action>
    1. In `internal/engine/bookmark.go`: Add `Size int64` field to Bookmark struct (bytes; 0 means unknown/not applicable).

    2. In `internal/engine/actions.go`:
       - Add `Size int64` to ActionResult struct.
       - Add a `humanSize(bytes int64) string` helper that formats bytes as human-readable (B, KB, MB, GB) with 1 decimal place. Use 1024-based units.
       - Update `bookmarkSummary(b Bookmark)`: if `b.Size > 0`, append ` size=X.X MB` (using humanSize) to the output string. If zero, omit entirely.
       - Set `result.Size = bookmark.Size` in ExecuteAction.

    3. In `internal/karakeep/client.go` `toEngineBookmark`:
       - Try `b.Content.AsBookmarkContent2()`. If err == nil and content.Size != nil, set `Size: int64(*content.Size)`.
       - For all other content types (link, text, unknown), Size stays 0.
       - This is best-effort: if the union parse fails, just leave Size as 0 (no error propagation).

    4. In `internal/engine/actions_test.go`:
       - Add test for humanSize: 0 -> "0 B", 500 -> "500 B", 1024 -> "1.0 KB", 1536 -> "1.5 KB", 1048576 -> "1.0 MB", 1073741824 -> "1.0 GB".
       - Add test: bookmarkSummary with Size=2048 includes "size=2.0 KB" in log output.
       - Add test: bookmarkSummary with Size=0 does NOT contain "size=" in log output.
       - Update TestExecuteAction_DryRunLogOutput: set Size on bookmark, verify log contains size string.
       - Update TestExecuteAction_LiveLogOutput: set Size on bookmark, verify log contains size string.
  </action>
  <verify>
    <automated>cd /var/home/lm/git/karaclean && go test ./internal/engine/ -run "TestHumanSize|TestBookmarkSummary|TestExecuteAction" -v</automated>
  </verify>
  <done>Per-bookmark log lines include human-readable size when available, omit when zero. ActionResult carries Size. toEngineBookmark extracts size from asset bookmarks. All new and updated tests pass.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Add TotalBytes to RunSummary and accumulate during Run</name>
  <files>internal/engine/run.go, internal/engine/run_test.go</files>
  <behavior>
    - RunSummary.TotalBytes accumulates Size from all successful (non-error) archived+deleted bookmarks
    - RunSummary.String() includes "total_size=X.X MB" when TotalBytes > 0, omits when 0
    - Dry-run actions also accumulate TotalBytes (user wants to see "would free X MB")
    - Excepted and NoMatch bookmarks do NOT contribute to TotalBytes
    - Error actions do NOT contribute to TotalBytes
  </behavior>
  <action>
    1. In `internal/engine/run.go`:
       - Add `TotalBytes int64 `json:"total_bytes"`` to RunSummary.
       - Update `String()`: if TotalBytes > 0, append ` total_size=X.X MB` using the humanSize helper from actions.go (export it as `HumanSize` or keep internal and call from same package).
       - In the Run() loop, after ExecuteAction returns with no error and action is archive/delete, add `summary.TotalBytes += result.Size`.

    2. In `internal/engine/run_test.go`:
       - Update TestRunSummary_String: set TotalBytes=1048576, verify output contains "total_size=1.0 MB".
       - Add TestRunSummary_String_ZeroBytes: TotalBytes=0, verify output does NOT contain "total_size".
       - Update existing Run test cases: add Size to test bookmarks where actions succeed, verify summary.TotalBytes equals expected sum.
       - Verify error actions do not accumulate TotalBytes.
  </action>
  <verify>
    <automated>cd /var/home/lm/git/karaclean && go test ./internal/engine/ -run "TestRunSummary_String|TestRun" -v</automated>
  </verify>
  <done>RunSummary.String() displays total_size when > 0. Run() accumulates TotalBytes from successful action results. All tests pass including updated run tests.</done>
</task>

</tasks>

<verification>
```bash
cd /var/home/lm/git/karaclean && go test ./... -count=1
cd /var/home/lm/git/karaclean && go vet ./...
```
</verification>

<success_criteria>
- `go test ./...` passes with zero failures
- Per-bookmark log lines show human-readable size for asset bookmarks (e.g. "size=1.2 MB")
- Per-bookmark log lines omit size for link/text bookmarks (Size=0)
- RunSummary String() shows "total_size=X.X MB" when bookmarks were processed
- RunSummary String() omits total_size when no bytes were processed
</success_criteria>

<output>
After completion, create `.planning/quick/260320-emk-display-bookmark-size-in-deletion-logs-a/260320-emk-SUMMARY.md`
</output>
