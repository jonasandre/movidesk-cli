## movidesk-cli persons

Manage Movidesk persons (/persons)

### Synopsis

Manage Movidesk persons. The /persons endpoint serves agents,
clients, companies, and departments — disambiguated by personType
(1=Pessoa, 2=Empresa, 4=Departamento) and profileType (1=Agente, 2=Cliente,
3=Both).

OData filters and projection apply on list. Custom field values follow the
same read-merge-patch semantics as tickets to avoid Movidesk's "delete
missing entries" trap.

### Options

```
  -h, --help   help for persons
```

### Options inherited from parent commands

```
      --compact         compact JSON output (no indentation)
      --no-color        disable colored output
      --no-retry        disable automatic retry on 429/5xx
  -o, --output string   output format: json|table|csv (default: tenant or 'json')
      --tenant string   tenant name (overrides current tenant; env: MOVIDESK_TENANT)
      --user string     default user id (Cod. Ref.) for createdBy on writes; overrides tenant config; env: MOVIDESK_USER
  -v, --verbose         verbose logging to stderr
```

### SEE ALSO

* [movidesk-cli](movidesk-cli.md)	 - CLI for the Movidesk REST API
* [movidesk-cli persons create](movidesk-cli_persons_create.md)	 - Create a person from a JSON body, template, or --set overrides
* [movidesk-cli persons customfields](movidesk-cli_persons_customfields.md)	 - Read and write person custom fields (read-merge-patch)
* [movidesk-cli persons delete](movidesk-cli_persons_delete.md)	 - Permanently delete a person (DELETE /persons?id=)
* [movidesk-cli persons get](movidesk-cli_persons_get.md)	 - Get one person by id (Cod. Ref.)
* [movidesk-cli persons list](movidesk-cli_persons_list.md)	 - List persons
* [movidesk-cli persons update](movidesk-cli_persons_update.md)	 - Patch a person by id

