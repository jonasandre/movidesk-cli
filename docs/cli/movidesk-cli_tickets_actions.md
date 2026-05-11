## movidesk-cli tickets actions

Inspeciona e modifica ações de chamados

### Options

```
  -h, --help   help for actions
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
* [movidesk-cli tickets actions add](movidesk-cli_tickets_actions_add.md)	 - Adiciona uma nova ação a um chamado
* [movidesk-cli tickets actions delete](movidesk-cli_tickets_actions_delete.md)	 - Marca uma ação como excluída (soft delete: isDeleted: true)
* [movidesk-cli tickets actions get](movidesk-cli_tickets_actions_get.md)	 - Obtém uma ação pelo id
* [movidesk-cli tickets actions list](movidesk-cli_tickets_actions_list.md)	 - Lista as ações de um chamado
* [movidesk-cli tickets actions update](movidesk-cli_tickets_actions_update.md)	 - Edita uma ação existente pelo id

