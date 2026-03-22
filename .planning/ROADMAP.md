# Roadmap: Karaclean

## Milestones

- ✅ **v1.0 MVP** — Phases 1–10 (shipped 2026-03-19)
- ✅ **v1.1 Notifications** — Phase 01 (shipped 2026-03-20)
- ✅ **v1.2 List-Based Exclusion** — Phase 01 (shipped 2026-03-22)

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

<details>
<summary>✅ v1.1 Notifications (Phase 01) — SHIPPED 2026-03-20</summary>

- [x] Phase 01: Notification System (3/3 plans) — completed 2026-03-20

Full archive: `.planning/milestones/v1.1-ROADMAP.md`

</details>

<details>
<summary>✅ v1.2 List-Based Exclusion (Phase 01) — SHIPPED 2026-03-22</summary>

- [x] Phase 01: List-based bookmark exclusion (3/3 plans) — completed 2026-03-22

Full archive: `.planning/milestones/v1.2-ROADMAP.md`

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
| 01. Notification System | v1.1 | 3/3 | Complete | 2026-03-20 |
| 01. List-based bookmark exclusion | v1.2 | 3/3 | Complete | 2026-03-22 |

### Phase 1: Error notification on invalid config

**Goal:** Send error notification to default channel when config validation fails at startup, toggleable via notifyOnError field, with lenient fallback for YAML syntax errors
**Requirements**: [ERRNOTIF-01, ERRNOTIF-02, ERRNOTIF-03, ERRNOTIF-04]
**Depends on:** Phase 0
**Plans:** 1 plan

Plans:
- [ ] 01-01-PLAN.md — NotifyOnError field, two-pass Load, lenient fallback, SendConfigError, tests, and docs
