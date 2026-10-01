# Cul-de-Chat: Technical Requirements Specification
Last Updated: September 16, 2026

## 1. Core Architecture
The application will be a containerized system running in a Docker environment on a local, self-hosted server.

- **Backend**: Go (Gin framework)
- **Database**: PostgreSQL
- **Frontend**: React Native (Expo, `mobile/`)
- **API Style**: REST
- **Real-time**: Socket.IO
- **Deployment**: Docker containers

## 2. Platform & Deployment
- **Target Platform**: Desktop web first (Expo web), then the same codebase on iOS/Android. Layout is adaptive: sidebar above 800px, bottom tabs below.
- **Hosting**: Self-hosted on a server within the apartment complex. Requires network configuration (static IP or DDNS) and physical security.

## 3. Backend Architecture
- **Framework**: Gin for the REST API in Go.
- **API documentation**: OpenAPI/Swagger 2.0 is generated from handler comments (`make docs`) and served at `/api/docs/index.html` only when `CULDECHAT_DOCS=true`.
- **Sessions**: Short-lived access JWTs plus hashed refresh tokens (30 days). Auth middleware reloads the user from the database on every request. `AdminRequired` authenticates and checks `is_admin` without nesting `AuthRequired` (which would call `Next` and run the handler before the admin check).
- **Real-time Features**: Socket.IO will be implemented on the Go backend to manage real-time messaging and notifications (not in MVP).

## 4. Database & Data Management
- **Database**: PostgreSQL running in a Docker container.
- **Media Storage**: User-uploaded files will be stored on the local server's filesystem, with strict backend validation for file type and size.
- **Email**: Outbound mail goes through `Mailer` (`SMTPMailer`) using one community mailbox (`SMTP_HOST` / `SMTP_PORT` / `SMTP_FROM`, plus `SMTP_USER` / `SMTP_PASS` when the provider requires auth). Callers build a `MailMessage` (to, subject, plain-text body) and `Send`. Invites use `InviteMail`. STARTTLS is used when credentials are set or `SMTP_STARTTLS=true`. Local development uses Mailpit as the SMTP sink. Do not send as the logged-in admin's personal inbox (no per-admin OAuth). Optional `CULDECHAT_APP_URL` adds a `/register?token=` link on invites.
- **Data Retention Policies**:
  - **User Data**: A soft delete policy will be used. Data is flagged as inactive for 30 days before a scheduled job performs a permanent hard delete.
  - **Chat Messages**: A Time-to-Live (TTL) of 6 months will be enforced via a scheduled job.

## 5. Frontend Architecture
- **Framework**: React Native (Expo) in `mobile/`. Expo Router for screens. Local preview via `npx expo start --web`.
- **Styling**: One theme object (`src/theme/theme.ts`) wrapped by `ThemeProvider`. Components and screens style through `useTheme` / `useStyles` and the UI kit (`Button`, `Card`, `Stack`, …). Do not hard-code colors or type sizes in pages.
- **State Management**: React Context API + Hooks. `AuthProvider` holds the session.
- **Page shell**: Authenticated screens render inside `AppShell`. Wide viewports use a left sidebar (no duplicate top header); compact viewports use a slim header + bottom tabs. Login is outside the shell. Nav tabs: Home, Boards, People, Messages, You; Admin is inserted before You when `is_admin`. Ionicons replace the old unicode glyphs. The Admin nav item and `/admin` page are shown only when the session user is `is_admin`.
- **Direct messages**: REST under `/api/messages`. Conversations are unique per user pair (`user_low_id` / `user_high_id`). Open threads poll about every 5s; Socket.IO remains a follow-up.

## 6. Authentication & Security
- **Login Method**: Standard Email & Password (`POST /api/auth/login`).
- **Session Management**: Access JWT plus refresh token stored in SecureStore on native and `localStorage` on web. The API client refreshes on 401 and hydrates `/auth/me` on launch.
- **Security MVP**:
  - All traffic will be served over HTTPS (using a Let's Encrypt certificate) in front of Gin (optional Caddy profile).
  - User passwords will be hashed using bcrypt and must be at least 8 characters.
  - `JWT_SECRET` must be set to a unique value; the process refuses the well-known default unless `CULDECHAT_ALLOW_INSECURE_JWT=true`.
  - Invite tokens and refresh tokens are stored hashed. Login, complete-registration, and refresh are rate limited (local compose sets `CULDECHAT_RATE_LIMIT=off` so `make seed` can create several accounts).
  - The API allows CORS from local Expo web origins (`http://localhost` / `http://127.0.0.1` any port, plus `CULDECHAT_CORS_ORIGINS`).
  - Postgres and the API bind to localhost in local compose. Optional Caddy (`--profile tls`) terminates HTTPS when `CULDECHAT_DOMAIN` is set.

## 7. Operations & Maintenance
- **Local seed**: `make seed` (`infra/dev/seed.sh`) invites and completes a handful of fake residents (Maya, Jordan, Priya, Sam, Riley, plus `seeduser@example.com`). Password is `changeme123`. Most opt into the directory; Riley stays hidden so unit-number DMs can be tested. Safe to re-run.
- **Initial Scale**: The system will be architected for an initial load of ~100 users.
- **Logging**: The PLG Stack (Promtail, Loki, Grafana) will be used for a self-hosted, real-time log monitoring solution.
- **Backups**: A daily, automated backup of the PostgreSQL database is strongly recommended. This can be achieved with a simple cron job in a Docker container that runs pg_dump.


