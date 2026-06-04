# Security Policy

## Supported versions

Karaclean is distributed as a rolling release. Security fixes land on `main`
and the `latest` container image. Please run a recent build.

## Reporting a vulnerability

Please **do not** open a public issue for security problems.

Instead, report privately via GitHub's
[private vulnerability reporting](https://github.com/lmgarret/karaclean/security/advisories/new),
or email the maintainer at lm@codingarret.dev.

Include enough detail to reproduce the issue (affected version, configuration,
and steps). You can expect an initial acknowledgement within a few days. Once a
fix is available, we will coordinate disclosure with you.

## Scope

Karaclean talks to your Karakeep instance using an API key and can archive or
delete bookmarks. Of particular interest:

- Handling of the `KARAKEEP_API_KEY` and other secrets.
- Any path where dry-run or exception clauses could be bypassed, causing
  unintended deletions.
- Config parsing and validation issues.
