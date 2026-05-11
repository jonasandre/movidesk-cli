## movidesk-cli contracts consumption

Lê /timeAgreementConsumption

### Synopsis

Ao filtrar por startPeriod/endPeriod, o Movidesk exige o nome do
contrato no $filter OData (ex.: --filter "name eq 'Default' and period gt
2026-01-01T00:00:00Z"). Evite combinar $select com filtros de período.

### Options

```
  -h, --help   help for consumption
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
* [movidesk-cli contracts consumption list](movidesk-cli_contracts_consumption_list.md)	 - Lista linhas de consumo

