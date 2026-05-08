## movidesk-cli tickets customfields clear

Remove a custom field value (read-merge-patch)

```
movidesk-cli tickets customfields clear <ticket-id> [flags]
```

### Options

```
      --field int            numeric custom field id
      --field-label string   label from the catalog
  -h, --help                 help for clear
      --line int             specific line; omit to clear every line
      --rule int             rule id (omit with catalog)
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

* [movidesk-cli tickets customfields](movidesk-cli_tickets_customfields.md)	 - Read and write ticket custom fields (with read-merge-patch safety)

