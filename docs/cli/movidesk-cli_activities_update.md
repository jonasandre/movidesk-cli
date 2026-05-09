## movidesk-cli activities update

Patch an activity by id

```
movidesk-cli activities update <id> [flags]
```

### Options

```
  -f, --file string                 path to JSON patch body
      --from-template string        load ~/.movidesk/templates/<name>.json
      --from-template-file string   load template from a specific path
  -h, --help                        help for update
      --set strings                 override fields inline, e.g. --set name=Foo
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

* [movidesk-cli activities](movidesk-cli_activities.md)	 - Manage Movidesk activities (/activity)

