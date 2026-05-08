## movidesk-cli tickets assets list

List assets attached to a ticket

```
movidesk-cli tickets assets list <ticket-id> [flags]
```

### Options

```
      --columns strings   comma-separated columns for table/csv output (dot-paths supported)
  -h, --help              help for list
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

* [movidesk-cli tickets assets](movidesk-cli_tickets_assets.md)	 - Inspect ticket assets

