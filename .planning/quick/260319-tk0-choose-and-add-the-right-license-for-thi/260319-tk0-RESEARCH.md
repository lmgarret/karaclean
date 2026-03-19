# Quick Task: Choose and Add License - Research

**Researched:** 2026-03-19
**Domain:** Open source licensing / Go dependency license compatibility
**Confidence:** HIGH

## Summary

Karaclean's Go dependencies use exclusively permissive licenses: MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, and BSD-0-Clause. There are zero copyleft (GPL/LGPL/AGPL/MPL) dependencies. This gives maximum freedom in license choice.

Karakeep itself is AGPL-3.0, but Karaclean only communicates with it over HTTP API -- there is no code linking, so AGPL does not propagate. Karaclean is an independent tool.

**Primary recommendation:** Use **MIT License**. It is the simplest permissive license, compatible with all dependency licenses, and is the most common choice in the Go ecosystem.

## Dependency License Audit

### Direct Dependencies (used at runtime)

| Module | License | Confidence |
|--------|---------|------------|
| `go.yaml.in/yaml/v3` | MIT + Apache-2.0 | HIGH |
| `github.com/robfig/cron/v3` | MIT | HIGH |
| `github.com/oapi-codegen/runtime` | Apache-2.0 | HIGH |

### Indirect Dependencies (transitive, pulled by oapi-codegen tooling)

| Module | License | Confidence |
|--------|---------|------------|
| `github.com/oapi-codegen/oapi-codegen/v2` | Apache-2.0 | HIGH |
| `github.com/getkin/kin-openapi` | MIT | HIGH |
| `github.com/apapsch/go-jsonmerge/v2` | MIT | HIGH |
| `github.com/dprotaso/go-yit` | MIT | HIGH |
| `github.com/go-openapi/jsonpointer` | Apache-2.0 | HIGH |
| `github.com/go-openapi/swag` | Apache-2.0 | HIGH |
| `github.com/google/uuid` | BSD-3-Clause | HIGH |
| `github.com/josharian/intern` | MIT | HIGH |
| `github.com/mailru/easyjson` | MIT | HIGH |
| `github.com/mohae/deepcopy` | MIT | HIGH |
| `github.com/oasdiff/yaml` | MIT + Apache-2.0 | HIGH |
| `github.com/oasdiff/yaml3` | MIT + Apache-2.0 | HIGH |
| `github.com/perimeterx/marshmallow` | MIT | HIGH |
| `github.com/speakeasy-api/jsonpath` | Apache-2.0 | HIGH |
| `github.com/speakeasy-api/openapi-overlay` | MIT | HIGH |
| `github.com/vmware-labs/yaml-jsonpath` | BSD-2-Clause | HIGH |
| `github.com/woodsbury/decimal128` | BSD-0-Clause | HIGH |
| `golang.org/x/mod` | BSD-3-Clause | HIGH |
| `golang.org/x/sync` | BSD-3-Clause | HIGH |
| `golang.org/x/text` | BSD-3-Clause | HIGH |
| `golang.org/x/tools` | BSD-3-Clause | HIGH |
| `gopkg.in/yaml.v2` | Apache-2.0 | HIGH |
| `gopkg.in/yaml.v3` | MIT + Apache-2.0 | HIGH |

### License Summary

| License | Count | Copyleft? |
|---------|-------|-----------|
| MIT | 12 | No |
| Apache-2.0 | 6 | No |
| MIT + Apache-2.0 (dual) | 4 | No |
| BSD-3-Clause | 5 | No |
| BSD-2-Clause | 1 | No |
| BSD-0-Clause | 1 | No |

**Zero copyleft dependencies.** No GPL, LGPL, AGPL, or MPL licenses present.

## License Recommendation

### MIT License (Recommended)

**Why MIT:**
- Compatible with ALL dependency licenses (MIT, Apache-2.0, BSD variants)
- Simplest and most widely understood permissive license
- Dominant license in Go ecosystem
- No patent clause complexity (unlike Apache-2.0)
- Users can freely use Karaclean in any context

