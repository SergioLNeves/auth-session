# Auth Session

Servico de autenticacao em Go com JWT (RS256) e gerenciamento de sessoes persistidas em banco de dados. Construido com o framework Echo, SQLite (GORM) e injecao de dependencias via `samber/do`.

Este repositorio tem fins de estudo e documentacao de um fluxo completo de autenticacao: criacao de conta, login, gerenciamento de sessoes e logout com invalidacao server-side.

## Requisitos

- Go 1.25.4+
- OpenSSL (para geracao das chaves RSA)
- Make

## Configuracao Inicial

```bash
make setup
```

Este comando instala as dependencias do projeto, ferramentas de desenvolvimento (mockery, golangci-lint, air) e gera o par de chaves RSA necessario para assinatura dos tokens.

### Variaveis de Ambiente

Crie um arquivo `.env` na raiz do projeto:

```env
ENV=development
PORT=8080
LOG_LEVEL=DEBUG

PRIVATE_KEY_PATH=private-key.pem
PUBLIC_KEY_PATH=public-key.pem

ACCESS_TOKEN_EXPIRY=60
REFRESH_TOKEN_EXPIRY=10080

DB_PATH=./data/auth-session.db
DB_MAX_CONN=10
DB_MAX_IDLE=5
DB_MAX_LIFETIME=1h
```

| Variavel | Descricao | Padrao |
|---|---|---|
| `ENV` | Ambiente de execucao (`development` ou `production`) | `development` |
| `PORT` | Porta do servidor HTTP | `8080` |
| `LOG_LEVEL` | Nivel de log (`debug`, `info`, `warn`, `error`) | `debug` |
| `PRIVATE_KEY_PATH` | Caminho para a chave privada RSA (.pem) | - |
| `PUBLIC_KEY_PATH` | Caminho para a chave publica RSA (.pem) | - |
| `ACCESS_TOKEN_EXPIRY` | Tempo de expiracao do access token (minutos) | `60` |
| `REFRESH_TOKEN_EXPIRY` | Tempo de expiracao do refresh token (minutos) | `10080` (7 dias) |
| `DB_PATH` | Caminho do banco SQLite | `./data/auth-session.db` |
| `DB_MAX_CONN` | Numero maximo de conexoes abertas | `10` |
| `DB_MAX_IDLE` | Numero maximo de conexoes ociosas | `5` |
| `DB_MAX_LIFETIME` | Tempo de vida maximo de uma conexao | `1h` |

## Execucao

```bash
make run
```

O servidor inicia com hot reload via Air na porta configurada.

## Comandos Disponiveis

| Comando | Descricao |
|---|---|
| `make setup` | Instala dependencias, ferramentas e gera chaves RSA |
| `make run` | Executa a aplicacao com hot reload (Air) |
| `make gen-key` | Gera par de chaves RSA (private-key.pem e public-key.pem) |
| `make mocks` | Gera mocks para testes com Mockery |
| `make lint` | Executa o linter (golangci-lint) |
| `make help` | Exibe os comandos disponiveis |

## Arquitetura

O projeto segue uma arquitetura em camadas com separacao estrita de responsabilidades:

```
cmd/api/main.go                  -> Ponto de entrada, DI e rotas
internal/
  |- handler/                    -> Camada HTTP (validacao, bind, cookies)
  |- middleware/                  -> Middleware de autenticacao de sessao
  |- service/                    -> Logica de negocio
  |- repository/                 -> Acesso a dados
  |- storage/sqlite/             -> Implementacao SQLite (GORM)
  |- domain/                     -> Entidades, DTOs e interfaces
  |- security/                   -> JWT (RS256) e bcrypt
  |- config/                     -> Configuracao e ambiente
  +- pkg/                        -> Utilitarios (logging, validacao, erros)
assets/
  |- html/                       -> Paginas HTML (login, criar conta, etc.)
  |- css/                        -> Estilos
  +- js/                         -> Scripts (auth, formularios)
```

### Fluxo de uma Requisicao

```
HTTP Request -> Middleware (SessionAuth) -> Handler -> Service -> Repository -> Storage (SQLite)
                                             |
                                             +-- Resposta retorna pelo mesmo caminho
```

### Injecao de Dependencias

Todas as dependencias sao registradas em `cmd/api/main.go` usando `samber/do`:

