# Arquitetura do Projeto

Este documento descreve a arquitetura de alto nivel do projeto de autenticacao. A estrutura segue uma abordagem de design em camadas (Layered Architecture), separando responsabilidades para promover modularidade, testabilidade e manutenibilidade.

## Visao Geral das Camadas

```
cmd/api/main.go                  -> Ponto de entrada, DI e rotas
internal/
  |- handler/                    -> Camada de Apresentacao (HTTP)
  |- middleware/                  -> Middleware de autenticacao
  |- service/                    -> Logica de negocio
  |- repository/                 -> Acesso a dados
  |- storage/sqlite/             -> Implementacao SQLite (GORM)
  |- domain/                     -> Entidades, DTOs e interfaces
  |- security/                   -> JWT (RS256) e bcrypt
  |- config/                     -> Configuracao e ambiente
  +- pkg/                        -> Utilitarios (logging, validacao, erros)
assets/
  |- html/                       -> Paginas HTML
  |- css/                        -> Estilos
  +- js/                         -> Scripts
```

## Descricao das Camadas

### `cmd/` - Ponto de Entrada

O diretorio `cmd/api/` contem a funcao `main` que:
- Carrega as variaveis de ambiente
- Inicializa o logger (Zap)
- Configura o Echo com middlewares (RequestLogger, Recover, CORS)
- Registra todas as dependencias via `samber/do`
- Configura rotas e assets estaticos
- Inicia o servidor HTTP

### `internal/` - Logica da Aplicacao

Este diretorio nao e importavel por outros projetos Go, garantindo encapsulamento.

#### `handler/` (Camada de Apresentacao)

Responsavel por lidar com requisicoes HTTP. Os handlers:
- Fazem bind e validacao dos dados de entrada
- Delegam a logica para os services
- Formatam respostas HTTP (JSON, status codes, cookies)
- Tratam erros com ProblemDetails (RFC 7807)

Handlers existentes:
- `AuthHandlerImpl`: CreateAccount, Login, Logout
- `HealthCheckHandlerImpl`: Check

#### `middleware/` (Camada de Middleware)

Contem o middleware `SessionAuth` que protege rotas autenticadas:
- Parseia o access token (permite expirado)
- Valida a sessao no banco de dados
- Verifica o refresh token:
  - Expirado: deleta a sessao e limpa cookies
  - Valido: regenera ambos os tokens
- Injeta `user_id`, `email` e `session_id` no contexto Echo

#### `service/` (Camada de Servico)

Contem a logica de negocio central:
- Orquestra operacoes entre repositories, token provider e password hasher
- Desacopla os handlers dos detalhes de acesso a dados

Services existentes:
- `AuthServiceImpl`: CreateAccount, Login, Logout
- `HealthCheckServiceImpl`: Check

#### `repository/` (Camada de Repositorio)

Implementa o acesso a dados usando a interface `Storage`:
- Usa constantes de nome de tabela (`TableUser`, `TableSession`)
- Converte erros do GORM (ex.: `ErrRecordNotFound` -> `nil`)

Repositories existentes:
- `AuthRepositoryImpl`: CreateUser, FindUserByEmail, FindUserByID
- `SessionRepositoryImpl`: CreateSession, FindSessionByID, DeleteSession

#### `storage/` (Camada de Armazenamento)

Implementacao concreta do acesso ao banco de dados:

```go
type Storage interface {
    Ping(ctx context.Context) error
    Writer   // Insert, Update, FindOneAndDelete
    Reader   // GetDB
    Querier  // FindByEmail, FindByID
}
```

A implementacao SQLite (`storage/sqlite/`) usa GORM. Modelos GORM sao definidos em `models.go` e registrados em `GetModelsToMigrate()` para migracao automatica.

Se o banco fosse trocado (ex.: PostgreSQL), apenas esta camada precisaria ser modificada.

#### `domain/` (Camada de Dominio)

Contem as estruturas de dados e interfaces centrais:
- Entidades: `User`, `Session`
- DTOs: `CreateAccountRequest`, `LoginRequest`, `AuthResponse`, `TokenClaims`
- Interfaces: `AuthHandler`, `AuthService`, `AuthRepository`, `SessionRepository`, `TokenProvider`, `PasswordHasher`, `HealthCheckHandler`, `HealthCheckService`
- Erros de dominio: `ErrEmailAlreadyExists`, `ErrInvalidCredentials`
- Configuracao: `Config`, `KeysConfig`, `TokenConfig`, `SQLConfig`

#### `security/` (Camada de Seguranca)

- `JWTProvider`: geracao e parsing de tokens JWT com RS256 (chaves RSA)
- `BcryptHasher`: hash e verificacao de senhas com bcrypt (cost 12)

#### `config/` (Configuracao)

- Carrega variaveis de ambiente do `.env` via `godotenv` e `go-env`
- Configura o servidor Echo com timeouts

#### `pkg/` (Utilitarios)

- `logging/`: logger estruturado com Zap
- `validator/`: validacao de structs com `go-playground/validator`
- `error/`: ProblemDetails (RFC 7807) para respostas de erro HTTP

### `assets/` - Frontend

Arquivos estaticos servidos pelo Echo:
- HTML: paginas de criacao de conta, login, recuperacao de senha e sucesso
- CSS: estilos por pagina
- JS: utilitarios de autenticacao (`auth.js`) e handlers de formulario

## Fluxo de uma Requisicao

1. Requisicao HTTP chega ao servidor Echo
2. Middlewares globais executam (RequestLogger, Recover, CORS)
3. Para rotas protegidas, o middleware `SessionAuth` valida a sessao
4. O roteador direciona para o handler apropriado
5. O handler faz bind, valida a requisicao e chama o service
6. O service executa a logica de negocio usando repositories e providers
7. O repository acessa o banco via interface `Storage`
8. A implementacao SQLite executa a operacao via GORM
9. O resultado retorna pela mesma cadeia ate o handler enviar a resposta HTTP

## Injecao de Dependencias

Todas as dependencias sao registradas em `cmd/api/main.go` usando `samber/do`:

```
Logger -> SQLite -> AuthRepository -> SessionRepository -> JWTProvider -> BcryptHasher -> Services -> Handlers
```

Cada componente recebe um `*do.Injector` no construtor e resolve suas dependencias via `do.MustInvoke`.
