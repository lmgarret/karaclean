# Roadmap: Karaclean

## Milestones

- ✅ **v1.0 MVP** — Phases 1–10 (shipped 2026-03-19)
- ✅ **v1.1 Notifications** — Phase 01 (shipped 2026-03-20)

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

### Phase 1: List-based bookmark exclusion — exclude bookmarks from rule actions if they belong to specified lists (e.g. exclude 'Read Later' list from archive cleanup rules)

**Goal:** Add list-based filtering to rules: `conditions.inList` to target bookmarks in specific lists, and `unless.inList` to protect bookmarks in specific lists from a rule's action. Lists referenced by name, preloaded from Karakeep API.
**Requirements**: D-01, D-02, D-03, D-04, D-05, D-06, D-07, D-08, D-09, D-10, D-11, D-12, D-13
**Depends on:** Phase 0
**Plans:** 3 plans

Plans:
- [ ] 01-01-PLAN.md — Config types (StringOrSlice, InList fields), structural validation, API interface extension
- [ ] 01-02-PLAN.md — API client wrappers (ListLists, GetListBookmarks) and ValidateListNames startup step
- [ ] 01-03-PLAN.md — Matcher inList integration, Run() preloading, end-to-end tests
