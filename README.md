## Cul-de-Chat

Private, modern town square for apartments and small communities. Community-driven and non‑commercial; focused on helping neighbors connect.

### Vision
Create a private, verified space for residents to discover boards by interest, post and comment openly, and strengthen local connection. Admins can publish official bulletins that stay pinned to the top of the feed.

### Core Tech
- **Backend**: Go (Gin)
- **API**: REST + Socket.IO for real-time
- **Database**: PostgreSQL
- **Frontend**: React Native (iOS/Android)
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

### Operations
- HTTPS (Let's Encrypt), JWT auth, bcrypt for password hashing.
- Logs via Promtail/Loki/Grafana; daily PostgreSQL backups recommended.

### License
See [LICENSE](./LICENSE).

