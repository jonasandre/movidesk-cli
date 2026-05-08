## movidesk-cli tickets attach

Upload a file to a ticket action via /ticketFileUpload

```
movidesk-cli tickets attach <ticket-id> [flags]
```

### Options

```
      --action-id int   action id to attach the file to (required)
      --file string     path to the local file (required)
  -h, --help            help for attach
      --name string     filename to record on Movidesk (default: basename of --file)
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

