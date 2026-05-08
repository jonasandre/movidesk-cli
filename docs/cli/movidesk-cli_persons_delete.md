## movidesk-cli persons delete

Permanently delete a person (DELETE /persons?id=)

```
movidesk-cli persons delete <id> [flags]
```

### Options

```
      --force   skip confirmation prompt
  -h, --help    help for delete
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

* [movidesk-cli persons](movidesk-cli_persons.md)	 - Manage Movidesk persons (/persons)

