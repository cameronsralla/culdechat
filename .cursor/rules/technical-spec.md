# Cul-de-Chat: Technical Requirements Specification
Last Updated: October 4, 2026

**Framing:** This file describes **what the code in `server/` and `web/` does today**. Target topology, locked stack, and cross-cutting decisions are authoritative in [docs/architecture.md](../../docs/architecture.md). The earlier Gin + Expo mock lives in `lab/` for reference only.

## 1. Core Architecture
- **Backend**: Go 1.26, `server/` — `chi` router, `log/slog`, `pgx/v5` + `sqlc` (generated package `internal/db/dbq`), `goose` embedded migrations.
- **Database**: PostgreSQL 16 (`citext` for emails, `pgcrypto` for UUIDs).
- **Frontend**: `web/` — Vite 7, React 19, TypeScript, React Router 7, TanStack Query 5, Tailwind v4 (`@theme` tokens), Radix primitives, `vite-plugin-pwa`, Vitest + Testing Library.
- **API style**: REST JSON under `/api`. Errors are always `{"error":{"code","message"}}`.
- **Deployment**: `deploy/dev` Compose for development; `deploy/prod` Compose (Caddy → web + server, db, backup) for a community instance behind public DNS + HTTPS.

## 2. Platform & Deployment
- **Client**: responsive PWA. `AppShell` renders a sidebar at ≥800px and a bottom tab bar below (`useViewport`). Installable; `registerType: 'prompt'` shows an update banner on new releases; network-first (no API caching).
- **Dev**: `make up` starts db (`127.0.0.1:5432`), Mailpit (`:8025`), server with `air` hot reload (`:8080`), web Vite (`:5173`, override `WEB_PORT`). Local admin `admin@culdechat.local` / `changeme123`. `RATE_LIMIT=false` in dev.
- **Prod**: copy `deploy/prod/.env.example` → `.env`, set `DOMAIN`, `ACME_EMAIL`, `DB_PASSWORD`, `JWT_SECRET`, `SMTP_*`, `BOOTSTRAP_ADMIN_*`; `docker compose up -d`. Server image is distroless static; web image is nginx with SPA fallback.

