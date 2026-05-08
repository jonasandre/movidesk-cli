## movidesk-cli tickets actions get

Get one action by id

```
movidesk-cli tickets actions get <ticket-id> --action-id N [flags]
```

### Options

```
      --action-id int   action id (required)
  -h, --help            help for get
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

