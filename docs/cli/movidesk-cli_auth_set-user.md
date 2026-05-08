## movidesk-cli auth set-user

Set or clear the default user (Cod. Ref.) for the current tenant

### Synopsis

Sets the default user that the CLI auto-injects as createdBy on writes that
need attribution (e.g. tickets create, tickets actions add). Override per
command with --user <id>.

Pass --clear to remove the configured default.

```
movidesk-cli auth set-user [<id>] [flags]
```

### Options

```
      --clear              remove the configured default user
  -h, --help               help for set-user
      --skip-verify-user   skip existence check
      --tenant string      tenant to update (default: current)
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

