## movidesk-cli tickets customfields

Read and write ticket custom fields (with read-merge-patch safety)

### Synopsis

Movidesk's PATCH /tickets deletes any customFieldValues entry not
present in the body. This subcommand uses read-merge-patch internally so
you only describe the change you want, never the whole list.

A local catalog at ~/.movidesk/<tenant>/customfields.yaml maps human-friendly
labels to numeric field IDs and types so you can use --field-label "Severidade"
instead of --field 125529.

### Options

```
  -h, --help   help for customfields
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
* [movidesk-cli tickets customfields catalog](movidesk-cli_tickets_customfields_catalog.md)	 - Manage the local catalog of custom fields per tenant
* [movidesk-cli tickets customfields clear](movidesk-cli_tickets_customfields_clear.md)	 - Remove a custom field value (read-merge-patch)
* [movidesk-cli tickets customfields set](movidesk-cli_tickets_customfields_set.md)	 - Set a custom field value (read-merge-patch)
* [movidesk-cli tickets customfields show](movidesk-cli_tickets_customfields_show.md)	 - List a ticket's customFieldValues

