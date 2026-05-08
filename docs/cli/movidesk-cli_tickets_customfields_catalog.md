## movidesk-cli tickets customfields catalog

Manage the local catalog of custom fields per tenant

### Options

```
  -h, --help   help for catalog
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
* [movidesk-cli tickets customfields catalog add](movidesk-cli_tickets_customfields_catalog_add.md)	 - Register a custom field in the local catalog
* [movidesk-cli tickets customfields catalog list](movidesk-cli_tickets_customfields_catalog_list.md)	 - List catalog entries for the current tenant
* [movidesk-cli tickets customfields catalog remove](movidesk-cli_tickets_customfields_catalog_remove.md)	 - Remove a label from the catalog

