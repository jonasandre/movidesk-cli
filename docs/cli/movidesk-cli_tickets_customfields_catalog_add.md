## movidesk-cli tickets customfields catalog add

Registra um campo personalizado no catálogo local

```
movidesk-cli tickets customfields catalog add [flags]
```

### Options

```
      --field int         customFieldId numérico do Movidesk (obrigatório)
  -h, --help              help for add
      --label string      rótulo legível, ex.: "Severidade" (obrigatório)
      --options strings   opções permitidas para tipos de lista
      --rule int          customFieldRuleId numérico do Movidesk (obrigatório)
      --type string       tipo do campo (obrigatório)
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

* [movidesk-cli tickets customfields catalog](movidesk-cli_tickets_customfields_catalog.md)	 - Gerencia o catálogo local de campos personalizados por tenant

