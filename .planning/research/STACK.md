# Stack Research

**Domain:** Go Docker sidecar -- REST API client with YAML config and cron scheduling
**Researched:** 2026-03-18
**Confidence:** HIGH

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.24+ (build with 1.26.x) | Language runtime | Single static binary, zero runtime deps, tiny Docker image. Project constraint. |
| `net/http` (stdlib) | Go stdlib | HTTP client for Karakeep API | Go's stdlib HTTP client is production-grade. No need for third-party HTTP frameworks for a simple REST consumer. Keeps dependencies minimal. |
| `go.yaml.in/yaml/v3` | v3.0.4 | YAML config parsing | Maintained successor to `gopkg.in/yaml.v3` (which is archived/unmaintained since April 2025). Drop-in compatible. v4 exists but is still RC (v4.0.0-rc.4), not production-ready. |
| `github.com/netresearch/go-cron` | v0.13.1 | Cron scheduling | Maintained fork of `robfig/cron` (unmaintained since 2020, 50+ open PRs, critical panic bugs). Drop-in replacement with bug fixes for TZ parsing panics, DST handling, and Go 1.25+ support. |
| `log/slog` (stdlib) | Go 1.21+ stdlib | Structured logging | Standard library structured logging since Go 1.21. JSON and text output handlers built-in. No reason to pull in zerolog/zap for a sidecar this simple. |
| `encoding/json` (stdlib) | Go stdlib | JSON encode/decode for API payloads | Karakeep API speaks JSON. Stdlib is sufficient for this use case. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/stretchr/testify` | v1.10.x | Test assertions and mocking | Use `assert` and `require` packages for readable test assertions. Optional -- stdlib `testing` works fine, but testify reduces boilerplate for table-driven tests with complex structs. |
| `context` (stdlib) | Go stdlib | Request cancellation, timeouts | Every HTTP call and cron job should be context-aware for graceful shutdown. |
| `os/signal` (stdlib) | Go stdlib | Signal handling (SIGTERM/SIGINT) | Graceful shutdown in Docker (handle container stop signals). |
| `time` (stdlib) | Go stdlib | Duration parsing, age calculations | Rule matching on bookmark age (e.g., "older than 30d"). |
| `flag` (stdlib) | Go stdlib | CLI flags (--dry-run, --config) | Simple flag parsing. No need for cobra/urfave-cli for a sidecar with 3-4 flags. |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| `golangci-lint` | Linting | Run `golangci-lint run` in CI. Aggregates 50+ linters. Use default config to start. |
| `go vet` | Static analysis | Built-in, always run. |
| `go test -race` | Race detector | Run in CI to catch concurrency bugs in cron/HTTP interactions. |
| `goreleaser` | Build + release (optional) | Only if distributing outside Docker. For Docker-only, a simple Makefile suffices. |

## Installation

```bash
# Initialize module
go mod init github.com/your-org/karaclean

# Core dependencies
go get go.yaml.in/yaml/v3@v3.0.4
go get github.com/netresearch/go-cron@v0.13.1

# Test dependencies
go get github.com/stretchr/testify@latest
```

Everything else is Go standard library -- no additional dependencies needed.

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| `net/http` (stdlib) | `github.com/go-resty/resty` | If you want fluent builder-pattern HTTP client with built-in retry, backoff, middleware. Adds complexity for marginal benefit in this use case. |
| `go.yaml.in/yaml/v3` | `go.yaml.in/yaml/v4` (RC) | When v4 reaches stable release. It is the recommended future path but currently v4.0.0-rc.4 -- not production-ready. |
| `go.yaml.in/yaml/v3` | `github.com/goccy/go-yaml` | If you need better error messages with line/column info. More complex API. Overkill for config-file parsing. |
| `netresearch/go-cron` | `robfig/cron/v3` | Never -- robfig/cron is unmaintained since 2020 with known panic bugs. netresearch/go-cron is a drop-in replacement. |
| `netresearch/go-cron` | OS-level cron (no library) | If running outside Docker with system cron. Inside a container, embedding the scheduler is cleaner -- no cron daemon needed. |
| `log/slog` (stdlib) | `go.uber.org/zap` or `rs/zerolog` | If you need extreme logging performance (millions of log lines/sec). Irrelevant for a sidecar that logs a few hundred lines per run. |
| `flag` (stdlib) | `github.com/spf13/cobra` | If building a multi-command CLI with subcommands and shell completion. Karaclean has one command with a few flags -- cobra is overkill. |
| `encoding/json` (stdlib) | `github.com/goccy/go-json` or `github.com/bytedance/sonic` | If JSON decode performance is a bottleneck. It will not be -- API responses are small. |
| `stretchr/testify` | stdlib `testing` only | Viable. Testify is optional sugar. If you prefer zero test deps, stdlib table-driven tests work fine. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `gopkg.in/yaml.v3` | Archived and unmaintained since April 2025. No security patches. | `go.yaml.in/yaml/v3` (drop-in replacement) |
| `robfig/cron/v3` | Unmaintained since 2020. Known panic bugs in TZ parsing and chain decorators. 50+ unmerged PRs. | `github.com/netresearch/go-cron` (maintained fork, drop-in) |
| `github.com/spf13/viper` | Massive dependency tree for config management. Karaclean reads one YAML file -- viper's env/consul/etcd support is wasted complexity. | Direct `yaml.v3` unmarshal into a Go struct |
| `github.com/spf13/cobra` | Multi-command CLI framework. Karaclean is a single-purpose daemon, not a CLI toolkit. | `flag` stdlib package |
| `github.com/gin-gonic/gin` / `echo` / `fiber` | Web frameworks. Karaclean is an API *client*, not a server. | `net/http` client |
| `go.yaml.in/yaml/v4` | Still in RC (v4.0.0-rc.4 as of Jan 2026). API may change before stable. | `go.yaml.in/yaml/v3` until v4 is stable |
| Alpine base image | Adds ~5 MiB for a shell and musl libc that a static Go binary does not need. Larger attack surface. | `scratch` or `gcr.io/distroless/static-debian12` |

## Docker Image Strategy

### Recommended: Multi-stage build with `scratch`

```dockerfile
# Build stage
FROM golang:1.24-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /karaclean ./cmd/karaclean

