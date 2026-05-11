## movidesk-cli customfields options

Adiciona/renomeia/remove valores do conjunto de opções

### Options

```
  -h, --help   help for options
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

* [movidesk-cli customfields](movidesk-cli_customfields.md)	 - Gerencia o conjunto de opções de campos personalizados do tipo lista (por tenant)
* [movidesk-cli customfields options add](movidesk-cli_customfields_options_add.md)	 - Insere valores de opção no conjunto de um campo tipo lista
* [movidesk-cli customfields options remove](movidesk-cli_customfields_options_remove.md)	 - Remove valores de opção do conjunto de um campo tipo lista
* [movidesk-cli customfields options rename](movidesk-cli_customfields_options_rename.md)	 - Renomeia valores existentes via --pair ANTIGO=NOVO (repetível)

