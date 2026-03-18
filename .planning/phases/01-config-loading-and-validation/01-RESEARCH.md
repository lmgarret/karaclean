# Phase 1: Config Loading and Validation - Research

**Researched:** 2026-03-18
**Domain:** Go YAML config parsing, struct-based validation, project initialization
**Confidence:** HIGH

## Summary

Phase 1 is a greenfield Go project setup: initialize the module, define typed config structs matching the v1 YAML schema, parse with strict unknown-field rejection, and validate semantics (required fields, enum values) with collected error reporting. No Go code exists yet -- this phase creates the foundation.

The Go YAML ecosystem recently shifted: `gopkg.in/yaml.v3` was marked unmaintained in April 2025; the YAML organization forked it to `go.yaml.in/yaml/v3` (currently v3.0.4, security-fixes only) and is developing v4 (still RC4 as of January 2026). For a new project starting now, `go.yaml.in/yaml/v3` is the correct import path -- it is the maintained fork with the same API as `gopkg.in/yaml.v3` but under active security maintenance.

The critical feature for CONF-02 is `Decoder.KnownFields(true)`, which rejects any YAML key that does not map to a struct field. This is available in both `gopkg.in/yaml.v3` and `go.yaml.in/yaml/v3`. Semantic validation (enum checking, required fields, error collection) must be hand-written since yaml.v3 only handles structural parsing, not business logic constraints.

**Primary recommendation:** Use `go.yaml.in/yaml/v3` with `Decoder.KnownFields(true)` for strict parsing, and implement a custom `Validate() []error` method on the config struct for semantic validation with collected errors.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- API token and server URL are NOT in the YAML config file -- they are passed as env vars (`KARACLEAN_API_URL`, `KARACLEAN_API_TOKEN`)
- Top-level config structure: `timezone`, `schedule`, `rules`
- Rule anatomy: `name` (optional label), `conditions`, `unless`, `action`
- Condition semantics: conditions AND together, unless exceptions OR together
- Phase 1 defines the FULL v1 schema (all condition fields, exception fields, actions, source enum values)
- All v1 condition fields: `olderThan` (days, int), `source` (enum), `archived` (bool), `favourited` (bool), `hasTag` (string), `lacksTag` (string)
- All v1 exception fields: `favourited` (bool), `hasTag` (string), `hasNote` (bool), `archived` (bool)
- Both v1 actions: `archive`, `delete` (string enum)
- Valid `source` values: `rss`, `web`, `api`, `mobile`, `extension`, `cli`, `import`
- Unknown fields in YAML cause a validation error (strict parsing -- CONF-02)
- Config file discovery precedence: `--config` flag > `KARACLEAN_CONFIG` env var > `/config/karaclean.yaml` (default)
- All validation errors collected and reported at once with field paths matching YAML structure
- Error format example: `rules[0].action: invalid value "remove" (must be archive or delete)`

### Claude's Discretion
- YAML parsing library choice (gopkg.in/yaml.v3 is standard; go.yaml.in/yaml/v4 if stable by dev time)
- Whether `name` is required or optional on rules
- Exact Go struct layout and package organization within `internal/config/`
- Whether to use a custom validator or implement validation inline

### Deferred Ideas (OUT OF SCOPE)
- AND/OR logical combinators within a single rule -- planned as v2 RULE-02
- `--validate` flag (validate config without running) -- planned as v2 TOOL-01
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| CONF-01 | User can define rules in a YAML config file mounted into the container | go.yaml.in/yaml/v3 Decoder parses YAML into typed Go structs; config discovery logic resolves file path from flag/env/default |
| CONF-02 | Config validation rejects unknown fields at startup (strict YAML parsing) | `Decoder.KnownFields(true)` rejects unknown YAML keys; custom `Validate()` method collects semantic errors (enum values, required fields) |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go.yaml.in/yaml/v3 | v3.0.4 | YAML parsing with strict field checking | Maintained fork of gopkg.in/yaml.v3 by YAML org; `KnownFields(true)` provides CONF-02 unknown field rejection; 490+ importers |
| Go stdlib `os` | (stdlib) | Environment variable and file reading | Config discovery (env vars, file paths) |
| Go stdlib `fmt/strings` | (stdlib) | Error message formatting | Field path error messages like `rules[0].action: ...` |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Go stdlib `testing` | (stdlib) | Unit tests | Table-driven tests for parsing and validation |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| go.yaml.in/yaml/v3 | go.yaml.in/yaml/v4 (RC4) | v4 has a new config API but is not stable yet -- stay on v3 |
| go.yaml.in/yaml/v3 | gopkg.in/yaml.v3 | Same code but gopkg.in is unmaintained since April 2025 -- use the maintained fork |
| Custom validation | go-playground/validator | Adds a dependency for simple enum/required checks that are easy to write by hand; YAML field paths need custom formatting anyway |
| stdlib testing | testify | Adds dependency; for config validation, stdlib `testing` with table-driven tests and helper functions is sufficient |

