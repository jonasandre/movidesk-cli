## movidesk-cli surveys questions get

Get a single survey question by id

```
movidesk-cli surveys questions get <id> [flags]
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

* [movidesk-cli surveys questions](movidesk-cli_surveys_questions.md)	 - Survey questions

