## movidesk-cli tickets actions add

Adiciona uma nova ação a um chamado

```
movidesk-cli tickets actions add <ticket-id> [flags]
```

### Options

```
      --description string        corpo da ação (HTML em escritas)
      --description-file string   lê o corpo de um arquivo
  -h, --help                      help for add
      --internal                  alias para --type 1
      --justification string      justificativa do status
      --public                    alias para --type 2
      --status string             transiciona o chamado para um status
      --tag strings               tag (repetível)
      --type int                  tipo da ação: 1=interna, 2=pública
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

