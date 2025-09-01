# Cul-de-Chat: Functional Requirements Specification
Last Updated: August 30, 2025

## Project Vision & Guiding Principles
In a world where social interaction has moved increasingly online, it has become paradoxically difficult to build meaningful relationships with the people right around us. This project is a direct response to the trend of social atomization, where local connections are often overlooked.

"Cul-de-Chat" is a non-commercial, community-driven project designed to create a private, modern 'town square' for our neighborhood. Its purpose is not to generate profit, but to leverage technology to help reverse the trend of local disconnection. We aim to provide a tool that fosters genuine connection, helps neighbors find common ground, and adds a little more value and texture to our everyday lives.

## 1. Core Concept & Philosophy
The app will serve as a private, modern "town square" exclusively for verified residents of the townhome complex. The primary goal is to foster community discovery and open interaction. The system is built around user-created, interest-based "Boards" rather than closed-off private groups, encouraging exploration and connection.

## 2. User Roles & Permissions
**Resident (Standard User)**: A verified member of the community. Can create boards, post on boards, comment, react, subscribe to boards, and send direct messages.

**Business Admin (Apartment Staff)**: Manages the community. Has all Resident permissions plus:
- Onboard and offboard users.
- Create special "Bulletin Posts."
- Pin posts.
- Deactivate/delete user accounts.

**Dev Admin (Technical Staff)**: For system maintenance. Has restricted access to user data and day-to-day functions unless required for technical support.

## 3. Onboarding & Offboarding Workflow
### Onboarding
1. A resident provides their email address to the Business Admin.
2. The Admin enters the email and associated unit number into the system, which sends a unique registration link to the resident.
3. The Admin separately provides the resident with a temporary passcode.
4. The resident clicks the link and enters the passcode to verify their identity and complete account setup.

### Offboarding
1. When a resident moves out, the Business Admin deactivates their user account.
2. This action permanently deletes the user's account and all associated personal data. The unit number is then free to be reassigned.

## 4. Core Feature: Boards & Feeds
- **Boards**: Residents can create public (within the community) "Boards" based on specific interests (e.g., "Dog Lovers," "Book Club," "For Sale").
- **General Feed**: The app's home screen. It aggregates and displays all posts from all boards in the community for broad discovery.
- **Posts & Interactions**: A post on a board creates a "thread." Other users can write comments within the thread and react to the initial post using a pre-defined set of emojis.
- **Bulletin Posts (Admin-Only)**: Business Admins can create special "Bulletin Posts" for official announcements. These posts are automatically pinned to the top of the General Feed, and comments are disabled.

## 5. Core Feature: User Profiles & Directory
- **Profile Information**: Users can optionally add a profile picture.
- **Directory & Privacy**: An opt-in directory allows residents to make their Name and Unit Number visible. If a user opts out, their details are hidden, but their account can still be referenced by unit number for messaging.

## 6. Communication
**Direct Messaging (DM)**: Users can send private, one-on-one messages. A user can initiate a message by referencing another user's Unit Number, allowing essential communication even if the recipient is not in the public directory.

## 7. Moderation (MVP)
For the initial version, users will report issues or inappropriate content by sending a direct message to a Business Admin account. A formal "report" button will be a future addition.

## 8. Development Phasing
**Version 1.0 (MVP)**: User management, profiles, boards, feeds, posting, commenting, and reacting. Admin Bulletin Post feature.

**Version 1.1 (Fast Follow)**: One-on-one Direct Messaging system.


