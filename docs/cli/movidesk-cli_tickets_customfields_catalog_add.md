## movidesk-cli tickets customfields catalog add

Register a custom field in the local catalog

```
movidesk-cli tickets customfields catalog add [flags]
```

### Options

```
      --field int         numeric customFieldId from Movidesk (required)
  -h, --help              help for add
      --label string      human label, e.g. "Severidade" (required)
      --options strings   allowed options for list types
      --rule int          numeric customFieldRuleId from Movidesk (required)
      --type string       field type (required)
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

* [movidesk-cli tickets customfields catalog](movidesk-cli_tickets_customfields_catalog.md)	 - Manage the local catalog of custom fields per tenant

