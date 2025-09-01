# Cul-de-Chat: Technical Requirements Specification
Last Updated: August 30, 2025

## 1. Core Architecture
The application will be a containerized system running in a Docker environment on a local, self-hosted server.

- **Backend**: Go (Gin framework)
- **Database**: PostgreSQL
- **Frontend**: React Native
- **API Style**: REST
- **Real-time**: Socket.IO
- **Deployment**: Docker containers

## 2. Platform & Deployment
- **Target Platform**: Mobile-first, using React Native for iOS and Android.
- **Hosting**: Self-hosted on a server within the apartment complex. Requires network configuration (static IP or DDNS) and physical security.

## 3. Backend Architecture
- **Framework**: Gin for the REST API in Go.
- **Real-time Features**: Socket.IO will be implemented on the Go backend to manage real-time messaging and notifications.

## 4. Database & Data Management
- **Database**: PostgreSQL running in a Docker container.
- **Media Storage**: User-uploaded files will be stored on the local server's filesystem, with strict backend validation for file type and size.
- **Data Retention Policies**:
  - **User Data**: A soft delete policy will be used. Data is flagged as inactive for 30 days before a scheduled job performs a permanent hard delete.
  - **Chat Messages**: A Time-to-Live (TTL) of 6 months will be enforced via a scheduled job.

## 5. Frontend Architecture
- **Framework**: React Native.
- **Styling**: Tailwind CSS (via a library like NativeWind).
- **State Management**: React Context API + Hooks will be used for managing application state in the MVP.

## 6. Authentication & Security
- **Login Method**: Standard Email & Password.
- **Session Management**: JSON Web Tokens (JWTs) will be issued by the server and stored securely on the mobile device's local storage (e.g., Keychain/Keystore).
- **Security MVP**:
  - All traffic will be served over HTTPS (using a Let's Encrypt certificate).
  - User passwords will be hashed using a strong algorithm (e.g., bcrypt).

## 7. Operations & Maintenance
- **Initial Scale**: The system will be architected for an initial load of ~100 users.
- **Logging**: The PLG Stack (Promtail, Loki, Grafana) will be used for a self-hosted, real-time log monitoring solution.
- **Backups**: A daily, automated backup of the PostgreSQL database is strongly recommended. This can be achieved with a simple cron job in a Docker container that runs pg_dump.


