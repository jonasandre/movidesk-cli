## movidesk-cli services

Manage Movidesk service catalog (/services)

### Synopsis

Manage entries from the Movidesk service catalog.

Note: PATCH replaces array-valued fields like "categories" — when updating,
send the complete list you want to keep, not just the additions.

### Options

```
  -h, --help   help for services
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
* [movidesk-cli services create](movidesk-cli_services_create.md)	 - Create a service from a JSON body, template, or --set overrides
* [movidesk-cli services delete](movidesk-cli_services_delete.md)	 - Permanently delete a service (DELETE /services?id=)
* [movidesk-cli services get](movidesk-cli_services_get.md)	 - Get one service by id
* [movidesk-cli services list](movidesk-cli_services_list.md)	 - List services
* [movidesk-cli services update](movidesk-cli_services_update.md)	 - Patch a service by id

