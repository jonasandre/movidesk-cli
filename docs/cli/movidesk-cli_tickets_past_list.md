## movidesk-cli tickets past list

List archived tickets (older than 90 days)

```
movidesk-cli tickets past list [flags]
```

### Options

```
      --all               fetch every page (auto-paginate)
      --columns strings   comma-separated columns for table/csv output (dot-paths supported)
      --expand strings    comma-separated $expand expressions
      --filter string     OData $filter expression
  -h, --help              help for list
      --max int           with --all, stop after this many records
      --orderby strings   comma-separated $orderby clauses (e.g. "id desc")
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

* [movidesk-cli tickets past](movidesk-cli_tickets_past.md)	 - Manage tickets older than 90 days (/tickets/past)

