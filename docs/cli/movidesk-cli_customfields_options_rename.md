## movidesk-cli customfields options rename

Renomeia valores existentes via --pair ANTIGO=NOVO (repetível)

```
movidesk-cli customfields options rename [flags]
```

### Options

```
      --field string   customFieldId numérico (obrigatório)
  -h, --help           help for rename
      --pair strings   ANTIGO=NOVO (repetível)
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

* [movidesk-cli customfields options](movidesk-cli_customfields_options.md)	 - Adiciona/renomeia/remove valores do conjunto de opções

