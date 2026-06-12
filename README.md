<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go" alt="Go Version" />
  <img src="https://img.shields.io/badge/Next.js-14-000000?style=for-the-badge&logo=next.js" alt="Next.js Version" />
  <img src="https://img.shields.io/badge/PostgreSQL-16-4169E1?style=for-the-badge&logo=postgresql" alt="PostgreSQL Version" />
  <img src="https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker" alt="Docker Compose" />
  <img src="https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge" alt="License" />
</p>

<h1 align="center">📋 Enterprise Task Management Application</h1>

<p align="center">
  <strong>Go Clean Architecture (Handler → Service → Repository) + Next.js 14 Standalone</strong>
  <br />
  A production-grade, full-stack task management system built with battle-tested patterns,
  enterprise security practices, and developer-first DX.
</p>

---

## Table of Contents

- [Overview](#overview)
- [Key Architectural Decisions](#key-architectural-decisions)
- [Bonus Features Implemented](#bonus-features-implemented)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Environment Configuration](#environment-configuration)
- [How to Run (One-Command Setup)](#how-to-run-one-command-setup)
- [API Endpoints](#api-endpoints)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

This is an **enterprise-grade task management application** designed with scalability, security, and maintainability as first-class concerns. The backend follows **Go Clean Architecture** with a strict layer separation pattern, while the frontend leverages **Next.js 14 Standalone** for optimal rendering strategies and deployment efficiency.

The application enables users to:
- Register and authenticate securely via **JWT stored in HTTP-Only cookies**
- Create, read, update, and delete tasks with rich metadata (priority, status, due dates, categories)
- Filter, sort, and paginate through tasks with sub-second response times
- View real-time activity logs (audit trail) for every task mutation
- Experience instant UI updates via **optimistic updates** with **TanStack Query v5**
- Receive real-time changes pushed from the server via **Server-Sent Events (SSE)**

---

## Key Architectural Decisions

### 1. Strict Layer Separation (Handler → Service → Repository)

```
┌─────────────────────────────────────────────────────┐
│                    HTTP Layer                        │
│              (Handlers / Middleware)                 │
│  - Request parsing, validation, response writing    │
│  - JWT extraction via middleware                     │
│  - No business logic                                │
├─────────────────────────────────────────────────────┤
│                  Service Layer                       │
│           (Business Logic Orchestration)             │
│  - Transaction management                           │
│  - Authorization checks                             │
│  - Audit logging (activity trail)                   │
│  - SSE event broadcasting                           │
├─────────────────────────────────────────────────────┤
│                 Repository Layer                     │
│            (Data Access / Persistence)               │
│  - PostgreSQL queries with prepared statements      │
│  - Pagination, filtering, sorting                   │
│  - Index-optimized access patterns                  │
└─────────────────────────────────────────────────────┘
```

Each layer communicates only with the layer directly below it. Handlers never touch the database; repositories never handle HTTP concerns. This separation yields:

- **Testability**: Each layer can be unit-tested in isolation with mocks
- **Maintainability**: Changing the database driver or HTTP framework requires zero changes outside the affected layer
- **Security**: Authorization logic lives in the service layer, ensuring it cannot be bypassed regardless of HTTP implementation

### 2. JWT via HTTP-Only Cookies (Advanced Security)

Authentication is implemented using **signed JWT tokens stored exclusively in HTTP-Only, Secure, SameSite=Strict cookies**. This approach eliminates the risk of XSS-based token theft — JavaScript in the browser never has programmatic access to the authentication token.

| Property | Value |
|---|---|
| Storage | HTTP-Only Cookie (not localStorage) |
| XSS Protection | ✅ Immune — JS cannot read the cookie |
| CSRF Protection | ✅ SameSite=Strict + Origin validation |
| Token Signature | HMAC-SHA256 via `golang-jwt/jwt/v5` |
| Password Hashing | bcrypt via `golang.org/x/crypto` |

### 3. PostgreSQL Indexing Strategy

The database schema is intentionally indexed to support the most common query patterns:

- **Composite index** on `(user_id, created_at DESC)` — powers the main task listing with pagination
- **Partial index** on `(status)` for status-based filtering queries
- **GIN index** on `(to_tsvector('english', title || ' ' || description))` — enables full-text search without external search infrastructure
- **Unique index** on `(id, user_id)` — used by the repository for ownership-guaranteed lookups

---

## Bonus Features Implemented

### 🚀 Optimistic UI (TanStack Query v5)

The frontend leverages **TanStack Query v5 (React Query)** to implement optimistic updates. When a user performs a mutation (create, update, delete a task), the UI immediately reflects the expected result **before the server responds**. If the server confirms the mutation, the cache is updated seamlessly. If the server rejects it, the UI automatically rolls back to the previous state with zero user-perceived latency.

```
User Action → Optimistic Cache Update → UI Renders Instantly
                                   ↓
                      Server Mutation API Call
                                   ↓
              ✅ Success → Cache Sync    ❌ Failure → Rollback
```

### 🔄 Real-Time Sync via Server-Sent Events (SSE)

Instead of polling or WebSocket overhead, the backend pushes real-time updates to connected clients using **Server-Sent Events**. When a task is created, updated, or deleted by one user, all other authenticated clients receive the change instantly.

- Lightweight, unidirectional protocol over standard HTTP
- Automatic reconnection handled by the browser's native `EventSource` API
- Scoped to authenticated user sessions — users only receive events for tasks they own

### 📜 Audit Trail (Activity Logs)

Every state mutation to a task is recorded in an immutable activity log. The `activity_logs` table stores:

- **Actor** — the user who performed the action
- **Action** — the type of mutation (created, updated status, deleted, etc.)
- **Entity** — the affected task ID
- **Diff / Snapshot** — the before/after state of the changed fields
- **Timestamp** — when the mutation occurred (server time, monotonic)

This provides a complete, forensically sound history of every change in the system.

### 🐳 Distroless / Standalone Multi-Stage Docker Optimization

The build pipeline uses **multi-stage Docker builds** with distroless runtime images to minimize attack surface and deployment size:

**Backend (Go):**
| Stage | Image | Size |
|---|---|---|
| Build | `golang:1.22-alpine` | ~900 MB |
| Runtime | `alpine:3.19` (distroless, non-root user) | ~18 MB ✅ |

- Statically linked binary with `CGO_ENABLED=0`
- Stripped debug symbols (`-ldflags="-w -s"`)
- Runs as `appuser` (UID 1001) — not root

**Frontend (Next.js):**
| Stage | Image | Size |
|---|---|---|
| Build | `node:20-alpine` | ~1.2 GB |
| Runtime | `node:20-alpine` (standalone output) | ~150 MB ✅ |

- Uses Next.js `output: 'standalone'` — copies only the minimal required files
- No development dependencies included in the final image

---

## Tech Stack

### Backend

| Technology | Purpose |
|---|---|
| **Go 1.22** | Primary language — compiled, statically typed, high performance |
| **Chi Router v5** | Lightweight, idiomatic HTTP router with middleware chaining |
| **pgx / lib/pq** | PostgreSQL driver with prepared statement support |
| **golang-jwt v5** | JWT creation and verification |
| **go-playground/validator** | Struct-level request validation |
| **golang-migrate** | Database schema migrations |
| **google/uuid** | UUID v4 generation for task and user IDs |

### Frontend

| Technology | Purpose |
|---|---|
| **Next.js 14 (App Router)** | React framework with SSR, SSG, and standalone output |
| **TypeScript** | Type-safe development |
| **TanStack Query v5** | Server state management, caching, optimistic updates |
| **Tailwind CSS** | Utility-first styling |
| **shadcn/ui** | Accessible, composable component primitives |

### Infrastructure

| Technology | Purpose |
|---|---|
| **Docker** | Containerization for both services |
| **Docker Compose** | Multi-service orchestration |
| **PostgreSQL 16** | Primary data store with full-text search support |

---

## Project Structure

```
task-management-app/
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go                 # Application entrypoint
│   ├── internal/
│   │   ├── handlers/
│   │   │   ├── auth_handler.go         # Authentication endpoints
│   │   │   ├── task_handler.go         # CRUD + activity endpoints
│   │   │   └── helpers.go              # Response helpers (JSON, errors)
│   │   ├── middleware/
│   │   │   └── auth.go                 # JWT extraction + context injection
│   │   ├── models/
│   │   │   ├── auth.go                 # Auth request/response DTOs
│   │   │   ├── task.go                 # Task domain model + DTOs
│   │   │   └── api.go                  # Generic API response envelope
│   │   ├── repository/
│   │   │   ├── user_repository.go      # User persistence
│   │   │   ├── task_repository.go      # Task CRUD + search queries
│   │   │   └── activity_repository.go  # Activity log persistence
│   │   └── service/
│   │       ├── auth_service.go         # Registration, login, JWT issuance
│   │       ├── task_service.go         # Business logic + authorization
│   │       └── sse_service.go          # Server-Sent Events broadcasting
│   ├── migrations/
│   │   └── 000001_init.up.sql          # Schema + indexes
│   ├── go.mod
│   ├── Dockerfile                      # Multi-stage distroless build
│   └── .dockerignore
├── frontend/
│   ├── src/
│   │   ├── app/                        # Next.js App Router pages
│   │   ├── components/                 # UI components
│   │   ├── hooks/                      # Custom hooks (mutations, SSE)
│   │   └── lib/                        # API client, utilities
│   ├── public/
│   ├── next.config.js                  # `output: 'standalone'`
│   ├── Dockerfile                      # Standalone production build
│   ├── package.json
│   └── tsconfig.json
├── docker-compose.yml                  # Production orchestration
├── .env.example                        # Environment variable reference
└── README.md                           # This file
```

---

## Environment Configuration

Copy `.env.example` to `.env` and configure the following variables:

| Variable | Description | Default |
|---|---|---|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://taskmanager:secretpassword@db:5432/taskmanager?sslmode=disable` |
| `POSTGRES_USER` | Database username | `taskmanager` |
| `POSTGRES_PASSWORD` | Database password | `secretpassword` |
| `POSTGRES_DB` | Database name | `taskmanager` |
| `JWT_SECRET` | HMAC signing key for JWT tokens | Must be changed in production |
| `JWT_EXPIRATION` | Token TTL (Go duration format) | `24h` |
| `API_PORT` | Backend server port | `8080` |
| `API_ALLOWED_ORIGINS` | CORS allowed origins | `http://localhost:3000` |
| `NEXT_PUBLIC_API_URL` | Public-facing API URL (browser) | `http://localhost:8080/api` |
| `NEXT_SERVER_API_URL` | Internal API URL (server-side) | `http://api:8080/api` |

> **⚠️ Production Requirements:** Change `JWT_SECRET` to a cryptographically random string (minimum 256 bits). Use strong, unique credentials for the database. Enable TLS by terminating SSL at a reverse proxy (e.g., nginx, Caddy, or a cloud load balancer).

---

## How to Run (One-Command Setup)

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (v24+) and [Docker Compose](https://docs.docker.com/compose/install/) (v2.20+)
- Git

### Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/your-org/task-management-app.git
cd task-management-app

# 2. Configure environment (optional — defaults work for local dev)
cp .env.example .env

# 3. Build and launch everything with a single command
docker compose up --build
```

This single command will:

1. Build the Go backend binary using the multi-stage Dockerfile
2. Build the Next.js frontend with `output: 'standalone'`
3. Start a PostgreSQL 16 instance with health checks
4. Run database migrations automatically (via the backend's startup process)
5. Expose the API on `http://localhost:8080` and the frontend on `http://localhost:3000`

### Verify It's Running

```bash
# Health check — API
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/auth/me

# Health check — Frontend
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000

# View logs
docker compose logs -f
```

### Shut Down

```bash
docker compose down

# To also delete the database volume (⚠️ destroys all data)
docker compose down -v
```

---

## Vercel Deployment Adaptation

The frontend is fully adapted for native Vercel deployment via the **Next.js standalone output mode**. Key configuration points:

| Concern | Implementation | File |
|---|---|---|
| **Output mode** | `output: 'standalone'` in `next.config.js` — produces a minimal self-contained build | `frontend/next.config.js` |
| **API routing** | `NEXT_PUBLIC_API_URL` and `NEXT_SERVER_API_URL` control client vs server-side API base URLs | `.env` / Vercel Environment Variables |
| **Static asset handling** | Public assets served from `/public`, bundled into standalone output | `frontend/public/` |
| **Image optimization** | Configured via `next.config.js`; uses default Vercel image optimizer when deployed | `frontend/next.config.js` |

### Environment Variables for Vercel

When deploying to Vercel, set the following in the Vercel dashboard (Settings → Environment Variables):

```
NEXT_PUBLIC_API_URL=https://api.yourdomain.com/api
NEXT_SERVER_API_URL=https://api.yourdomain.com/api
```

> **Note:** For local Docker Compose development, `NEXT_SERVER_API_URL=http://api:8080/api` resolves via the internal Docker network. On Vercel, both variables point to the same public API URL because there is no internal Docker network.

### Cross-Domain Cookie Configuration (Frontend on Vercel, Backend on Render)

When the frontend is hosted on **Vercel** and the backend on **Render** (or any separate domain), the browser's Same-Origin Policy and cookie scoping rules require explicit opt-in for cross-domain cookie delivery:

| Attribute | Required Value | Rationale |
|---|---|---|
| `SameSite` | `None` | Allows the cookie to be sent with cross-origin requests. `SameSite=Lax` (the browser default in 2025) would block the cookie on the actual cross-origin `fetch()` call. |
| `Secure` | `true` | Required by the Fetch Standard when `SameSite=None` is set. The backend endpoint must be served over HTTPS with a valid TLS certificate. |
| `Domain` | _Not set_ | Leaving the domain attribute unspecified restricts the cookie to the issuing origin (the backend). The browser will still attach it to cross-origin `fetch()` requests because `SameSite=None` is in effect. |
| `HttpOnly` | `true` | Prevents JavaScript from reading the cookie, eliminating XSS-based token theft. |

**Backend cookie-setting code example (Go):**
```go
http.SetCookie(w, &http.Cookie{
    Name:     "auth_token",
    Value:    tokenString,
    Path:     "/",
    HttpOnly: true,
    Secure:   true,           // Required when frontend and backend are on different origins
    SameSite: http.SameSiteNoneMode, // Allows cross-origin delivery
    MaxAge:   86400,          // 24 hours
})
```

> ⚠️ **Important**: `http.SameSiteNoneMode` requires Go 1.16+. The backend binary uses Go 1.22, so this value is fully supported. Ensure `Secure: true` is only enabled in production — local development over HTTP should use `SameSite=http.SameSiteStrictMode` with `Secure: false`.

### Custom Domain Considerations

1. **CORS Origin**: Set `API_ALLOWED_ORIGINS` in the backend deployment (e.g., Render, Railway, Fly.io, or your own VPS) to match the **exact** Vercel frontend domain (e.g., `https://task-manager.vercel.app` or your custom domain). The Go CORS middleware (`backend/internal/middleware/cors.go`) automatically merges this with the built-in defaults (`http://localhost:3000` and `https://task-management-application-omega-lyart.vercel.app`).
2. **Cookie Domain**: For cross-domain cookies (frontend on Vercel, backend on another host), ensure the backend sets `SameSite=None; Secure` on the JWT cookie and the backend domain is served over HTTPS with a valid TLS certificate.

---

## API Endpoints

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `POST` | `/api/v1/auth/register` | Create a new user account | ❌ |
| `POST` | `/api/v1/auth/login` | Authenticate and receive JWT cookie | ❌ |
| `POST` | `/api/v1/auth/logout` | Invalidate session (clear cookie) | ✅ |
| `GET` | `/api/v1/auth/me` | Get current authenticated user | ✅ |
| `GET` | `/api/v1/tasks` | List tasks (paginated, filterable, sortable) | ✅ |
| `POST` | `/api/v1/tasks` | Create a new task | ✅ |
| `GET` | `/api/v1/tasks/{id}` | Get a single task by ID | ✅ |
| `PUT` | `/api/v1/tasks/{id}` | Update a task | ✅ |
| `DELETE` | `/api/v1/tasks/{id}` | Delete a task | ✅ |
| `GET` | `/api/v1/tasks/{id}/activity` | Get activity log for a task | ✅ |
| `GET` | `/api/v1/events` | SSE stream for real-time updates | ✅ |

> **🔧 CORS Preflight Optimization**
>
> The server explicitly handles `OPTIONS` preflight requests at the middleware layer and returns
> **HTTP 200 OK** immediately — without entering the route execution pipeline. This behaviour
> eliminates unnecessary latency for cross-origin requests (the preflight round-trip completes
> in a single network hop) and ensures full compatibility with credentialed requests (cookies)
> by returning the exact `Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials: true`
> headers. The preflight cache is set to `3600` seconds (`Access-Control-Max-Age: 3600`), so
> browsers only send one `OPTIONS` request per origin per hour.
>
> **Implementation details (`backend/internal/middleware/cors.go`):**
> - The middleware reads `API_ALLOWED_ORIGINS` from the environment and merges it with two
>   **built-in default origins**: `http://localhost:3000` (local development) and
>   `https://task-management-application-omega-lyart.vercel.app` (production Vercel domain).
> - Wildcard origins (`*`) are filtered out at startup — never returned to clients — because
>   the `Access-Control-Allow-Credentials: true` header required for cookie-based auth is
>   **incompatible with wildcard origins** per the Fetch Standard.
> - Preflight requests are terminated at the middleware boundary **before** reaching any
>   authentication guards or route handlers, ensuring that cross-origin cookie flows are
>   never blocked by a missing or expired JWT during the preflight phase.

### Query Parameters for `GET /api/v1/tasks`

| Parameter | Type | Description | Default |
|---|---|---|---|
| `page` | `int` | Page number (1-based) | `1` |
| `per_page` | `int` | Items per page (max 100) | `20` |
| `status` | `string` | Filter by status (`pending`, `in_progress`, `completed`) | — |
| `priority` | `string` | Filter by priority (`low`, `medium`, `high`, `critical`) | — |
| `search` | `string` | Full-text search across title and description | — |
| `sort_by` | `string` | Sort field (`created_at`, `updated_at`, `due_date`, `priority`) | `created_at` |
| `sort_dir` | `string` | Sort direction (`asc` or `desc`) | `desc` |

---

## Testing

### Backend (Go Table-Driven Tests)

The backend uses Go's standard `testing` package with table-driven test patterns for comprehensive test coverage:

```bash
# Run all backend tests with verbose output
docker compose exec api go test ./... -v -count=1

# Run tests with race detection
docker compose exec api go test ./... -race -count=1

# Run tests for a specific package
docker compose exec api go test ./internal/... -v

# Generate coverage report
docker compose exec api go test ./... -coverprofile=coverage.out
docker compose exec api go tool cover -html=coverage.out -o coverage.html
```

### Frontend (Integration Tests)

```bash
# Run frontend tests with Jest and React Testing Library
docker compose exec frontend npm test

# Run tests in watch mode
docker compose exec frontend npm test -- --watch

# Run tests with coverage
docker compose exec frontend npm test -- --coverage
```

---

## Contributing

We welcome contributions that align with the architectural principles of this project:

1. **Fork** the repository
2. **Create a feature branch** (`git checkout -b feat/your-feature`)
3. **Write tests** for any new functionality
4. **Ensure all existing tests pass** before submitting
5. **Open a pull request** with a clear description of changes

### Code Style

- **Go**: Follow standard `gofmt` formatting. Use `golangci-lint` for linting.
- **TypeScript / React**: Use the project's ESLint and Prettier configurations.
- **SQL**: Use `golang-migrate` for schema changes. Never modify existing migrations.

---

## License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.

---

<p align="center">
  Built with ❤️ by the Task Management Team
  <br />
<sub>Go Clean Architecture · Next.js 14 Standalone · PostgreSQL 16 · Vercel Deploy Ready</sub>
</p>