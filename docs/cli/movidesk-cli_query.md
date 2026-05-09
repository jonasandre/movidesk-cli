## movidesk-cli query

Raw HTTP call against any Movidesk endpoint

### Synopsis

Raw escape hatch. Path is relative to the API base, e.g. "tickets",
"/tickets", "/persons". Token is injected automatically.

Examples:
  movidesk-cli query /tickets --filter "id eq 1"
  movidesk-cli query /persons --select id,businessName --top 5
  movidesk-cli query /persons --method GET --param id=abc

Output format follows --output (json|table|csv); use -o json with jq for further slicing.

```
movidesk-cli query <path> [flags]
```

### Options

```
      --all               fetch every page (auto-paginate)
      --expand strings    comma-separated $expand expressions
      --filter string     OData $filter expression
  -h, --help              help for query
      --max int           with --all, stop after this many records
      --method string     HTTP method (GET or DELETE only) (default "GET")
      --orderby strings   comma-separated $orderby clauses (e.g. "id desc")
      --param strings     extra query param key=value (repeatable)
      --select strings    comma-separated $select fields
      --skip int          $skip: server-side offset
      --top int           $top: page size or single-page limit
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

