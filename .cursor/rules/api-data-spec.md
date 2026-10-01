# Cul-de-Chat: API & Data Specifications
Last Updated: September 16, 2026

Interactive, generated API docs live with the server at `/api/docs/index.html` when `CULDECHAT_DOCS=true`. They are produced from handler annotations via `make docs`. Treat Swagger as the request/response source of truth; this file is the schema and behavior summary.

## 1. Database Schema (PostgreSQL)

Assumptions:
- `id` is the primary key on entity tables
- `created_at` and `updated_at` timestamps exist on all entity tables

### users
Stores information about each resident.

- `id` (uuid) - Primary Key
- `unit_number` (varchar) - The resident's unit number. Unique among `active` and `pending` users so a unit can be reassigned after offboarding.
- `email` (varchar, unique) - Used for login and notifications.
- `name` (varchar) - Display name, set when the resident completes registration.
- `hashed_password` (varchar) - The securely hashed password (placeholder hash until registration is completed).
- `profile_picture_url` (varchar, nullable) - Link to their profile picture.
- `is_directory_opt_in` (boolean, default: false) - If true, their name/unit are public in the directory.
- `is_admin` (boolean, default: false) - Differentiates Business Admins.
- `status` (varchar, default: 'active') - `active`, `inactive` (soft delete), or `pending`.
- `invite_token` (varchar, nullable, unique) - SHA-256 hex of the registration token. Cleared after completion.
- `passcode_hash` (varchar, nullable) - bcrypt hash of the invite passcode.
- `invite_expires_at` (timestamptz, nullable) - Invite TTL (7 days).

### refresh_tokens
Hashed refresh sessions. Revoked on logout, logout-all, password change, and offboard.

- `id` (uuid) - Primary Key
- `user_id` (uuid) - Foreign Key to `users.id`
- `token_hash` (varchar, unique) - SHA-256 of the opaque refresh token
- `expires_at` (timestamptz) - 30 days from issue
- `revoked_at` (timestamptz, nullable)
- `created_at` (timestamptz)

### boards
Stores the user-created communities.

- `id` (uuid) - Primary Key
- `creator_id` (uuid) - Foreign Key to `users.id`
- `name` (varchar, unique) - The name of the board (e.g., "Dog Lovers").
- `description` (text, nullable) - A short description of the board.

Creating a board auto-subscribes the creator.

### posts
The individual threads started on a board.

- `id` (uuid) - Primary Key
- `author_id` (uuid) - Foreign Key to `users.id`
- `board_id` (uuid) - Foreign Key to `boards.id`
- `title` (varchar) - The title of the post.
- `content` (text) - The body of the post.
- `post_type` (varchar, default: `standard`) - `standard` or `bulletin`.
- `is_pinned` (boolean, default: false) - Pinned posts sort first on feeds. Bulletin posts are always pinned.

### comments
Replies to a specific post.

- `id` (uuid) - Primary Key
- `author_id` (uuid) - Foreign Key to `users.id`
- `post_id` (uuid) - Foreign Key to `posts.id`
- `content` (text) - The body of the comment.

Comments are rejected on bulletin posts.

### board_subscriptions (junction)
Tracks which users are subscribed to which boards.

- `user_id` (uuid) - Foreign Key to `users.id`
- `board_id` (uuid) - Foreign Key to `boards.id`
- Primary Key: composite (`user_id`, `board_id`)

### post_reactions (junction)
Tracks user reactions to posts. One reaction per user per post.

- `id` (uuid) - Primary Key
- `user_id` (uuid) - Foreign Key to `users.id`
- `post_id` (uuid) - Foreign Key to `posts.id`
- `type` (varchar) - One of `like`, `love`, `laugh`, `wow`, `sad`, `angry`.
- Unique: (`post_id`, `user_id`)

### conversations
Unique 1:1 DM thread between two users. Pair is stored as ordered (`user_low_id`, `user_high_id`) so there is only one row per pair.

- `id` (uuid) - Primary Key
- `user_low_id` (uuid) - Foreign Key to `users.id` (lexicographically smaller id)
- `user_high_id` (uuid) - Foreign Key to `users.id`
- Unique: (`user_low_id`, `user_high_id`)

### messages
Text messages inside a conversation. Rows are created only when someone sends; the conversation is created with the first message.

