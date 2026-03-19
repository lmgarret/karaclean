---
phase: quick
plan: 260319-uni
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/config/validate.go
  - internal/config/validate_test.go
  - internal/engine/actions.go
  - internal/engine/actions_test.go
  - internal/engine/run.go
  - internal/engine/run_test.go
  - cmd/karaclean/main.go
  - cmd/karaclean/main_test.go
autonomous: true
requirements: [per-rule-dryrun, richer-logs]

must_haves:
  truths:
    - "A rule with dryRun: true skips API mutation even when global dryRun is false"
    - "A rule with dryRun: false executes normally even when global dryRun is true"
    - "A rule with dryRun omitted inherits the global dryRun setting"
    - "Action logs include bookmark title/URL/source/tags, not just the ID"
    - "DRY-RUN logs include bookmark title/URL/source/tags, not just the ID"
  artifacts:
    - path: "internal/config/config.go"
      provides: "Rule.DryRun *bool field"
      contains: "DryRun *bool"
    - path: "internal/engine/actions.go"
      provides: "ExecuteAction accepts Bookmark for richer logging"
    - path: "internal/engine/run.go"
      provides: "Per-rule dryRun resolution logic"
  key_links:
    - from: "internal/engine/run.go"
      to: "internal/config/config.go"
      via: "rule.DryRun pointer check"
      pattern: "rule\\.DryRun"
    - from: "internal/engine/actions.go"
      to: "internal/engine/bookmark.go"
      via: "Bookmark fields in log output"
      pattern: "bookmark\\."
---

<objective>
Add per-rule dryRun override and enrich action/dry-run log output with bookmark details.

Purpose: Currently dryRun is global-only (config-level, env, flag). Users need per-rule granularity to test individual rules while others run live. Additionally, logs only show bookmark IDs which are opaque -- users need title, URL, source, and tags at a glance to understand what was (or would be) affected.

Output: Updated config model, engine actions, orchestrator, and comprehensive tests.
</objective>

<execution_context>
@/var/home/lm/git/karaclean/.claude/get-shit-done/workflows/execute-plan.md
@/var/home/lm/git/karaclean/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/config/config.go
@internal/config/validate.go
@internal/engine/actions.go
@internal/engine/run.go
@internal/engine/bookmark.go
@cmd/karaclean/main.go

<interfaces>
<!-- Key types and contracts the executor needs -->

From internal/config/config.go:
```go
type Rule struct {
	Name       string      `yaml:"name"`
	Conditions *Conditions `yaml:"conditions"`
	Unless     *Exceptions `yaml:"unless"`
	Action     string      `yaml:"action"`
}

type Config struct {
	Timezone string `yaml:"timezone"`
	Schedule string `yaml:"schedule"`
	DryRun   bool   `yaml:"dryRun"`
	Rules    []Rule `yaml:"rules"`
}
```

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

