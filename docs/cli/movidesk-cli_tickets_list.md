## movidesk-cli tickets list

Lista chamados (últimos 90 dias; mais antigos em `tickets past list`)

### Synopsis

Lista chamados via GET /tickets. Cobre apenas os últimos 90 dias —
para chamados arquivados use 'tickets past list' com os mesmos filtros.

A sintaxe completa de --filter, --select e --orderby está em:
  movidesk-cli topics filters

```
movidesk-cli tickets list [flags]
```

### Examples

```
  # tickets em atendimento, ordenados pelo mais recente
  movidesk-cli tickets list --filter "baseStatus eq 'InAttendance'" --orderby "createdDate desc" --top 20

  # tickets de abril/2026 do time "Qlik" (ownerTeam é string, não navegação)
  movidesk-cli tickets list --all \
    --filter "createdDate ge 2026-04-01T00:00:00.000Z and createdDate lt 2026-05-01T00:00:00.000Z and ownerTeam eq 'Qlik'" \
    --select "id,protocol,subject"
```

### Options

```
      --all               busca todas as páginas (auto-paginação)
      --columns strings   colunas separadas por vírgula para saída table/csv (suporta dot-paths)
      --expand strings    expressões $expand separadas por vírgula
      --filter string     expressão OData $filter
  -h, --help              help for list
      --include-deleted   inclui ações/clientes/relações excluídas
      --max int           com --all, interrompe após este número de registros
      --orderby strings   cláusulas $orderby separadas por vírgula (ex.: "id desc")
      --select strings    campos $select separados por vírgula
      --skip int          $skip: offset no servidor
      --top int           $top: tamanho da página ou limite de uma única página
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

