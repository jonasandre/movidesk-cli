## movidesk-cli customfields

Gerencia o conjunto de opções de campos personalizados do tipo lista (por tenant)

### Synopsis

Estes comandos encapsulam os endpoints /ticketCustomFieldValue/{InsertValues,UpdateValues,DeleteValues},
que gerenciam o CONJUNTO DE OPÇÕES dos campos personalizados do tipo lista —
os valores que aparecem no dropdown pros agentes selecionarem.

Pra definir um valor em um CHAMADO ou PESSOA específica, use:
  tickets customfields set ...
  persons customfields set ...


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

* [movidesk-cli](movidesk-cli.md)	 - CLI para a API REST do Movidesk
* [movidesk-cli customfields options](movidesk-cli_customfields_options.md)	 - Adiciona/renomeia/remove valores do conjunto de opções

