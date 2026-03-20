# Roadmap: Karaclean

## Milestones

- ✅ **v1.0 MVP** — Phases 1–10 (shipped 2026-03-19)

## Phases

<details>
<summary>✅ v1.0 MVP (Phases 1–10) — SHIPPED 2026-03-19</summary>

- [x] Phase 1: Config Loading and Validation (2/2 plans) — completed 2026-03-18
- [x] Phase 2: API Client and Authentication (3/3 plans) — completed 2026-03-18
- [x] Phase 3: Age and Source Conditions (2/2 plans) — completed 2026-03-18
- [x] Phase 4: Status and Tag Conditions (2/2 plans) — completed 2026-03-18
- [x] Phase 5: Exception Evaluation (2/2 plans) — completed 2026-03-18
- [x] Phase 6: Actions and Dry-Run (2/2 plans) — completed 2026-03-18
- [x] Phase 7: Run Orchestrator and Observability (2/2 plans) — completed 2026-03-18
- [x] Phase 8: Scheduler and Deployment (2/2 plans) — completed 2026-03-19
- [x] Phase 9: Documentation (1/1 plan) — completed 2026-03-19
- [x] Phase 10: CI: Tests, Lint, Docker (2/2 plans) — completed 2026-03-19

Full archive: `.planning/milestones/v1.0-ROADMAP.md`

</details>

## Progress

| Phase | Milestone | Plans | Status | Completed |
|-------|-----------|-------|--------|-----------|
| 1. Config Loading and Validation | v1.0 | 2/2 | Complete | 2026-03-18 |
| 2. API Client and Authentication | v1.0 | 3/3 | Complete | 2026-03-18 |
| 3. Age and Source Conditions | v1.0 | 2/2 | Complete | 2026-03-18 |
| 4. Status and Tag Conditions | v1.0 | 2/2 | Complete | 2026-03-18 |
| 5. Exception Evaluation | v1.0 | 2/2 | Complete | 2026-03-18 |
| 6. Actions and Dry-Run | v1.0 | 2/2 | Complete | 2026-03-18 |
| 7. Run Orchestrator and Observability | v1.0 | 2/2 | Complete | 2026-03-18 |
| 8. Scheduler and Deployment | v1.0 | 2/2 | Complete | 2026-03-19 |
| 9. Documentation | v1.0 | 1/1 | Complete | 2026-03-19 |
| 10. CI: Tests, Lint, Docker | v1.0 | 2/2 | Complete | 2026-03-19 |

### Phase 1: Notification system: send per-rule action summaries to configurable channels (Slack, ntfy, Telegram, etc.) with global default channel and per-rule channel override

**Goal:** Add per-rule notification dispatch via Shoutrrr to configurable channels (Slack, ntfy, Telegram, etc.) with global default and per-rule override, best-effort delivery, and specific message format with counts
**Requirements:** [NOTIF-CFG, NOTIF-VAL, NOTIF-FMT, NOTIF-SEND, NOTIF-RUN, NOTIF-SILENT, NOTIF-FAIL]
**Depends on:** v1.0 (complete)
**Plans:** 3 plans

Plans:
- [ ] 01-01-PLAN.md — Config extension: Notifications struct, Notify field on Rule, Shoutrrr URL validation
- [ ] 01-02-PLAN.md — Notification engine: RuleSummary, FormatNotification, Notifier interface, channel resolution
- [ ] 01-03-PLAN.md — Integration: wire Run() with per-rule summaries, dispatch notifications, update main.go