### Discretion Recommendations

**YAML library:** Use `go.yaml.in/yaml/v3` (v3.0.4). It is the maintained fork with security fixes. v4 is RC4 and not production-ready.

**Rule `name` field:** Make it optional. Rules are identifiable by their index in the array. A name is helpful for logging but should not be required.

**Validation approach:** Custom `Validate() []ValidationError` method on the config struct. A dedicated `ValidationError` type with `Field` and `Message` allows collecting all errors and formatting them with YAML-matching field paths. This is simpler than wiring a third-party validator for this use case.

**Installation:**
```bash
go mod init github.com/user/karaclean  # or appropriate module path
go get go.yaml.in/yaml/v3@v3.0.4
```

## Architecture Patterns

### Recommended Project Structure
```
karaclean/
├── cmd/
│   └── karaclean/
│       └── main.go           # Entry point: resolve config path, load, validate, exit on error
├── internal/
│   └── config/
│       ├── config.go         # Config struct types, Load() function
│       ├── config_test.go    # Table-driven tests for loading
│       ├── validate.go       # Validate() method, ValidationError type
│       └── validate_test.go  # Table-driven tests for validation
├── go.mod
└── go.sum
```

### Pattern 1: Config Loading Pipeline
**What:** A `Load(path string) (*Config, error)` function that opens the file, creates a Decoder with `KnownFields(true)`, decodes into the struct, then calls `Validate()`.
**When to use:** Always -- this is the single entry point for config loading.
**Example:**
```go
// Source: gopkg.in/yaml.v3 official docs (KnownFields)
func Load(path string) (*Config, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("opening config: %w", err)
    }
    defer f.Close()

    var cfg Config
    decoder := yaml.NewDecoder(f)
    decoder.KnownFields(true) // CONF-02: reject unknown fields
    if err := decoder.Decode(&cfg); err != nil {
        return nil, fmt.Errorf("parsing config: %w", err)
    }

    if errs := cfg.Validate(); len(errs) > 0 {
        return nil, &ValidationErrors{Errors: errs}
    }
    return &cfg, nil
}
```

### Pattern 2: Config File Discovery
**What:** Resolve config file path from flag, env var, or default in priority order.
**When to use:** At startup before calling `Load()`.
**Example:**
```go
func ResolvePath(flagValue string) string {
    if flagValue != "" {
        return flagValue
    }
    if envPath := os.Getenv("KARACLEAN_CONFIG"); envPath != "" {
        return envPath
    }
    return "/config/karaclean.yaml"
}
```

### Pattern 3: Collected Validation Errors
**What:** A `ValidationError` struct with `Field` and `Message`, and a `Validate()` method that returns a slice of all errors found.
**When to use:** Semantic validation after YAML parsing succeeds.
**Example:**
```go
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors struct {
    Errors []ValidationError
}

func (e *ValidationErrors) Error() string {
    var b strings.Builder
    b.WriteString("config validation failed:\n")
    for _, err := range e.Errors {
        fmt.Fprintf(&b, "  - %s\n", err.Error())
    }
    return b.String()
}
```

