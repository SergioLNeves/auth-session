# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based authentication service implementing JWT authentication with RS256 (asymmetric signing) and database-backed session management. Built with Echo framework, SQLite (GORM) for data persistence, and `samber/do` for dependency injection. On logout, sessions are deleted from the database (not just cookie clearing).

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
cmd/api/main.go              -> Entry point; DI container setup; route configuration
internal/
  |- handler/                -> HTTP handlers (presentation layer)
  |- middleware/              -> Session authentication middleware
  |- service/                -> Business logic orchestration
  |- repository/             -> Data access (auth and session repositories)
  |- storage/sqlite/         -> SQLite implementation (GORM)
  |- domain/                 -> Core entities, DTOs, and interface definitions
  |- config/                 -> Configuration and environment management
  |- pkg/                    -> Reusable utilities (logging, validation, error handling)
  +- security/               -> JWT (RS256 signing/parsing) and bcrypt password hashing
```

### Request Flow
1. HTTP request -> **Middleware** (SessionAuth for protected routes)
2. Middleware -> **Handler** (validates input, converts to DTOs)
3. Handler -> **Service** (executes business logic)
4. Service -> **Repository interface** (defined in domain)
5. Repository -> **Storage implementation** (SQLite)
6. Response flows back through the same layers

### Dependency Injection

The project uses `github.com/samber/do` for dependency injection. All dependencies are registered in `cmd/api/main.go:initDependencies()`.

**Registration order:**
```
Logger -> SQLite -> AuthRepository -> SessionRepository -> JWTProvider -> BcryptHasher -> Services -> Handlers
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

GORM is used for SQLite interaction. Storage models live in `internal/storage/sqlite/models.go`:

```go
type UserTable struct {
    ID        uuid.UUID  `gorm:"type:uuid;primary_key"`
    Email     string     `gorm:"uniqueIndex;not null"`
    Password  string     `gorm:"not null"`
    DeletedAt *time.Time `gorm:"index"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type SessionTable struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key"`
    UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
    ExpiresAt time.Time `gorm:"not null;index"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

Repository implementations use table name constants (e.g., `TableUser = "user"`, `TableSession = "session"`) when calling storage methods.

### Storage Interface

The storage layer exposes generic methods with table name parameters:

```go
type Storage interface {
    Ping(ctx context.Context) error
    Writer   // Insert, Update, FindOneAndDelete
    Reader   // GetDB
    Querier  // FindByEmail, FindByID
}

type Writer interface {
    Insert(ctx context.Context, table string, data any) error
    Update(ctx context.Context, table string, data any) error
    FindOneAndDelete(ctx context.Context, table string, id any, dest any) error
}

type Querier interface {
    FindByEmail(ctx context.Context, table, email string, dest any) error
    FindByID(ctx context.Context, table string, id any, dest any) error
}
```

## Authentication Implementation

### Current State
- **Account creation**: `POST /v1/user/create-account` -- full flow with validation, bcrypt hash, session creation and JWT generation
- **Login**: `POST /v1/auth/login` -- full flow with email/password verification, session creation and JWT generation
- **Logout**: `POST /v1/auth/logout` -- protected by SessionAuth middleware, deletes session from DB, clears cookies
- **Me**: `GET /v1/auth/me` -- protected by SessionAuth middleware, returns user data from session context
- **Update password**: `PATCH /v1/user/password` -- protected by SessionAuth middleware, validates current password
- **Update profile**: `PATCH /v1/user/profile` -- protected by SessionAuth middleware, updates user fields
- **Delete user**: `DELETE /v1/user` -- protected by SessionAuth middleware, deletes all sessions and user (self-deletion)
- **Session auth middleware**: `internal/middleware/session_auth.go` -- validates session, handles refresh token rotation
- **Password hashing**: bcrypt cost 12 (`internal/security/bcrypt.go`) via `PasswordHasher` interface

### JWT (RS256)

Token generation and parsing is handled by `internal/security/jwt.go` (`JWTProvider`):
- Loads both private key (signing) and public key (verification) from PEM files at startup
- Access token claims: `sub` (sessionID), `iat`, `exp` -- pure authentication token
- Refresh token claims: `sub` (userID), `session_id`, `iat`, `exp`
- `ParseAccessToken` uses `jwt.WithoutClaimsValidation()` to allow parsing expired tokens (needed for middleware refresh flow)
- `ParseAccessToken` returns `*AccessTokenClaims` (SessionID only)
- `ParseRefreshToken` returns `*RefreshTokenClaims` (UserID + SessionID)
- Expiry durations are configured via environment variables in **minutes**
- User information (email, name, avatar) is NOT stored in tokens; it comes from the `/me` endpoint

### Sessions

Sessions are persisted in the `session` table (SQLite). Each login/account creation produces a new session row with a UUID. The session ID is embedded as `sub` in the access token JWT claims. On logout, the session is **deleted** from the database via `FindOneAndDelete` (not deactivated -- the session table has no `active` field).

Key interfaces:
- `domain.SessionRepository`: `CreateSession`, `FindSessionByID`, `DeleteSession`, `UpdateSessionExpiry`, `DeleteExpiredSessions`, `DeleteSessionsByUserID`
- Implemented in `internal/repository/session.go`

### Session Auth Middleware

`internal/middleware/session_auth.go` protects authenticated routes:
1. Parses access token (allows expired via `WithoutClaimsValidation`) -- extracts `sessionID` from `sub`
2. Validates session exists in DB
3. Checks refresh token:
   - If expired: deletes session from DB, clears cookies, returns 401
   - If valid: finds user by `session.UserID` (from DB, not token), regenerates both tokens, sets new cookies
4. Injects `user_id`, `email`, `name`, `avatar`, `session_id` into Echo context via `c.Set()`

Applied per-route: `authGroup.POST("/logout", handler.Logout, sessionAuth)`

### Cookies

Auth cookies are set via `setAuthCookies()` in both handler and middleware:
- `access_token`: HttpOnly, SameSite=Strict, Secure in production
- `refresh_token`: HttpOnly, SameSite=Strict, Secure in production
- Both cookies are **HttpOnly** -- user data is accessed via the `/me` endpoint instead of decoding tokens client-side
- MaxAge is derived from env variables (minutes x 60 = seconds)
- Cleared via `clearAuthCookies()` on logout (MaxAge=-1)

## Configuration

Environment variables are loaded from `.env` and parsed into `internal/domain/env.go`. Configuration is accessed via the global `config.Env` variable.

**Key environment variables:**
- `ENV`: `development` or `production` (affects cookie Secure flag)
- `PORT`: Server port (default 8080)
- `LOG_LEVEL`: Log level - debug, info, warn, error (default debug)
- `PRIVATE_KEY_PATH` / `PUBLIC_KEY_PATH`: RSA key file paths
- `ACCESS_TOKEN_EXPIRY`: Access token lifetime in minutes (default 60)
- `REFRESH_TOKEN_EXPIRY`: Refresh token lifetime in minutes (default 10080)
- `DB_PATH`: SQLite database file path
- `DB_MAX_CONN`: Max open connections (default 10)
- `DB_MAX_IDLE`: Max idle connections (default 5)
- `DB_MAX_LIFETIME`: Max connection lifetime (default 1h)

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
- Test files follow `*_test.go` naming, same package as source
- Arrange/Act/Assert pattern with `t.Parallel()` in subtests
- See `.claude/rules/Test_Example.md` for detailed conventions

### Existing Tests
- `internal/handler/auth_test.go`
- `internal/service/auth_test.go`
- `internal/middleware/session_auth_test.go`

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
- Password recovery is not yet implemented
