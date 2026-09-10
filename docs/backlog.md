# Cul-de-Chat: Backlog
Last Updated: September 9, 2026

Living list of work after the current MVP slice. Behavior lives in `.cursor/rules/`; this file is priority and status only. When an item ships, mark it **done** here and update the matching spec.

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
| One-to-one DMs | planned | Person-to-person threads |
| Start a chat from a person | planned | Directory row, author on a post |
| Message-by-unit | planned | Lands on the **primary** for that unit (see Households). Hidden members stay hidden |
| Report via DM to an admin | planned | Formal report table can wait |
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
2. v1.1 DMs (person + unit-to-primary, then realtime / badge)
3. v1.2 Households
4. Then (reset, search, post images, real admin, host + backups)
