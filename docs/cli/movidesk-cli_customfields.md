## movidesk-cli customfields

Manage list-type custom field option pools (tenant-wide)

### Synopsis

These commands wrap the /ticketCustomFieldValue/{InsertValues,UpdateValues,DeleteValues}
endpoints, which manage the OPTION POOL of list-type custom fields — the set
of values agents can pick from in the dropdown.

To set a value on a SPECIFIC ticket or person, use:
  tickets customfields set ...
  persons customfields set ...


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

* [movidesk-cli](movidesk-cli.md)	 - CLI for the Movidesk REST API
* [movidesk-cli customfields options](movidesk-cli_customfields_options.md)	 - Add/rename/remove option-pool values

