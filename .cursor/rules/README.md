## Cul-de-Chat: Cursor Rules & Project Specs

This directory contains living specifications for Cul-de-Chat. Keep these up to date as the product evolves so contributors can quickly understand scope and decisions.

### Files
- [Functional Requirements Specification](./functional-spec.md)
- [Technical Requirements Specification](./technical-spec.md)
- [API & Data Specifications](./api-data-spec.md) — schema/behavior summary (written against the `lab/` mock; update as features port to `server/`)
- [UI Screens Specification](./ui-screens.md) — screen inventory (written against the `lab/` Expo app; the product client is `web/`)

Product intent (living):
- [Vision & Mission](../../docs/vision.md) — why the product exists; drives major decisions
- [Architecture & stack](../../docs/architecture.md) — target topology; Docker-first + PWA; mock vs target
- [Admin Guide](../../docs/admin-guide.md) — operator why + manuals (stub; grow as we go)
- [Feature backlog](../../docs/feature-backlog.md) — high-leverage product features (calendar, templates, …)
- [Non-goals](../../docs/non-goals.md) — what we refuse while building
- [Competitive landscape](../../docs/competitive-landscape.md) — Nextdoor, Towne, WhatsApp, portals, …
- [Design Direction](../../docs/design-direction.md) — UI feel answers
- [Backlog](../../docs/backlog.md) — near-term engineering priority

Code layout: `server/` (Go API), `web/` (Vite + React PWA), `deploy/` (dev + prod Compose), `lab/` (earlier mock, reference only). Technical-spec describes what `server/` and `web/` do today; architecture.md holds the target topology and cross-cutting decisions.

### Conventions
- Each spec includes a "Last Updated" date. Update it when making substantive changes.
- Prefer incremental edits over large rewrites; preserve original intent where possible.
- Keep examples concise and runnable where applicable.