- `id` (uuid) - Primary Key
- `conversation_id` (uuid) - Foreign Key to `conversations.id`
- `sender_id` (uuid) - Foreign Key to `users.id`
- `content` (text) - Plain text, max 2000 characters

---

## 2. REST API Endpoints

All community endpoints require `Authorization: Bearer <jwt>` unless noted. Errors use `{ "error": "message" }`.

List endpoints that return posts are cursor-paginated: `limit` (default 20, max 50) and opaque `cursor`. Response includes `next_page_cursor` when more rows exist. Pinned posts appear first, then newest.

On posts and comments, `author.name` is only present when that user has opted into the directory. `unit_number` is always shown.

### Meta

#### GET /api/health
Public. `{ "status": "ok" }`

#### GET /api/docs/*
Swagger UI, only when `CULDECHAT_DOCS=true`.

Login, complete-registration, and refresh are rate limited (5/15min for login, 10/15min for complete-registration and refresh; disabled when `CULDECHAT_RATE_LIMIT=off`, which local compose sets). Passwords must be at least 8 characters. Invite passcodes are 10 characters from `abcdefghijkmnpqrstuvwxyz23456789`. Access tokens are 1-hour JWTs; `refresh_token` is an opaque 30-day token returned on login, complete-registration, and refresh. Auth middleware reloads the user from the database and requires `status=active`. Residents may post only on boards they are subscribed to (admins may post anywhere). `my_reaction` is included on feed and post detail. JSON bodies are capped at 2 MiB. Profile picture URLs must be `http(s)` or `/api/media/...`. Unique constraint violations return 409 `"already exists"`.

### Authentication

#### POST /api/auth/register
Admin-only. Creates a `pending` user, emails the invite when SMTP is configured, and always returns the token and passcode so the admin can share them if mail fails. Re-inviting an existing pending email refreshes the token and passcode.

Local compose sends through Mailpit (`http://127.0.0.1:8025`). A real instance uses one community mailbox over SMTP (`SMTP_HOST`, `SMTP_PORT`, `SMTP_FROM`, optional `SMTP_USER`/`SMTP_PASS`). `email_sent` is true when the message was accepted by the mail server. Invites go through the shared mailer (`InviteMail` → `Mailer.Send`); other outbound messages should do the same.

Request Body:

```json
{
  "email": "new.resident@example.com",
  "unit_number": "101"
}
```

Response Body (201 Created):

```json
{
  "message": "Invite emailed to new.resident@example.com",
  "email": "new.resident@example.com",
  "registration_token": "uuid",
  "passcode": "a3k9wm2p7x",
  "invite_expires_at": "timestamp",
  "email_sent": true
}
```

#### POST /api/auth/complete-registration
Public. Resident sets name and password using the stub token and passcode, then receives access and refresh tokens.

Request Body:

```json
{
  "token": "uuid",
  "passcode": "a3k9wm2p7x",
  "password": "user_password",
  "name": "Alex Rivera"
}
```

Response Body (200 OK): same as login.

#### POST /api/auth/login
Public. Authenticates an `active` user.

Request Body:

```json
{
  "email": "resident@example.com",
  "password": "user_password"
}
```

Response Body (200 OK):

```json
{
  "token": "your_jwt_token_here",
  "refresh_token": "opaque_refresh_token",
  "user": {
    "id": "user_uuid",
    "name": "Alex Rivera",
    "unit_number": "101",
    "is_admin": false
  }
}
```

#### POST /api/auth/refresh
Public. Rotates the refresh token and issues a new access JWT. Body: `{ "refresh_token": "..." }`. Response: same as login.

#### POST /api/auth/logout
Public. Revokes one refresh token. Body: `{ "refresh_token": "..." }`. 204.

#### POST /api/auth/logout-all
Revokes every refresh token for the current user. 204.

#### POST /api/auth/change-password
Body: `{ "current_password": "...", "new_password": "..." }`. Revokes all refresh tokens. 204.

#### GET /api/auth/me
Returns the current user.

### Boards

#### GET /api/boards
Lists all boards, including `subscriber_count` and `is_subscribed` for the current user.

#### GET /api/boards/{boardId}
One board with `subscriber_count` and `is_subscribed`.