### Pattern 4: Config Struct Design
**What:** Typed Go structs with yaml tags matching the YAML schema exactly.
**When to use:** Define all v1 config types.
**Example:**
```go
type Config struct {
    Timezone string `yaml:"timezone"`
    Schedule string `yaml:"schedule"`
    Rules    []Rule `yaml:"rules"`
}

type Rule struct {
    Name       string      `yaml:"name"`
    Conditions *Conditions `yaml:"conditions"`
    Unless     *Exceptions `yaml:"unless"`
    Action     string      `yaml:"action"`
}

type Conditions struct {
    OlderThan  *int    `yaml:"olderThan"`
    Source     *string `yaml:"source"`
    Archived   *bool   `yaml:"archived"`
    Favourited *bool   `yaml:"favourited"`
    HasTag     *string `yaml:"hasTag"`
    LacksTag   *string `yaml:"lacksTag"`
}

type Exceptions struct {
    Favourited *bool   `yaml:"favourited"`
    HasTag     *string `yaml:"hasTag"`
    HasNote    *bool   `yaml:"hasNote"`
    Archived   *bool   `yaml:"archived"`
}
```

**Key design choice:** Use pointer types (`*int`, `*string`, `*bool`) for optional fields so `nil` means "not specified" vs zero-value means "explicitly set to zero/empty/false". This matters for conditions like `archived: false` (user wants non-archived) vs omitted (user does not filter on archive status).

### Anti-Patterns to Avoid
- **Global config variable:** Do not store config in a package-level `var`. Pass it explicitly from `main()` to consumers. This makes testing trivial.
- **Validation during parsing:** Keep YAML structural parsing (Decoder) separate from business validation (Validate). They produce different error types and are tested independently.
- **String enums without validation:** Do not leave `action` and `source` as unchecked strings. Validate them in `Validate()` against known values.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| YAML parsing | Custom YAML tokenizer/parser | go.yaml.in/yaml/v3 | YAML spec is complex; anchors, multiline strings, type coercion all handled |
| Unknown field detection | Manual reflection over struct fields | `Decoder.KnownFields(true)` | Built into yaml.v3, handles nested structs, returns clear errors |
| Config file reading | Custom buffered reader | `os.Open` + `yaml.NewDecoder` | Streaming decoder handles large files efficiently |

**Key insight:** The yaml.v3 `KnownFields` feature handles CONF-02's structural validation. Only semantic validation (enum checking, required fields, cross-field rules) needs custom code.

## Common Pitfalls

### Pitfall 1: Zero Values vs Absent Fields
**What goes wrong:** Using `bool` or `int` instead of `*bool` or `*int` means you cannot distinguish "user set false" from "user omitted the field". A rule with `archived: false` (match non-archived) looks identical to a rule that does not filter on archive status.
**Why it happens:** Go defaults are zero values. YAML unmarshaling fills omitted fields with zero values.
**How to avoid:** Use pointer types for all optional condition and exception fields. Check for `nil` (absent) vs dereferenced value (present).
**Warning signs:** Tests pass with explicitly-set values but fail when fields are omitted.

