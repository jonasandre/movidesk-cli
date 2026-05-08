## movidesk-cli surveys questions list

List survey questions

```
movidesk-cli surveys questions list [flags]
```

### Options

```
      --columns strings   comma-separated columns for table/csv output (dot-paths supported)
  -h, --help              help for list
      --type int          filter by question type (1=satisf, 2=faces, 3=NPS, 4=yes/no)
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

