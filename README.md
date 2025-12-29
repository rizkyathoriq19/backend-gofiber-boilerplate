# Go Fiber Boilerplate API

A production-ready Go Fiber boilerplate with RBAC, Redis caching, and PostgreSQL.

## 📚 API Documentation

**Swagger UI**: [http://localhost:8000/swagger/](http://localhost:8000/swagger/)

## 🏗️ Architecture

![RBAC Architecture](docs/images/rbac_architecture.png)

<details>
<summary>View as Mermaid (for GitHub)</summary>

```mermaid
graph TB
    subgraph "Entry Point"
        MAIN["main.go"]
    end

    subgraph "Transport Layer"
        FIBER["Fiber App"]
        MW["Middleware Stack"]
        HANDLERS["Handlers/Controllers"]
    end

    subgraph "Business Layer"
        USECASE["Use Cases / Services"]
        DOMAIN["Domain Interfaces"]
    end

    subgraph "Data Layer"
        REPO["Repositories"]
        CACHE["Cache Helper"]
    end

    subgraph "Infrastructure"
        PG["PostgreSQL"]
        REDIS["Redis"]
        JWT["JWT Manager"]
    end

    MAIN --> FIBER
    FIBER --> MW
    MW --> HANDLERS
    HANDLERS --> USECASE
    USECASE --> DOMAIN
    USECASE --> REPO
    REPO --> PG
    REPO --> CACHE
    CACHE --> REDIS
    USECASE --> JWT
    JWT --> REDIS
```

</details>

## 📁 Project Structure

```
├── cmd/server/main.go       # Entry point
├── internal/
│   ├── config/              # Configuration
│   ├── database/            # PostgreSQL & Redis
│   ├── middleware/          # Auth, CORS, Logger, Rate Limit
│   ├── module/auth/         # Auth module (Clean Architecture)
│   └── pkg/                 # Shared packages
├── migrations/              # Database migrations
└── tests.http              # HTTP test file
```

## 🚀 Quick Start

```bash
# Clone & setup
cp .env.example .env

# Run database migrations
# (setup PostgreSQL and Redis first)

# Run development
go run ./cmd/server/main.go

# Run with hot reload
air

# Run tests
go test ./... -v
```

## 🔐 Authentication Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant API as API Server
    participant DB as Database
    participant RD as Redis

    C->>API: POST /register
    API->>DB: Create user
    API->>RD: Store refresh token
    API-->>C: Return tokens

    C->>API: POST /login
    API->>DB: Verify credentials
    API->>RD: Store refresh token
    API-->>C: Return tokens

    C->>API: GET /profile (with token)
    API->>RD: Check blacklist
    API->>DB: Get user
    API-->>C: Return profile

    C->>API: POST /logout
    API->>RD: Blacklist token
    API-->>C: Success
```

---

## 📡 API Reference

### Base URL
```
http://localhost:8000
```

---

## Health Check

### Ping
```bash
curl -X GET http://localhost:8000/ping
```

**Response (200):**
```
pong
```

### Health Status
```bash
curl -X GET http://localhost:8000/api/v1/health
```

**Response (200):**
```json
{
  "status": "ok",
  "message": "Server is running"
}
```

---

## Authentication Endpoints

### Register
```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

**Success Response (201):**
```json
{
  "success": true,
  "code": 201,
  "message": "Data berhasil dibuat",
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "John Doe",
      "email": "user@example.com",
      "role": "user",
      "created_at": "2025-12-16T08:00:00Z",
      "updated_at": "2025-12-16T08:00:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  },
  "timestamp": "2025-12-16T08:00:00Z"
}
```

**Error - Email Already Exists (409):**
```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "existing@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```
```json
{
  "success": false,
  "code": 409,
  "message": "Email sudah terdaftar",
  "error_code": -1101,
  "timestamp": "2025-12-16T08:00:00Z"
}
```

**Error - Validation Failed (400):**
```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalid-email",
    "password": "123"
  }'
```
```json
{
  "success": false,
  "code": 400,
  "message": "Validasi gagal",
  "error_code": -1009,
  "errors": [
    {"field": "email", "message": "email harus berupa email yang valid"},
    {"field": "password", "message": "password minimal 6 karakter"},
    {"field": "name", "message": "name harus diisi"}
  ],
  "timestamp": "2025-12-16T08:00:00Z"
}
```

---

### Login
```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

**Success Response (200):**
```json
{
  "success": true,
  "code": 200,
  "message": "Login berhasil",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  },
  "timestamp": "2025-12-16T08:00:00Z"
}
```

**Error - Account Not Found (404):**
```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "notfound@example.com",
    "password": "password123"
  }'
```
```json
{
  "success": false,
  "code": 404,
  "message": "Akun tidak ditemukan",
  "error_code": -1104,
  "timestamp": "2025-12-16T08:00:00Z"
}
```

**Error - Password Mismatch (422):**
```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "wrongpassword"
  }'
```
```json
{
  "success": false,
  "code": 422,
  "message": "Password tidak cocok",
  "error_code": -1106,
  "timestamp": "2025-12-16T08:00:00Z"
}
```

---

### Refresh Token
```bash
curl -X POST http://localhost:8000/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Success Response (200):**
```json
{
  "success": true,
  "code": 200,
  "message": "Token berhasil diperbarui",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  },
  "timestamp": "2025-12-16T08:00:00Z"
}
```

