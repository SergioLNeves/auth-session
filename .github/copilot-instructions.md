# Copilot Instructions

Go authentication service with JWT (RS256) + database-backed sessions. Echo framework, SQLite/GORM, `samber/do` for DI.

## Commands

```bash
make setup          # Install deps, tools (mockery, golangci-lint, air), generate RSA keys
make run            # Run with Air (hot reload)
make lint           # golangci-lint with .golangci.yml config
make mocks          # Generate mocks (outputs to mock/)
go test ./...       # Run all tests
```

## Architecture

Layered architecture with strict separation: Handler -> Service -> Repository -> Storage.

All interfaces are defined in `internal/domain/`. Implementations live in their respective packages. Dependencies are wired via `samber/do` in `cmd/api/main.go:initDependencies()`, registered in order: Logger -> SQLite -> Repositories -> JWTProvider -> BcryptHasher -> Services -> Handlers.

**Adding a new component** always follows this pattern:
1. Define interface in `internal/domain/`
2. Implement with a `New*` constructor that takes `*do.Injector`
3. Register with `do.Provide(injector, ...)` in `main.go`
4. Run `make mocks` if interfaces changed

### Session Auth Middleware

`internal/middleware/session_auth.go` protects routes. It parses the access token (allows expired via `WithoutClaimsValidation`), validates the session exists in DB, then checks the refresh token. If refresh is expired, it **deletes the session** from the DB and clears cookies. If valid, it regenerates both tokens and sets `user_id`, `email`, `session_id` in the Echo context via `c.Set()`. Apply it per-route: `authGroup.POST("/logout", handler.Logout, sessionAuth)`.

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

Repositories use table name constants (`TableUser = "user_tables"`, `TableSession = "session_tables"`) and call generic storage methods (`Insert`, `Update`, `FindByID`, `FindByEmail`, `FindOneAndDelete`). New GORM models must be registered in `GetModelsToMigrate()`.

### JWT

- Access token: `sub`, `email`, `session_id`, `iat`, `exp` -- parsed with `WithoutClaimsValidation()` (allows expired)
- Refresh token: `sub`, `session_id`, `iat`, `exp` -- parsed with expiration validation
- Expiry values in env are in **minutes**
- RSA keys loaded from PEM files at startup

### Cookies

- `access_token`: **not** HttpOnly (JS reads claims for UI)
- `refresh_token`: HttpOnly
- Both: SameSite=Strict, Secure in production, MaxAge = env minutes x 60

### Validation

`go-playground/validator/v10` tags on domain structs. Validate in handlers with `validatorpkg.NewValidator().Validate(request)`.

### Sessions

Sessions are **deleted** from the database on logout (not deactivated). The `session_tables` has no `active` field -- it only stores `id`, `user_id`, `created_at`, `updated_at`. The `DeleteSession` method uses `FindOneAndDelete` to atomically find and remove the session row.

### Password Hashing

Passwords are hashed with bcrypt (cost 12) via the `PasswordHasher` interface. The `BcryptHasher` implementation lives in `internal/security/bcrypt.go`.

## Current State

- **Implemented**: create account (with validation, bcrypt hash, session creation, JWT generation), login (email/password verification, session creation, JWT generation), logout (session deletion from DB, cookie clearing), session auth middleware (with refresh token rotation), health check, frontend pages (create-account, login, password, success)
- **Not implemented**: password recovery (only has static HTML page)

## Testing

- Test files live alongside source: `auth.go` -> `auth_test.go`, same package
- Pattern: `TestFunctionName` with `t.Run` subtests, always `t.Parallel()`
- Arrange/Act/Assert structure
- Mocks via mockery + testify: `mock.NewMock<Interface>(t)`
- Existing tests: `handler/auth_test.go`, `service/auth_test.go`, `middleware/session_auth_test.go`
