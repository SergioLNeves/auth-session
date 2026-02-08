# Copilot Instructions

Go authentication service with JWT (RS256) + database-backed sessions. Echo framework, SQLite/GORM, `samber/do` for DI.

## Commands

```bash
make setup          # Install deps, tools (mockery, golangci-lint, air), generate RSA keys
make run            # Run with Air (hot reload)
make lint           # golangci-lint with .golangci.yml config
make mocks          # Generate mocks (outputs to mock/)
go test ./internal/service/... -run TestLogout -v  # Run a single test
```

## Architecture

Layered architecture with strict separation: Handler → Service → Repository → Storage.

All interfaces are defined in `internal/domain/`. Implementations live in their respective packages. Dependencies are wired via `samber/do` in `cmd/api/main.go:initDependencies()`, registered in order: SQLite → Repositories → JWTProvider → Services → Handlers.

**Adding a new component** always follows this pattern:
1. Define interface in `internal/domain/`
2. Implement with a `New*` constructor that takes `*do.Injector`
3. Register with `do.Provide(injector, ...)` in `main.go`
4. Run `make mocks` if interfaces changed

### Session Auth Middleware

`internal/middleware/session_auth.go` protects routes. It parses the access token (allows expired), validates the session is active in DB, then checks the refresh token. If refresh is expired, it deactivates the session and clears cookies. If valid, it regenerates both tokens and sets `user_id`, `email`, `session_id` in the Echo context via `c.Set()`. Apply it per-route: `authGroup.POST("/logout", handler.Logout, sessionAuth)`.

## Conventions

### Error Responses

All HTTP errors use ProblemDetails (RFC 7807):

```go
problemDetails := errorpkg.NewProblemDetails().
    WithType("auth", "validation-error").
    WithTitle("Validation Failed").
    WithStatus(http.StatusBadRequest).
    WithDetail("One or more fields failed validation").
    WithInstance(c.Request().URL.Path)
```

### Logging

Structured logging with Zap. Always add context:

```go
logger := logging.With(zap.String("handler", "AuthHandler.CreateAccount"))
```

### Imports

Three groups separated by blank lines: stdlib, external deps, internal (`github.com/SergioLNeves/auth-session/...`).

### Database

Repositories use table name constants (`TableUser = "user_tables"`, `TableSession = "session_tables"`) and call generic storage methods (`Insert`, `Update`, `FindByID`, `FindByEmail`). New GORM models must be registered in `GetModelsToMigrate()`.

### JWT

- Access token: `sub`, `email`, `session_id`, `iat`, `exp` — parsed with `WithoutClaimsValidation()` (allows expired)
- Refresh token: `sub`, `session_id`, `iat`, `exp` — parsed with expiration validation
- Expiry values in env are in **minutes**
- RSA keys loaded from PEM files at startup

### Cookies

- `access_token`: **not** HttpOnly (JS reads claims for UI)
- `refresh_token`: HttpOnly
- Both: SameSite=Strict, Secure in production, MaxAge = env minutes × 60

### Validation

`go-playground/validator/v10` tags on domain structs. Validate in handlers with `validatorpkg.NewValidator().Validate(request)`.

## Current State

- **Implemented**: create account, logout (with session deactivation), session auth middleware with refresh flow, health check
- **Stub**: login handler (`POST /v1/auth/login` returns 200 with no logic)
- **Not implemented**: password recovery
