# Guia de Contribuicao

Este documento fornece diretrizes para contribuir com o projeto, com foco nas convencoes de codigo Go.

## Configuracao do Ambiente

```bash
make setup    # Instala dependencias, ferramentas (mockery, golangci-lint, air) e gera chaves RSA
make run      # Executa com hot reload
make lint     # Executa o linter
make mocks    # Gera mocks para testes
```

## Estilo de Codigo e Padroes (Golang)

Aderimos as praticas padrao da comunidade Go para garantir que o codigo seja limpo, legivel e consistente.

### 1. Formatacao

Todo o codigo Go **deve** ser formatado com `gofmt`. Antes de submeter qualquer alteracao, execute `gofmt -s -w .` no diretorio do projeto. A maioria dos IDEs pode ser configurada para fazer isso automaticamente ao salvar.

### 2. Linting

O projeto usa `golangci-lint` com configuracao em `.golangci.yml`:

- Linters habilitados: gocritic, misspell, revive, unconvert, unparam, whitespace
- Formatters: gofmt (com simplify), goimports

Execute `make lint` antes de submeter alteracoes.

### 3. Organizacao de Imports

Imports devem ser organizados em tres grupos separados por linhas em branco:

```go
import (
    // 1. Biblioteca padrao
    "context"
    "fmt"

    // 2. Dependencias externas
    "github.com/labstack/echo/v4"
    "github.com/samber/do"

    // 3. Pacotes internos
    "github.com/SergioLNeves/auth-session/internal/domain"
)
```

### 4. Nomenclatura

- **Pacotes**: nomes curtos, concisos e em minusculas. Evite `under_scores` ou `mixedCaps`.
- **Variaveis**: nomes curtos, mas descritivos. Para escopo limitado, nomes de uma ou duas letras sao aceitaveis.
- **Funcoes e Metodos**: use `camelCase`. Maiuscula inicial = exportado, minuscula = privado.
- **Interfaces**: interfaces de um unico metodo frequentemente usam o sufixo "er" (ex.: `Reader`, `Writer`).

### 5. Tratamento de Erros

- Erros devem ser tratados explicitamente. Nao os ignore com `_`.
- Mensagens de erro nao devem ser capitalizadas ou terminar com pontuacao.
- Use `fmt.Errorf` com `%w` para encapsular erros, preservando o contexto.
- Use o padrao ProblemDetails (RFC 7807) para respostas HTTP de erro.

### 6. Comentarios

- Comente todo membro exportado. O comentario deve comecar com o nome do membro.
- Use comentarios para explicar o *porque* de uma logica complexa, nao o *o que*.

### 7. Simplicidade

Prefira codigo simples e direto. A legibilidade e fundamental. "Clear is better than clever."

## Padrao para Adicionar Novos Componentes

### Novo Endpoint

1. Defina DTOs de request/response em `internal/domain/`
2. Adicione metodo na interface do domain (ex.: `AuthHandler`, `AuthService`)
3. Implemente no service e handler correspondentes
4. Registre a rota em `cmd/api/main.go`
5. Execute `make mocks` se interfaces foram alteradas

### Nova Operacao de Banco

1. Defina o metodo na interface do repository em `internal/domain/`
2. Implemente em `internal/repository/` usando constantes de tabela e a interface `Storage`
3. Se necessario, adicione metodo na interface `Storage` em `internal/storage/storage.go`
4. Implemente na versao SQLite em `internal/storage/sqlite/sqlite.go`
5. Modelos GORM ficam em `internal/storage/sqlite/models.go` e devem ser registrados em `GetModelsToMigrate()`

### Novo Provider/Servico

1. Defina a interface em `internal/domain/`
2. Implemente com construtor `New*` que recebe `*do.Injector`
3. Registre com `do.Provide(injector, ...)` em `cmd/api/main.go`
4. Execute `make mocks`

## Testes

Os testes seguem convencoes especificas documentadas em `.claude/rules/Test_Example.md`:

- Testes ficam ao lado do arquivo fonte (`auth.go` -> `auth_test.go`)
- Mesmo pacote (nao `_test`)
- Padrao Arrange/Act/Assert com `t.Parallel()` em subtestes
- Mocks gerados pelo Mockery com framework testify
- Cubra happy path e error paths

```bash
go test ./...                                     # Todos os testes
go test ./internal/service/... -run TestLogin -v  # Teste especifico
```
