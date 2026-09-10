## Cul-de-Chat

Private, modern town square for apartments and small communities. Community-driven and non‑commercial; focused on helping neighbors connect.

### Vision
Create a private, verified space for residents to discover boards by interest, post and comment openly, and strengthen local connection. Admins can publish official bulletins that stay pinned to the top of the feed.

### Core Tech
- **Backend**: Go (Gin)
- **API**: REST + Socket.IO for real-time
- **Database**: PostgreSQL
- **Frontend**: React Native (Expo, in `mobile/`)
- **Deployment**: Docker

### MVP Features
- **Boards & General Feed**: Interest-based boards; feed aggregates posts from all boards.
- **Posts, Comments, Reactions**: Threaded discussions and emoji reactions.
- **Admin Bulletins**: Pinned announcements; comments disabled.
- **Profiles & Directory**: Optional profile pictures; opt-in directory by name/unit.

### Fast Follow
- **Direct Messages**: One-on-one private messaging (v1.1).

### Specifications
Authoritative, living specs are kept in `.cursor/rules/`:
- [Functional Requirements Specification](.cursor/rules/functional-spec.md)
- [Technical Requirements Specification](.cursor/rules/technical-spec.md)
- [API & Data Specifications](.cursor/rules/api-data-spec.md)
- [UI Screens Specification](.cursor/rules/ui-screens.md)
- [Backlog](docs/backlog.md)

Generated HTTP docs (Swagger UI) are served at `/api/docs/index.html` when `CULDECHAT_DOCS=true`. Regenerate with `make docs` after changing handlers.

### Local development
```bash
docker compose -f infra/dev/docker-compose.yml up --build
cd mobile && npx expo start --web
make test
```

Local admin: `admin@culdechat.local` / `changeme123`. Seed fake residents with `make seed` (same password). The Expo web app talks to the API at `http://127.0.0.1:8080/api` (override with `EXPO_PUBLIC_API_URL`).

Postgres and the API bind to localhost only. The compose stack bootstraps a local admin (`admin@culdechat.local` / `changeme123`) and allows the local JWT secret via `CULDECHAT_ALLOW_INSECURE_JWT=true`. Invite emails are caught by Mailpit at [http://127.0.0.1:8025](http://127.0.0.1:8025) unless you put real SMTP settings in gitignored `infra/dev/.env` (`SMTP_HOST`, `SMTP_PORT`, `SMTP_FROM`, `SMTP_USER`, `SMTP_PASS`, `SMTP_STARTTLS=true`). Branding artwork lives in `assets/branding/logo.jpg`.

Optional TLS (needs a real hostname):

```bash
CULDECHAT_DOMAIN=your.domain docker compose -f infra/dev/docker-compose.yml --profile tls up --build
```

### Operations
- HTTPS (Let's Encrypt), JWT auth, bcrypt for password hashing.
- Logs via Promtail/Loki/Grafana; daily PostgreSQL backups recommended.

### License
See [LICENSE](./LICENSE).

