## movidesk-cli auth login

Add or update a tenant and store its API token

### Synopsis

Add or update a Movidesk tenant. The token is read from a hidden prompt
(or stdin when not a TTY) and saved to the OS keychain when available, otherwise
to an encrypted file under ~/.movidesk.

By default, login validates the token by issuing GET /persons?$top=1 against
the configured base URL. Use --skip-verify to bypass.

After validating the token, login optionally prompts for a default user
(Cod. Ref.) that will be auto-injected as createdBy on writes that need
attribution. Pass --user <id> to set it non-interactively, or skip the prompt
by leaving the answer empty. Use --skip-verify-user to skip the existence
check (handy when the token's permissions can't read the persons API).

```
movidesk-cli auth login [flags]
```

### Options

```
      --base-url string    override API base URL (sandbox)
  -h, --help               help for login
      --label string       human label, e.g. "Acme Prod"
      --make-default       set this tenant as the current one
      --skip-verify        do not validate the token against the API
      --skip-verify-user   skip existence check on the default user
      --tenant string      tenant name (required)
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

