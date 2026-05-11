## movidesk-cli activities create

Cria uma atividade

```
movidesk-cli activities create [flags]
```

### Options

```
  -f, --file string                 caminho do corpo JSON
      --from-template string        carrega ~/.movidesk/templates/<nome>.json
      --from-template-file string   carrega template de um caminho específico
  -h, --help                        help for create
      --set strings                 sobrescreve campos, ex.: --set name="Atividade" --set isActive=true
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