**Error - Invalid Token (401):**
```json
{
  "success": false,
  "code": 401,
  "message": "Token tidak valid",
  "error_code": -1010,
  "timestamp": "2025-12-16T08:00:00Z"
}
```

---

### Get Profile (Protected)
```bash
curl -X GET http://localhost:8000/api/v1/auth/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200):**
```json
{
  "success": true,
  "code": 200,
  "message": "Profil berhasil diambil",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "user@example.com",
    "role": "user",
    "created_at": "2025-12-16T08:00:00Z",
    "updated_at": "2025-12-16T08:00:00Z"
  },
  "timestamp": "2025-12-16T08:00:00Z"
}
```

**Error - Unauthorized (401):**
```bash
curl -X GET http://localhost:8000/api/v1/auth/profile
```
```json
{
  "success": false,
  "code": 401,
  "message": "Tidak memiliki akses",
  "error_code": -1005,
  "timestamp": "2025-12-16T08:00:00Z"
}
```

**Error - Invalid Token (401):**
```bash
curl -X GET http://localhost:8000/api/v1/auth/profile \
  -H "Authorization: Bearer invalid-token"
```
```json
{
  "success": false,
  "code": 401,
  "message": "Token tidak valid",
  "error_code": -1010,
  "timestamp": "2025-12-16T08:00:00Z"
}
```

---

### Update Profile (Protected)
```bash
curl -X PUT http://localhost:8000/api/v1/auth/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Updated"
  }'
```

**Success Response (200):**
```json
{
  "success": true,
  "code": 200,
  "message": "Profil berhasil diperbarui",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Updated",
    "email": "user@example.com",
    "role": "user",
    "created_at": "2025-12-16T08:00:00Z",
    "updated_at": "2025-12-16T08:01:00Z"
  },
  "timestamp": "2025-12-16T08:01:00Z"
}
```

---

### Logout (Protected)
```bash
curl -X POST http://localhost:8000/api/v1/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200):**
```json
{
  "success": true,
  "code": 200,
  "message": "Logout berhasil",
  "data": null,
  "timestamp": "2025-12-16T08:00:00Z"
}
```

---

## RBAC (Role-Based Access Control)

### Get My Roles (Protected)
```bash
curl -X GET http://localhost:8000/api/v1/auth/my-roles \
  -H "Authorization: Bearer <access_token>"
```

**Success Response (200):**
```json
{
  "success": true,
  "code": 200,
  "message": "Your roles retrieved successfully",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "roles": [
      {"id": "...", "name": "user", "description": "Standard user access"}
    ]
  }
}
```

### Get My Permissions (Protected)
```bash
curl -X GET http://localhost:8000/api/v1/auth/my-permissions \
  -H "Authorization: Bearer <access_token>"
```

---

## Admin Endpoints (Requires admin/super_admin role)

### Get User Roles
```bash
curl -X GET http://localhost:8000/api/v1/admin/users/{userId}/roles \
  -H "Authorization: Bearer <admin_token>"
```

### Assign Role to User
```bash
curl -X POST http://localhost:8000/api/v1/admin/users/{userId}/roles \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"role_id": "<role_uuid>"}'
```

### Remove Role from User
```bash
curl -X DELETE http://localhost:8000/api/v1/admin/users/{userId}/roles/{roleId} \
  -H "Authorization: Bearer <admin_token>"
```

---

## Super Admin Endpoints (Requires super_admin role)

### Get All Roles
```bash
curl -X GET http://localhost:8000/api/v1/super-admin/roles \
  -H "Authorization: Bearer <super_admin_token>"
```

### Create Role
```bash
curl -X POST http://localhost:8000/api/v1/super-admin/roles \
  -H "Authorization: Bearer <super_admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "moderator", "description": "Content moderator"}'
```

### Get All Permissions
```bash
curl -X GET http://localhost:8000/api/v1/super-admin/permissions \
  -H "Authorization: Bearer <super_admin_token>"
```

### Assign Permission to Role
```bash
curl -X POST http://localhost:8000/api/v1/super-admin/roles/{roleId}/permissions \
  -H "Authorization: Bearer <super_admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"permission_id": "<permission_uuid>"}'
```

**Error - Forbidden (403):**
```json
{
  "success": false,
  "code": 403,
  "message": "Akses ditolak",
  "error_code": -1006,
  "timestamp": "2025-12-16T08:00:00Z"
}
```

---

## Rate Limiting

**Error - Rate Limit Exceeded (429):**
```json
{
  "success": false,
  "code": 429,
  "message": "Terlalu banyak permintaan, coba lagi nanti",
  "error_code": -1012,
  "timestamp": "2025-12-16T08:00:00Z"
}
```

---

## 🧪 Testing

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./internal/module/auth/
```

---

## 📝 Environment Variables

```env
# Server
APP_NAME=Go Fiber Auth API
APP_ENV=development
APP_PORT=8000
APP_PREFORK=false

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=boilerplate_dev
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h

# Rate Limiting
RATE_LIMIT_MAX=100
RATE_LIMIT_WINDOW=1m
```

---

## 📜 License

MIT License
