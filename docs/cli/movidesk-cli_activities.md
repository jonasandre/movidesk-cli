## movidesk-cli activities

Gerencia atividades do Movidesk (/activity)

### Synopsis

Atividades usam paginação por cursor (limit/startingAfter) — não OData.
Use --name para filtrar por substring e --all para percorrer todas as páginas.

### Options

```
  -h, --help   help for activities
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
* [movidesk-cli activities add-teams](movidesk-cli_activities_add-teams.md)	 - Adiciona equipes a uma atividade (POST /addTeamsToActivity)
* [movidesk-cli activities create](movidesk-cli_activities_create.md)	 - Cria uma atividade
* [movidesk-cli activities delete](movidesk-cli_activities_delete.md)	 - Exclui uma atividade
* [movidesk-cli activities get](movidesk-cli_activities_get.md)	 - Obtém uma atividade pelo id
* [movidesk-cli activities list](movidesk-cli_activities_list.md)	 - Lista atividades (paginação por cursor)
* [movidesk-cli activities update](movidesk-cli_activities_update.md)	 - Aplica patch em uma atividade por id

