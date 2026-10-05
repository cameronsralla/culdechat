# Cul-de-Chat: Vision & Mission
Last Updated: September 30, 2026

Status: **stable enough to drive product/design** — still editable after review. This doc stays at the *why* and *what kind of thing this is*.

**Stack / architecture:** locked in [architecture.md](./architecture.md) — Docker-first community test, responsive PWA, classic server–client. Current repo is a mock/lab toward that target.

UI feel: best-guess pass applied to the mock; assumptions and next phases in [design-direction.md](./design-direction.md).

---

## Problem
The internet sorted people into global, low-context communities that compete with local life. Real-world social connection has thinned: fewer casual places to meet neighbors, less shared context on the same block or building. The infrastructure that fragmented us can also be aimed the other way — toward the people who actually share a place.

## Thesis (internal for now)
Cul-de-Chat is a shot across the bow at that stratification: a free, open-source tool that uses networked software to rebuild **local** community. A public manifesto can come later, after a real community instance is running and showing positive movement.

## Vision
**Build a modern digital town square for bringing local communities together.**

One line for product, docs, and design: private to a place, social in spirit, aimed at neighbors who share real geography — not a global feed of strangers.

## Mission
Give a real local community unit a trusted, self-contained social space — boards, a shared feed, private messages, and light official tools — exclusive to people who belong to that place. Open source from day one. Non-commercial. Runnable on modest hardware, with richer capabilities when the host has more resources. Provisioning helpers (not managed hosting) make standing up an instance easier.

## Who we serve first
**Residents** who want a better way to know and interact with neighbors. Operators/admins enable membership and stewardship; success is neighbor use.

## Proof path
1. Build toward a **test-bed instance** in our own apartment community (timing is a judgment call, not a fixed checklist date).
2. Run it, learn, look for positive movement.
3. Then consider growth: public manifesto / thesis, and ways for other communities to discover and run instances.

## What success looks like
**North star:** stronger local connection and community — neighbors who might not have bumped into each other meet, discover shared context, and keep interacting. Time-on-app and infinite-scroll habit are not success.

**Connection-validating signals** (examples; exact tracking defined at launch):
- **Jobs completed** — a real neighbor need resolved through the tool (ask answered, item found, coordination done)
- **Boards with cross-user life** — a resident-created board that draws posts/replies from others (shared interest, not a dead shelf)
- **Calendar with RSVPs** — events people commit to (intent to show up / coordinate in real life)

These are leading indicators of connection, not the goal themselves. Quiet periods are fine if important things still reach people and the community feels denser over months. Do not optimize a feed to inflate these artificially.

---

## Product shape

### Local by design
- One instance ↔ one local community unit (apartment/townhome complex, HOA/neighborhood, dorm, or other place-based group the operator defines).
- **Scale is practical:** whatever the operator is willing to support. Guidance lives in the Admin Guide; the product does not hard-code a max community size.
- **No global social graph.** Instances do not federate, share feeds, or interconnect users across communities.

### Free / open / non-commercial
- Application is **open source from day one** (already).
- Mothership / provisioning tooling may stay **private** and separately owned.
- No ads as the intended path. No engagement-maximizing machinery.
- Growth strategy deferred until the test-bed shows life.

### Capability tiers (working model)
Design so the **lowest practical hardware** can run a meaningful core, and stronger hosts unlock more.

| Tier | Intent | Examples (direction, not final) |
|---|---|---|
| **Core** | Must work on modest hosts | Auth, boards, feed, text posts/comments/reactions, DMs, directory, **email invites**, admin membership tools |
| **Standard** | Comfortable small server | Images (profile / posts), fuller retention, weekly digest |
| **Plus** | Stronger host or external services | Broader media (video, audio, albums), search, automated backups, etc. |

**Constraints & flexibility**
- **Email is Core** — onboarding depends on it.
- **Media direction:** eventually support **all useful media types** the deployment can handle; tiers and retention limit *what’s enabled*, not a permanent text-only product. On low resources, prefer strict lifetime / retention over a hard “no media” wall — exact rules TBD in the tech pass.
- A **system check** (deploy/runtime) should determine which capabilities to enable.
- **Architecture:** classic server/client is locked; see [architecture.md](./architecture.md). P2P parked as an appendix.

### Provisioning & mothership
- **Mothership helps you set up; it does not run your community for you.**
- Examples of intent: optional CNAME under a culdechat domain; tools to provision resources *in the operator’s accounts* (e.g. object storage bucket).
- Not in scope now: fully managed hosting of every instance.

### Install / first-run (product intent)
Setup should feel closer to an **OS install wizard** than a pile of env vars:
1. Choose how much to self-manage vs offload to mothership provisioning aids.
2. **Define the first admin** in that workflow.
3. Later: in-app tools to **promote additional admins**; possible **election** flows if we build governance for communities without a natural staff admin.

Default governance remains: at least one admin who can gate membership (invite / offboard).

---

## Product principles
1. **Local containers only** — one place, one instance.
2. **Connection over engagement** — neighbors meeting and coordinating beats time-on-app.
3. **Private by design** — membership gated; exclusivity is the point.
4. **Open within the walls** — boards/feed for discovery; DMs for 1:1.
5. **Trust and personal privacy** — directory opt-in; unit reach for essentials.
6. **Non-commercial** — clean tool; community service ethic.
7. **Core on small machines** — scale features (and retention) with available resources.
8. **Provisioning over babysitting** — make setup easy; operators own the instance.
9. **Get out of the way** — simple, intentional, calm when idle.
10. **Warm professionalism** — polished and clear; neighborly, not childish or corporate-cold.
11. **Teach the why** — operator docs lead with vision, then how-to.

## Out of scope / refuse
Authoritative avoid list (expanded): [non-goals.md](./non-goals.md).

- Cross-instance / global connection between communities
- Ads-driven design
- Engagement-optimized algorithmic feeds
- Mothership as default managed hosting for all communities
- Letting architecture debates block the higher-level plan

---

## Documentation we will grow

### Vision (this file)
Why it exists, non-goals, principles.

### Admin Guide (to create — living)
Intended structure:
1. **Why this exists** — vision/thesis for operators (first and foremost)
2. **Admin instruction manual** — run and govern the instance
3. **Resident-facing guidance** — what to tell users / simple user how-to

Build this **as we go**, not as a big-bang doc at the end.

### Architecture & stack
[architecture.md](./architecture.md) — target topology, Docker-first deploy, PWA client, future wizard/OS installers.

### Design direction
[design-direction.md](./design-direction.md) — UI feel answers.

### Backlog
[backlog.md](./backlog.md) — engineering build order. Product features: [feature-backlog.md](./feature-backlog.md).

### Proof path (infra)
First community test: **Docker Compose + public DNS + HTTPS** (see architecture). OS-specific installers and install wizard when going wide.

---

## Deferred (explicitly later)
- Exact capability-tier feature matrix and retention rules
- Mothership service catalog and install wizard **implementation** (shape outlined in architecture)
- OS-specific server installer packages
- Native mobile apps (PWA first)
- Deep P2P (parked; see architecture appendix)
- Admin election / voting governance
- Public manifesto and growth channels
- Hard “ready for building launch” checklist (owner gut feel)
- Exact success KPIs / instrumentation (direction above; numbers at launch)

## Related
- [architecture.md](./architecture.md)
- [../.cursor/rules/functional-spec.md](../.cursor/rules/functional-spec.md)
- [design-direction.md](./design-direction.md)
- [backlog.md](./backlog.md)
