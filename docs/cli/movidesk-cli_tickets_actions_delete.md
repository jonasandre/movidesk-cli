## movidesk-cli tickets actions delete

Marca uma ação como excluída (soft delete: isDeleted: true)

```
movidesk-cli tickets actions delete <ticket-id> [flags]
```

### Options

```
      --action-id int   id da ação (obrigatório)
  -h, --help            help for delete
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

* [movidesk-cli tickets actions](movidesk-cli_tickets_actions.md)	 - Inspeciona e modifica ações de chamados

