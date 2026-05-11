## movidesk-cli contracts

Gerencia contratos de horas do Movidesk (/timeAgreement)

### Options

```
  -h, --help   help for contracts
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
* [movidesk-cli contracts consumption](movidesk-cli_contracts_consumption.md)	 - Lê /timeAgreementConsumption
* [movidesk-cli contracts create](movidesk-cli_contracts_create.md)	 - Cria um contrato
* [movidesk-cli contracts delete](movidesk-cli_contracts_delete.md)	 - Exclui um contrato
* [movidesk-cli contracts get](movidesk-cli_contracts_get.md)	 - Obtém um contrato pelo id
* [movidesk-cli contracts list](movidesk-cli_contracts_list.md)	 - Lista contratos de horas
* [movidesk-cli contracts update](movidesk-cli_contracts_update.md)	 - Aplica patch em um contrato por id

