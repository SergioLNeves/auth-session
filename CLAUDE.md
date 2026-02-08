# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based authentication service implementing JWT authentication with RS256 (asymmetric signing) and database-backed session management. Built with Echo framework, SQLite (GORM) for data persistence, and `samber/do` for dependency injection. On logout, sessions are deactivated server-side (not just cookie clearing).

## Development Commands

### Initial Setup
```bash
make setup          # Install dependencies, tools (mockery, golangci-lint, air), and generate RSA keys
```

### Running the Application
```bash
make run            # Run with Air (hot reload enabled)
```

The server runs on the port specified in the `.env` file (see `internal/config/enviroment.go` for configuration).

### Code Quality
```bash
make lint           # Run golangci-lint with project-specific configuration
make mocks          # Generate test mocks using mockery (outputs to mock/ directory)
```

### Key Generation
```bash
make gen-key        # Generate RSA key pair (private-key.pem and public-key.pem)
```

The keys are required for JWT signing/verification. Never commit `.pem` files (already in `.gitignore`).

## Architecture

The project follows a **layered architecture** with strict separation of concerns:

```
cmd/api/main.go              → Entry point; DI container setup; route configuration
internal/
  ├─ handler/                → HTTP handlers (presentation layer)
  ├─ service/                → Business logic orchestration
  ├─ repository/             → Data access (auth and session repositories)
  ├─ storage/sqlite/         → SQLite implementation (GORM)
  ├─ domain/                 → Core entities, DTOs, and interface definitions
  ├─ config/                 → Configuration and environment management
  ├─ pkg/                    → Reusable utilities (logging, validation, error handling)
  └─ security/               → JWT (RS256 signing/parsing) and bcrypt password hashing
assets/
  ├─ html/                   → Pages (create-account, login, password, success)
  ├─ css/                    → Stylesheets
  └─ js/                     → Scripts (auth utilities, form handlers)
```

### Request Flow
1. HTTP request → **Handler** (validates input, converts to DTOs)
2. Handler → **Service** (executes business logic)
3. Service → **Repository interface** (defined in domain)
4. Repository → **Storage implementation** (SQLite)
5. Response flows back through the same layers

### Dependency Injection

The project uses `github.com/samber/do` for dependency injection. All dependencies are registered in `cmd/api/main.go:initDependencies()`.

**Registration order:**
```
SQLite → AuthRepository → SessionRepository → JWTProvider → Services → Handlers
```

**Pattern for new components:**
```go
// In domain package: define interface
type MyService interface {
    DoSomething(ctx context.Context) error
}

// In service package: implement
type MyServiceImpl struct {
    repo domain.MyRepository
}

func NewMyService(i *do.Injector) (domain.MyService, error) {
    repo := do.MustInvoke[domain.MyRepository](i)
    return &MyServiceImpl{repo: repo}, nil
}

// In main.go: register
do.Provide(injector, service.NewMyService)
```

## Key Patterns and Conventions

### Error Handling

Use the **ProblemDetails** pattern (RFC 7807) for HTTP error responses:

```go
problemDetails := errorpkg.NewProblemDetails().
    WithType("auth", "validation-error").
    WithTitle("Validation Failed").
    WithStatus(http.StatusBadRequest).
    WithDetail("One or more fields failed validation").
    WithInstance(c.Request().URL.Path)
return c.JSON(http.StatusBadRequest, problemDetails)
```

For validation errors, use `AddFieldErrors()` to include field-specific details.

### Logging

Use the centralized Zap logger from `internal/pkg/logging`:

```go
logger := logging.With(zap.String("handler", "AuthHandler.CreateAccount"))
logger.Error("failed to create account", zap.Error(err))
```

### Validation

Request validation uses `go-playground/validator/v10`. Add validation tags to domain structs:

```go
type CreateAccountRequest struct {
    Email    string `form:"email" validate:"required,email"`
    Password string `form:"password" validate:"required,min=8"`
}
```

Validate in handlers using:
```go
if err := validatorpkg.NewValidator().Validate(request); err != nil {
    // Handle validation errors with ProblemDetails
}
```

### Database Models

GORM is used for SQLite interaction. Domain entities use GORM tags. Storage models live in `internal/storage/sqlite/models.go`:

