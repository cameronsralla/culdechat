# Cul-de-Chat: UI Screens Specification
Last Updated: August 30, 2025

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
  - Counts for comments and reactions.
- Floating Action Button (FAB): Circular "+" button bottom-right.

### User Interactions
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