### Pitfall 2: KnownFields Does Not Apply to Node.Decode
**What goes wrong:** If you use custom `UnmarshalYAML(value *yaml.Node)` methods, `Node.Decode()` creates a new decoder internally without `KnownFields` enabled. Unknown fields in nested structs decoded via Node.Decode will NOT be caught.
**Why it happens:** Known issue in yaml.v3 (GitHub issue #460).
**How to avoid:** Avoid custom `UnmarshalYAML` methods when relying on `KnownFields`. Use plain struct tags and let the decoder handle everything. If custom unmarshaling is needed, validate unknown fields manually.
**Warning signs:** Unknown fields in nested structs (conditions, unless) silently pass.

### Pitfall 3: Partial Unmarshal on TypeError
**What goes wrong:** yaml.v3 returns `*yaml.TypeError` when fields cannot be decoded to the target type, but the struct is still partially populated. If you check the struct values after a TypeError, you get a mix of valid and invalid data.
**Why it happens:** yaml.v3 continues unmarshaling after type mismatches.
**How to avoid:** On any decode error, do NOT use the partially populated struct. Return the error immediately.
**Warning signs:** Config loads "successfully" with wrong types in some fields.

### Pitfall 4: Error Messages Without YAML Context
**What goes wrong:** Validation errors say "invalid action" without specifying which rule. Users with 20+ rules cannot find the problem.
**Why it happens:** Validation iterates rules but does not track the index.
**How to avoid:** Always include the field path: `rules[2].action`, `rules[0].conditions.source`. Pass the rule index through validation.
**Warning signs:** User complaints about unhelpful error messages.

### Pitfall 5: Forgetting to Validate Empty Rules List
**What goes wrong:** Config with `rules: []` or missing `rules` key parses successfully but produces a no-op application.
**Why it happens:** An empty slice is valid Go; yaml.v3 will not error.
**How to avoid:** Validate that `len(cfg.Rules) > 0` or at minimum emit a warning. Decide on policy: error or warn.
**Warning signs:** Application starts and does nothing.

## Code Examples

### Complete Config File (for test fixtures)
```yaml
# Source: CONTEXT.md decisions
timezone: America/New_York
schedule: "0 3 * * *"
rules:
  - name: old-rss-cleanup
    conditions:
      olderThan: 30
      source: rss
    unless:
      favourited: true
    action: archive

  - name: delete-ancient-archived
    conditions:
      olderThan: 90
      archived: true
    unless:
      hasTag: keep-forever
      hasNote: true
    action: delete
```

### Table-Driven Validation Test
```go
// Source: Go wiki TableDrivenTests pattern
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr []string // expected error substrings
    }{
        {
            name: "valid config",
            config: Config{
                Rules: []Rule{{
                    Conditions: &Conditions{OlderThan: intPtr(30)},
                    Action:     "archive",
                }},
            },
            wantErr: nil,
        },
        {
            name: "invalid action",
            config: Config{
                Rules: []Rule{{
                    Conditions: &Conditions{OlderThan: intPtr(30)},
                    Action:     "remove",
                }},
            },
            wantErr: []string{`rules[0].action: invalid value "remove"`},
        },
        {
            name: "invalid source",
            config: Config{
                Rules: []Rule{{
                    Conditions: &Conditions{Source: strPtr("feed")},
                    Action:     "archive",
                }},
            },
            wantErr: []string{`rules[0].conditions.source: invalid value "feed"`},
        },
        {
            name: "missing action",
            config: Config{
                Rules: []Rule{{
                    Conditions: &Conditions{OlderThan: intPtr(30)},
                }},
            },
            wantErr: []string{`rules[0]: missing required field "action"`},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            errs := tt.config.Validate()
            if len(tt.wantErr) == 0 {
                if len(errs) > 0 {
                    t.Errorf("unexpected errors: %v", errs)
                }
                return
            }
            for _, want := range tt.wantErr {
                found := false
                for _, err := range errs {
                    if strings.Contains(err.Error(), want) {
                        found = true
                        break
                    }
                }
                if !found {
                    t.Errorf("expected error containing %q, got %v", want, errs)
                }
            }
        })
    }
}

func intPtr(i int) *int    { return &i }
func strPtr(s string) *string { return &s }
```

### Strict Parsing Test (CONF-02)
```go
func TestStrictParsing_RejectsUnknownFields(t *testing.T) {
    input := `
timezone: UTC
schedule: "0 * * * *"
unknownField: bad
rules:
  - action: archive
    conditions:
      olderThan: 7
`
    var cfg Config
    dec := yaml.NewDecoder(strings.NewReader(input))
    dec.KnownFields(true)
    err := dec.Decode(&cfg)
    if err == nil {
        t.Fatal("expected error for unknown field, got nil")
    }
    if !strings.Contains(err.Error(), "unknownField") {
        t.Errorf("error should mention unknown field, got: %s", err)
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| gopkg.in/yaml.v3 | go.yaml.in/yaml/v3 | April 2025 | Old import path unmaintained; new path gets security fixes |
| yaml.v2 `SetStrict(true)` | yaml.v3 `KnownFields(true)` | yaml.v3 release | Different method name, same concept |
| go-yaml single maintainer | YAML org team maintenance | April 2025 | More sustainable long-term |

**Deprecated/outdated:**
- `gopkg.in/yaml.v3`: Unmaintained since April 2025. Use `go.yaml.in/yaml/v3` instead (same API, same code, different import path)
- `gopkg.in/yaml.v2`: Frozen, security fixes only. v3 API is different (Node-based)

## Open Questions

1. **Module path for go.mod**
   - What we know: The project is called karaclean, hosted presumably on GitHub
   - What is unclear: Exact GitHub org/user for the module path
   - Recommendation: Use a placeholder like `github.com/user/karaclean` and let the implementer set the real path; does not affect config package design

2. **Empty rules: error or warning?**
   - What we know: An empty rules list is technically valid YAML and Go
   - What is unclear: Whether users should be blocked from starting with no rules
   - Recommendation: Treat as validation error ("at least one rule required") -- a config with no rules serves no purpose and likely indicates a misconfiguration

3. **Schedule field validation depth in Phase 1**
   - What we know: `schedule` holds a cron expression; Phase 8 handles scheduling
   - What is unclear: Whether Phase 1 should validate cron syntax or just check non-empty
   - Recommendation: Phase 1 should validate that `schedule` is a non-empty string. Cron syntax validation belongs in Phase 8 when the cron library is added.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (Go 1.22+) |
| Config file | None needed -- `go test ./...` works out of the box |
| Quick run command | `go test ./internal/config/ -v -count=1` |
| Full suite command | `go test ./... -v -count=1` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CONF-01 | YAML file parsed into typed Go structs | unit | `go test ./internal/config/ -run TestLoad -v -count=1` | No -- Wave 0 |
| CONF-01 | Config discovery: flag > env > default | unit | `go test ./internal/config/ -run TestResolvePath -v -count=1` | No -- Wave 0 |
| CONF-02 | Unknown YAML fields rejected | unit | `go test ./internal/config/ -run TestStrictParsing -v -count=1` | No -- Wave 0 |
| CONF-02 | Invalid enum values rejected | unit | `go test ./internal/config/ -run TestValidate -v -count=1` | No -- Wave 0 |
| CONF-02 | Missing required fields rejected | unit | `go test ./internal/config/ -run TestValidate -v -count=1` | No -- Wave 0 |
| CONF-02 | All errors collected (not fail-fast) | unit | `go test ./internal/config/ -run TestValidate -v -count=1` | No -- Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/config/ -v -count=1`
- **Per wave merge:** `go test ./... -v -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `go.mod` -- module initialization (`go mod init`)
- [ ] `internal/config/config_test.go` -- covers CONF-01 (parsing, loading, discovery)
- [ ] `internal/config/validate_test.go` -- covers CONF-02 (strict parsing, enum validation, required fields, error collection)
- [ ] Test fixture YAML files in `internal/config/testdata/` -- valid and invalid config examples

## Sources

### Primary (HIGH confidence)
- [go.yaml.in/yaml/v3 on pkg.go.dev](https://pkg.go.dev/go.yaml.in/yaml/v3) -- v3.0.4, KnownFields API, Decoder type, TypeError behavior
- [gopkg.in/yaml.v3 on pkg.go.dev](https://pkg.go.dev/gopkg.in/yaml.v3) -- v3.0.1, same API reference (frozen)
- [YAML org go-yaml repository](https://github.com/yaml/go-yaml) -- maintenance status, v3 frozen/security-only, v4 RC
- [Go wiki: TableDrivenTests](https://go.dev/wiki/TableDrivenTests) -- standard test patterns
- [Go modules layout](https://go.dev/doc/modules/layout) -- cmd/internal project structure

### Secondary (MEDIUM confidence)
- [KnownFields issue #460](https://github.com/go-yaml/yaml/issues/460) -- KnownFields limitation with Node.Decode
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout) -- community project structure conventions

### Tertiary (LOW confidence)
- None -- all findings verified against official sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- go.yaml.in/yaml/v3 is the maintained fork, verified on pkg.go.dev with v3.0.4 published June 2025
- Architecture: HIGH -- standard Go cmd/internal layout with well-documented patterns
- Pitfalls: HIGH -- KnownFields limitation documented in official GitHub issues; pointer-type pattern is well-established Go idiom

**Research date:** 2026-03-18
**Valid until:** 2026-04-18 (30 days -- stable domain, yaml.v3 is frozen/security-only)
