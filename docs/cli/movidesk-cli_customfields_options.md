## movidesk-cli customfields options

Add/rename/remove option-pool values

### Options

```
  -h, --help   help for options
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

* [movidesk-cli customfields](movidesk-cli_customfields.md)	 - Manage list-type custom field option pools (tenant-wide)
* [movidesk-cli customfields options add](movidesk-cli_customfields_options_add.md)	 - Insert option values into a list-type field's pool
* [movidesk-cli customfields options remove](movidesk-cli_customfields_options_remove.md)	 - Remove option values from a list-type field's pool
* [movidesk-cli customfields options rename](movidesk-cli_customfields_options_rename.md)	 - Rename existing option values via --pair OLD=NEW (repeatable)

