## movidesk-cli auth logout

Remove o token armazenado de um tenant (e opcionalmente o registro do tenant)

```
movidesk-cli auth logout [flags]
```

### Options

```
      --all             faz logout de todos os tenants configurados
  -h, --help            help for logout
      --tenant string   tenant para fazer logout (padrão: atual)
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

