# Cul-de-Chat: API & Data Specifications
Last Updated: August 30, 2025

## 1. Database Schema (PostgreSQL)
A relational database like PostgreSQL is perfect for this. Here’s a logical breakdown of the tables we'll need for the MVP. We'll use a simplified notation here to show columns and relationships.

Assumptions:
- `id` is the primary key on entity tables
- `created_at` and `updated_at` timestamps exist on all tables

### users
Stores information about each resident.

- `id` (uuid) - Primary Key
- `unit_number` (varchar) - The resident's unit number.
- `email` (varchar, unique) - Used for login and notifications.
- `hashed_password` (varchar) - The securely hashed password.
- `profile_picture_url` (varchar, nullable) - Link to their profile picture.
- `is_directory_opt_in` (boolean, default: false) - If true, their name/unit are public.
- `is_admin` (boolean, default: false) - Differentiates Business Admins.
- `status` (varchar, default: 'active') - Can be active, inactive (soft delete), pending.

### boards
Stores the user-created communities.

- `id` (uuid) - Primary Key
- `creator_id` (uuid) - Foreign Key to `users.id`
- `name` (varchar) - The name of the board (e.g., "Dog Lovers").
- `description` (text, nullable) - A short description of the board.

### posts
The individual threads started on a board.

- `id` (uuid) - Primary Key
- `author_id` (uuid) - Foreign Key to `users.id`
- `board_id` (uuid) - Foreign Key to `boards.id`
- `title` (varchar) - The title of the post.
- `content` (text) - The body of the post.
- `post_type` (varchar, default: 'standard') - Can be `standard` or `bulletin`.
- `is_pinned` (boolean, default: false) - For admin posts.

### comments
Replies to a specific post.

- `id` (uuid) - Primary Key
- `author_id` (uuid) - Foreign Key to `users.id`
- `post_id` (uuid) - Foreign Key to `posts.id`
- `content` (text) - The body of the comment.

### board_subscriptions (junction)
Tracks which users are subscribed to which boards.

- `user_id` (uuid) - Foreign Key to `users.id`
- `board_id` (uuid) - Foreign Key to `boards.id`
- Primary Key: composite (`user_id`, `board_id`)

### post_reactions (junction)
Tracks user reactions to posts.

- `user_id` (uuid) - Foreign Key to `users.id`
- `post_id` (uuid) - Foreign Key to `posts.id`
- `reaction_type` (varchar) - The emoji used (e.g., 'like', 'laugh').
- Primary Key: composite (`user_id`, `post_id`)

---

## 2. REST API Endpoints

### Authentication

#### POST /api/auth/register
Business Logic: An admin-only endpoint used to initiate the onboarding process for a new resident. It creates a user in a pending state and sends the registration email.

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
  "message": "Registration link sent successfully to new.resident@example.com"
}
```

#### POST /api/auth/login
Business Logic: Authenticates a user with their email and password. If successful, it returns a JWT for session management.

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
  "user": {
    "id": "user_uuid",
    "unit_number": "101"
  }
}
```

### Boards

#### GET /api/boards
Business Logic: Fetches a list of all available boards in the community.

Request Body: None

Response Body (200 OK):

```json
[
  {
    "id": "board_uuid_1",
    "name": "Dog Lovers",
    "description": "A place for all things canine.",
    "subscriber_count": 25
  },
  {
    "id": "board_uuid_2",
    "name": "For Sale",
    "description": "Buy and sell items with your neighbors.",
    "subscriber_count": 40
  }
]
```

#### POST /api/boards
Business Logic: Allows a logged-in user to create a new board.

Request Body:

```json
{
  "name": "Book Club",
  "description": "Let's read and discuss!"
}
```

Response Body (201 Created):

```json
{
  "id": "new_board_uuid",
  "name": "Book Club",
  "description": "Let's read and discuss!",
  "creator_id": "user_uuid"
}
```

#### POST /api/boards/{boardId}/subscribe
Business Logic: Allows the logged-in user to subscribe to (or unsubscribe from) a specific board.

Request Body: None

Response Body (200 OK):

```json
{
  "message": "Successfully subscribed to the board."
}
```

### Posts

#### GET /api/posts
Business Logic: This is the main endpoint for the "General Feed." It fetches a paginated list of all posts from all boards, sorted by creation date. Pinned bulletin posts appear first.

Request Body: None

Response Body (200 OK):

```json
{
  "posts": [
    {
      "id": "post_uuid_1",
      "title": "Pool Maintenance on Friday",
      "author": { "id": "admin_uuid", "unit_number": "Admin" },
      "board": { "id": "board_uuid_general", "name": "Announcements" },
      "comment_count": 0,
      "reaction_count": 5,
      "is_pinned": true,
      "created_at": "timestamp"
    },
    {
      "id": "post_uuid_2",
      "title": "Anyone have a ladder I can borrow?",
      "author": { "id": "user_uuid", "unit_number": "101" },
      "board": { "id": "board_uuid_ask", "name": "Neighborly Help" },
      "comment_count": 3,
      "reaction_count": 8,
      "is_pinned": false,
      "created_at": "timestamp"
    }
  ],
  "next_page_cursor": "encrypted_cursor_for_pagination"
}
```

#### GET /api/boards/{boardId}/posts
Business Logic: Fetches a paginated list of posts from a specific board.

Request Body: None

Response Body (200 OK): Same structure as `/api/posts` but filtered for the board.

#### POST /api/boards/{boardId}/posts
Business Logic: Creates a new post on a specific board. Admins can additionally set `post_type` and `is_pinned`.

Request Body:

```json
{
  "title": "New book for September!",
  "content": "We'll be reading 'The Midnight Library'. First meeting is next Tuesday."
}
```

Response Body (201 Created):

```json
{
  "id": "new_post_uuid",
  "title": "New book for September!",
  "content": "We'll be reading 'The Midnight Library'. First meeting is next Tuesday.",
  "author_id": "user_uuid",
  "board_id": "board_uuid_book_club"
}
```

#### GET /api/posts/{postId}
Business Logic: Fetches the full details of a single post, including all its comments.

Request Body: None

Response Body (200 OK):

```json
{
  "id": "post_uuid_2",
  "title": "Anyone have a ladder I can borrow?",
  "content": "Just need it for an hour to change a lightbulb!",
  "author": { "id": "user_uuid", "unit_number": "101" },
  "comments": [
    {
      "id": "comment_uuid_1",
      "content": "I have one you can use!",
      "author": { "id": "user_uuid_neighbor", "unit_number": "205" }
    }
  ]
}
```

#### POST /api/posts/{postId}/comments
Business Logic: Adds a new comment to a specific post.

Request Body:

```json
{
  "content": "Awesome, I'll swing by in 10 minutes!"
}
```

Response Body (201 Created):

```json
{
  "id": "new_comment_uuid",
  "content": "Awesome, I'll swing by in 10 minutes!",
  "author_id": "user_uuid"
}
```


