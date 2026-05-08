## movidesk-cli tickets actions

Inspect and modify ticket actions

### Options

```
  -h, --help   help for actions
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

* [movidesk-cli tickets](movidesk-cli_tickets.md)	 - Manage Movidesk tickets
* [movidesk-cli tickets actions add](movidesk-cli_tickets_actions_add.md)	 - Append a new action to a ticket
* [movidesk-cli tickets actions delete](movidesk-cli_tickets_actions_delete.md)	 - Soft-delete an action (sets isDeleted: true)
* [movidesk-cli tickets actions get](movidesk-cli_tickets_actions_get.md)	 - Get one action by id
* [movidesk-cli tickets actions list](movidesk-cli_tickets_actions_list.md)	 - List actions of a ticket
* [movidesk-cli tickets actions update](movidesk-cli_tickets_actions_update.md)	 - Edit an existing action by id

