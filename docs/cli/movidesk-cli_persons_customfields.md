## movidesk-cli persons customfields

Lê e escreve campos personalizados de pessoa (read-merge-patch)

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

* [movidesk-cli persons](movidesk-cli_persons.md)	 - Gerencia pessoas do Movidesk (/persons)
* [movidesk-cli persons customfields clear](movidesk-cli_persons_customfields_clear.md)	 - Remove o valor de um campo personalizado de pessoa (read-merge-patch)
* [movidesk-cli persons customfields set](movidesk-cli_persons_customfields_set.md)	 - Define o valor de um campo personalizado de pessoa (read-merge-patch)
* [movidesk-cli persons customfields show](movidesk-cli_persons_customfields_show.md)	 - Lista os customFieldValues de uma pessoa

