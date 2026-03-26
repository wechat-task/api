# WeChat Task API

User management API service with Passkeys (WebAuthn) authentication built with Golang + Gin + GORM + PostgreSQL.

## Features

- 🔐 **Passwordless Authentication** - Passkeys/WebAuthn support for secure, passwordless login
- 👤 **User Management** - Simple user profile management with unique usernames
- 📝 **YAML Configuration** - Flexible config with Viper, supports environment variable overrides
- 🪵 **Structured Logging** - Logrus-based logging with JSON output for production
- 🐳 **Docker Support** - Easy deployment with Docker and Docker Compose

## Local Development

### Prerequisites

- Go 1.21+
- PostgreSQL 16+
- Docker & Docker Compose (optional)

### Quick Start

```bash
# Clone the repository
git clone https://github.com/wechat-task/api.git
cd api

# Copy config template
cp config.example.yaml config.yaml

# Start with Docker (includes PostgreSQL)
docker-compose up --build

# Or run locally (requires PostgreSQL)
go mod tidy
go run main.go
```

### Configuration

Copy `config.example.yaml` to `config.yaml` and customize:

```yaml
server:
  port: 8080
  mode: debug  # debug, release, test

database:
  url: postgres://postgres:postgres@localhost:5432/wechat_task?sslmode=disable

webauthn:
  rp_display_name: WeChat Task
  rp_id: localhost
  rp_origins:
    - http://localhost:8080
  timeout: 5m
```

## API Endpoints

### Authentication

All authentication endpoints are public and use Passkeys (WebAuthn) for passwordless authentication.

#### Begin Authentication

```http
POST /api/v1/auth/start
Content-Type: application/json
```

**Response:**
```json
{
  "publicKey": {
    "challenge": "base64url-encoded-challenge",
    "rp": {
      "name": "WeChat Task",
      "id": "localhost"
    },
    "user": {
      "id": "user-id",
      "name": "user-id",
      "displayName": "User"
    },
    "pubKeyCredParams": [...],
    "timeout": 60000
  }
}
```

#### Finish Authentication

```http
POST /api/v1/auth/finish
Content-Type: application/json
Cookie: session_id=<session-from-start>

{
  "id": "credential-id",
  "rawId": "base64url-credential-id",
  "response": {
    "clientDataJSON": "base64url-data",
    "attestationObject": "base64url-attestation"
  },
  "type": "public-key"
}
```

**Success Response (201 Created for new users):**
```json
{
  "user_id": 1,
  "username": null,
  "is_new_user": true
}
```

**Success Response (200 OK for existing users):**
```json
{
  "user_id": 1,
  "username": "john_doe",
  "is_new_user": false
}
```

### User Management

All user endpoints require authentication (session cookie from `/api/v1/auth/finish`).

#### Get Current User

```http
GET /api/v1/user/me
Cookie: session_id=<valid-session>
```

**Response:**
```json
{
  "id": 1,
  "username": "john_doe",
  "icon": "",
  "created_at": "2026-03-26T10:30:00Z",
  "updated_at": "2026-03-26T10:30:00Z"
}
```

#### Set Username

```http
PUT /api/v1/user/username
Content-Type: application/json
Cookie: session_id=<valid-session>

{
  "username": "john_doe"
}
```

**Success Response (200 OK):**
```json
{
  "id": 1,
  "username": "john_doe",
  "icon": "",
  "created_at": "2026-03-26T10:30:00Z",
  "updated_at": "2026-03-26T10:35:00Z"
}
```

**Error Response (409 Conflict):**
```json
{
  "error": "username already taken"
}
```

## Authentication Flow

### First Time User (Registration)

1. **Client** calls `POST /api/v1/auth/start`
2. **Server** returns WebAuthn challenge
3. **Client** creates Passkeys credential using browser's WebAuthn API
4. **Client** calls `POST /api/v1/auth/finish` with credential
5. **Server** validates credential and creates new user
6. **Response** returns `"is_new_user": true`

### Returning User (Login)

