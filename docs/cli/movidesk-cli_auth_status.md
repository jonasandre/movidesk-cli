## movidesk-cli auth status

Validate the current (or specified) tenant token

```
movidesk-cli auth status [flags]
```

### Options

```
  -h, --help            help for status
      --tenant string   tenant to check (default: current)
```

### Options inherited from parent commands

```
      --compact         compact JSON output (no indentation)
      --no-color        disable colored output
      --no-retry        disable automatic retry on 429/5xx
  -o, --output string   output format: json|table|csv (default: tenant or 'json')
      --user string     default user id (Cod. Ref.) for createdBy on writes; overrides tenant config; env: MOVIDESK_USER
  -v, --verbose         verbose logging to stderr
```

### SEE ALSO

* [movidesk-cli auth](movidesk-cli_auth.md)	 - Manage Movidesk tokens and tenants

