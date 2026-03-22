# Milestones

## v1.3 Error Notification on Invalid Config (Shipped: 2026-03-22)

**Phases completed:** 1 phases, 1 plans, 2 tasks

**Key accomplishments:**

- NotifyOnError opt-in field with two-pass Load, lenient fallback parse, and SendConfigError dispatch to default channel via ConfigErrorNotifier interface

---

## v1.2 List-Based Exclusion (Shipped: 2026-03-22)

**Phases completed:** 2 phases, 6 plans, 11 tasks

**Key accomplishments:**

- StringOrSlice custom YAML type for inList fields on Conditions/Exceptions, structural validation, CollectListNames helper, and KarakeepAPI extended with ListLists/GetListBookmarks
- ListLists and GetListBookmarks API wrappers with cursor pagination, plus validateListNames startup check that fails fast on misconfigured list names
- inList OR-semantics matcher checks with preloaded list membership data wired through Run() to MatchesConditions and MatchesExceptions
- Config structs extended with Notifications/NotificationChannel types and Shoutrrr URL validation at startup via validateNotifications
- RuleSummary type, FormatNotification with conditional lines, Notifier interface with ShoutrrrNotifier, and ResolveChannelURL for channel resolution
- Per-rule notification dispatch wired into Run() with summary accumulation, channel resolution, and non-fatal error handling

---

## v1.1 Notifications (Shipped: 2026-03-20)

**Phases completed:** 1 phases, 3 plans, 0 tasks

**Key accomplishments:**

- (none recorded)

---

## v1.0 MVP (Shipped: 2026-03-19)

**Phases completed:** 10 phases, 20 plans, 0 tasks

**Key accomplishments:**

- (none recorded)

---
