## movidesk-cli auth set-user

Define ou remove o usuário padrão (Cod. Ref.) do tenant atual

### Synopsis

Define o usuário padrão que o CLI injeta como createdBy nas escritas que
exigem atribuição (ex.: tickets create, tickets actions add). Sobreponha por
comando com --user <id>.

Use --clear para remover o padrão configurado.

```
movidesk-cli auth set-user [<id>] [flags]
```

### Options

```
      --clear              remove o usuário padrão configurado
  -h, --help               help for set-user
      --skip-verify-user   pula a checagem de existência
      --tenant string      tenant a atualizar (padrão: atual)
```

### Options inherited from parent commands

```
      --compact         JSON compacto (sem indentação)
      --no-color        desativa cores na saída
      --no-retry        desativa retentativa automática em 429/5xx
  -o, --output string   formato de saída: json|table|csv (padrão: do tenant ou 'json')
      --user string     id do usuário padrão (Cod. Ref.) usado em createdBy nas escritas; sobrepõe a configuração do tenant; env: MOVIDESK_USER
  -v, --verbose         log detalhado em stderr
```

### SEE ALSO

* [movidesk-cli auth](movidesk-cli_auth.md)	 - Gerencia tokens e tenants do Movidesk

