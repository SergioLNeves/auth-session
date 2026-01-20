# Implementação de Login com JWT (RS256)

Este documento descreve como implementar um sistema de autenticação "stateless" (sem estado) baseado em **JSON Web Tokens (JWT)** com assinatura assimétrica (RS256), utilizando o framework Echo.

Esta abordagem é mais segura e escalável do que as sessões tradicionais baseadas em banco de dados.

## Visão Geral

A autenticação com JWT funcionará da seguinte maneira:
1.  **Geração de Chaves:** Um par de chaves (privada e pública) é gerado para o servidor.
2.  **Login:** O usuário envia as credenciais. O servidor as valida e, se estiverem corretas, cria um JWT contendo as informações do usuário (chamadas "claims"). Este JWT é **assinado com a chave privada** e enviado de volta ao cliente.
3.  **Armazenamento:** O cliente armazena o JWT (em um cookie `HttpOnly` ou no `localStorage`).
4.  **Requisições Autenticadas:** Para cada requisição a uma rota protegida, o cliente envia o JWT (geralmente no cabeçalho `Authorization: Bearer <token>`).
5.  **Verificação (Middleware):** Um middleware no servidor intercepta a requisição, extrai o JWT e **verifica sua assinatura usando a chave pública**. Se a assinatura for válida, o acesso é concedido. Como a chave pública não pode criar tokens, a segurança é garantida.
6.  **Logout:** Sendo "stateless", o logout no servidor não é estritamente necessário. O cliente simplesmente descarta o token.

---

## 1. Dependências

Primeiro, adicione a biblioteca JWT mais popular para Go:
```bash
go get github.com/golang-jwt/jwt/v5
```

## 2. Geração das Chaves (RSA)

Você precisa gerar um par de chaves RSA. Execute os seguintes comandos no seu terminal para criar os arquivos `private.pem` e `public.pem`. Mantenha o `private.pem` em segredo absoluto!

```bash
# Gerar a chave privada RSA de 2048 bits
openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048

# Extrair a chave pública da chave privada
openssl rsa -pubout -in private.pem -out public.pem
```
Adicione `*.pem` ao seu arquivo `.gitignore` para nunca comitar as chaves. Em produção, carregue-as a partir de variáveis de ambiente ou um sistema de "secrets".

## 3. Fluxo de Login (Gerando o JWT)

Modifique o `internal/handler/auth.go` para criar e assinar um JWT após a autenticação bem-sucedida. O `AuthService` não precisará mais de métodos para criar sessões no banco de dados.

```go
// internal/handler/auth.go
package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// ... (NewAuthHandler)

// Estrutura para as "claims" do nosso token
type JwtCustomClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func (h *AuthHandlerImpl) Login(c echo.Context) error {
	var params domain.LoginRequest
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}
	if err := c.Validate(&params); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	user, err := h.AuthService.Authenticate(c.Request().Context(), params.Email, params.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid credentials")
	}

	// Ler a chave privada para assinar o token
	privateKeyBytes, err := os.ReadFile("private.pem")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Could not read private key")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Could not parse private key")
	}

	// Definir as "claims" customizadas
	claims := &JwtCustomClaims{
		user.ID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)), // Token expira em 3 dias
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Criar o token usando o método de assinatura RS256
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Gerar o token assinado
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to sign token")
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": signedToken,
	})
}
```

## 4. Middleware de Autenticação (Verificando o JWT)

Crie/modifique o arquivo `internal/handler/middleware.go` para verificar o JWT. Ele não precisa mais do `AuthService`.

```go
// internal/handler/middleware.go
package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JwtAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Ler a chave pública para verificar a assinatura
			publicKeyBytes, err := os.ReadFile("public.pem")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, "Could not read public key")
			}
			
			publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, "Could not parse public key")
			}
			
			// Extrair o token do cabeçalho Authorization
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, "Missing Authorization header")
			}
			
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader { // Se não havia o prefixo "Bearer "
				return c.JSON(http.StatusUnauthorized, "Invalid Authorization header format")
			}

			// Parse e validação do token
			token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
				// Valida se o método de assinatura é o esperado (RS256)
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unexpected signing method")
				}
				return publicKey, nil
			})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, "Invalid token: "+err.Error())
			}

			if claims, ok := token.Claims.(*JwtCustomClaims); ok && token.Valid {
				// Adiciona o ID do usuário e as claims ao contexto para uso posterior
				c.Set("userID", claims.UserID)
				c.Set("userClaims", claims)
				return next(c)
			}

			return c.JSON(http.StatusUnauthorized, "Invalid token")
		}
	}
}
```

## 5. Fluxo de Logout

Com JWTs, o logout é gerenciado principalmente pelo cliente: ele simplesmente apaga o token. Não é necessário um endpoint de logout no servidor. Se você precisar de um logout forçado (por exemplo, se um token for roubado), a abordagem comum é criar uma "blocklist" (lista de bloqueio) em um cache como o Redis, o que adiciona complexidade e torna o sistema "stateful" novamente. Para a maioria dos casos, o logout do lado do cliente é suficiente.

## 6. Configuração das Rotas

Finalmente, atualize o `cmd/api/main.go` para usar o novo middleware JWT.

```go
// cmd/api/main.go
package main

import (
	// ... outros imports
	"github.com/SergioLNeves/auth-session/internal/repository" // Manter import
)

// ...

func configureSessionRouters(e *echo.Echo, db *storage.SQLiteStorage) {
	authRepository, _ := repository.NewAuthRepository(db)
	// O AuthService agora é mais simples, sem métodos de sessão
	authService, _ := service.NewAuthService(authRepository)
	authHandler, _ := handler.NewAuthHandler(authService)

    // ... (servir arquivos estáticos)
    
	v1 := e.Group("/v1")
	auth := v1.Group("/auth")
	auth.POST("/sign-up", authHandler.CreateAccount)
	auth.POST("/login", authHandler.Login)
    // O endpoint de logout não é mais necessário

    // Grupo de rotas protegidas com o novo middleware JWT
    protected := e.Group("/app", handler.JwtAuthMiddleware())
    protected.GET("/dashboard", func(c echo.Context) error {
        userID := c.Get("userID").(string)
        return c.String(http.StatusOK, "Welcome to your Dashboard, User "+userID)
    })
}
```

Com estes passos, seu sistema usará uma autenticação JWT robusta, segura e stateless. Lembre-se de implementar a lógica de negócio no `AuthService` (como a validação de senha) e tratar os erros de forma adequada.
