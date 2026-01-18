# Visão Geral do Projeto

Este documento fornece uma descrição do propósito, tecnologia e fases de desenvolvimento do projeto de autenticação.

## Propósito

O objetivo principal deste projeto é criar um sistema de autenticação robusto e seguro utilizando Go para o backend. O sistema evoluirá para implementar autenticação baseada em JSON Web Tokens (JWT), fornecendo uma solução completa para gerenciamento de sessões e acesso seguro a APIs.

## Tecnologia

- **Backend**: API desenvolvida em **Go (Golang)**.
- **Banco de Dados**: **SQLite** é utilizado para a persistência de dados, como tokens de sessão e informações de usuário. A escolha pelo SQLite simplifica a configuração e o deploy inicial.
- **Frontend**: Páginas simples em **HTML, CSS e JavaScript** (`assets/`) são fornecidas para interação com a API (e.g., telas de login).

## Fases de Desenvolvimento

O projeto será desenvolvido em duas fases principais:

### Fase 1: Autenticação baseada em Sessão
Nesta fase inicial, o foco é criar um sistema de gerenciamento de sessão tradicional. Quando um usuário faz login com sucesso, uma sessão é criada e seu identificador é armazenado no banco de dados. Um cookie de sessão é enviado ao cliente para manter o estado de autenticação entre as requisições.

### Fase 2: Autenticação com JWT
A segunda fase expandirá o sistema para utilizar **JSON Web Tokens (JWT)**. Após o login, em vez de um simples token de sessão, a API gerará um JWT assinado. Este token conterá as informações (claims) do usuário e será enviado ao cliente. Para requisições subsequentes, o cliente enviará o JWT no cabeçalho `Authorization`, permitindo que a API verifique a autenticidade e autorize o acesso sem a necessidade de consultar o banco de dados a cada requisição, tornando o sistema mais escalável e stateless.