1. **Client** calls `POST /api/v1/auth/start`
2. **Server** returns WebAuthn challenge
3. **Client** signs challenge with existing Passkeys credential
4. **Client** calls `POST /api/v1/auth/finish` with assertion
5. **Server** validates assertion and authenticates user
6. **Response** returns `"is_new_user": false` with user data

### Setting Username

After authentication, users can set a unique username:

1. **Client** calls `PUT /api/v1/user/username` with `{"username": "desired_name"}`
2. **Server** checks uniqueness and updates user profile
3. Username can only be set once and must be unique

## Environment Variables

Configuration can be overridden using environment variables. Environment variables take precedence over config file values.

### Database

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:postgres@localhost:5432/wechat_task?sslmode=disable` |

### WebAuthn

| Variable | Description | Default |
|----------|-------------|---------|
| `WEBAUTHN_RP_DISPLAY_NAME` | Display name for WebAuthn | `WeChat Task` |
| `WEBAUTHN_RP_ID` | Relying Party ID (domain) | `localhost` |
| `WEBAUTHN_RP_ORIGINS` | Allowed origins (comma-separated) | `http://localhost:8080` |

### Server

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `GIN_MODE` | Server mode (debug/release/test) | `debug` |

### Example: Docker Compose with Environment Variables

```yaml
services:
  api:
    image: wechat-task-api:latest
    environment:
      - DATABASE_URL=postgres://user:pass@db:5432/wechat_task?sslmode=disable
      - WEBAUTHN_RP_ID=example.com
      - WEBAUTHN_RP_ORIGINS=https://example.com
      - PORT=8080
      - GIN_MODE=release
    ports:
      - "8080:8080"
```

## Development

### Code Verification

After making changes, run these commands:

```bash
go fmt ./...
go mod tidy
go build -o server .
```

### Running Tests

```bash
go test ./...
```

### Project Structure

```
.
├── config.yaml              # Configuration file (not in git)
├── config.example.yaml      # Configuration template
├── main.go                  # Application entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── database/           # Database initialization and migrations
│   ├── handler/            # HTTP request handlers
│   ├── logger/             # Centralized logging
│   ├── middleware/         # Gin middleware
│   ├── model/              # Data models
│   ├── repository/         # Data access layer
│   └── service/            # Business logic layer
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## Architecture

The codebase follows a clean three-layer architecture:

1. **Handler Layer** (`internal/handler/`) - HTTP request/response handling
2. **Service Layer** (`internal/service/`) - Business logic
3. **Repository Layer** (`internal/repository/`) - Database operations with GORM

## Technology Stack

- **Go 1.21+** - Programming language
- **Gin** - HTTP web framework
- **GORM** - ORM for database operations
- **PostgreSQL** - Primary database
- **WebAuthn/Passkeys** - Passwordless authentication (go-webauthn/webauthn)
- **Viper** - Configuration management
- **Logrus** - Structured logging
- **UUID** - Unique identifier generation

## Security

- Passkeys provide phishing-resistant authentication
- Session cookies are HTTP-only and secure
- CSRF protection recommended for production
- Rate limiting recommended for production
- Credentials are stored with proper cryptographic protection

## Production Deployment

### Using Docker

```bash
docker build -t wechat-task-api:latest .
docker run -p 8080:8080 \
  -e DATABASE_URL=$DATABASE_URL \
  -e WEBAUTHN_RP_ID=yourdomain.com \
  -e WEBAUTHN_RP_ORIGINS=https://yourdomain.com \
  -e GIN_MODE=release \
  wechat-task-api:latest
```

### Environment Checklist

Before deploying to production:

- [ ] Set `GIN_MODE=release`
- [ ] Configure production database URL
- [ ] Set proper `WEBAUTHN_RP_ID` (your domain)
- [ ] Set `WEBAUTHN_RP_ORIGINS` to HTTPS URLs
- [ ] Configure HTTPS/TLS termination
- [ ] Set up log aggregation for JSON logs
- [ ] Configure rate limiting
- [ ] Enable CORS if needed for frontend

## License

[Your License Here]

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
