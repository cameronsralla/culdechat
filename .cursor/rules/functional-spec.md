# Cul-de-Chat: Functional Requirements Specification
Last Updated: September 30, 2026

## Project Vision & Guiding Principles
Authoritative vision and mission live in [docs/vision.md](../../docs/vision.md). Summary:

In a world where social interaction has moved increasingly online, it has become paradoxically difficult to build meaningful relationships with the people right around us. This project is a direct response to the trend of social atomization, where local connections are often overlooked.

"Cul-de-Chat" is the modern digital town square for bringing local communities together — a non-commercial, community-driven project for our neighborhood. Its purpose is not to generate profit, but to leverage technology to help reverse the trend of local disconnection — a trusted space for verified residents to communicate, find common ground, and build community where they live.

## 1. Core Concept & Philosophy
The app will serve as a private, modern "town square" exclusively for verified residents of the townhome complex. The primary goal is to foster community discovery and open interaction. The system is built around user-created, interest-based "Boards" rather than closed-off private groups, encouraging exploration and connection.

## 2. User Roles & Permissions
**Resident (Standard User)**: A verified member of the community. Can create boards, post on boards, comment, react, subscribe to boards, and send direct messages. A unit is its own record. An admin creates units first, then assigns one primary resident to each. Additional non-primary residents on a unit come later (see [backlog.md](../../docs/backlog.md)).

**Business Admin (Apartment Staff)**: Manages the community. Has all Resident permissions plus:
- Onboard and offboard users.
- Create special "Bulletin Posts."
- Pin posts.
- Deactivate/delete user accounts.

**Dev Admin (Technical Staff)**: For system maintenance. Has restricted access to user data and day-to-day functions unless required for technical support.

## 3. Onboarding & Offboarding Workflow
### Onboarding
1. A resident provides their email address to the Business Admin.
2. The Admin creates the unit if it does not exist, then enters the email and assigns that unit. The system issues a registration token and a temporary passcode. The invite makes that person the unit's primary resident. A unit that already has a primary cannot take another until non-primary members exist.
3. The instance sends the invite from one configured community mailbox over SMTP (apartment Microsoft 365 / Google Workspace, or a dedicated free Gmail/Outlook mailbox). There is no per-instance SaaS signup. Locally, Mailpit catches mail at `http://127.0.0.1:8025`.
4. The admin API also returns the token and passcode so they can be shared by hand if email fails.
5. The resident submits the token, passcode, chosen password, and display name to complete account setup.

### Offboarding
1. When a resident moves out, the Business Admin offboards their user account.
2. The account is soft-deleted (`inactive`) immediately so the unit number can be reassigned. Their sessions are revoked. Permanent deletion of personal data after 30 days is a later retention job.

## 4. Core Feature: Boards & Feeds
- **Boards**: Residents can create public (within the community) "Boards" based on specific interests (e.g., "Dog Lovers," "Book Club," "For Sale").
- **General Feed**: The app's home screen. It aggregates and displays all posts from all boards in the community for broad discovery.
- **Posts & Interactions**: A post on a board creates a "thread." Other users can write comments within the thread and react to the initial post using a pre-defined set of emojis.
- **Bulletin Posts (Admin-Only)**: Business Admins can create special "Bulletin Posts" for official announcements. These posts are automatically pinned to the top of the General Feed, and comments are disabled.

## 5. Core Feature: User Profiles & Directory
- **Profile Information**: Users have a display name (set at registration) and can optionally add a profile picture.
- **Directory & Privacy**: An opt-in directory lets residents show their name and email. If they opt out, their name stays hidden, but their unit still appears in People while they are active so a neighbor can reach them. Inactive and invited residents do not appear.

## 6. Communication
**Direct Messaging (DM)**: Users send private, one-on-one text messages. A conversation belongs to the two people in it. When a unit's primary resident changes, the new resident starts with an empty chat history. The previous resident keeps theirs, unless they were unlisted: deactivating an unlisted resident deletes those threads. A listed resident's threads stay, under their name. Chat is a conversation list beside the thread. The first message is typed in that thread, the same way later replies are. People opens that thread for a neighbor or a unit. Messaging a listed neighbor opens the thread immediately. Messaging a unit where the resident is not listed sends a request: they read it in the thread and accept or decline before anyone can continue. Until they accept, the sender waits and they cannot reply. A declined request can be sent again from the same thread. The unit's primary resident receives it when they are active and not listed. Directory opt-out hides name and email. Admins participate as ordinary users. No media, edit, or delete in the first slice; the inbox polls.

## 7. Moderation (MVP)
For the initial version, users will report issues or inappropriate content by sending a direct message to a Business Admin account. A formal "report" button will be a future addition.

## 8. Development Phasing
Priority and status live in [docs/backlog.md](../../docs/backlog.md).

**Version 1.0 (MVP)**: User management, profiles (including photo upload on web), boards, feeds, posting, commenting, reacting, edit/delete, admin pin, invite copy + register deep link. Admin Bulletin Post feature.

**Version 1.1 (Fast Follow)**: One-on-one Direct Messaging (inbox tab, search by person/unit, create-on-send). Realtime and unread badge follow.

**Version 1.2 (Households)**: A primary resident per unit may invite family members onto that unit. Offboarding the primary offboards the household. The primary cannot be transferred.


