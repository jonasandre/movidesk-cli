## movidesk-cli services update

Aplica patch em um serviço por id

```
movidesk-cli services update <id> [flags]
```

### Options

```
  -f, --file string                 caminho do corpo JSON de patch
      --from-template string        carrega ~/.movidesk/templates/<nome>.json
      --from-template-file string   carrega template de um caminho específico
  -h, --help                        help for update
      --set strings                 sobrescreve campos inline
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

