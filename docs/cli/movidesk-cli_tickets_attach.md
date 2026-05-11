## movidesk-cli tickets attach

Envia um arquivo para uma ação de chamado via /ticketFileUpload

```
movidesk-cli tickets attach <ticket-id> [flags]
```

### Options

```
      --action-id int   id da ação onde anexar o arquivo (obrigatório)
      --file string     caminho do arquivo local (obrigatório)
  -h, --help            help for attach
      --name string     nome do arquivo a registrar no Movidesk (padrão: basename de --file)
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

