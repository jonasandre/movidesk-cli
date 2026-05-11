## movidesk-cli services delete

Exclui um serviço de forma permanente (DELETE /services?id=)

```
movidesk-cli services delete <id> [flags]
```

### Options

```
      --force   pula o prompt de confirmação
  -h, --help    help for delete
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

* [movidesk-cli services](movidesk-cli_services.md)	 - Gerencia o catálogo de serviços do Movidesk (/services)

