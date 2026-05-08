## movidesk-cli activities get

Get one activity by id

```
movidesk-cli activities get <id> [flags]
```

### Options

```
      --columns strings   comma-separated columns for table/csv output (dot-paths supported)
  -h, --help              help for get
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

* [movidesk-cli activities](movidesk-cli_activities.md)	 - Manage Movidesk activities (/activity)

