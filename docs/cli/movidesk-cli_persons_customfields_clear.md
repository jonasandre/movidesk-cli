## movidesk-cli persons customfields clear

Remove o valor de um campo personalizado de pessoa (read-merge-patch)

```
movidesk-cli persons customfields clear <person-id> [flags]
```

### Options

```
      --field int            id numérico do campo personalizado
      --field-label string   rótulo do catálogo
  -h, --help                 help for clear
      --line int             linha específica; omita para limpar todas
      --rule int             id da regra (omita se usar catálogo)
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

* [movidesk-cli persons customfields](movidesk-cli_persons_customfields.md)	 - Lê e escreve campos personalizados de pessoa (read-merge-patch)

