# Changelog

Todas as alterações notáveis deste projeto serão documentadas neste
arquivo.

O formato é baseado em
[Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/) e este
projeto adere ao
[Versionamento Semântico](https://semver.org/lang/pt-BR/).

## [Unreleased]

## [1.2.1] — 2026-05-11

### Corrigido

- MCP: parâmetros `select` / `expand` / `orderby` passam a aceitar três
  formas no wire — JSON array canônico (`["id","protocol"]`),
  JSON-string (`"[\"id\",\"protocol\"]"`) e comma-string
  (`"id,protocol"`) — convertidas automaticamente para o array
  canônico. O schema declarado para o LLM mantém array como forma
  preferida, mas a validação aceita as três variantes para evitar
  falhas de tipo geradas por LLMs mal-comportados.
- MCP: `tickets_list` e `tickets_past_list` aplicam clamp implícito em
  `top > 250`, reduzindo a 250 e anexando aviso visível na resposta.
  Movidesk retorna HTTP 500 esporádicos em páginas maiores; o clamp
  evita que a falha se propague ao usuário.

### Atualizado

- MCP: descrições de todas as tools `*_list` (`tickets_list`,
  `tickets_past_list`, `persons_list`, `services_list`,
  `contracts_list`, `contracts_consumption_list`) ganham exemplos
  literais (`{"select": [...], "filter": "...", "top": N}`) e, em
  tickets, notas explícitas de `$select` obrigatório e cap de `top`.
- MCP: descrição da tool `query` reforça que o `path` deve apontar
  para endpoints reais do Movidesk e enumera os principais; alerta
  que `/$count`, `/count` e variantes não existem.
- SKILL `movidesk-mcp`: bloco "Regras críticas" no topo (formato de
  lista, `$select` obrigatório, cap de `top`, contagem via `top: 1`,
  paths permitidos). Cheatsheet de erros expandido com `$select`
  required (HTTP 400), HTTP 500 em volumes altos e aviso de clamp.

## [1.2.0] — 2026-05-11

### Adicionado

- Servidor MCP (Model Context Protocol) embutido: `movidesk-cli mcp`
  inicia um servidor stdio que expõe a API do Movidesk como ferramentas
  estruturadas para chat apps compatíveis (Claude Desktop, Cline,
  Continue, etc.). 22 tools read-only cobrindo tickets, persons,
  services, contracts, knowledge base, activities, surveys e um escape
  hatch `query` para qualquer endpoint OData. Resources expõem a
  sintaxe completa de filtros OData e o catálogo de custom fields do
  tenant para resolução label↔id pelo modelo. Caps de segurança:
  `all=true` sem `max` limita em 500 linhas para preservar o budget de
  10 req/min; respostas >256 KiB são truncadas com aviso. Erros 429 são
  traduzidos em mensagem clara para o LLM fazer backoff.

## [1.1.1] — 2026-05-11

### Adicionado

- Novo comando `movidesk-cli topics filters`: referência embutida da
  sintaxe OData aceita pelo Movidesk em `--filter` / `--select` /
  `--orderby` — operadores, funções, formatos de literal (datas em UTC
  com sufixo `Z`), campos comuns por recurso (`tickets`, `persons`,
  `services`, `contracts`) e armadilhas frequentes (notadamente
  `ownerTeam` ser um campo string, não navegação, o que causava HTTP
  400 "segment isn't Navigation/Structural/Complex/Collections" quando
  filtros eram gerados por LLMs). Cada `list` que aceita `--filter`
  ganhou bloco de exemplos próprio e ponteiro para o tópico.

### Corrigido

- `.golangci.yml`: removida chave `linters.settings` vazia que quebrava
  `golangci-lint v2.12 config verify` no GitHub Actions com
  `"linters.settings" ... got null, want object`.

## [1.1.0] — 2026-05-11

### Alterado

- Toda a interface do CLI agora é em português brasileiro: helpers de
  comando (Cobra `Short`/`Long`), descrições de flag, prompts
  interativos, mensagens de sucesso e mensagens de erro user-facing
  foram traduzidas. Comportamento, nomes de comandos/flags, env vars
  e chaves YAML permanecem em inglês — só o texto exibido ao usuário
  mudou. `README.md` e este `CHANGELOG.md` também foram traduzidos
  para pt-BR.
- Removido o linter `misspell` (dicionário en-US apenas) do
  `.golangci.yml`. As strings agora são bilíngues (identificadores Go
  em inglês, texto user-facing em pt-BR) e o linter passou a gerar
  falsos positivos em todo arquivo.

## [1.0.2] — 2026-05-08

### Corrigido

- A instalação via Cask do Homebrew na v1.0.1 abortava porque o cask
  declarava caminhos `bash_completion`/`zsh_completion`/`fish_completion`
  mas o tarball publicado não incluía os scripts de completion. O
  GoReleaser agora gera `completions/movidesk-cli.{bash,zsh,fish}` via
  um hook `before` e empacota os arquivos em todos os archives.

## [1.0.1] — 2026-05-08

### Segurança

- A retentativa HTTP agora é restrita a métodos seguros/idempotentes
  (`GET`, `HEAD`, `OPTIONS`). Antes, falhas transitórias em `POST`,
  `PATCH` e `DELETE` podiam ser repetidas automaticamente, com risco
  de gravações duplicadas (chamados/pessoas criados em duplicidade,
  mudanças de estado repetidas). Operações de escrita não retentam
  mais; use os helpers explícitos ou defina `--no-retry` para
  desativar retentativa completamente.
- O rate limiter não desfaz mais um slot quando o contexto é
  cancelado. O código anterior removia um stamp que ainda não havia
  sido reservado de fato (a reserva é tomada antes da espera), o que
  podia permitir rajadas além da janela configurada de 10 req/min
  após uma requisição cancelada.

## [1.0.0] — 2026-05-08

Primeiro release público. Cobre todas as APIs listadas no menu de
integração do Movidesk.

### Adicionado

#### Fundação
- Configuração multi-tenant em `~/.movidesk/config.yaml` (chmod 0600).
- Tokens armazenados no chaveiro do sistema operacional (macOS Keychain,
  Windows Credential Manager, Linux libsecret/kwallet) via
  `zalando/go-keyring`, com fallback em arquivo criptografado
  AES-GCM/PBKDF2 para ambientes headless.
- Cliente HTTP com injeção de token via query-param, rate limiter de
  janela deslizante (10 req/min conforme limite publicado do Movidesk),
  retentativas que respeitam `Retry-After` e backoff exponencial em
  5xx.
- Builder de consulta OData para
  `$filter`/`$select`/`$expand`/`$orderby`/`$top`/`$skip`/`$count`.
- Formatters de saída JSON / tabela / CSV com colunas padrão por
  recurso e suporte a dot-path para campos aninhados.
- Flags persistentes: `--tenant`, `--output`, `--user`, `--no-color`,
  `--verbose`, `--no-retry`, `--compact`. Overrides via env:
  `MOVIDESK_TENANT`, `MOVIDESK_TOKEN`, `MOVIDESK_USER`,
  `MOVIDESK_HOME`, `MOVIDESK_PASSPHRASE`.
- Workflow de CI: vet + test + lint + build em todo PR.
- Configuração GoReleaser compilando darwin/linux/windows × amd64/arm64
  com publicação na tap do Homebrew em cada tag.

#### Auth
- `auth login --tenant` com prompt oculto de token e validação ao vivo
  contra a API.
- `auth list` (token nunca exibido), `auth switch`, `auth status`,
  `auth logout [--all]`, `auth token` (impressão explícita para piping).
- `auth set-user` mais prompt de usuário padrão no `auth login` para
  injeção de `createdBy` por tenant.

#### Tickets
- `tickets list/get/create/update/html/past list/merged list/attach`
  com auto-paginação (`--all` + `--max`) e criação por template JSON
  (`--file`, `--from-template`, `--from-template-file`, `--set k=v`).
- Tipagem completa do schema (~80 campos) com
  `Extra json.RawMessage` em todo tipo aninhado para compatibilidade
  com futuras evoluções.
- Subcomandos de coleção: `actions list/get/add/update/delete`,
  `clients list`, `relations`, `timeline`, `assets list`,
  `histories list`.
- Subsistema de campos personalizados:
  `customfields show/set/clear` com semântica read-merge-patch, mais
  catálogo por tenant em
  `~/.movidesk/<tenant>/customfields.yaml`
  (`catalog list/add/remove`).
- Injeção do `createdBy` padrão em `create` e `actions add` quando o
  tenant tem um usuário padrão definido, com override via `--user <id>`.

#### Pessoas
- `persons list/get/create/update/delete` com confirmação de exclusão.
- `persons customfields show/set/clear` reutilizando o catálogo dos
  tickets.

#### Serviços
- `services list/get/create/update/delete` com confirmação de exclusão.

#### Atividades
- `activities list/get/create/update/delete/add-teams`, com paginação
  por cursor (`limit`/`startingAfter`/`name`) — a API de atividades do
  Movidesk não é OData.

#### Contratos (`/timeAgreement`)
- `contracts list/get/create/update/delete`.
- `contracts consumption list` (leitura de `/timeAgreementConsumption`).

#### Pesquisas
- `surveys questions list/get` (somente leitura).
- `surveys responses list` com paginação por cursor.

#### Base de conhecimento
- `kb articles get <id>` (leitura de artigo único; não há endpoint
  público de listagem).

#### Telefonia
- `telephony queue --event <nome>` para
  `/asterisk_{received,transfered,completed,lost,canceled}Call`.
- `telephony nonqueue --event <nome>` para as variantes sem fila.
- `telephony made-call-link` para `/setMadeCallLink`.

#### Conjunto de opções de campos personalizados
- `customfields options add/rename/remove` encapsulando
  `/ticketCustomFieldValue/{Insert,Update,Delete}Values` para
  gerenciar os dropdowns de campos tipo lista no tenant inteiro.

#### Escape hatch
- `query <path>` para qualquer chamada GET/DELETE OData contra
  endpoints arbitrários, com auto-paginação opcional via `--all`.

### Notas

- PATCH em `/tickets` substitui campos com valor de array como
  `customFieldValues`; os helpers tipados de escrita sempre passam por
  read-merge-patch para que o chamador descreva só a alteração
  desejada.
- O CLI é seguro em round-trip via `--output json`: o handler
  subjacente reemite a resposta sem alterações, então mesmo campos
  que o SDK ainda não tipa são preservados de ponta a ponta.

[Unreleased]: https://github.com/jonasandre/movidesk-cli/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/jonasandre/movidesk-cli/compare/v1.0.2...v1.1.0
[1.0.2]: https://github.com/jonasandre/movidesk-cli/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/jonasandre/movidesk-cli/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/jonasandre/movidesk-cli/releases/tag/v1.0.0
