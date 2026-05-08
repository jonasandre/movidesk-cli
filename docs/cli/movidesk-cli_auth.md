## movidesk-cli auth

Manage Movidesk tokens and tenants

### Options

```
  -h, --help   help for auth
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
* [movidesk-cli auth list](movidesk-cli_auth_list.md)	 - List configured tenants
* [movidesk-cli auth login](movidesk-cli_auth_login.md)	 - Add or update a tenant and store its API token
* [movidesk-cli auth logout](movidesk-cli_auth_logout.md)	 - Remove a tenant's stored token (and optionally the tenant entry)
* [movidesk-cli auth set-user](movidesk-cli_auth_set-user.md)	 - Set or clear the default user (Cod. Ref.) for the current tenant
* [movidesk-cli auth status](movidesk-cli_auth_status.md)	 - Validate the current (or specified) tenant token
* [movidesk-cli auth switch](movidesk-cli_auth_switch.md)	 - Switch the current tenant
* [movidesk-cli auth token](movidesk-cli_auth_token.md)	 - Print a tenant's token to stdout (use with care; for piping)

