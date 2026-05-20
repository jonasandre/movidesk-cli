## movidesk-cli tickets bulk-update

Aplica o mesmo PATCH a vários chamados (com seletor interativo)

### Synopsis

Atualiza vários chamados de uma só vez. Os alvos podem vir de --ids,
--ids-file ou de um --filter OData (com seletor TUI por padrão quando em TTY).

O corpo do PATCH é montado por --file, --from-template[-file] ou --set chave=valor
(mesmas regras de 'tickets update').

Respeita o limite de 10 req/min: lotes grandes podem demorar. Use --report para
gravar o status de cada chamado em um JSONL e --continue-on-error para não abortar.

```
movidesk-cli tickets bulk-update [flags]
```

### Examples

```
  # encerra em lote por filtro com confirmação interativa
  movidesk-cli tickets bulk-update \
    --filter "baseStatus eq 'Stopped' and ownerTeam eq 'Qlik'" --top 100 \
    --set status=Resolvido --set justification=Resolvido

  # reaproveita IDs de uma listagem prévia
  movidesk-cli tickets list --filter "..." --output json | jq -r '.[].id' > /tmp/ids
  movidesk-cli tickets bulk-update --ids-file /tmp/ids --set ownerTeam=Suporte
```

### Options

```
      --all                         busca todas as páginas (auto-paginação)
      --all-from-filter             usa todos os resultados do --filter sem abrir o seletor (necessário em ambientes sem TTY)
      --continue-on-error           continua o lote mesmo após uma falha individual
      --dry-run                     mostra o que seria feito sem chamar a API
      --expand strings              expressões $expand separadas por vírgula
  -f, --file string                 caminho do corpo JSON de patch
      --filter string               expressão OData $filter
      --force                       pula o prompt de confirmação (obrigatório fora de TTY)
      --from-template string        carrega ~/.movidesk/templates/<nome>.json
      --from-template-file string   carrega template de um caminho específico
  -h, --help                        help for bulk-update
      --ids ints                    lista de ids separados por vírgula (pula a listagem)
      --ids-file string             arquivo com um id por linha (# inicia comentário); aceita também IDs separados por vírgula
      --max int                     com --all, interrompe após este número de registros
      --orderby strings             cláusulas $orderby separadas por vírgula (ex.: "id desc")
      --pick                        abre seletor TUI mesmo quando --ids/--ids-file forem informados
      --report string               grava resultado por ticket em arquivo JSONL
      --select strings              campos $select separados por vírgula
      --set strings                 sobrescreve campos inline, ex.: --set status=Resolvido
      --skip int                    $skip: offset no servidor
      --source string               fonte da listagem: live (/tickets, últimos 90d), past (/tickets/past, arquivados), both (mescla as duas) (default "live")
      --top int                     $top: tamanho da página ou limite de uma única página
```

### Options inherited from parent commands

```
      --compact         JSON compacto (sem indentação)
      --no-color        desativa cores na saída
      --no-retry        desativa retentativa automática em 429/5xx
  -o, --output string   formato de saída: json|table|csv (padrão: do tenant ou 'json')
      --tenant string   nome do tenant (sobrepõe o tenant atual; env: MOVIDESK_TENANT)
      --user string     id do usuário padrão (Cod. Ref.) usado em createdBy nas escritas; sobrepõe a configuração do tenant; env: MOVIDESK_USER
  -v, --verbose         log detalhado em stderr
```

### SEE ALSO

* [movidesk-cli tickets](movidesk-cli_tickets.md)	 - Gerencia chamados (tickets) do Movidesk

