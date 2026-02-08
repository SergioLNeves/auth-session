# Auth Session

Serviço de autenticação em Go com JWT (RS256) e gerenciamento de sessões persistidas em banco de dados. Construído com o framework Echo, SQLite (GORM) e injeção de dependências via `samber/do`.

Este repositório tem fins de estudo e documentação de um fluxo completo de autenticação: criação de conta, login automático, gerenciamento de sessões e logout com invalidação server-side.

## Requisitos

- Go 1.25.4+
- OpenSSL (para geração das chaves RSA)
- Make

## Configuração Inicial

```bash
make setup
```

Este comando instala as dependências do projeto, ferramentas de desenvolvimento (mockery, golangci-lint, air) e gera o par de chaves RSA necessário para assinatura dos tokens.

### Variáveis de Ambiente

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

| Variável | Descrição | Padrão |
|---|---|---|
| `ENV` | Ambiente de execução (`development` ou `production`) | `development` |
| `PORT` | Porta do servidor HTTP | `8080` |
| `LOG_LEVEL` | Nível de log (`debug`, `info`, `warn`, `error`) | `debug` |
| `PRIVATE_KEY_PATH` | Caminho para a chave privada RSA (.pem) | - |
| `PUBLIC_KEY_PATH` | Caminho para a chave pública RSA (.pem) | - |
| `ACCESS_TOKEN_EXPIRY` | Tempo de expiração do access token (minutos) | `60` |
| `REFRESH_TOKEN_EXPIRY` | Tempo de expiração do refresh token (minutos) | `10080` (7 dias) |
| `DB_PATH` | Caminho do banco SQLite | `./data/auth-session.db` |

## Execução

```bash
make run
```

O servidor inicia com hot reload via Air na porta configurada.

## Comandos Disponíveis

| Comando | Descrição |
|---|---|
| `make setup` | Instala dependências, ferramentas e gera chaves RSA |
| `make run` | Executa a aplicação com hot reload (Air) |
| `make gen-key` | Gera par de chaves RSA (private-key.pem e public-key.pem) |
| `make mocks` | Gera mocks para testes com Mockery |
| `make lint` | Executa o linter (golangci-lint) |

## Arquitetura

O projeto segue uma arquitetura em camadas com separação estrita de responsabilidades:

```
cmd/api/main.go                  → Ponto de entrada, DI e rotas
internal/
  ├─ handler/                     → Camada HTTP (validação, bind, cookies)
  ├─ service/                     → Lógica de negócio
  ├─ repository/                  → Acesso a dados
  ├─ storage/sqlite/              → Implementação SQLite (GORM)
  ├─ domain/                      → Entidades, DTOs e interfaces
  ├─ security/                    → JWT (RS256) e bcrypt
  ├─ config/                      → Configuração e ambiente
  └─ pkg/                         → Utilitários (logging, validação, erros)
assets/
  ├─ html/                        → Páginas HTML (login, criar conta, etc.)
  ├─ css/                         → Estilos
  └─ js/                          → Scripts (auth, formulários)
```

### Fluxo de uma Requisição

```
HTTP Request → Handler → Service → Repository → Storage (SQLite)
                 │
                 └── Resposta retorna pelo mesmo caminho
```

### Injeção de Dependências

Todas as dependências são registradas em `cmd/api/main.go` usando `samber/do`:

```
SQLite → Repositories → JWTProvider → Services → Handlers
```

## Endpoints da API

### Páginas

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/` | Página de sucesso (requer autenticação) |
| `GET` | `/create-account` | Formulário de criação de conta |
| `GET` | `/login` | Formulário de login |
| `GET` | `/password` | Formulário de recuperação de senha |

### API REST

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/health` | Health check |
| `POST` | `/v1/user/create-account` | Criação de conta |
| `POST` | `/v1/auth/login` | Login |
| `POST` | `/v1/auth/logout` | Logout (invalida sessão) |

### Exemplos de Requisição

**Criar conta:**
```bash
curl -X POST http://localhost:8080/v1/user/create-account \
  -d "email=usuario@exemplo.com" \
  -d "password=senha12345"
```

**Resposta (201 Created):**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIs..."
}
```

Os tokens também são setados automaticamente como cookies na resposta.

**Logout:**
```bash
curl -X POST http://localhost:8080/v1/auth/logout \
  --cookie "access_token=eyJhbGciOiJSUzI1NiIs..."
