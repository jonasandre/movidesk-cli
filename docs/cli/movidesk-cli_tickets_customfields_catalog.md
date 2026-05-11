## movidesk-cli tickets customfields catalog

Gerencia o catálogo local de campos personalizados por tenant

### Options

```
  -h, --help   help for catalog
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
* [movidesk-cli tickets customfields catalog add](movidesk-cli_tickets_customfields_catalog_add.md)	 - Registra um campo personalizado no catálogo local
* [movidesk-cli tickets customfields catalog list](movidesk-cli_tickets_customfields_catalog_list.md)	 - Lista as entradas do catálogo do tenant atual
* [movidesk-cli tickets customfields catalog remove](movidesk-cli_tickets_customfields_catalog_remove.md)	 - Remove um rótulo do catálogo

