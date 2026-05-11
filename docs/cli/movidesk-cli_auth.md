## movidesk-cli auth

Gerencia tokens e tenants do Movidesk

### Options

```
  -h, --help   help for auth
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
* [movidesk-cli auth list](movidesk-cli_auth_list.md)	 - Lista os tenants configurados
* [movidesk-cli auth login](movidesk-cli_auth_login.md)	 - Adiciona ou atualiza um tenant e armazena seu token de API
* [movidesk-cli auth logout](movidesk-cli_auth_logout.md)	 - Remove o token armazenado de um tenant (e opcionalmente o registro do tenant)
* [movidesk-cli auth set-user](movidesk-cli_auth_set-user.md)	 - Define ou remove o usuário padrão (Cod. Ref.) do tenant atual
* [movidesk-cli auth status](movidesk-cli_auth_status.md)	 - Valida o token do tenant atual (ou do informado)
* [movidesk-cli auth switch](movidesk-cli_auth_switch.md)	 - Troca o tenant atual
* [movidesk-cli auth token](movidesk-cli_auth_token.md)	 - Imprime o token de um tenant no stdout (use com cuidado; para piping)

