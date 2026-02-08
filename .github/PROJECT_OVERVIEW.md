# Visão Geral do Projeto

Este documento fornece uma descrição do propósito, tecnologia e estado atual do projeto de autenticação.

## Propósito

O objetivo deste projeto é criar um sistema de autenticação seguro utilizando Go para o backend. O sistema implementa autenticação baseada em JWT (RS256) com gerenciamento de sessões persistidas em banco de dados, fornecendo uma solução completa para criação de conta, login, logout com invalidação server-side e acesso seguro a APIs.

Este repositório tem fins de estudo e documentação.

## Tecnologia

- **Backend**: API desenvolvida em **Go (Golang)** com o framework **Echo**.
- **Banco de Dados**: **SQLite** com **GORM** para persistência de usuários e sessões.
- **Autenticação**: **JWT com RS256** (assinatura assimétrica) para access tokens e refresh tokens.
- **Segurança**: Senhas hasheadas com **bcrypt** (cost 12), cookies com **HttpOnly**, **SameSite=Strict** e **Secure** em produção.
- **Injeção de Dependências**: **samber/do** para registro e resolução de dependências.
- **Frontend**: Páginas simples em **HTML, CSS e JavaScript** (`assets/`) para interação com a API (criação de conta, login, logout).

## Estado Atual

### Implementado

- **Criação de conta** (`POST /v1/user/create-account`): validação de campos, verificação de email duplicado, hash de senha com bcrypt, criação de usuário e sessão no banco, geração de access token e refresh token (RS256), cookies setados na resposta.
- **Logout** (`POST /v1/auth/logout`): leitura do cookie access_token, extração do `session_id` dos claims JWT, desativação da sessão no banco (`active=false`), limpeza de cookies. Idempotente e funciona mesmo com token expirado.
- **Sessões persistidas**: tabela `session_tables` no banco com estado ativo/inativo, vinculada aos tokens via claim `session_id`.
- **Frontend**: páginas de criação de conta, login, recuperação de senha e página de sucesso com email do usuário e botão de logout. Verificação de autenticação via JavaScript em todas as páginas.
- **Tratamento de erros**: padrão ProblemDetails (RFC 7807) com suporte a erros de validação por campo.
- **Health check** (`GET /health`).

### Pendente

- **Login** (`POST /v1/auth/login`): handler parcialmente implementado (stub).
- **Middleware de autenticação**: validação de sessão ativa nas rotas protegidas.
- **Refresh token**: endpoint para renovação do access token usando o refresh token.
- **Recuperação de senha**: lógica de envio de email e reset.

## Arquitetura

O projeto segue uma arquitetura em camadas:

```
Handler (HTTP) → Service (negócio) → Repository (dados) → Storage (SQLite)
```

Interfaces são definidas no pacote `domain` e implementadas nas camadas correspondentes. A injeção de dependências é configurada em `cmd/api/main.go`.

Para mais detalhes, consulte o [README.md](../README.md).
