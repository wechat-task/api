# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the API service for a WeChat clawbot-based task management system. It provides a RESTful API for task CRUD operations with PostgreSQL persistence.

## Development Commands

```bash
# Local development with Docker (starts PostgreSQL + API)
docker-compose up --build

# Build the Go binary
go build -o server .

# Run tests
go test ./...

# Run specific package tests
go test ./internal/handler
```

## Code Verification

**MANDATORY**: After writing or modifying any code, you MUST run these commands to verify correctness:

```bash
# Format code
go fmt ./...

# Update dependencies
go mod tidy

# Build to verify compilation
go build -o server .
```

Never skip these steps. They ensure code quality, dependency hygiene, and catch compilation errors before committing.

The codebase follows a strict three-layer architecture within the `internal/` package:

1. **Handler** (`internal/handler/`) - HTTP request/response handling, input validation
2. **Service** (`internal/service/`) - Business logic
3. **Repository** (`internal/repository/`) - Database operations using GORM

Dependency injection is manual in `main.go`. When adding new features, wire dependencies in this order: `repository → service → handler`, then register routes in the `/api/v1` group.

**Key architectural decisions:**
- Database migrations run automatically on startup via `database.Migrate(db)`
- Configuration uses environment variables with sensible defaults (see `internal/config/`)
- All models are auto-migrated; add new models to the `Migrate()` function
- Logging uses logrus with JSON format, applied via middleware

## Environment Variables

- `DATABASE_URL` - PostgreSQL connection string (default: `postgres://postgres:postgres@localhost:5432/wechat_task?sslmode=disable`)

## API Design Patterns

- All routes are prefixed with `/api/v1`
- HTTP handlers return JSON responses with appropriate status codes
- Use `gin.Context` methods for parameter binding and error responses
- Follow existing handler patterns for consistency (see `internal/handler/task.go`)

## Database

- PostgreSQL 16 with GORM ORM
- Connection is initialized in `database.Init()` and auto-migrated on startup
- Models in `internal/model/` use GORM struct tags for schema definition
