## movidesk-cli services create

Create a service from a JSON body, template, or --set overrides

```
movidesk-cli services create [flags]
```

### Options

```
  -f, --file string                 path to JSON body
      --from-template string        load ~/.movidesk/templates/<name>.json
      --from-template-file string   load template from a specific path
  -h, --help                        help for create
      --return-all                  ask Movidesk to return the full service
      --set strings                 override fields, e.g. --set name="Suporte" --set isActive=true
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

* [movidesk-cli services](movidesk-cli_services.md)	 - Manage Movidesk service catalog (/services)

