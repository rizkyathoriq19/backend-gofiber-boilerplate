# Go Fiber Boilerplate

A minimal, production-ready Go Fiber boilerplate with **flat RBAC**, Redis caching, and PostgreSQL.

## Features

- 🔐 **JWT Authentication** - Register, login, logout, refresh tokens
- 👥 **Flat RBAC** - Roles & permissions (super_admin, user)
- ⚡ **Redis** - Caching, rate limiting, token blacklisting
- 🐘 **PostgreSQL** - Database with migrations
- 📝 **Swagger** - Auto-generated API docs
- 🔒 **Security** - CORS, Helmet, rate limiting

## Project Structure

```
├── cmd/server/main.go       # Entry point
├── internal/
│   ├── config/              # App configuration
│   ├── database/            # PostgreSQL & Redis
│   ├── middleware/          # Auth, CORS, Logger, Rate Limit
│   ├── module/
│   │   ├── auth/            # Authentication
│   │   └── rbac/            # Role-Based Access Control
│   └── pkg/                 # Shared utilities
├── migrations/              # Database migrations
└── docs/                    # Swagger docs
```

## Quick Start

### Prerequisites
- Go 1.24+
- PostgreSQL 16+
- Redis 7+

### Setup

```bash
# Clone & configure
cp .env.example .env
# Edit .env with your database credentials

# Run migrations
# (Use golang-migrate or similar tool)

# Start server
go run ./cmd/server

# With hot reload
air
```

### Docker

```bash
docker-compose up -d
```

## API Endpoints

**Swagger UI**: http://localhost:8000/swagger/

### Public
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/refresh` | Refresh token |

### Protected (Auth Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/auth/profile` | Get profile |
| PUT | `/api/v1/auth/profile` | Update profile |
| POST | `/api/v1/auth/logout` | Logout |
| GET | `/api/v1/auth/my-roles` | Get my roles |
| GET | `/api/v1/auth/my-permissions` | Get my permissions |

### Super Admin Only
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/super-admin/roles` | List roles |
| POST | `/api/v1/super-admin/roles` | Create role |
| GET | `/api/v1/super-admin/permissions` | List permissions |
| POST | `/api/v1/super-admin/roles/:id/permissions` | Assign permission |

## Environment Variables

```env
# Server
APP_NAME=Go Fiber API
APP_ENV=development
APP_PORT=8000

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=boilerplate_dev

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h

# Rate Limiting
RATE_LIMIT_MAX=100
RATE_LIMIT_WINDOW=1m
```

## Default Users

| Email | Password | Role |
|-------|----------|------|
| superadmin@example.com | password123 | super_admin |
| user@example.com | password123 | user |

## Testing

```bash
go test ./... -v
```

## License

MIT
