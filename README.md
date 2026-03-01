# Go Production API

A 12-factor compliant, production-grade REST API built with Go (1.23+), following Clean Architecture principles for maximum maintainability and testability.

## 1. Executive Summary

This API serves as a robust foundation for modern web projects. It is engineered with a focus on:
- **Clean Architecture**: Decoupling business logic from external dependencies (DBs, frameworks).
- **Scalability**: Stateless JWT authentication and distributed Redis caching/rate-limiting.
- **Security**: Argon2/Bcrypt password hashing, dual-token strategy (Access/Refresh), RBAC, and IP-based rate limiting.
- **Observability**: Structured JSON logging with `request_id` propagation, panic recovery, and **Deep Health Diagnostics** (reporting raw errors in `/health`).
- **Reliability**: Graceful degradation—if Redis fails, the system continues to serve requests via Postgres. Supports `REDIS_URL` for seamless cloud integration (Railway, Render).

## 2. High-Level Architecture Diagram

```mermaid
graph TD
    Client[Client] --> G[Gateway/Router]
    
    subgraph Middleware Layer
        G --> RL["IP Rate Limiter\n(Redis/Memory)"]
        RL --> L[Structured Logger]
        L --> Rec[Panic Recovery]
        Rec --> Auth["JWT Auth\n(Dual Token + Blacklist)"]
        Auth --> RB[Role-Based Access Control]
    end

    subgraph Delivery Layer
        RB --> H[HTTP Handlers]
    end

    subgraph Business Logic Layer
        H --> U[User / Admin UseCases]
    end

    subgraph Data Layer
        U --> R[Repositories]
        R --> DB[(PostgreSQL)]
        U -.-> C[(Redis Cache-Aside)]
    end




## 3. The 'Clean Architecture' Implementation

The project follows the flow: **Delivery -> Usecase -> Repository**.

- **/cmd/api**: Entry point. Handles configuration loading, dependency injection (DI), and server startup.
- **/internal/domain**: Core entities and interface definitions. The "Heart" of the app, containing NO external dependencies.
- **/internal/usecase**: Pure business logic. Implements the `Usecase` interfaces.
- **/internal/repository**: Data access implementations. Maps domain requests to SQLC-generated queries.
- **/internal/delivery/http**: Framework-specific code (Chi router, handlers, DTOs).

## 4. Security & Performance Features

### Authentication & Authorization
- **Argon2/Bcrypt**: Industry-standard password hashing (cost 12).
- **Dual Token JWT Strategy**: 
    - **Access Token**: Short-lived (15m).
    - **Refresh Token**: Long-lived (7d).
- **Session Revocation**: A distributed Redis blacklist stores `jti` (JWT ID) on logout, invalidating tokens across the cluster.
- **RBAC**: Fine-grained access control using `user`, `moderator`, and `admin` roles.

### Performance & Hardening
- **Cache-Aside Pattern**: User profiles are cached in Redis with a **15-minute TTL**.
- **Active Invalidation**: Profile cache keys are purged on updates (`PUT /me`) or deletes (`DELETE /admin/users/{id}`).
- **Global Rate Limiting**: Distributed IP-based limiting using Redis (falls back to local memory if Redis is down).
- **Secure Headers**: `X-Content-Type-Options`, `X-Frame-Options`, and strict `CORS` policies.

## 5. The Testing Fortress

Quality is baked in through:
- **Unit Testing**: 100% logic coverage in `/internal/usecase` using manual mocks.
- **Integration/E2E Testing**: Utilizing `testcontainers-go` to spin up a real PostgreSQL container for API flow validation.
- **CI/CD**: GitHub Actions pipeline including:
    - `golangci-lint` for static analysis.
    - `govulncheck` for security vulnerabilities.
    - `-race` detector for concurrency safety.

## 6. Getting Started

### Prerequisites
- Go 1.23+
- Docker & Docker Compose

### Quick Start
1. **Clone & Setup**:
   ```bash
   git clone <repo-url>
   cd go-production-api
   cp .env.example .env
   ```
2. **Launch Infrastructure**:
   ```bash
   docker-compose up -d
   ```
3. **Run Tests**:
   ```bash
   make test
   ```
4. **Development Run**:
   ```bash
   make run
   ```

## 7. API Documentation (Sneak Peek)

| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/auth/register` | Register new user | No |
| `POST` | `/api/v1/auth/login` | Obtain Access & Refresh tokens | No |
| `POST` | `/api/v1/auth/logout` | Revoke current session (Blacklist JTI) | Yes |
| `GET` | `/api/v1/users/me` | Fetch current user profile (Cached) | Yes |
| `PUT` | `/api/v1/users/me` | Update profile (Invalidates Cache) | Yes |
| `GET` | `/api/v1/admin/users` | List all users | Yes (Admin) |
| `DELETE` | `/api/v1/admin/users/{id}` | Soft delete user (Invalidates Cache) | Yes (Admin) |
| `GET` | `/health` | API Health Check | No |

## 8. Testing with Postman & cURL

All commands below use cURL, which can be **imported directly into Postman** (File > Import > Paste Raw Text).

### 1. Health Check
```bash
curl -i https://go-production-api-production-35b9.up.railway.app/health
```

### 2. Register User
```bash
curl -i -X POST https://go-production-api-production-35b9.up.railway.app/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

**JSON Body (Postman):**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

### 3. Login
```bash
curl -i -X POST https://go-production-api-production-35b9.up.railway.app/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

**JSON Body (Postman):**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

### 4. Fetch Profile (Protected)
*Replace `{{ACCESS_TOKEN}}` with the token returned from the login response.*
```bash
curl -i -X GET https://go-production-api-production-35b9.up.railway.app/api/v1/users/me \
  -H "Authorization: Bearer {{ACCESS_TOKEN}}"
```

### 5. Logout (Revoke Session)
```bash
curl -i -X POST https://go-production-api-production-35b9.up.railway.app/api/v1/auth/logout \
  -H "Authorization: Bearer {{ACCESS_TOKEN}}"
```

