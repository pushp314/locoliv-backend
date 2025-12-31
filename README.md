# LocoLive Backend

Production-grade Go backend for LocoLive Mobile (React-Native Expo).

## Stack

- **Language**: Go 1.22+
- **Router**: Chi
- **Database**: PostgreSQL
- **Cache**: Redis
- **Auth**: JWT (access + refresh tokens)
- **OAuth**: Google (Expo-compatible)

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make (optional)

### Setup

```bash
# Clone and setup
cd locolive-backend
cp .env.example .env
# Edit .env with your values

# Start everything
make dev-stack

# Or manually:
docker-compose up -d postgres redis
docker-compose --profile migrate run --rm migrate
go run ./cmd/api
```

### API Endpoints

#### Auth (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Email/password registration |
| POST | `/auth/login` | Email/password login |
| POST | `/auth/refresh` | Token refresh |
| POST | `/auth/logout` | Logout (revoke token) |
| POST | `/auth/google` | Google OAuth |

#### Protected

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/me` | Get current user |
| POST | `/api/v1/auth/logout-all` | Logout all devices |

#### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/health/ready` | Readiness check |
| GET | `/health/live` | Liveness check |

## Development

```bash
# Run with hot reload
make dev

# Run tests
make test

# Generate SQLC code
make sqlc

# Run linter
make lint
```

## Docker

```bash
# Build image
make docker-build

# Start all services
make docker-up

# View logs
make docker-logs

# Stop all services
make docker-down
```

## Database Migrations

```bash
# Run migrations
make migrate-up

# Rollback one migration
make migrate-down

# Create new migration
make migrate-create
```

## Expo Integration

### Google OAuth Flow

1. User taps "Sign in with Google" in Expo app
2. Expo uses `expo-auth-session` to get Google `id_token`
3. Expo sends `id_token` to `POST /auth/google`
4. Backend verifies token, creates/finds user, returns JWT pair
5. Expo stores tokens:
   - Access token: Memory
   - Refresh token: SecureStore

```javascript
// Expo example
const response = await fetch(`${API_URL}/auth/google`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ id_token: googleIdToken })
});
const { access_token, refresh_token, user } = await response.json();
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | 8080 |
| `ENV` | Environment | development |
| `DATABASE_URL` | PostgreSQL URL | - |
| `REDIS_URL` | Redis URL | - |
| `JWT_SECRET` | JWT signing key | - |
| `JWT_ACCESS_EXPIRY` | Access token TTL | 15m |
| `JWT_REFRESH_EXPIRY` | Refresh token TTL | 168h |
| `GOOGLE_CLIENT_ID` | Google OAuth Client ID | - |

## Project Structure

```
locolive-backend/
├── cmd/api/              # Entry point
├── internal/
│   ├── api/              # HTTP handlers
│   ├── auth/             # JWT, OAuth
│   ├── config/           # Configuration
│   ├── domain/           # Business logic
│   ├── middleware/       # HTTP middleware
│   └── repository/       # Data access
├── db/
│   ├── migrations/       # SQL migrations
│   └── queries/          # SQLC queries
├── pkg/                  # Shared utilities
├── Dockerfile
└── docker-compose.yml
```

## License

MIT
# locoliv-backend
