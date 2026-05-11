## movidesk-cli tickets customfields set

Define o valor de um campo personalizado (read-merge-patch)

```
movidesk-cli tickets customfields set <ticket-id> [flags]
```

### Options

```
      --field int             id numérico do campo personalizado (ou use --field-label)
      --field-label string    rótulo registrado no catálogo
  -h, --help                  help for set
      --item strings          rótulo de item da lista de valores (repetível)
      --item-client strings   id do cliente (repetível)
      --item-person strings   id da pessoa (repetível)
      --item-team strings     nome da equipe (repetível)
      --line int              número da linha (padrão 1)
      --rule int              id da regra (vem do catálogo se omitido)
      --value string          valor para tipos texto/numérico/data/etc.
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

* [movidesk-cli tickets customfields](movidesk-cli_tickets_customfields.md)	 - Lê e escreve campos personalizados de chamados (com read-merge-patch seguro)

