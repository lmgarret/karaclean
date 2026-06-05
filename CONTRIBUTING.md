# Contributing to Karaclean

Thanks for your interest in improving Karaclean! Contributions of all kinds are
welcome -- bug reports, feature ideas, documentation fixes, and code.

## Getting started

```sh
git clone --recurse-submodules https://github.com/lmgarret/karaclean.git
cd karaclean
go build ./...
```

The `karakeep-upstream` git submodule pins the Karakeep source used as a
reference for the API client. Run `git submodule update --init` if you cloned
without `--recurse-submodules`.

## Before you open a pull request

Run the same checks CI runs:

```sh
go test -race ./...
~/go/bin/golangci-lint run ./...   # golangci-lint v2.11
```

- **Tests must pass** with `-race`. Add tests for new behavior.
- **Lint must be clean.** Config lives in `.golangci.yml`.
- **Update docs in the same change.** If you add or modify a feature, update
  `README.md` and `karaclean.example.yaml` -- don't defer docs to a follow-up.

## Project conventions

- Config parsing uses `go.yaml.in/yaml/v3` (the maintained fork, not
  `gopkg.in`) with `KnownFields(true)` for strict parsing.
- Optional config fields use pointer types to distinguish nil from zero value.
- Matching is case-sensitive throughout.
- Keep changes safety-first: dry-run and exception clauses exist to protect
  users' bookmarks.

## Commit and PR style

Keep commits focused and write a clear description of *what* changed and *why*.
Link any related issues. That's it -- no rigid template required.

The release changelog is generated from commit messages since the previous tag, so
clear, descriptive commit subjects make for a better changelog.

## Releasing

Releases are tag-driven. To cut a release (maintainers):

```sh
git tag v1.4.0
git push origin v1.4.0
```

Pushing a `vX.Y.Z` tag triggers [`.github/workflows/release.yml`](.github/workflows/release.yml),
which:

1. Builds the multi-arch (`amd64` + `arm64`) Docker image.
2. Publishes the tag ladder to GHCR: `1.4.0`, `1.4`, `1`, and `latest`.
3. Generates a changelog from the commits since the previous tag and creates a
   GitHub Release.

**Prerequisite:** the changelog step uses an LLM and requires an `ANTHROPIC_API_KEY`
repository secret (Settings > Secrets and variables > Actions). Builds on `main`
(not tags) publish the rolling `edge` tag and a commit-`<sha>` tag instead.
