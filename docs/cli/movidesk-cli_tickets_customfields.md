## movidesk-cli tickets customfields

Lê e escreve campos personalizados de chamados (com read-merge-patch seguro)

### Synopsis

O PATCH /tickets do Movidesk apaga qualquer entrada de customFieldValues
ausente no corpo. Este subcomando usa read-merge-patch internamente, então
você só descreve a alteração desejada, nunca a lista completa.

Um catálogo local em ~/.movidesk/<tenant>/customfields.yaml mapeia rótulos
legíveis para os ids numéricos e tipos dos campos, permitindo usar
--field-label "Severidade" no lugar de --field 125529.

### Options

```
  -h, --help   help for customfields
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
* [movidesk-cli tickets customfields catalog](movidesk-cli_tickets_customfields_catalog.md)	 - Gerencia o catálogo local de campos personalizados por tenant
* [movidesk-cli tickets customfields clear](movidesk-cli_tickets_customfields_clear.md)	 - Remove o valor de um campo personalizado (read-merge-patch)
* [movidesk-cli tickets customfields set](movidesk-cli_tickets_customfields_set.md)	 - Define o valor de um campo personalizado (read-merge-patch)
* [movidesk-cli tickets customfields show](movidesk-cli_tickets_customfields_show.md)	 - Lista os customFieldValues de um chamado