#### POST /api/boards
Creates a board. The creator is subscribed automatically.

Request Body:

```json
{
  "name": "Book Club",
  "description": "Let's read and discuss!"
}
```

#### POST /api/boards/{boardId}/subscribe
Toggles subscription. Response: `{ "subscribed": true, "message": "Successfully subscribed to the board." }`

### Posts

#### GET /api/posts
General feed. Pinned posts first, then newest. Each item includes title, snippet, author, board, counts, `post_type`, `is_pinned`, `created_at`.

#### GET /api/boards/{boardId}/posts
Same page structure, filtered to one board.

#### POST /api/boards/{boardId}/posts
Creates a post. Admins may set `post_type` to `bulletin` (auto-pinned, comments disabled) or `is_pinned` to pin a standard post.

Request Body:

```json
{
  "title": "New book for September!",
  "content": "We'll be reading 'The Midnight Library'. First meeting is next Tuesday."
}
```

#### GET /api/posts/{postId}
Full post, comments, reaction counts, and `my_reaction`. Bulletin posts return `comments_disabled: true` and an empty comments list.

#### PATCH /api/posts/{postId}
Author or admin. Updatable: `title`, `content`, `is_pinned` (admin). Bulletins cannot be unpinned. Title ≤200, content ≤10000.

#### DELETE /api/posts/{postId}
Author or admin. 204.

#### POST /api/posts/{postId}/comments
Adds a comment. Forbidden on bulletin posts. Content ≤2000.

#### GET /api/posts/{postId}/comments
Chronological comments.

#### PATCH /api/posts/{postId}/comments/{commentId}
Author or admin. Body: `{ "content": "..." }`.

#### DELETE /api/posts/{postId}/comments/{commentId}
Author or admin. 204.

#### PUT /api/posts/{postId}/reactions
Sets the current user's reaction (`type` required). 204 No Content.

#### DELETE /api/posts/{postId}/reactions
Removes the current user's reaction. 204 No Content.

### Profile & Directory

#### GET /api/profile/me
#### PATCH /api/profile/me
Updatable fields: `name` (≤80), `profile_picture_url` (`http(s)` or `/api/media/...`), `directory_opt_in`.

#### POST /api/profile/me/photo
Multipart field `photo`. jpeg/png/webp, max 2 MiB. Stores under `CULDECHAT_UPLOAD_DIR` and sets `profile_picture_url` to `/api/media/profile/{filename}`.

#### GET /api/media/profile/{filename}
Authenticated. Serves an uploaded profile photo.

#### GET /api/directory
Active users who opted in: `id`, `name`, `unit_number`, `profile_picture_url`.

### Direct messages

One conversation per pair of users. Created on the first sent message. Text only (≤2000). No edit/delete in this slice. Hidden (directory opt-out) peers expose unit only — name and photo are omitted. Messaging a unit number delivers to the active resident for that unit (today: the primary / only resident).

#### GET /api/messages/conversations
Inbox for the current user. Each row: `id`, `peer`, `last_message`, `updated_at`.

#### GET /api/messages/conversations/{id}
Conversation peer plus recent messages (default 50). Optional `before` message id loads older history.

#### POST /api/messages
Send. Body requires `content` and exactly one of `user_id` or `unit_number`. Returns `{ conversation_id, message, created }`.

#### GET /api/messages/recipients?q=
Search to start a chat. Visible users match name or unit (`kind: "user"` with id/name/photo). Hidden residents match unit only (`kind: "unit"`, no id/name/photo).

### Admin

#### GET /api/admin/users
Roster of residents: id, email, name, unit, status, is_admin, directory opt-in.

#### POST /api/admin/users/{userId}/offboard
Soft-deletes a non-admin user (`status=inactive`) so the unit can be reassigned. Revokes their refresh tokens; subsequent requests with an old access JWT fail because auth reloads status from the database. Hard delete of personal data remains a later retention job.

### Local bootstrap
If no admin exists, migrate/API startup will create one when `BOOTSTRAP_ADMIN_EMAIL` and `BOOTSTRAP_ADMIN_PASSWORD` are set (`BOOTSTRAP_ADMIN_NAME`, `BOOTSTRAP_ADMIN_UNIT` optional). This is a development stub until a real admin lifecycle exists.
