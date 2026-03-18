# PF2E Companion — Project Conventions

This document captures the conventions Claude should follow when working in this codebase.

---

## Project Structure

```
pf2e-companion/
├── backend/                 # Go Echo API server
│   ├── database/            # DB connection setup (database.go)
│   ├── handlers/            # HTTP route handlers + shared helpers
│   ├── models/              # GORM structs and response DTOs
│   └── main.go              # Entry point, route wiring
├── database/
│   └── migrations/          # Flyway SQL migration files
├── docker-compose.yml
├── Makefile
└── .env                     # Local environment variables (not committed)
```

---

## Go Conventions

- Follow [Effective Go](https://go.dev/doc/effective_go).
- Always run `gofmt` — code must be formatted before committing.
- **Exported** identifiers use `PascalCase` (e.g. `UserResponse`, `Connect`).
- **Unexported** identifiers use `camelCase` (e.g. `getEnv`, `parseBody`).
- Prefer explicit error handling over panics.
- No global mutable state outside of the DB handle passed via closure or dependency injection.

---

## JSON Naming

All request and response JSON fields use **snake_case**, enforced via struct tags:

```go
type Example struct {
    FieldName string `json:"field_name"`
}
```

---

## Response Envelope

All API responses follow a consistent envelope:

**Success:**
```json
{ "data": <payload> }
```

**Error:**
```json
{ "code": <http_status_int>, "message": "<human-readable description>" }
```

Use the shared helpers in `handlers/helpers.go`:
- `SuccessResponse(c, status, data)`
- `ErrorResponse(c, status, message)`

---

## HTTP Status Codes

| Situation                        | Status |
|----------------------------------|--------|
| Resource created                 | 201    |
| Read / update / delete success   | 200    |
| Malformed request (bad JSON, bad UUID, etc.) | 400 |
| Resource not found               | 404    |
| Validation failure (missing required fields, business rule) | 422 |
| Unexpected server error          | 500    |

---

## PATCH Semantics

- PATCH endpoints perform **partial updates** — only fields present in the request body are applied.
- **JSONB columns** (e.g. `foundry_data`, `skills`, `content`) are replaced **wholesale** when provided. There is no partial merge of nested JSON objects.

---

## Sensitive Fields

- `User` model includes `password_hash` (stored in DB, never returned in responses).
- Always use the `UserResponse` DTO (defined in `models/responses.go`) when returning user data.
- `FromUser(user User) UserResponse` strips `password_hash`.

---

## Backend Server

- Default port: **8080**
- Override with the `PORT` environment variable.
- Start locally with: `make backend-run`

---

## Database Configuration

The backend reads database credentials from these environment variables (defined in `.env`):

| Variable            | Description              |
|---------------------|--------------------------|
| `POSTGRES_USER`     | Database username        |
| `POSTGRES_PASSWORD` | Database password        |
| `POSTGRES_DB`       | Database name            |
| `POSTGRES_HOST`     | Host (default: localhost)|
| `POSTGRES_PORT`     | Port (default: 5432)     |

Schema migrations are managed by **Flyway** (not GORM AutoMigrate). Never call `AutoMigrate` in application code.
