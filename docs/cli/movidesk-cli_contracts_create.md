## movidesk-cli contracts create

Cria um contrato

```
movidesk-cli contracts create [flags]
```

### Options

```
  -f, --file string                 caminho do corpo JSON
      --from-template string        carrega ~/.movidesk/templates/<nome>.json
      --from-template-file string   carrega template de um caminho específico
  -h, --help                        help for create
      --return-all                  pede ao Movidesk pra retornar o contrato completo
      --set strings                 sobrescreve campos inline, ex.: --set status=2
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

* [movidesk-cli contracts](movidesk-cli_contracts.md)	 - Gerencia contratos de horas do Movidesk (/timeAgreement)

