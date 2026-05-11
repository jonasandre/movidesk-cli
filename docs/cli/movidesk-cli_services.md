## movidesk-cli services

Gerencia o catálogo de serviços do Movidesk (/services)

### Synopsis

Gerencia entradas do catálogo de serviços do Movidesk.

Atenção: PATCH substitui campos com valor de array como "categories" — ao
atualizar, envie a lista completa que deseja manter, não apenas os adicionais.

### Options

```
  -h, --help   help for services
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
* [movidesk-cli services create](movidesk-cli_services_create.md)	 - Cria um serviço a partir de corpo JSON, template ou substituições --set
* [movidesk-cli services delete](movidesk-cli_services_delete.md)	 - Exclui um serviço de forma permanente (DELETE /services?id=)
* [movidesk-cli services get](movidesk-cli_services_get.md)	 - Obtém um serviço pelo id
* [movidesk-cli services list](movidesk-cli_services_list.md)	 - Lista serviços
* [movidesk-cli services update](movidesk-cli_services_update.md)	 - Aplica patch em um serviço por id

