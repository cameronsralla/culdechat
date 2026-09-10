# Cul-de-Chat: UI Screens Specification
Last Updated: September 9, 2026

The Expo app lives in `mobile/`. One codebase: Expo web is the desktop experience; the same screens ship to iOS/Android. `AppShell` uses a left sidebar when the window is ≥800px and bottom tabs when it is narrower. Visual language lives in `src/theme/theme.ts`. Screens are wired to the live API: login, complete registration (`/register?token=` pre-fills), feed, boards, board detail, create board/post, post detail (comments, reactions, edit/delete, admin pin), directory with photos, You (profile), and Admin (`/admin`, nav item and page only for `is_admin`).

## 1. Login Screen
Path: `/login`

### Purpose & Layout
Entry point for existing users. Simple and clean, featuring:
- The "Cul-de-Chat" logo at the top.
- An input field for Email.
- An input field for Password.
- A prominent "Login" button.

### User Interactions
- Users enter their credentials and tap "Login."
- On success, they are redirected to the General Feed (`/`).
- On failure, an error message appears (e.g., "Invalid email or password").

## 1b. Complete Registration Screen
Path: `/register`

### Purpose & Layout
First-time residents finishing an admin invite.
- Token (from the invite email or `/register?token=` link) and 10-character passcode fields.
- Display name, password (at least 8 characters), and a submit button.

### User Interactions
- On success, they are logged in and redirected to the General Feed (`/`).
- On failure, show invalid/expired invite errors.

## 2. General Feed (Home Screen)
Path: `/`

### Purpose & Layout
Main "town square" and the first screen after login.
- Header: A simple header with the app name.
- Pinned Posts: Admin "Bulletin Posts" displayed at the top in a highlighted section.
- Main Content: Vertically scrolling list of posts from all boards, sorted by most recent. Each item is a "Post Card" showing:
  - Author's unit number (or name if opted-in).
  - The board it was posted on (e.g., "in Dog Lovers").
  - Post title.
  - Snippet of the post content.
  - Relative timestamp.
  - Counts for comments and reactions.
  - Badge: Bulletin vs Pinned (not the same thing).
- Floating Action Button (FAB): Circular "+" button bottom-right.

### User Interactions
- Pull to refresh. Empty and error states offer a retry or write-a-post action.
- Infinite scroll through the feed.
- Tap a Post Card → Post Detail (`/posts/{postId}`).
- Tap FAB (+) → Create Post (`/posts/new`).

## 3. Boards List Screen
Path: `/boards` (via main navigation tab)

### Purpose & Layout
Discover all communities within the app.
- Header: Titled "Boards."
- Main Content: Scrolling list of all boards. Each list item displays:
  - Board Name (e.g., "Book Club").
  - Board Description.
  - Count of subscribers.
  - "Subscribe" / "Unsubscribe" button.
- Create Board Button: Prominent button at the top labeled "Create New Board."

### User Interactions
- Tap "Subscribe" to join a board.
- Tap a Board Name → Board Feed (`/boards/{boardId}`).
- Tap "Create New Board" → Create Board (`/boards/new`).

## 3b. Board Feed
Path: `/boards/{boardId}`

Posts on one board, subscribe toggle, and a create-post action when subscribed.

## 3c. Create Board
Path: `/boards/new`

Name and optional description. Creator is subscribed. Redirects to the new board feed.

## 3d. Directory
Path: `/directory`

Opted-in neighbors: photo, name, and unit. Pull to refresh.

## 4. Post Detail Screen
Path: `/posts/{postId}`

### Purpose & Layout
Displays a single post and its entire comment thread.
- Main Post Content: Full title, content, author, timestamp.
- Reactions: Reaction emojis and counts below the post.
- Comment Input: Text box at the bottom to write a new comment.
- Comment List: Chronological list of all comments with author and text.

### User Interactions
- Tap an emoji to add/remove reaction to the main post.
- Type in the comment input and hit "Send" to add a comment.
- Scroll through all existing comments.
- Author (or admin) can edit or delete the post. Admins can pin/unpin a standard post. Bulletins stay pinned and comments stay off.
- Comment author can edit; author or admin can delete.

## 5. Create Post Screen
Path: `/posts/new`

### Purpose & Layout
Form for creating a new post.
- Board Selector: Dropdown/selector listing subscribed boards.
- Title Input: Text field for the post's title.
- Content Input: Larger text area for the post body.
- Submit Button: "Post" to submit.

### User Interactions
- Select a board, fill title and content, tap "Post."
- On success, redirect to the new post's detail screen (`/posts/{newPostId}`).
- Admins can mark the post as a bulletin.

## 6. You
Path: `/you`

Profile photo (web upload), name, directory opt-in, password change, logout.

## 7. Admin
Path: `/admin`

Visible only to users with `is_admin`. The Admin nav item is omitted for residents. Hitting `/admin` without the role redirects home.

- Invite a resident (copy token, passcode, and `/register?token=` link).
- Roster with offboard for non-admins.


