## movidesk-cli contracts consumption

Read /timeAgreementConsumption

### Synopsis

When filtering by startPeriod/endPeriod, Movidesk requires the contract
name in the OData $filter (e.g. --filter "name eq 'Default' and period gt
2026-01-01T00:00:00Z"). Avoid combining $select with period filters.

### Options

```
  -h, --help   help for consumption
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
* [movidesk-cli contracts consumption list](movidesk-cli_contracts_consumption_list.md)	 - List consumption rows

