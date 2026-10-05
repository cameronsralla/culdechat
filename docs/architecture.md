# Cul-de-Chat: Architecture & Stack
Last Updated: October 4, 2026

Status: **stack locked and core framework built** (`server/`, `web/`, `deploy/`). Feature work lands on this framework next. Product why: [vision.md](./vision.md). Avoids: [non-goals.md](./non-goals.md).

---

## Purpose of this doc

Define **how the system is shaped** so product direction can be implemented without guessing. This is not a feature backlog and not an install manual.

---

## Repo layout

| Path | What it is |
|---|---|
| `server/` | **The product API.** Go module `github.com/cameronsralla/culdechat/server`. |
| `web/` | **The product client.** Vite + React + TypeScript responsive PWA. |
| `deploy/dev/` | Local Compose: Postgres, Mailpit, server (hot reload), web (Vite). `make up`. |
| `deploy/prod/` | Production Compose: Caddy (TLS) → web + server, Postgres, backup sidecar. |
| `lab/` | The earlier mock (Gin API, Expo app, old compose). Reference only; not shipped, not built by CI. |
| `docs/`, `.cursor/rules/` | Product and engineering docs. |

Treat `lab/` as a **lab**: useful for remembering UX decisions and endpoint shapes while features are ported to `server/` + `web/`. Nothing in it is required at runtime.

---

## Product → system mapping

| Product rule | Architecture consequence |
|---|---|
| One place, one community | **One deployable instance** = one community. No federation, no shared user graph across instances |
| Invite-only, admin-stewarded | Server holds membership; admin can invite / offboard |
| Connection over engagement | Chronological / simple feeds; no ranking machinery |
| Neighbors off-property still participate | Instance reachable on the **open internet** via **HTTPS + DNS** |
| Non-commercial, self-hosted | Operator (or steward) runs the instance; mothership is not in the runtime path |
| Modest hardware later | Capability tiers (Core / Standard / Plus) — exact matrix later; stack must stay lean |

**Topology:** classic **server–client**. The server is source of truth for membership, boards, feed, DMs, and (later) calendar/digest jobs.

