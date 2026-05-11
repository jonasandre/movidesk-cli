## movidesk-cli tickets past list

Lista chamados arquivados (mais de 90 dias)

### Synopsis

Lista chamados arquivados via GET /tickets/past. Aceita os mesmos
operadores e funções OData de 'tickets list'.

Sintaxe completa em: movidesk-cli topics filters

```
movidesk-cli tickets past list [flags]
```

### Examples

```
  movidesk-cli tickets past list --filter "createdDate lt 2024-01-01T00:00:00.000Z and ownerTeam eq 'Qlik'" --top 50
```

### Options

```
      --all               busca todas as páginas (auto-paginação)
      --columns strings   colunas separadas por vírgula para saída table/csv (suporta dot-paths)
      --expand strings    expressões $expand separadas por vírgula
      --filter string     expressão OData $filter
  -h, --help              help for list
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

* [movidesk-cli tickets past](movidesk-cli_tickets_past.md)	 - Gerencia chamados com mais de 90 dias (/tickets/past)

