# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based authentication service implementing JWT authentication with RS256 (asymmetric signing). The project is currently transitioning from session-based authentication to stateless JWT authentication. Built with Echo framework and SQLite for data persistence.

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
  ├─ repository/             → Data access interfaces
  ├─ storage/sqlite/         → SQLite implementation of repositories
  ├─ domain/                 → Core entities, DTOs, and interface definitions
  ├─ config/                 → Configuration and environment management
  ├─ pkg/                    → Reusable utilities (logging, validation, error handling)
  └─ security/               → Security utilities (bcrypt password hashing)
```

### Request Flow
1. HTTP request → **Handler** (validates input, converts to DTOs)
2. Handler → **Service** (executes business logic)
3. Service → **Repository interface** (defined in domain)
4. Repository → **Storage implementation** (SQLite)
5. Response flows back through the same layers

### Dependency Injection

The project uses `github.com/samber/do` for dependency injection. All dependencies are registered in `cmd/api/main.go:initDependencies()`.

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
    Name     string `form:"name" validate:"required"`
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

GORM is used for SQLite interaction. Domain entities use GORM tags:

```go
type User struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
    Email     string    `gorm:"type:varchar(100);uniqueIndex;not null"`
    Active    bool      `gorm:"not null;default:true"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## Authentication Implementation

### Current State
- Account creation endpoint: `POST /v1/user/create-account`
- Login endpoint: `POST /v1/auth/login` (partially implemented)
- Password hashing: bcrypt (`internal/security/bcrypt.go`)

### JWT Authentication (Planned/In Progress)
The README.md contains detailed implementation guidance for JWT with RS256:

**JWT Flow:**
1. Generate RSA key pair (private-key.pem, public-key.pem)
2. Login handler signs JWT with private key
3. Middleware validates JWT signature with public key
4. Claims include user ID and standard fields (exp, iat)

**JWT Claims Structure:**
```go
type JwtCustomClaims struct {
    UserID string `json:"user_id"`
    jwt.RegisteredClaims
}
```

When implementing JWT features, refer to the Portuguese documentation in README.md for the complete implementation pattern.

## Configuration

Environment variables are loaded from `.env` and parsed into `internal/domain/env.go`. Configuration is accessed via the global `config.Env` variable.

**Important environment variables:**
- `PORT`: Server port (default configured in env struct)
- Database path and other SQLite settings in storage layer

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

1. Define repository method in `internal/domain/` interface (e.g., `AuthRepository`)
2. Implement in `internal/repository/` using the storage interface
3. Add storage method to `internal/storage/storage.go` if needed
4. Implement concrete SQLite version in `internal/storage/sqlite/`

## Notes

- Air configuration (`.air.toml`) watches `.go`, `.html`, and template files
- The project uses Go 1.25.4
- SQLite database location is configured via the storage layer
- Assets (HTML, CSS, JS) are in `assets/` directory for login/password pages
