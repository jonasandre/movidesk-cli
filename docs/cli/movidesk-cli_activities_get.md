## movidesk-cli activities get

Obtém uma atividade pelo id

```
movidesk-cli activities get <id> [flags]
```

### Options

```
      --columns strings   colunas separadas por vírgula para saída table/csv (suporta dot-paths)
  -h, --help              help for get
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

* [movidesk-cli activities](movidesk-cli_activities.md)	 - Gerencia atividades do Movidesk (/activity)

