## movidesk-cli tickets bulk-cancel

Cancela vários chamados em lote, registrando uma ação com o motivo

### Synopsis

Atualiza o status para 'Cancelado' (configurável) e adiciona uma ação
com o motivo informado em cada chamado selecionado. Mesma mecânica do
'tickets bulk-close', porém pensado para descartes (duplicados, abandonados,
fora de escopo) — não para encerramentos com solução.

Use --public para registrar a ação como pública (visível pelo cliente). Sem --public
a ação é interna (type=1). O nome exato do status deve bater com o configurado
no tenant (padrão: "Cancelado"). O campo justification é sempre enviado (o Movidesk
exige ao mudar Status); fica vazio quando --justification não é informado, o que
funciona pra status sem motivos cadastrados.

```
movidesk-cli tickets bulk-cancel [flags]
```

### Examples

```
  movidesk-cli tickets bulk-cancel \
    --filter "baseStatus eq 'New' and createdDate lt 2026-04-01T00:00:00.000Z" \
    --message "Cancelado por falta de retorno do cliente"

  # variação pública e com justificativa customizada
  movidesk-cli tickets bulk-cancel --ids 12,34,56 \
    --message "Duplicado do #99" --public --justification "Duplicado"
```

### Options

```
      --action-type int        tipo da ação: 1=interna, 2=pública (padrão: 1)
      --all                    busca todas as páginas (auto-paginação)
      --all-from-filter        usa todos os resultados do --filter sem abrir o seletor (necessário em ambientes sem TTY)
      --continue-on-error      continua o lote mesmo após uma falha individual
      --dry-run                mostra o que seria feito sem chamar a API
      --expand strings         expressões $expand separadas por vírgula
      --filter string          expressão OData $filter
      --force                  pula o prompt de confirmação (obrigatório fora de TTY)
  -h, --help                   help for bulk-cancel
      --ids ints               lista de ids separados por vírgula (pula a listagem)
      --ids-file string        arquivo com um id por linha (# inicia comentário); aceita também IDs separados por vírgula
      --justification string   justificativa do ticket (padrão: igual a --status)
      --max int                com --all, interrompe após este número de registros
      --message string         texto da ação de cancelamento (obrigatório)
      --orderby strings        cláusulas $orderby separadas por vírgula (ex.: "id desc")
      --pick                   abre seletor TUI mesmo quando --ids/--ids-file forem informados
      --public                 atalho para --action-type=2
      --report string          grava resultado por ticket em arquivo JSONL
      --select strings         campos $select separados por vírgula
      --skip int               $skip: offset no servidor
      --source string          fonte da listagem: live (/tickets, últimos 90d), past (/tickets/past, arquivados), both (mescla as duas) (default "live")
      --status string          nome do status final (padrão: Cancelado)
      --top int                $top: tamanho da página ou limite de uma única página
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