```go
type UserTable struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key"`
    Email     string    `gorm:"uniqueIndex;not null"`
    Password  string    `gorm:"not null"`
    Active    bool      `gorm:"default:true"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

Repository implementations use table name constants (e.g., `TableUser = "user_tables"`, `TableSession = "session_tables"`) when calling storage methods.

### Storage Interface

The storage layer exposes generic methods with table name parameters:

```go
type Writer interface {
    Insert(ctx context.Context, table string, data any) error
    Update(ctx context.Context, table string, data any) error
}
type Querier interface {
    FindByEmail(ctx context.Context, table, email string, dest any) error
    FindByID(ctx context.Context, table string, id any, dest any) error
}
```

## Authentication Implementation

### Current State
- **Account creation**: `POST /v1/user/create-account` — full flow with session creation and JWT generation
- **Login**: `POST /v1/auth/login` — handler stub (needs implementation)
- **Logout**: `POST /v1/auth/logout` — reads access_token cookie, deactivates session in DB, clears cookies
- **Password hashing**: bcrypt cost 12 (`internal/security/bcrypt.go`)

### JWT (RS256)

Token generation and parsing is handled by `internal/security/jwt.go` (`JWTProvider`):
- Loads both private key (signing) and public key (verification) from PEM files at startup
- Access token claims: `sub` (userID), `email`, `session_id`, `iat`, `exp`
- Refresh token claims: `sub` (userID), `session_id`, `iat`, `exp`
- `ParseAccessToken` uses `jwt.WithoutClaimsValidation()` to allow parsing expired tokens (needed for logout)
- Expiry durations are configured via environment variables in **minutes**

### Sessions

Sessions are persisted in `session_tables` (SQLite). Each login/account creation produces a new session row with a UUID. The session ID is embedded as `session_id` in JWT claims. On logout, the session is marked `active=false`.

Key interfaces:
- `domain.SessionRepository`: `CreateSession`, `FindSessionByID`, `DeactivateSession`
- Implemented in `internal/repository/session.go`

### Cookies

Auth cookies are set via `setAuthCookies()` in the handler:
- `access_token`: **not** HttpOnly (readable by JS for claim extraction), SameSite=Strict, Secure in production
- `refresh_token`: HttpOnly, SameSite=Strict, Secure in production
- MaxAge is derived from env variables (minutes × 60 = seconds)
- Cleared via `clearAuthCookies()` on logout

### Frontend Auth

`assets/js/auth.js` provides shared utilities:
- `getUser()`: decodes JWT from cookie, returns `{id, email}` or null
- `requireAuth()`: redirects to `/login` if not authenticated
- `requireGuest()`: redirects to `/` if already authenticated
- `logout()`: calls `POST /v1/auth/logout` and redirects to `/login`

## Configuration

Environment variables are loaded from `.env` and parsed into `internal/domain/env.go`. Configuration is accessed via the global `config.Env` variable.

**Key environment variables:**
- `ENV`: `development` or `production` (affects cookie Secure flag)
- `PORT`: Server port (default 8080)
- `PRIVATE_KEY_PATH` / `PUBLIC_KEY_PATH`: RSA key file paths
- `ACCESS_TOKEN_EXPIRY`: Access token lifetime in minutes (default 60)
- `REFRESH_TOKEN_EXPIRY`: Refresh token lifetime in minutes (default 10080)
- `DB_PATH`: SQLite database file path

## Testing

### Mock Generation
```bash
make mocks
```

Generates mocks for interfaces in:
- `internal/domain`
- `internal/repository`
- `internal/service`
- `internal/storage`

Configuration: `.mockery.yml`

### Testing Conventions
- Mocks use testify framework (`github.com/stretchr/testify`)
- Test files follow `*_test.go` naming (excluded from Air builds)

## Code Style and Linting

Linting configuration in `.golangci.yml`:
- Enabled linters: gocritic, misspell, revive, unconvert, unparam, whitespace
- Formatters: gofmt (with simplify), goimports
- Local import prefix: `github.com/SergioLNeves/auth-session`

**Import organization:** Group imports as:
1. Standard library
2. External dependencies
3. Internal packages (prefixed with module path)

## Common Development Patterns

### Adding a New Endpoint

1. Define request/response DTOs in `internal/domain/`
2. Add interface method to appropriate domain interface (e.g., `AuthHandler`, `AuthService`)
3. Implement in corresponding service and handler packages
4. Register route in `cmd/api/main.go` (e.g., in `configureAuthRoute`)
5. Run `make mocks` if new interfaces were added

### Adding Database Operations

1. Define repository method in `internal/domain/` interface (e.g., `AuthRepository`, `SessionRepository`)
2. Implement in `internal/repository/` using the storage interface with table name constants
3. Add storage method to `internal/storage/storage.go` if needed
4. Implement concrete SQLite version in `internal/storage/sqlite/sqlite.go`
5. Add GORM model to `internal/storage/sqlite/models.go` and register in `GetModelsToMigrate()`

## Notes

- Air configuration (`.air.toml`) watches `.go`, `.html`, and template files
- The project uses Go 1.25.4
- SQLite database location is configured via the `DB_PATH` env variable
- Assets (HTML, CSS, JS) are in `assets/` directory, served as static files via Echo
- Static assets are served at `/css`, `/js`; HTML pages at `/`, `/create-account`, `/login`, `/password`
