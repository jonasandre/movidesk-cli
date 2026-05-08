## movidesk-cli surveys responses list

List survey responses (cursor pagination)

```
movidesk-cli surveys responses list [flags]
```

### Options

```
      --all                     walk every page
      --columns strings         comma-separated columns for table/csv output (dot-paths supported)
  -h, --help                    help for list
      --limit int               page size (1..100, default 100)
      --max int                 with --all, stop after this many records
      --starting-after string   cursor (id of last item from previous page)
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

* [movidesk-cli surveys responses](movidesk-cli_surveys_responses.md)	 - Survey responses