**P2P:** Deep peer-to-peer mesh is **parked**. Optional later appendix: peer assist for media fan-out only — never authoritative membership or DM store on resident devices. See [Appendix: P2P](#appendix-p2p-parked).

---

## Locked stack

| Layer | Choice | Notes |
|---|---|---|
| Server API | **Go** — `chi` router, `log/slog`, `pgx` + `sqlc`, `goose` migrations | Small footprint; single static binary; cross-compiles when we add OS installers |
| Database | **PostgreSQL 16** | Docker-friendly; backups; concurrent community use |
| Media | **Local volume** on the instance | Object storage (S3-compatible) later / Plus |
| Edge / TLS | **Caddy** in Compose | Public HTTPS with real DNS; proxies `/api` to server, everything else to the static web image |
| Resident UI | **Vite + React 19 + TypeScript PWA** — React Router, TanStack Query, Tailwind v4 tokens, Radix primitives, `vite-plugin-pwa` | One responsive UI for phone and desktop; no native tooling |
| Auth | Email/password (argon2id) + invite link **and** out-of-band passcode; 15-min JWT access + 30-day rotating refresh tokens with reuse detection | Enough for invite-only; SSO later if needed |
| Mothership | **Not required at runtime** | Future: downloads, docs, optional DNS/provisioning aids |

**Non-choices (for now):** native App Store / Play apps, Electron server GUI, true P2P mesh, install wizard implementation, OS-specific installer packages.

---

## Near-term: Docker community test

First real community proof runs as **Docker Compose** on a host the operator controls, published with **public DNS and HTTPS**.

```text
Internet (HTTPS)
       │
       ▼
   [ Caddy ]  ←── public DNS points here
    │     │
    │     └──► web (nginx, static PWA build, SPA fallback)
    │
    └──► /api, /healthz, /readyz ──► server (Go) ──► PostgreSQL
                                        └──► media volume (+ SMTP out for invites)
```

### Container responsibilities

| Service | Responsibility |
|---|---|
| **caddy** | Terminate TLS (ACME); route `/api/*` + probes to server, everything else to web; HSTS |
| **server** | Auth, users/directory, settings, (next) boards, feed, DMs, media; embedded migrations run on boot |
| **web** | nginx serving the built PWA; hashed assets immutable, `sw.js`/`index.html` no-cache |
| **db** | PostgreSQL 16 with persistent volume |
| **backup** | Nightly `pg_dump` + media tarball into `BACKUP_DIR`, pruned after `BACKUP_KEEP_DAYS` |
| *(not a container)* | Mail is outbound SMTP from server — no local mail server in Core |

### Config surface

`deploy/prod/.env.example` is the operator contract (wizard writes the same file at go-wide):

- `DOMAIN`, `ACME_EMAIL` — public hostname and TLS contact  
- `DB_PASSWORD`, `JWT_SECRET` (≥32 bytes; server refuses shorter)  
- `SMTP_*` — required in prod (invites are Core)  
- `BOOTSTRAP_ADMIN_*` — first admin, applied only when the users table is empty  
- `BACKUP_DIR`, `BACKUP_KEEP_DAYS`  

Server-side validation: `APP_ENV=prod` requires https `PUBLIC_URL`, no `*` in `CORS_ORIGINS`, and SMTP configured. `TRUSTED_PROXIES` limits who may set `X-Forwarded-For` (Caddy's compose network).  

### Client requirements (near-term)

- **Responsive PWA** — `AppShell` swaps a sidebar (≥800px) for bottom tabs via one `useViewport()` hook; styling differences use Tailwind `md:`; UI controls are breakpoint-agnostic.  
- Installable to home screen where the browser allows.  
- No dependency on native store builds for the community test.  
- Native apps remain an optional later shell if iOS push proves necessary.

### Quality bar

- **Fast perceived UX** — snappy navigation and loads on modest links.  
- **Reliability:** single-node solid (supervised containers, backups) — not multi-node HA.  
- **Low-resource conscious:** prefer lean defaults; capability tiers and retention knobs come with feature work, not this doc.

---

## Server framework (how features are built)

Request pipeline, applied to every route in `internal/app/app.go`:

```text
RequestID → RealIP → Logger → Recover → Timeout → BodyLimit → SecurityHeaders → CORS
   → /api: Authenticate (bearer, never rejects) → RateLimit(per user, IP fallback)
        → feature routes, each guarded with RequireUser / RequireAdmin
```

- **Feature module** = one package under `internal/features/<name>/` exporting `Module` with `Routes(chi.Router)`, a `Service` (business rules, returns `httpx.Error`s), and SQL in `internal/db/queries/<name>.sql` compiled by `sqlc` into `internal/db/dbq`. Handlers bind → call service → `httpx.JSON` / `httpx.Fail`. Register the module in `app.New`.
- **Errors**: `httpx.Error{Status, Code, Message}`; `Fail` maps anything else to a logged 500. Clients always get `{"error":{"code","message"}}`.
- **Binding**: `httpx.Bind` rejects unknown fields and trailing data and calls `Validate()` on the request type.
- **Schema**: add `internal/db/migrations/0000N_*.sql` (goose up/down). Migrations are embedded and run on boot; the server refuses to start on a newer schema. `/readyz` fails while migrations are pending.
- **Settings**: `internal/settings` — JSONB `settings` table with a 30s in-process cache, public vs admin-only keys, per-key validators.
- **Tests**: `internal/app/app_test.go` boots the real `app.New` against a throwaway Postgres database (`TEST_DATABASE_URL`, default: the dev compose DB) and drives it with `httptest`. `make test`.

## Web framework (how screens are built)

- **Tokens** live in `web/src/theme/tokens.css` (`@theme`): colors, type scale, radii, shadows, breakpoint, layout widths. Tailwind utilities are generated from them; no hard-coded values in components.
- **Layout** owns responsiveness: `AppShell` (sidebar vs tab bar), `Screen` (compact header, content width, tab-bar padding), `AuthShell`. Pages supply content only.
- **UI kit** in `components/ui/` is breakpoint-agnostic: `Text`, `Button`, `Card`, `TextField`, `Avatar`, `Badge`, `Stack`, `ListRow`, `EmptyState`, `ErrorBanner`, `PageHeader`, `Fab`, `Toast`, `Icon`.
- **Data**: `api/client.ts` (bearer + single-flight refresh-on-401 + sign-out on failure), typed endpoint wrappers in `api/endpoints.ts`, TanStack Query for server state, `AuthProvider` for the session, route guards (`RequireUser`, `RequireAdmin`, `RequireAnonymous`).
- **Add a feature**: `features/<name>/<Name>Page.tsx`, a route in `app/router.tsx`, a nav entry in `layout/nav.ts` if top-level.
- **PWA**: `vite-plugin-pwa` with `registerType: 'prompt'`; precache shell only, never `/api`; `UpdatePrompt` offers a reload on new release.

---

## Cross-cutting decisions

Locked so feature work does not reinvent them. Each is single-node and lives inside the server container unless stated.

### Realtime
- **WebSocket from the Go API** (SSE acceptable fallback). No separate Socket.IO service or broker.
- One connection per client session; server fans out events: new DM, read receipt, reply to my post, unread counts, RSVP, bulletin.
- Lab may keep polling until the realtime slice ships; the contract above is the target.

### Background jobs
- **In-process scheduler in the API.** Postgres is the queue and lock (no Redis, no worker container in Core).
- Jobs: weekly digest, event reminders, retention/TTL sweeps, 30-day hard delete, news feed fetch, backup trigger hooks.
- Must be safe to restart mid-run; one leader when multiple API replicas exist (not expected in Core).

### Notification delivery
Ordered channels, each per-user configurable:
1. **In-app** — Activity inbox + badges over the WebSocket.
2. **Email** — immediate for DMs/replies (user choice) or batched into the digest.
3. **Web Push** — VAPID keys generated per instance by the operator/API; subscriptions stored in Postgres. iOS requires Home Screen install; we accept that limit.
Native APNs/FCM is out of scope (no native apps).

### PWA behavior
- Web app manifest + service worker whose jobs are **install** and **push** only.
- **Not offline-first.** Network-first; light cache of shell assets. Content requires connectivity.
- Service worker update: detect new release → prompt "Refresh for the latest" (no silent swaps mid-session).
- Client and API ship from the same release so versions always match.

### Media pipeline
- Uploads go **through the API** (multipart), never direct-to-disk from the client.
- Server-side: type sniffing, EXIF strip, resize to fixed variants (thumb / display / original), stored on the media volume.
- Serving: authenticated media routes through Caddy → API (membership checked). No public bucket.
- Limits by tier: max size, max count per post/DM, and retention TTL on Core. Video/audio only at Standard+.

### Configuration & settings
- **Deploy-time:** `.env` + Compose for secrets and infrastructure (DB creds, `JWT_SECRET`, domain, SMTP host).
- **Runtime:** a `settings` table in Postgres for operator-editable values (community name, invite policy, digest schedule, feature flags, tier, news feeds). Admin UI edits this.
- The future install wizard writes the same two stores; nothing new to invent at go-wide.

### Capability tier
- Operator sets **`CORE` / `STANDARD` / `PLUS`** in settings. Auto-detection of hardware is later.
- Features read the tier to gate media limits, retention, digest AI, video, search depth.

### Data lifecycle
- Users: soft delete on offboard → hard delete after 30 days (scheduled job).
- DMs: TTL sweep (default 6 months; operator-adjustable in settings).
- Media: TTL on Core; retained on Standard+ until the post is deleted.
- Resident export (own posts/DMs) is a later feature, not v1 architecture.

### Backups & restore
- Core Compose ships a **backup sidecar**: nightly `pg_dump` + media volume snapshot to a local backup directory (operator may point it at mounted external storage).
- Restore procedure is documented in the Admin Guide and tested before the community launch.
- Off-site copy is operator responsibility; object storage helpers are a Plus/mothership aid later.

### Migrations & upgrades
- **Versioned goose migrations** embedded in the server binary (`server/internal/db/migrations`).
- Server runs pending migrations on boot; refuses to start on an unknown newer schema.
- Images tagged per release; `docker compose pull && up` is the upgrade path.

### Observability
- Structured logs to stdout via `log/slog` — JSON in prod, text in dev; one line per request with `request_id`, status, duration, ip, user_id. No Loki/Grafana/Promtail in Core.
- `/healthz` (liveness) and `/readyz` (DB reachable, migrations current) — implemented.
- Admin **System status** page: version, tier, disk usage, last backup, last digest, job failures.

### Abuse & limits
- Credential endpoints (login, refresh, invite completion) are limited per IP (10/min, burst 5); all other `/api` routes per user (300/min, burst 60). Feature work adds per-action limits on posts, comments, DMs, uploads, invites.
- JSON bodies capped at 1 MiB; upload routes wrap with a larger limit. Request timeout 30s. Security headers + HSTS set by the server; Caddy handles TLS and HTTP→HTTPS.
- Invite-only remains the primary defense; no public signup surface exists.

---

## Go-wide later (reference only — do not build now)

When distributing beyond the first community:

### OS-specific server installs

Syncthing-like **native binaries** (Windows / macOS / Linux) that run the same server contract as Docker. Docker Compose remains a supported path (especially VPS).

### Install wizard (deferred)

First-run browser UI after install — intended fields (future):

1. Community name / identity  
2. First admin account  
3. Public base URL / how the world reaches the instance  
4. SMTP for invites  
5. Data directory / retention or tier preset  
6. Optional: tunnel vs “I already have DNS/HTTPS” (tunnel is optional aid; not required for the Docker community test)

GitHub clone/build remains available for operators who prefer it.

### Mothership boundary

| May do | Must not |
|---|---|
| Host download site and docs | Host community content or message bodies |
| Optional DNS / provisioning helpers in *operator* accounts | Be required for the instance to run |
| Optional anonymized update checks (if we ever add them) | Sell ads or build a global social graph |

---

## Lab → target evolution

1. ~~Lock the stack.~~ Done.  
2. ~~Build the core framework: server pipeline, auth, settings, web shell + kit, dev/prod Compose.~~ Done (users/auth/settings ship; `lab/` retained for reference).  
3. Port and redesign features onto the framework in product order: seeded boards + feed, profiles/onboarding, calendar, notifications (WebSocket + email + push), digest, post types, media. Then polls, search, news, trust & safety.  
4. Wizard + OS packages only when going wide.

Engineering backlog: [backlog.md](./backlog.md). Product features: [feature-backlog.md](./feature-backlog.md). Current code behavior: [../.cursor/rules/technical-spec.md](../.cursor/rules/technical-spec.md).

---

## Explicitly deferred

- Feature matrix (read receipts, media types, calendar, digest, …) — product passes  
- Capability-tier **auto-detection** (operator-set tier is locked above)  
- Install wizard implementation  
- OS-specific installer packages  
- Native mobile apps  
- Multi-replica API / HA  
- Object storage for media (Plus)  
- Resident data export  
- Deep P2P design  

---

## Appendix: P2P (parked)

We explored peer-to-peer for offloading resources. For Cul-de-Chat it conflicts with:

- Clear admin / invite / offboard  
- Digests, email invites, and always-reachable off-site neighbors  
- Simple backups and recovery on a small host  

**Revisit only if** media cost on a tiny host dominates and a **server-authoritative + peer media assist** model has a clear win. Not in the near-term stack.

---

## Related

- [vision.md](./vision.md)  
- [non-goals.md](./non-goals.md)  
- [feature-backlog.md](./feature-backlog.md)  
- [design-direction.md](./design-direction.md)  
- [admin-guide.md](./admin-guide.md)  