```
SQLite -> Repositories -> JWTProvider -> BcryptHasher -> Services -> Handlers
```

## Endpoints da API

### Paginas

| Metodo | Rota | Descricao |
|---|---|---|
| `GET` | `/` | Pagina de sucesso (requer autenticacao) |
| `GET` | `/create-account` | Formulario de criacao de conta |
| `GET` | `/login` | Formulario de login |
| `GET` | `/password` | Formulario de recuperacao de senha |

### API REST

| Metodo | Rota | Auth | Descricao |
|---|---|---|---|
| `GET` | `/health` | Nao | Health check |
| `POST` | `/v1/user/create-account` | Nao | Criacao de conta |
| `POST` | `/v1/auth/login` | Nao | Login com email e senha |
| `POST` | `/v1/auth/logout` | Sim (SessionAuth) | Logout (deleta sessao do banco) |

### Exemplos de Requisicao

**Criar conta:**
```bash
curl -X POST http://localhost:8080/v1/user/create-account \
  -d "email=usuario@exemplo.com" \
  -d "password=senha12345"
```

**Login:**
```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -d "email=usuario@exemplo.com" \
  -d "password=senha12345"
```

**Resposta (201 Created / 200 OK):**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIs..."
}
```

Os tokens tambem sao setados automaticamente como cookies na resposta.

**Logout:**
```bash
curl -X POST http://localhost:8080/v1/auth/logout \
  --cookie "access_token=eyJhbGciOiJSUzI1NiIs..." \
  --cookie "refresh_token=eyJhbGciOiJSUzI1NiIs..."
