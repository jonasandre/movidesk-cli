## movidesk-cli tickets get

Obtém um chamado por id (posicional) ou protocolo (--protocol)

```
movidesk-cli tickets get [id] [flags]
```

### Options

```
      --columns strings   colunas separadas por vírgula para saída table/csv (suporta dot-paths)
  -h, --help              help for get
      --include-deleted   inclui filhos excluídos
      --protocol string   protocolo do chamado (ex.: MOVI202109000001)
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

