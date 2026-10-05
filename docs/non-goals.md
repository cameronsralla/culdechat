# Cul-de-Chat: Non-Goals & Avoid List
Last Updated: October 1, 2026

Reference while building. If a feature request pulls toward these, pause and check [vision.md](./vision.md) and [architecture.md](./architecture.md).

---

## Hard avoids (product identity)

| Avoid | Why |
|---|---|
| **Cross-instance / global social graph** | One instance = one local community. No federation, shared feed, or “nearby neighborhoods” discovery |
| **Ads / paid boosts / business payola** | Non-commercial; recommendations must stay neighbor-trust, not bought placement |
| **Engagement-optimized algorithmic ranking** | Connection over addiction; chronological / simple ranking preferred |
| **Notification volume as growth tactic** | Utility alerts only; Nextdoor’s own turnaround includes cutting noisy pings |
| **Crime-feed / “suspicious person” culture as a default** | Damages trust; Nextdoor’s reputation problem; not our mission |
| **Mothership as full managed hosting for every community** | Provisioning aids yes; babysitting every deploy no (unless economics radically change) |
| **Mothership required at instance runtime** | Instance must run standalone after deploy; mothership is optional aids only ([architecture.md](./architecture.md)) |

---

## Soft avoids (usually wrong for us)

| Avoid | Why | Exception |
|---|---|---|
| Becoming a **property-ops OS** (rent, work orders, packages desk) | That’s BuildingLink/Funnel territory; dilutes neighbor connection | Tiny optional modules only if a community begs |
| **Star ratings / complaint boards** for local businesses | Invites toxicity; Towne explicitly avoids this | Positive “I’d call again” recommendations OK |
| **Anonymous posting** by default | Weakens accountability in a small place | Privacy via directory opt-out + unit identity is enough |
| **City-wide or interest-only communities with no place** | Breaks “local container” | Operator can define unit; we guide toward place-based |
| **Heavy gamification** (streaks, points, leaderboards) | Fun level 3 — warm, not arcade | Light badges only if they serve stewardship (maybe later) |
| **Opaque private group chats as the main social layer** | Cliques / side channels undercut the open town square | 1:1 DMs stay; N>2 groups **parked** — only revisit with transparency rules (see feature backlog) |
| **Native store apps as a requirement for the first community test** | Self-hosted URL + responsive PWA is the locked client path | Optional native shell later if iOS push proves necessary |
| **True P2P mesh as the primary architecture** | Breaks simple admin, digests, off-site reach, backups | Server stays source of truth; peer media-assist only as a parked appendix |

---

## Design / UX avoids (from design-direction)

- Sound effects  
- Emoji-heavy chrome (tabs, badges)  
- Purple SaaS / generic AI aesthetic clichés  
- Features that only increase clicks without helping two neighbors help or meet each other  

---

## When tempted anyway

Ask: *Does this make it easier for people who share a place to help each other, meet, or coordinate — or does it only increase engagement metrics / expand beyond the local container?*

If the second: don’t build it (or park under later with explicit rationale).
