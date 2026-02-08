# Visao Geral do Projeto

Este documento fornece uma descricao do proposito, tecnologia e estado atual do projeto de autenticacao.

## Proposito

O objetivo deste projeto e criar um sistema de autenticacao seguro utilizando Go para o backend. O sistema implementa autenticacao baseada em JWT (RS256) com gerenciamento de sessoes persistidas em banco de dados, fornecendo uma solucao completa para criacao de conta, login, logout com invalidacao server-side e acesso seguro a APIs.

Este repositorio tem fins de estudo e documentacao.

## Tecnologia

- **Backend**: API desenvolvida em **Go 1.25.4** com o framework **Echo v4**.
- **Banco de Dados**: **SQLite** com **GORM** para persistencia de usuarios e sessoes.
- **Autenticacao**: **JWT com RS256** (assinatura assimetrica) para access tokens e refresh tokens.
- **Seguranca**: Senhas hasheadas com **bcrypt** (cost 12), cookies com **SameSite=Strict** e **Secure** em producao. Access token legivel pelo JS (nao HttpOnly) para extracao de claims no frontend; refresh token HttpOnly.
- **Injecao de Dependencias**: **samber/do** para registro e resolucao de dependencias.
- **Logging**: **Zap** para logging estruturado.
- **Validacao**: **go-playground/validator v10** para validacao de structs.
- **Frontend**: Paginas simples em **HTML, CSS e JavaScript** (`assets/`) para interacao com a API.

## Estado Atual

### Implementado

- **Criacao de conta** (`POST /v1/user/create-account`): validacao de campos (email valido, senha min 8 chars), verificacao de email duplicado, hash de senha com bcrypt (cost 12), criacao de usuario e sessao no banco, geracao de access token e refresh token (RS256), cookies setados na resposta. Retorna 201 com tokens em JSON.

- **Login** (`POST /v1/auth/login`): validacao de campos, busca de usuario por email, verificacao de senha com bcrypt (`CompareHashAndPassword`), criacao de nova sessao, geracao de tokens, cookies setados. Retorna 200 com tokens ou 401 para credenciais invalidas.

- **Logout** (`POST /v1/auth/logout`): protegido pelo middleware `SessionAuth`. Le o `session_id` do contexto Echo, **deleta a sessao** do banco via `FindOneAndDelete`, limpa cookies (MaxAge=-1). Retorna 200 sem corpo.

- **Middleware de autenticacao** (`SessionAuth`): parseia access token (permite expirado), valida sessao no banco, verifica refresh token. Se refresh expirado: deleta sessao e limpa cookies. Se valido: regenera ambos os tokens e injeta `user_id`, `email`, `session_id` no contexto Echo.

- **Sessoes persistidas**: tabela `session_tables` no banco com campos `id` (UUID), `user_id` (UUID), `created_at`, `updated_at`. Sessoes sao **deletadas** no logout (nao possuem campo `active`).

- **Frontend**: paginas de criacao de conta, login, recuperacao de senha e pagina de sucesso com email do usuario e botao de logout. `auth.js` fornece utilitarios: `getUser()`, `requireAuth()`, `requireGuest()`, `logout()`.

- **Tratamento de erros**: padrao ProblemDetails (RFC 7807) com suporte a erros de validacao por campo.

- **Health check** (`GET /health`).

- **Testes unitarios**: testes para handlers (`auth_test.go`), services (`auth_test.go`) e middleware (`session_auth_test.go`) usando mockery + testify.

### Pendente

- **Recuperacao de senha**: apenas pagina HTML estatica, sem logica de envio de email e reset.

## Arquitetura

O projeto segue uma arquitetura em camadas:

```
Handler (HTTP) -> Service (negocio) -> Repository (dados) -> Storage (SQLite)
```

Interfaces sao definidas no pacote `domain` e implementadas nas camadas correspondentes. A injecao de dependencias e configurada em `cmd/api/main.go`. O middleware `SessionAuth` protege rotas autenticadas com refresh token rotation.

Para mais detalhes, consulte [ARCHITECTURE.md](ARCHITECTURE.md) e o [README.md](../README.md).
