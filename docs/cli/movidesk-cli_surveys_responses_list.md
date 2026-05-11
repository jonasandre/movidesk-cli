## movidesk-cli surveys responses list

Lista respostas de pesquisas (paginação por cursor)

```
movidesk-cli surveys responses list [flags]
```

### Options

```
      --all                     percorre todas as páginas
      --columns strings         colunas separadas por vírgula para saída table/csv (suporta dot-paths)
  -h, --help                    help for list
      --limit int               tamanho da página (1..100, padrão 100)
      --max int                 com --all, interrompe após este número de registros
      --starting-after string   cursor (id do último item da página anterior)
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

* [movidesk-cli surveys responses](movidesk-cli_surveys_responses.md)	 - Respostas de pesquisas

