# Cul-de-Chat: Backlog
Last Updated: September 30, 2026

Living list of **engineering / near-term delivery** work. Behavior lives in `.cursor/rules/`. Product feature wishlist (calendar, templates, polls, …): [feature-backlog.md](./feature-backlog.md). Avoid list: [non-goals.md](./non-goals.md). Competitors: [competitive-landscape.md](./competitive-landscape.md).

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
| Realtime for new DMs | planned | Socket.IO; polling is fine for the first slice |
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
| Native / installable client | planned | PWA or Expo sideload; same codebase |
| Production host | planned | Real hostname, Caddy/HTTPS, daily Postgres backups |
| 30-day hard-delete after offboard | planned | Soft-delete already lands |
| Chat message TTL (6 months) | planned | Spec already calls for a job |

## Later
| Item | Status | Notes |
|---|---|---|
| Nested comment threads | later | |
| Mentions | later | |
| Polls | later | |
| Events calendar | later | A board + dates is enough at first |
| Push notifications | later | In-app badge first |
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
2. **Vision & mission** — lock [vision.md](./vision.md) before deep UX (design pass paused in [design-direction.md](./design-direction.md))
3. **Admin Guide** — living stub in [admin-guide.md](./admin-guide.md); lead with why, then admin + resident manuals as features land
4. v1.1 DMs leftovers — unread badge, report-to-admin, Socket.IO
5. UX deep-dive resume — per design-direction phases (Messages → feed → People/You…)
6. Capability-tier / system-check design (core vs media-heavy hosts)
7. v1.2 Households
8. Then (reset, search, post images, real admin, host + backups, mothership provisioning tools)
