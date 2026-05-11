## movidesk-cli activities add-teams

Adiciona equipes a uma atividade (POST /addTeamsToActivity)

```
movidesk-cli activities add-teams <activity-id> [flags]
```

### Options

```
  -h, --help           help for add-teams
      --team strings   nome da equipe a adicionar (repetível)
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

* [movidesk-cli activities](movidesk-cli_activities.md)	 - Gerencia atividades do Movidesk (/activity)