# Runtime stage
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /karaclean /karaclean
ENTRYPOINT ["/karaclean"]
```

**Why scratch over distroless:** Go compiles to a fully static binary with `CGO_ENABLED=0`. No libc, no runtime needed. Scratch produces images under 10 MiB. The only extras needed are CA certificates (for HTTPS to Karakeep API) and timezone data (for cron schedule TZ support).

**Why not distroless:** `gcr.io/distroless/static-debian12` (~2 MiB base) is a reasonable alternative if you want pre-bundled CA certs and tzdata without manually copying them. Acceptable tradeoff for convenience. Either works.

### Build flags rationale

| Flag | Purpose |
|------|---------|
| `CGO_ENABLED=0` | Fully static binary, no libc dependency. Required for scratch. |
| `-ldflags="-s -w"` | Strip debug info and DWARF symbols. Reduces binary by ~30%. |
| `GOOS=linux` | Target Linux (container runtime). |

## Stack Patterns by Variant

**If adding a health endpoint later (e.g., for Docker healthcheck):**
- Use `net/http` stdlib server on a separate goroutine (e.g., `/healthz` on port 8080)
- Do not pull in a web framework for one endpoint

**If config grows complex enough to need env var overrides:**
- Use `os.Getenv()` for the 2-3 values that need override (API URL, API key, config path)
- Do NOT pull in viper. Env vars for secrets + YAML file for rules is sufficient.

**If migrating to yaml/v4 later:**
- Wait for v4.0.0 stable release
- Migration is straightforward: change import path, update struct tags if needed
- v3 and v4 share the same underlying repo (go-yaml/yaml)

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| `go.yaml.in/yaml/v3` v3.0.4 | Go 1.21+ | Maintained by yaml.org team |
| `netresearch/go-cron` v0.13.1 | Go 1.25+ | Requires Go 1.25 minimum. Build with Go 1.26.x to satisfy this. |
| `log/slog` | Go 1.21+ | Part of stdlib since 1.21 |
| `stretchr/testify` v1.10.x | Go 1.21+ | Widely compatible |

**Minimum Go version for this project: 1.25** (driven by netresearch/go-cron requirement).
**Recommended build toolchain: Go 1.26.x** (current stable as of March 2026).

## Sources

- [go.yaml.in/yaml/v3 on pkg.go.dev](https://pkg.go.dev/go.yaml.in/yaml/v3) -- verified v3.0.4, published Jun 2025 (HIGH confidence)
- [go.yaml.in/yaml/v4 on pkg.go.dev](https://pkg.go.dev/go.yaml.in/yaml/v4) -- verified v4.0.0-rc.4, pre-release (HIGH confidence)
- [netresearch/go-cron on GitHub](https://github.com/netresearch/go-cron) -- verified v0.13.1, published Mar 2026 (HIGH confidence)
- [netresearch/go-cron on pkg.go.dev](https://pkg.go.dev/github.com/netresearch/go-cron) -- verified version and Go 1.25+ requirement (HIGH confidence)
- [robfig/cron on GitHub](https://github.com/robfig/cron) -- confirmed unmaintained since 2020 (HIGH confidence)
- [Go slog official blog post](https://go.dev/blog/slog) -- stdlib structured logging since Go 1.21 (HIGH confidence)
- [Go release history](https://go.dev/doc/devel/release) -- Go 1.26.1 is current stable (HIGH confidence)
- [Alpine vs distroless vs scratch](https://medium.com/google-cloud/alpine-distroless-or-scratch-caac35250e0b) -- Docker image strategy (MEDIUM confidence)
- [Go ecosystem trends 2025 - JetBrains](https://blog.jetbrains.com/go/2025/11/10/go-language-trends-ecosystem-2025/) -- stdlib-first philosophy confirmed (MEDIUM confidence)
- [gopkg.in/yaml.v3 deprecation issues](https://github.com/go-task/task/issues/2171) -- confirmed unmaintained status (HIGH confidence)

---
*Stack research for: Go Docker sidecar (Karaclean)*
*Researched: 2026-03-18*
