## movidesk-cli auth logout

Remove a tenant's stored token (and optionally the tenant entry)

```
movidesk-cli auth logout [flags]
```

### Options

```
      --all             log out every configured tenant
  -h, --help            help for logout
      --tenant string   tenant to log out (default: current)
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