### Apache-2.0 (Alternative)

Would also work. Adds explicit patent grant, which is useful for larger projects. Slightly more complex text. Compatible with all deps. Reasonable if the user prefers patent protection.

### Why NOT Other Licenses

| License | Reason to Skip |
|---------|---------------|
| GPL/AGPL | Unnecessary -- no copyleft deps require it, and it would restrict users |
| MPL-2.0 | File-level copyleft adds complexity with no benefit here |
| BSD-2/3 | Functionally similar to MIT but less common for new projects |
| Unlicense/CC0 | Not recommended for software (legal uncertainty in some jurisdictions) |

## Compatibility Rules

All permissive licenses are compatible with MIT:
- **MIT deps** -- MIT project can use freely, just include their copyright notices
- **Apache-2.0 deps** -- MIT project can use freely; Apache-2.0 is compatible with MIT (Apache code retains its own license)
- **BSD deps** -- fully compatible with MIT

### Attribution Requirements

When distributing the binary, the following are technically required:
1. Include Karaclean's own LICENSE file
2. For Apache-2.0 dependencies: retain NOTICE files if any exist (most Go deps don't have them)
3. For BSD/MIT dependencies: their license texts are embedded in the Go module cache but not in the compiled binary

**Practical note:** For a Docker scratch image distributing a static binary, the standard practice is to include a LICENSE file in the repo. Some projects also include a `THIRD_PARTY_LICENSES` file, but this is optional for MIT-licensed projects and uncommon for small Go tools.

## Gotchas

1. **Karakeep is AGPL-3.0** -- but Karaclean communicates only via HTTP API. No code linking occurs. AGPL does not propagate over network API boundaries to clients. Karaclean is a separate, independent work.

2. **go.yaml.in/yaml/v3 dual license** -- portions are MIT, portions are Apache-2.0. Both are permissive and compatible with an MIT-licensed project.

3. **`// indirect` annotation** -- all deps in go.mod are marked `// indirect` including the 4 that are actually direct (noted as known debt in PROJECT.md). This is cosmetic and does not affect licensing.

## Action Items for Implementation

1. Create `LICENSE` file in repo root with MIT License text
2. Set copyright holder and year: `Copyright 2026 [user's name or handle]`
3. Optionally add `license` field to any package metadata if applicable
4. No THIRD_PARTY_LICENSES file needed (standard practice for small MIT Go projects)

## Sources

### Primary (HIGH confidence)
- [github.com/robfig/cron LICENSE](https://github.com/robfig/cron/blob/v3.0.1/LICENSE) - MIT
- [github.com/oapi-codegen/oapi-codegen LICENSE](https://github.com/oapi-codegen/oapi-codegen/blob/main/LICENSE) - Apache-2.0
- [go.yaml.in/yaml/v3 on pkg.go.dev](https://pkg.go.dev/go.yaml.in/yaml/v3) - MIT + Apache-2.0
- [github.com/getkin/kin-openapi](https://github.com/getkin/kin-openapi) - MIT
- [github.com/PerimeterX/marshmallow](https://github.com/PerimeterX/marshmallow) - MIT
- [github.com/vmware-labs/yaml-jsonpath](https://github.com/vmware-labs/yaml-jsonpath) - BSD-2-Clause
- [github.com/woodsbury/decimal128](https://github.com/woodsbury/decimal128) - BSD-0-Clause
- [github.com/speakeasy-api/jsonpath](https://github.com/speakeasy-api/jsonpath) - Apache-2.0
- [github.com/apapsch/go-jsonmerge](https://github.com/apapsch/go-jsonmerge) - MIT
- [github.com/mohae/deepcopy](https://github.com/mohae/deepcopy) - MIT
- [github.com/karakeep-app/karakeep](https://github.com/karakeep-app/karakeep) - AGPL-3.0
- golang.org/x/* packages are BSD-3-Clause (standard Go project licensing)