```

## Autenticacao

### JWT com RS256

O sistema utiliza tokens JWT assinados com chaves RSA assimetricas (RS256):

- **Chave privada** -- usada para assinar os tokens (mantida no servidor)
- **Chave publica** -- usada para validar assinaturas e parsear tokens

Os arquivos `.pem` sao gerados via `make gen-key` e **nunca devem ser comitados** (ja estao no `.gitignore`).

### Tokens

| Token | Expiracao Padrao | Claims | Cookie |
|---|---|---|---|
| Access Token | 60 min | `sub`, `email`, `session_id`, `iat`, `exp` | `access_token` (legivel pelo JS) |
| Refresh Token | 7 dias | `sub`, `session_id`, `iat`, `exp` | `refresh_token` (HttpOnly) |

O `access_token` e legivel pelo JavaScript para permitir a extracao de claims no frontend (ex.: exibir email do usuario). O `refresh_token` e HttpOnly, inacessivel via JS.

Ambos os cookies utilizam `SameSite=Strict` e `Secure=true` em producao.

### Middleware de Autenticacao (SessionAuth)

O middleware `SessionAuth` protege rotas que requerem autenticacao. Ele executa o seguinte fluxo:

1. Le o cookie `access_token` (permite tokens expirados via `WithoutClaimsValidation`)
2. Extrai o `session_id` dos claims JWT
3. Verifica se a sessao existe no banco de dados
4. Valida o `refresh_token`:
   - Se expirado: **deleta a sessao** do banco e limpa os cookies
   - Se valido: **regenera ambos os tokens** (access e refresh) e seta novos cookies
5. Injeta `user_id`, `email` e `session_id` no contexto do Echo via `c.Set()`

### Gerenciamento de Sessoes

As sessoes sao persistidas no banco de dados (tabela `session_tables`):

| Campo | Tipo | Descricao |
|---|---|---|
| `id` | UUID | Identificador unico da sessao |
| `user_id` | UUID | Referencia ao usuario |
| `created_at` | TIMESTAMP | Data de criacao |
| `updated_at` | TIMESTAMP | Ultima atualizacao |

O `session_id` e incluido nos claims de ambos os tokens JWT, vinculando cada token a uma sessao especifica no banco.

No logout, a sessao e **deletada** do banco (nao apenas desativada). Isso garante que tokens associados a sessao nao possam mais ser usados.

### Seguranca de Senhas

As senhas sao armazenadas com hash bcrypt (cost 12). Nunca sao armazenadas ou trafegadas em texto plano.

## Fluxos

### Criacao de Conta

1. Usuario preenche o formulario em `/create-account`
2. JavaScript envia `POST /v1/user/create-account` com email e senha
3. Handler valida os campos (email valido, senha minimo 8 caracteres)
4. Service verifica se o email ja existe no banco
5. Senha e hasheada com bcrypt (cost 12)
6. Usuario e criado no banco
7. Sessao e criada no banco com um UUID
8. Access token e refresh token sao gerados (RS256) com `session_id` nos claims
9. Tokens sao setados como cookies na resposta HTTP
10. Resposta retorna os tokens em JSON (status 201)

### Login

1. Usuario preenche o formulario em `/login`
2. JavaScript envia `POST /v1/auth/login` com email e senha
3. Handler valida os campos
4. Service busca usuario por email no banco
5. Senha e verificada com bcrypt (`CompareHashAndPassword`)
6. Se credenciais invalidas, retorna erro `401 Unauthorized`
7. Nova sessao e criada no banco com um UUID
8. Access token e refresh token sao gerados (RS256) com `session_id` nos claims
9. Tokens sao setados como cookies na resposta HTTP
10. Resposta retorna os tokens em JSON (status 200)

### Logout

1. Usuario clica em "Sair" na pagina de sucesso
2. JavaScript envia `POST /v1/auth/logout`
3. Middleware `SessionAuth` valida a sessao e injeta `session_id` no contexto
4. Handler le o `session_id` do contexto do Echo
5. Service parseia o UUID e **deleta a sessao** do banco via `FindOneAndDelete`
6. Cookies `access_token` e `refresh_token` sao limpos (MaxAge=-1)
7. Resposta retorna status `200 OK` sem corpo

## Tratamento de Erros

O projeto utiliza o padrao **ProblemDetails** (RFC 7807) para respostas de erro HTTP:

```json
{
  "type": "auth/email-already-exists",
  "title": "Email Already Registered",
  "status": 409,
  "detail": "An account with this email already exists",
  "instance": "/v1/user/create-account"
}
```

Erros de validacao incluem detalhes por campo:

```json
{
  "type": "auth/validation-error",
  "title": "Validation Failed",
  "status": 400,
  "detail": "One or more fields failed validation",
  "errors": [
    { "field": "email", "message": "Email is required" }
  ]
}
```

## Tecnologias

| Tecnologia | Utilizacao |
|---|---|
| [Go 1.25.4](https://go.dev/) | Linguagem |
| [Echo v4](https://echo.labstack.com/) | Framework HTTP |
| [GORM](https://gorm.io/) | ORM |
| [SQLite](https://www.sqlite.org/) | Banco de dados |
| [golang-jwt v5](https://github.com/golang-jwt/jwt) | Geracao e validacao de JWT (RS256) |
| [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | Hash de senhas (cost 12) |
| [samber/do](https://github.com/samber/do) | Injecao de dependencias |
| [Zap](https://github.com/uber-go/zap) | Logging estruturado |
| [validator v10](https://github.com/go-playground/validator) | Validacao de structs |
| [Air](https://github.com/air-verse/air) | Hot reload |
| [Mockery v3](https://github.com/vektra/mockery) | Geracao de mocks |
| [golangci-lint v2](https://golangci-lint.run/) | Linter |

## Banco de Dados

O projeto utiliza SQLite com GORM. As migracoes sao executadas automaticamente na inicializacao da aplicacao.

### Tabelas

**user_tables**

| Campo | Tipo | Restricoes |
|---|---|---|
| `id` | UUID | Primary Key |
| `email` | VARCHAR(100) | Unique, Not Null |
| `password` | TEXT | Not Null |
| `active` | BOOLEAN | Default: true |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

**session_tables**

| Campo | Tipo | Restricoes |
|---|---|---|
| `id` | UUID | Primary Key |
| `user_id` | UUID | Not Null, Index |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

> Sessoes nao possuem campo `active`. No logout, a sessao e fisicamente deletada do banco via `FindOneAndDelete`.

## Testes

```bash
go test ./...                                        # Todos os testes
go test ./internal/service/... -run TestLogin -v     # Teste especifico
```

O projeto utiliza mocks gerados pelo Mockery com o framework testify. Os testes seguem o padrao Arrange/Act/Assert com subtestes paralelos.

Arquivos de teste existentes:
- `internal/handler/auth_test.go`
- `internal/service/auth_test.go`
- `internal/middleware/session_auth_test.go`

## Licenca

Este projeto e destinado a fins de estudo.