```

## Autenticação

### JWT com RS256

O sistema utiliza tokens JWT assinados com chaves RSA assimétricas (RS256):

- **Chave privada** — usada para assinar os tokens (mantida no servidor)
- **Chave pública** — usada para validar assinaturas e parsear tokens

Os arquivos `.pem` são gerados via `make gen-key` e **nunca devem ser comitados** (já estão no `.gitignore`).

### Tokens

| Token | Expiração Padrão | Claims | Cookie |
|---|---|---|---|
| Access Token | 60 min | `sub`, `email`, `session_id`, `iat`, `exp` | `access_token` (legível pelo JS) |
| Refresh Token | 7 dias | `sub`, `session_id`, `iat`, `exp` | `refresh_token` (HttpOnly) |

O `access_token` é legível pelo JavaScript para permitir a extração de claims no frontend (ex.: exibir email do usuário). O `refresh_token` é HttpOnly, inacessível via JS.

Ambos os cookies utilizam `SameSite=Strict` e `Secure=true` em produção.

### Gerenciamento de Sessões

As sessões são persistidas no banco de dados (tabela `session_tables`):

| Campo | Tipo | Descrição |
|---|---|---|
| `id` | UUID | Identificador único da sessão |
| `user_id` | UUID | Referência ao usuário |
| `active` | boolean | Estado da sessão (`true`/`false`) |
| `created_at` | timestamp | Data de criação |
| `updated_at` | timestamp | Última atualização |

O `session_id` é incluído nos claims de ambos os tokens JWT, vinculando cada token a uma sessão específica no banco.

### Segurança de Senhas

As senhas são armazenadas com hash bcrypt (cost 12). Nunca são armazenadas ou trafegadas em texto plano.

## Fluxos

### Criação de Conta

{fluxo de criação de conta}

1. Usuário preenche o formulário em `/create-account`
2. JavaScript envia `POST /v1/user/create-account` com email e senha
3. Handler valida os campos (email válido, senha mínimo 8 caracteres)
4. Service verifica se o email já existe no banco
5. Senha é hasheada com bcrypt
6. Usuário é criado no banco
7. Sessão é criada no banco (`active=true`)
8. Access token e refresh token são gerados (RS256) com `session_id` nos claims
9. Tokens são setados como cookies na resposta HTTP
10. Usuário é redirecionado para `/` (página de sucesso)

### Login

{fluxo de login}

1. Usuário preenche o formulário em `/login`
2. `POST /v1/auth/login` com email e senha
3. Service busca usuário por email e verifica a senha com bcrypt
4. Nova sessão é criada no banco
5. Tokens são gerados e setados como cookies
6. Usuário é redirecionado

### Logout

{fluxo de logout}

1. Usuário clica em "Sair" na página de sucesso
2. JavaScript envia `POST /v1/auth/logout`
3. Handler lê o cookie `access_token`
4. Service parseia o JWT e extrai o `session_id` dos claims
5. Sessão é marcada como `active=false` no banco
6. Cookies `access_token` e `refresh_token` são limpos
7. Usuário é redirecionado para `/login`

O logout é idempotente: se não houver cookie, os cookies são limpos e a resposta é 200 OK. O `ParseAccessToken` utiliza `WithoutClaimsValidation` para permitir logout mesmo com token expirado.

## Tratamento de Erros

O projeto utiliza o padrão **ProblemDetails** (RFC 7807) para respostas de erro HTTP:

```json
{
  "type": "auth/email-already-exists",
  "title": "Email Already Registered",
  "status": 409,
  "detail": "An account with this email already exists",
  "instance": "/v1/user/create-account"
}
```

Erros de validação incluem detalhes por campo:

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

| Tecnologia | Utilização |
|---|---|
| [Go](https://go.dev/) | Linguagem |
| [Echo](https://echo.labstack.com/) | Framework HTTP |
| [GORM](https://gorm.io/) | ORM |
| [SQLite](https://www.sqlite.org/) | Banco de dados |
| [golang-jwt](https://github.com/golang-jwt/jwt) | Geração e validação de JWT (RS256) |
| [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | Hash de senhas |
| [samber/do](https://github.com/samber/do) | Injeção de dependências |
| [Zap](https://github.com/uber-go/zap) | Logging estruturado |
| [Air](https://github.com/air-verse/air) | Hot reload |
| [Mockery](https://github.com/vektra/mockery) | Geração de mocks |
| [golangci-lint](https://golangci-lint.run/) | Linter |

## Banco de Dados

O projeto utiliza SQLite com GORM. As migrações são executadas automaticamente na inicialização da aplicação.

### Tabelas

**user_tables**

| Campo | Tipo | Restrições |
|---|---|---|
| `id` | UUID | Primary Key |
| `email` | VARCHAR(100) | Unique, Not Null |
| `password` | TEXT | Not Null |
| `active` | BOOLEAN | Default: true |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

**session_tables**

| Campo | Tipo | Restrições |
|---|---|---|
| `id` | UUID | Primary Key |
| `user_id` | UUID | Not Null, Index |
| `active` | BOOLEAN | Default: true |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

## Licença

Este projeto é destinado a fins de estudo.
