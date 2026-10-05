# Cul-de-Chat: Backlog
Last Updated: October 1, 2026

Living list of **engineering / near-term delivery** work. Behavior lives in `.cursor/rules/` (lab). **Stack locked:** [architecture.md](./architecture.md) — Docker-first community test + responsive PWA; current repo is mock/lab. Product feature wishlist: [feature-backlog.md](./feature-backlog.md). Avoid list: [non-goals.md](./non-goals.md). Competitors: [competitive-landscape.md](./competitive-landscape.md).

**Next major lane after stack docs:** feature/build work against the locked Docker + PWA architecture (not OS installers or wizard yet).

When an item ships, mark it **done** here and update the matching spec.

Status: **next** (do now) · **planned** (scheduled) · **later** (only if people ask) · **skip** · **done**

## Now — finish the MVP
API already exists for most of this. Missing UI or polish.

| Item | Status | Notes |
|---|---|---|
| Profile photo upload in the app | done | Web file picker; native picker later |
| Edit / delete own post or comment | done | Edit is author-only on comments; delete is author or admin |
| Pin / unpin a standard post | done | Admin; bulletins stay pinned |
| Invite copy + register deep link | done | Copy token/passcode/link; `/register?token=` pre-fills |
| Bulletin vs pin in the feed | done | Separate badges |
| Empty states, errors, pull-to-refresh | done | Feed, boards, board detail, directory, You |
| Relative timestamps | done | Cards, post detail, comments |

## v1.1 — Direct messages
| Item | Status | Notes |
|---|---|---|
| One-to-one DMs | done | Single thread per pair; create on first send |
| Messages tab + inbox | done | Nav tab; FAB to compose |
| Start a chat from a person | done | Directory row, author on a post |
| Search visible people + hidden units | done | `/messages/recipients`; unit-only for opted-out |
| Message-by-unit | done | Lands on the active resident for that unit (primary later) |
| Report via DM to an admin | planned | Formal report table can wait; admins are regular DM peers |
| Realtime for new DMs | planned | WebSocket from the API (architecture); polling is fine for the first slice |
| In-app notification / badge | planned | Email-on-message can follow |

## v1.2 — Households
Primary resident per unit can invite family members onto that unit.

| Item | Status | Notes |
|---|---|---|
| Primary + member users | planned | One `users` row per person. `primary_user_id` / household link on members |
| Admin invite creates a primary | planned | Current invite flow, unchanged |
| Primary invites members by email | planned | Unit inherited. Members cannot invite. Cap: 4 members |
| Primary cannot be changed | planned | No transfer, no promote-from-member |
| Offboard primary cascades | planned | All members (active or pending) offboarded, sessions revoked, unit freed |
| Offboard member only | planned | Primary stays |
| Unique unit among primaries only | planned | Replace one-user-per-unit index |
| Admin roster grouped by unit | planned | Primary on top, members indented; offboard-unit + remove-member |
| Members are full residents | planned | Post, comment, boards, directory opt-in, DMs. Not admins |

If the leaseholder leaves and someone else stays, admin offboards the household and invites the remaining person as a **new** primary.

## Then — make it last
| Item | Status | Notes |
|---|---|---|
| Forgot password / admin reset | planned | Invites cover day one; lockout is week three |
| Search posts, boards, people | planned | Simple `q=` is enough at ~100 residents |
| Images on posts | planned | For Sale, lost items |
| Admin roster lifecycle | planned | Resend invite, change unit, promote/demote admin, pending vs active |
| First-week onboarding | planned | After register: photo, join General, say hi |
| Password / invite email copy | planned | Distinct member-invite copy from admin invite |
| Installable client | planned | **PWA** (manifest + service worker for install/push); no store apps for the community test |
| Production host | planned | Production Compose: Caddy/HTTPS, API, Postgres, PWA static, backup sidecar |
| Runtime settings table + admin settings page | planned | Community name, tier, digest schedule, feature flags (architecture) |
| Versioned migrations | planned | Replace auto-migrate before the community test |
| Health endpoints + admin System status | planned | `/healthz`, `/readyz`, version/disk/last backup |
| Per-user rate limits (posts, DMs, uploads, invites) | planned | Auth limits already exist |
| 30-day hard-delete after offboard | planned | Soft-delete already lands |
| Chat message TTL (6 months) | planned | Spec already calls for a job |

## Later
| Item | Status | Notes |
|---|---|---|
| Nested comment threads | later | |
| Mentions | later | |
| Polls | later | |
| Events calendar | later | A board + dates is enough at first |
| Web Push (VAPID) | later | In-app badge + email first; iOS needs Home Screen install |
| Board icons / covers | later | |
| Read receipts / typing | later | |
| Moderation queue / word filters | later | Report → admin DM first |
| Promtail / Loki / Grafana | later | After a box that stays up |
| i18n | later | |

## Skip
| Item | Status | Notes |
|---|---|---|
| Dev Admin role | skip | Not needed until a real instance has a second operator |
| Shared unit login / one password | skip | Households are separate people |
| Members inviting members | skip | |
| Transfer or promote primary | skip | Offboard household, invite a new primary |
| Admin account as a household | skip | Staff is not a unit |

## Suggested order
1. Now — done (photos, edit/delete/pin, invite copy, feed polish)
2. Vision — done enough in [vision.md](./vision.md)
3. **Architecture & stack** — locked in [architecture.md](./architecture.md) (Docker-first + PWA)
4. **Next:** feature/build against that stack (production Compose hardening, PWA, then product features)
5. Admin Guide — living stub in [admin-guide.md](./admin-guide.md); grow as features land
6. v1.1 DMs leftovers — unread badge, report-to-admin, realtime
7. Capability-tier / system-check design (with feature work)
8. v1.2 Households
9. Go-wide only: OS installers, install wizard, mothership provisioning aids