## 3. Backend Architecture
- **Wiring**: `internal/app.New` opens the pool, runs migrations, bootstraps the admin (only when `users` is empty), builds services, and mounts modules. `cmd/server/main.go` adds the `http.Server` with timeouts and SIGTERM graceful shutdown.
- **Middleware order**: `RequestID → RealIP(TRUSTED_PROXIES) → Logger → Recover → Timeout → BodyLimit(1 MiB) → SecurityHeaders(HSTS in prod) → CORS(exact allow-list)`; under `/api`: `Authenticate` (parses bearer, never rejects) → `RateLimit.ByUser`.
- **Feature module pattern**: `internal/features/<name>/{handler.go,service.go}` + `internal/db/queries/<name>.sql`. `Module.Routes(chi.Router)` applies `auth.RequireUser` / `auth.RequireAdmin` per group. Services return `httpx.Error` values (`ErrBadRequest`, `ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, `ErrConflict`, …) with `.WithMessage()` / `.Wrap()`; handlers call `httpx.Fail`.
- **Binding**: `httpx.Bind` requires `application/json`, rejects unknown fields and trailing data, and calls `Validate()` when the request type implements it (`httpx.Fields` collects field errors).
- **Config** (`internal/config`): env only. Required: `DATABASE_URL`, `JWT_SECRET` (≥32 bytes). Prod additionally requires https `PUBLIC_URL`, no `*` in `CORS_ORIGINS`, and `SMTP_HOST`/`SMTP_FROM`. `CORS_ORIGINS` defaults to `PUBLIC_URL`.
- **Mail** (`internal/mail`): SMTP with optional STARTTLS/auth; when `SMTP_HOST` is unset the message is logged instead of sent (dev).
- **Health**: `/healthz` liveness; `/readyz` pings the DB and fails while migrations are pending.
- **Realtime / jobs**: not yet built. Targets (WebSocket from the server, in-process scheduler with Postgres locks) are in the architecture doc.

## 4. Database & Data Management
- **Schema** (`00001_init.sql`): `users` (email citext unique, unit_number, display_name, password_hash nullable until activated, is_admin, status `invited|active|inactive`, directory_opt_in), `invites` (token_hash, passcode_hash, expires_at, consumed_at), `refresh_tokens` (token_hash, family_id, expires_at, revoked_at, user_agent, ip), `settings` (key, JSONB value, updated_by).
- **Migrations**: add `internal/db/migrations/0000N_name.sql` with `-- +goose Up` / `-- +goose Down`. Run on boot; server refuses to start if the DB is ahead of the binary.
- **Queries**: write SQL in `internal/db/queries/*.sql` with sqlc annotations, run `make sqlc`. Never hand-write scan code.
- **Settings** (`internal/settings`): keys `community_name`, `capability_tier` (`CORE|STANDARD|PLUS`), `invite_expiry_hours`; each has a validator. Public keys readable by any resident at `GET /api/settings`; full list/edit admin-only at `/api/admin/settings`. 30s in-process cache, invalidated on write.
- **Media**: not yet built (media volume and `MEDIA_DIR` are provisioned).
- **Retention**: not yet built. Targets: soft-delete → 30-day hard delete, DM TTL 6 months.

## 5. Frontend Architecture
- **Tokens**: `src/theme/tokens.css` is the only place colors, type scale, radii, shadows, breakpoints, and layout widths are defined. Components use the generated utilities (`bg-brand`, `text-title`, `rounded-md`, `w-sidebar`, …). `lib/cn.ts` wraps `tailwind-merge` and is taught the custom type scale.
- **Layout** (`src/layout/`): `AppShell` (sidebar vs tabs), `Sidebar`, `TabBar`, `Header` (compact only), `Screen` (page frame: compact header, content width, tab-bar padding), `AuthShell`, `LoadingScreen`, `nav.ts` (single nav list; admin item filtered by role).
- **UI kit** (`src/components/ui/`): `Text` (variants display/title/subtitle/body/label/caption), `Button` (primary/secondary/ghost/danger; loading; asChild), `Card`, `TextField` (label/hint/error wired for a11y), `Avatar`, `Badge`, `Stack`, `ListGroup`/`ListRow`, `EmptyState`, `ErrorBanner`, `PageHeader`, `Fab`, `Toast` (`useToast`), `Icon`, `Logo`. Kit components are breakpoint-agnostic.
- **Data**: `api/client.ts` attaches the bearer token, refreshes once on 401 (single-flight), retries, and signs out if refresh fails. Tokens persist in `localStorage` (`culdechat.session`). `api/endpoints.ts` holds typed wrappers; `api/types.ts` mirrors server shapes. TanStack Query for server state.
- **Auth**: `AuthProvider` restores the session on boot via `GET /api/me`; guards `RequireUser`, `RequireAdmin`, `RequireAnonymous` in `auth/guards.tsx`.
- **Routes**: `/login`, `/register?token=` (invite completion), `/` (Square), `/directory`, `/you`, `/admin`. Add pages under `src/features/<name>/`.

## 6. Authentication & Security
- **Passwords**: argon2id (m=64 MiB, t=3, p=2), PHC string. Minimum 10 characters. Unknown-user logins burn a dummy verification to equalize timing.
- **Sessions**: HS256 access JWT (15 min, `sub` = user id, `adm` flag) + opaque refresh token (32 random bytes, SHA-256 hashed at rest, 30 days). Refresh **rotates** within a family; presenting a revoked token revokes the whole family (reuse detection). Logout revokes the device family; `logout-all` and password change revoke every session.
- **Invites**: admin creates an invited user → emailed link `PUBLIC_URL/register?token=…` **plus** a 6-digit passcode shown to the admin for out-of-band delivery. Completion requires token + passcode, sets password and display name, consumes the invite. Token hashes only in the DB; `invite_url` is returned in the API response only in dev.
- **Guards**: `RequireUser` (401) and `RequireAdmin` (403). Admins cannot deactivate or demote themselves. Deactivation revokes all refresh tokens.
- **Rate limits**: credential routes 10/min burst 5 per IP; other `/api` routes 300/min burst 60 per user (IP fallback). `429` with `Retry-After`.
- **Headers**: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, `Cache-Control: no-store`, restrictive CSP, `Permissions-Policy`; HSTS in prod.
- **CORS**: exact-origin allow-list from `CORS_ORIGINS`; methods GET/POST/PUT/PATCH/DELETE/OPTIONS; headers Authorization, Content-Type; exposes Content-Length, X-Request-Id; credentials off (bearer tokens); preflight cached 10 min; disallowed preflights get 403.

## 7. Operations & Maintenance
- **Tests**: `make test` runs the Go suite (`internal/app/app_test.go`) against a throwaway `culdechat_test` database on the dev Postgres (`TEST_DATABASE_URL` to override). `make web-test` runs Vitest; `make lint` runs `go vet`, eslint, and `tsc`.
- **Logging**: `slog` to stdout — text in dev, JSON in prod; one `http` line per request with `request_id`, method, path, status, bytes, `dur_ms`, ip, `user_id`. Panics are recovered and logged with a stack.
- **Backups**: `deploy/prod` `backup` service runs on start and nightly at 03:00: `pg_dump --format=custom` + `media-*.tar.gz` into `BACKUP_DIR`, pruned after `BACKUP_KEEP_DAYS`.
- **Upgrade**: rebuild/pull images, `docker compose up -d`; migrations apply on server boot.
- **Initial scale**: ~100 users on a single node.