func ExecuteAction(ctx context.Context, api KarakeepAPI, action string, bookmarkID string, ruleName string, dryRun bool) ActionResult
```

From internal/engine/run.go:
```go
func Run(ctx context.Context, api KarakeepAPI, rules []config.Rule, dryRun bool) (RunSummary, error)
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Add per-rule dryRun to config and resolve in engine</name>
  <files>internal/config/config.go, internal/config/config_test.go, internal/config/validate.go, internal/config/validate_test.go, internal/engine/run.go, internal/engine/run_test.go, cmd/karaclean/main.go, cmd/karaclean/main_test.go</files>
  <behavior>
    - Config: Rule with `dryRun: true` parses to DryRun=ptr(true)
    - Config: Rule with `dryRun: false` parses to DryRun=ptr(false)
    - Config: Rule with dryRun omitted parses to DryRun=nil
    - Engine: rule.DryRun=nil inherits global dryRun value
    - Engine: rule.DryRun=ptr(true) overrides global dryRun=false to dry-run
    - Engine: rule.DryRun=ptr(false) overrides global dryRun=true to live
    - Run tests: per-rule dryRun=true with global=false still counts Archived/Deleted but makes no API calls for that rule
    - Run tests: per-rule dryRun=false with global=true makes real API calls for that rule
  </behavior>
  <action>
    1. In `internal/config/config.go`, add `DryRun *bool \`yaml:"dryRun"\`` to the `Rule` struct. Use pointer type to distinguish omitted (nil) from explicit false, consistent with existing pattern (Conditions uses *bool for Archived, Favourited).

    2. No validation changes needed for Rule.DryRun -- bool pointer with nil/true/false are all valid states. No new validation rule required.

    3. In `internal/config/config_test.go`, add tests:
       - `TestLoad_RuleDryRunTrue`: YAML with rule-level `dryRun: true`, assert `*cfg.Rules[0].DryRun == true`
       - `TestLoad_RuleDryRunFalse`: YAML with rule-level `dryRun: false`, assert `*cfg.Rules[0].DryRun == false`
       - `TestLoad_RuleDryRunOmitted`: YAML without rule-level dryRun, assert `cfg.Rules[0].DryRun == nil`

    4. In `internal/engine/run.go`, add a helper function:
       ```go
       // resolveRuleDryRun determines effective dry-run for a rule.
       // Per-rule setting (non-nil) overrides global; nil inherits global.
       func resolveRuleDryRun(ruleDryRun *bool, globalDryRun bool) bool {
           if ruleDryRun != nil {
               return *ruleDryRun
           }
           return globalDryRun
       }
       ```
       Update the `Run()` function to call `resolveRuleDryRun(rule.DryRun, dryRun)` and pass the result to `ExecuteAction` instead of the raw `dryRun` parameter.

    5. In `internal/engine/run_test.go`, add test cases to the existing `TestRun` table:
       - "per-rule dryRun true overrides global false": global dryRun=false, rule has DryRun=ptr(true). Expect Archived incremented but NO archiveBookmarkCalls (dry-run skips API).
       - "per-rule dryRun false overrides global true": global dryRun=true, rule has DryRun=ptr(false). Expect archiveBookmarkCalls to contain the bookmark ID (real API call despite global dry-run).
       - "per-rule dryRun nil inherits global true": global dryRun=true, rule has DryRun=nil. Expect NO archiveBookmarkCalls.

    6. Add a unit test `TestResolveRuleDryRun` in `internal/engine/run_test.go` covering all 4 combinations (nil+false, nil+true, ptr(true)+false, ptr(false)+true).

    7. In `cmd/karaclean/main.go`, add a log line after resolving global dry-run: when iterating is not needed here since per-rule resolution happens in engine.Run(). No changes needed to main.go beyond ensuring it still passes `dryRun` to `engine.Run()` (which it already does).
  </action>
  <verify>
    <automated>cd /var/home/lm/git/karaclean && go test ./internal/config/ ./internal/engine/ ./cmd/karaclean/ -run "DryRun|ResolveRule" -v -count=1</automated>
  </verify>
  <done>Rule struct has DryRun *bool field. resolveRuleDryRun correctly handles nil/true/false vs global. All new and existing tests pass. Per-rule dryRun=true prevents API calls; per-rule dryRun=false forces API calls regardless of global setting.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Enrich action and dry-run log output with bookmark details</name>
  <files>internal/engine/actions.go, internal/engine/actions_test.go, internal/engine/run.go, internal/engine/run_test.go</files>
  <behavior>
    - DRY-RUN log includes: action, bookmark ID, source, tags (comma-joined), and rule name
    - Live action success log includes: action, bookmark ID, source, tags, and rule name
    - Error log still includes bookmark ID and rule name plus the error
    - When tags are empty, log shows "tags=[]" (not a nil panic)
    - Bookmark with source="" logs "source=(unknown)"
  </behavior>
  <action>
    1. In `internal/engine/actions.go`, change `ExecuteAction` signature to accept a `Bookmark` instead of just `bookmarkID string`. New signature:
       ```go
       func ExecuteAction(ctx context.Context, api KarakeepAPI, action string, bookmark Bookmark, ruleName string, dryRun bool) ActionResult
       ```
       This avoids needing to pass individual fields and keeps the function clean.

    2. Build a `bookmarkSummary` helper function in `actions.go`:
       ```go
       func bookmarkSummary(b Bookmark) string {
           source := b.Source
           if source == "" {
               source = "(unknown)"
           }
           tags := "[]"
           if len(b.Tags) > 0 {
               tags = "[" + strings.Join(b.Tags, ", ") + "]"
           }
           return fmt.Sprintf("id=%s source=%s tags=%s", b.ID, source, tags)
       }
       ```
       Add `"strings"` to the import block.

    3. Update log lines in `ExecuteAction`:
       - DRY-RUN: `log.Printf("DRY-RUN %s: %s (rule: %s)", action, bookmarkSummary(bookmark), ruleName)`
       - Success (add a new log after the switch, before return, when err==nil): `log.Printf("%s: %s (rule: %s)", action, bookmarkSummary(bookmark), ruleName)`
       - Error: `result.Err = fmt.Errorf("%s failed: %s (rule: %s): %w", action, bookmarkSummary(bookmark), ruleName, err)` and `log.Printf("ERROR %s", result.Err)`

    4. Update `ActionResult.BookmarkID` to still be populated from `bookmark.ID`.

    5. In `internal/engine/run.go`, update the `ExecuteAction` call site (line 59) to pass the full `b` (Bookmark) instead of `b.ID`:
       ```go
       result := ExecuteAction(ctx, api, rule.Action, b, rule.Name, effectiveDryRun)
       ```

    6. In `internal/engine/actions_test.go`, update ALL existing test calls to pass a `engine.Bookmark{ID: "bk-1"}` (or "bk-2") instead of the string ID. Update `TestExecuteAction_DryRunLogOutput` to:
       - Pass a bookmark with ID="bk-99", Source="web", Tags=["cleanup", "old"]
       - Assert log contains "DRY-RUN", "archive", "bk-99", "web", "cleanup", "old", "cleanup-rule"
       - Add a new test `TestExecuteAction_LiveLogOutput` that verifies live (non-dry-run) successful action logs the bookmark summary.
       - Add test `TestBookmarkSummary_EmptySource` with Source="" verifying "(unknown)" appears.
       - Add test `TestBookmarkSummary_EmptyTags` with Tags=nil verifying "tags=[]" appears.

    7. In `internal/engine/run_test.go`, no changes needed to the existing `TestRun` table -- the mock API and Bookmark structs already work correctly since `Run()` passes `b` (a Bookmark) to ExecuteAction.
  </action>
  <verify>
    <automated>cd /var/home/lm/git/karaclean && go test ./internal/engine/ -v -count=1</automated>
  </verify>
  <done>All action logs (DRY-RUN, live success, error) include bookmark ID, source, and tags. Empty source shows "(unknown)", empty tags shows "tags=[]". All existing and new tests pass. No test in any package fails.</done>
