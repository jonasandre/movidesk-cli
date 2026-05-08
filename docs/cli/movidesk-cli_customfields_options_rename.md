## movidesk-cli customfields options rename

Rename existing option values via --pair OLD=NEW (repeatable)

```
movidesk-cli customfields options rename [flags]
```

### Options

```
      --field string   numeric customFieldId (required)
  -h, --help           help for rename
      --pair strings   OLD=NEW (repeatable)
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

* [movidesk-cli customfields options](movidesk-cli_customfields_options.md)	 - Add/rename/remove option-pool values

