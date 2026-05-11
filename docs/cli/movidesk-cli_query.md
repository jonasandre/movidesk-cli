## movidesk-cli query

Chamada HTTP bruta contra qualquer endpoint do Movidesk

### Synopsis

Escape hatch genérico. O caminho é relativo à base da API, ex.: "tickets",
"/tickets", "/persons". O token é injetado automaticamente.

Exemplos:
  movidesk-cli query /tickets --filter "id eq 1"
  movidesk-cli query /persons --select id,businessName --top 5
  movidesk-cli query /persons --method GET --param id=abc

O formato de saída segue --output (json|table|csv); use -o json com jq pra fatiar mais.

```
movidesk-cli query <path> [flags]
```

### Options

```
      --all               busca todas as páginas (auto-paginação)
      --expand strings    expressões $expand separadas por vírgula
      --filter string     expressão OData $filter
  -h, --help              help for query
      --max int           com --all, interrompe após este número de registros
      --method string     método HTTP (apenas GET ou DELETE) (default "GET")
      --orderby strings   cláusulas $orderby separadas por vírgula (ex.: "id desc")
      --param strings     query param extra chave=valor (repetível)
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

* [movidesk-cli](movidesk-cli.md)	 - CLI para a API REST do Movidesk

