## movidesk-cli auth token

Imprime o token de um tenant no stdout (use com cuidado; para piping)

```
movidesk-cli auth token [flags]
```

### Options

```
  -h, --help            help for token
      --tenant string   tenant (padrão: atual)
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

