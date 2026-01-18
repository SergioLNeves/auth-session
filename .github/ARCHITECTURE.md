# Arquitetura do Projeto

Este documento descreve a arquitetura de alto nível do projeto de autenticação. A estrutura do projeto segue uma abordagem de design em camadas (Layered Architecture), separando as responsabilidades para promover a modularidade, testabilidade e manutenibilidade do código.

## Visão Geral das Camadas

O código-fonte principal está localizado no diretório `internal/` e é organizado da seguinte forma:

- **`cmd/`**: Ponto de entrada da aplicação. O diretório `cmd/api/` contém a função `main` que inicializa e inicia o servidor da API.

- **`internal/`**: Contém toda a lógica de negócios e da aplicação. Este diretório não é importável por outros projetos Go, garantindo que a lógica interna permaneça encapsulada.

    - **`handler/` (Camada de Apresentação)**: Responsável por lidar com as requisições HTTP. Os handlers recebem os dados da requisição, os validam (utilizando pacotes como `pkg/validator`) e chamam os serviços apropriados. Eles são responsáveis por formatar as respostas HTTP (JSON, status codes, etc.).

    - **`service/` (Camada de Serviço)**: Contém a lógica de negócios central da aplicação. Os serviços orquestram as operações, interagem com os repositórios e executam as regras de negócio. Eles desacoplam os handlers dos detalhes de acesso a dados.

    - **`repository/` (Camada de Repositório)**: Define as interfaces para acesso a dados. Os contratos definidos aqui são implementados pela camada de armazenamento (`storage`), permitindo que a lógica de negócios seja independente da tecnologia de banco de dados utilizada.

    - **`storage/` (Camada de Armazenamento)**: Implementação concreta dos repositórios. No caso deste projeto, `storage/sqlite.go` contém a lógica para interagir com o banco de dados SQLite. Se o banco de dados fosse trocado (e.g., para PostgreSQL), apenas esta camada precisaria ser modificada.

    - **`domain/` (Camada de Domínio)**: Contém as estruturas de dados (structs) e os modelos de negócio principais da aplicação (e.g., `User`, `Token`). Essas são as entidades centrais que são manipuladas pelas outras camadas.

    - **`config/`**: Responsável por carregar e gerenciar as configurações da aplicação, como variáveis de ambiente (`.env`).

    - **`pkg/`**: Pacotes reutilizáveis que não possuem dependências da lógica de negócio específica do projeto (e.g., um utilitário de validação).

- **`assets/`**: Contém os arquivos estáticos para o frontend, como HTML, CSS e JavaScript.

## Fluxo de uma Requisição

1.  Uma requisição HTTP chega ao servidor.
2.  O roteador (configurado em `cmd/api/main.go`) direciona a requisição para o `handler` apropriado.
3.  O `handler` decodifica e valida a requisição.
4.  O `handler` chama um método na camada de `service`.
5.  O `service` executa a lógica de negócio, utilizando os `domain models`.
6.  Se necessário, o `service` solicita ou persiste dados através das interfaces do `repository`.
7.  A implementação do `storage` (SQLite) executa a operação no banco de dados.
8.  O resultado retorna pela mesma cadeia, com o `handler` finalmente enviando a resposta HTTP ao cliente.