</task>

<task type="auto">
  <name>Task 3: Full test suite validation and lint check</name>
  <files></files>
  <action>
    Run the complete test suite and linter to confirm nothing is broken across the entire project.

    1. Run `go test ./... -count=1` to verify all tests pass.
    2. Run `go vet ./...` to check for issues.
    3. If golangci-lint is available (check with `which golangci-lint`), run `golangci-lint run ./...`.

    No file changes expected -- this is a verification-only task.
  </action>
  <verify>
    <automated>cd /var/home/lm/git/karaclean && go test ./... -count=1 && go vet ./...</automated>
  </verify>
  <done>All tests pass across all packages. No vet warnings. Project is clean.</done>
</task>

</tasks>

<verification>
- `go test ./... -count=1` passes with zero failures
- `go vet ./...` reports no issues
- Per-rule dryRun config parsing works for true/false/omitted
- Per-rule dryRun correctly overrides or inherits global setting
- Log output for actions includes bookmark source and tags, not just ID
</verification>

<success_criteria>
1. A YAML config with `dryRun: true` on a specific rule causes only that rule to run in dry-run mode, regardless of global setting
2. A YAML config with `dryRun: false` on a specific rule forces live execution even when global dryRun is true
3. Omitting dryRun on a rule inherits the global dryRun behavior (backward compatible)
4. All action log lines (DRY-RUN, success, error) include bookmark source and tags alongside the ID
5. All tests pass, including new tests for both features
</success_criteria>

<output>
After completion, create `.planning/quick/260319-uni-per-rule-dryrun-and-richer-deletion-logs/260319-uni-SUMMARY.md`
</output>
