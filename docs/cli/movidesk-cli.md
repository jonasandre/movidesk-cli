## movidesk-cli

CLI for the Movidesk REST API

### Synopsis

movidesk-cli is a command-line interface for the Movidesk REST API.

It supports multiple tenants (each with its own token stored in the OS keychain),
formats output as JSON, table or CSV, and respects Movidesk's rate limit of
10 requests per minute with automatic retry on 429.

### Options

```
      --compact         compact JSON output (no indentation)
  -h, --help            help for movidesk-cli
      --no-color        disable colored output
      --no-retry        disable automatic retry on 429/5xx
  -o, --output string   output format: json|table|csv (default: tenant or 'json')
      --tenant string   tenant name (overrides current tenant; env: MOVIDESK_TENANT)
      --user string     default user id (Cod. Ref.) for createdBy on writes; overrides tenant config; env: MOVIDESK_USER
  -v, --verbose         verbose logging to stderr
```

### SEE ALSO

* [movidesk-cli activities](movidesk-cli_activities.md)	 - Manage Movidesk activities (/activity)
* [movidesk-cli auth](movidesk-cli_auth.md)	 - Manage Movidesk tokens and tenants
* [movidesk-cli contracts](movidesk-cli_contracts.md)	 - Manage Movidesk hour contracts (/timeAgreement)
* [movidesk-cli customfields](movidesk-cli_customfields.md)	 - Manage list-type custom field option pools (tenant-wide)
* [movidesk-cli kb](movidesk-cli_kb.md)	 - Read Movidesk knowledge base articles (/article/:id)
* [movidesk-cli persons](movidesk-cli_persons.md)	 - Manage Movidesk persons (/persons)
* [movidesk-cli query](movidesk-cli_query.md)	 - Raw HTTP call against any Movidesk endpoint
* [movidesk-cli services](movidesk-cli_services.md)	 - Manage Movidesk service catalog (/services)
* [movidesk-cli surveys](movidesk-cli_surveys.md)	 - Read Movidesk satisfaction survey data (/survey/...)
* [movidesk-cli telephony](movidesk-cli_telephony.md)	 - Dispatch Movidesk call events (asterisk_*)
* [movidesk-cli tickets](movidesk-cli_tickets.md)	 - Manage Movidesk tickets

