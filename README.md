<p align="center">
  <img src="assets/branding/logo.jpg" alt="Cul-de-Chat logo" width="120" />
</p>

<h1 align="center">Cul-de-Chat</h1>

<p align="center">
  <strong>A modern digital town square for bringing local communities together.</strong>
</p>

Cul-de-Chat is a free, open-source, non-commercial app for **one place at a time** — an apartment building, townhome complex, HOA, dorm, or similar community. It gives neighbors a private, invite-only space to know each other, coordinate, and build real local connection.

It is meant to replace the messy mix of building group chats, HOA email, and city-scale social apps with something smaller and more trusted: a shared feed and boards, a resident directory, private messages, and light admin tools — owned and run by the community, not monetized with ads or engagement ranking.

**Who it’s for:** residents first. An admin (or steward) invites people in and keeps membership tidy; success is neighbors actually using it to meet, help, and show up for each other.

**What it’s not:** a global social network, a Nextdoor clone, or property-management software. One instance serves one community. No cross-community feed. No ads.

---

### Product docs
- Vision & mission: [docs/vision.md](docs/vision.md)
- **Architecture & stack:** [docs/architecture.md](docs/architecture.md) — Docker-first community test, responsive PWA, classic server–client
- Feature backlog: [docs/feature-backlog.md](docs/feature-backlog.md)
- Non-goals: [docs/non-goals.md](docs/non-goals.md)
- Competitive landscape: [docs/competitive-landscape.md](docs/competitive-landscape.md)
- Design direction: [docs/design-direction.md](docs/design-direction.md)
- Admin guide (stub): [docs/admin-guide.md](docs/admin-guide.md)

### Repo layout
| Path | What |
|---|---|
| `server/` | Go API — chi, slog, pgx + sqlc, goose migrations, argon2id + JWT/refresh auth |
| `web/` | Responsive PWA — Vite, React 19, TypeScript, React Router, TanStack Query, Tailwind v4 tokens |
| `deploy/dev/` | Local Compose: Postgres, Mailpit, server (hot reload), web (Vite) |
| `deploy/prod/` | Community instance: Caddy (auto TLS) → web + server, Postgres, nightly backup sidecar |
| `lab/` | The earlier Gin + Expo mock. Reference only. |

The core framework (auth, users/directory, admin roster + invites, settings, logging, hardening, layout shell, component kit) is built. Boards, feed, calendar, messages, and the rest land on it next in the order set by [docs/feature-backlog.md](docs/feature-backlog.md).

### Specifications
Living specs are kept in `.cursor/rules/`:
- [Technical Requirements Specification](.cursor/rules/technical-spec.md) — what `server/` and `web/` do today
- [Functional Requirements Specification](.cursor/rules/functional-spec.md)
- [API & Data Specifications](.cursor/rules/api-data-spec.md)
- [UI Screens Specification](.cursor/rules/ui-screens.md)
- [Engineering backlog](docs/backlog.md)

### Local development
Requires Docker. Go 1.26 and Node 24 only if you want to run things outside Compose.

```bash
make up          # db + mailpit + server (:8080) + web (:5173)
make test        # Go suite against the dev Postgres
make web-test    # Vitest
make lint        # go vet, eslint, tsc
make sqlc        # regenerate internal/db/dbq after editing queries/*.sql
```

Local admin: `admin@culdechat.local` / `changeme123`. Invite emails land in Mailpit at [http://127.0.0.1:8025](http://127.0.0.1:8025); in dev the invite response also includes the registration link. Postgres and the server bind to localhost only. Set `WEB_PORT` if 5173 is taken.

### Running a community instance
```bash
cd deploy/prod
cp .env.example .env   # set DOMAIN, ACME_EMAIL, DB_PASSWORD, JWT_SECRET, SMTP_*, BOOTSTRAP_ADMIN_*
docker compose up -d --build
```
Point your domain's A/AAAA record at the host; Caddy obtains the certificate. Backups are written nightly to `BACKUP_DIR`. See [docs/admin-guide.md](docs/admin-guide.md).

### License
See [LICENSE](./LICENSE).
