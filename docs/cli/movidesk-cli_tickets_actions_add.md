## movidesk-cli tickets actions add

Append a new action to a ticket

```
movidesk-cli tickets actions add <ticket-id> [flags]
```

### Options

```
      --description string        action body (HTML on writes)
      --description-file string   read body from a file
  -h, --help                      help for add
      --internal                  alias for --type 1
      --justification string      status justification
      --public                    alias for --type 2
      --status string             transition the ticket to a status
      --tag strings               tag (repeatable)
      --type int                  action type: 1=internal, 2=public
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

* [movidesk-cli tickets actions](movidesk-cli_tickets_actions.md)	 - Inspect and modify ticket actions

