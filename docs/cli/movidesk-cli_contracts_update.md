## movidesk-cli contracts update

Patch a contract by id

```
movidesk-cli contracts update <id> [flags]
```

### Options

```
  -f, --file string                 
      --from-template string        
      --from-template-file string   
  -h, --help                        help for update
      --set strings                 
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

* [movidesk-cli contracts](movidesk-cli_contracts.md)	 - Manage Movidesk hour contracts (/timeAgreement)

