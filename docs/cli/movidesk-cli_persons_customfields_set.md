## movidesk-cli persons customfields set

Set a person custom field value (read-merge-patch)

```
movidesk-cli persons customfields set <person-id> [flags]
```

### Options

```
      --field int             numeric custom field id (or use --field-label)
      --field-label string    label registered in the catalog
  -h, --help                  help for set
      --item strings          list-of-values item label (repeatable)
      --item-client strings   client id (repeatable)
      --item-person strings   person id (repeatable)
      --item-team strings     team name (repeatable)
      --line int              row number (default 1)
      --rule int              rule id (taken from catalog if omitted)
      --value string          value for text/numeric/date types
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

* [movidesk-cli persons customfields](movidesk-cli_persons_customfields.md)	 - Read and write person custom fields (read-merge-patch)

