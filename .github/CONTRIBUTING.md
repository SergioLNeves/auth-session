# Guia de Contribuição

Este documento fornece diretrizes para contribuir com o projeto, com foco nas convenções de código Go.

## Estilo de Código e Padrões (Golang)

Aderimos às práticas padrão da comunidade Go para garantir que o código seja limpo, legível e consistente.

### 1. Formatação
Todo o código Go **deve** ser formatado com `gofmt`. Antes de submeter qualquer alteração, execute `gofmt -s -w .` no diretório do projeto. A maioria dos IDEs pode ser configurada para fazer isso automaticamente ao salvar.

### 2. Linting
Utilize `go vet .` para identificar construções suspeitas e `golangci-lint` (se disponível) para uma análise estática mais aprofundada. O código deve passar por essas verificações sem erros.

### 3. Nomenclatura
- **Pacotes**: Nomes de pacotes devem ser curtos, concisos e em minúsculas. Evite `under_scores` ou `mixedCaps`.
- **Variáveis**: Nomes de variáveis devem ser curtos, mas descritivos. Para variáveis de escopo muito limitado, nomes de uma ou duas letras (como `i` para um índice de loop) são aceitáveis.
- **Funções e Métodos**: Use `camelCase`. Nomes que começam com letra maiúscula são exportados (públicos), enquanto nomes que começam com letra minúscula são privados ao pacote.
- **Interfaces**: Interfaces que definem um único método são frequentemente nomeadas com o sufixo "er" (e.g., `Reader`, `Writer`, `Formatter`).

### 4. Comentários
- **Documentação**: Comente todo membro exportado (funções, tipos, constantes e variáveis). O comentário deve começar com o nome do membro que ele descreve. Ex: `// MinhaFuncao faz X e Y.`.
- **Clareza**: Use comentários para explicar o *porquê* de uma lógica complexa, não o *o quê*. O código deve ser autoexplicativo sempre que possível.

### 5. Tratamento de Erros
- Erros devem ser tratados explicitamente. Não os ignore com `_`.
- Mensagens de erro não devem ser capitalizadas ou terminar com pontuação, pois geralmente são encadeadas com outras informações de contexto.
- Use a função `fmt.Errorf` com a diretiva `%w` para encapsular (wrap) erros, preservando o contexto do erro original.

### 6. Organização de Pacotes
- Siga a estrutura de diretórios definida em `ARCHITECTURE.md`.
- Evite dependências circulares entre pacotes.
- Pacotes devem ter uma responsabilidade clara e coesa.

### 7. Simplicidade
Prefira código simples e direto a soluções excessivamente complexas ou "inteligentes". A legibilidade é fundamental. Como diz o provérbio Go: "Clear is better than clever."
