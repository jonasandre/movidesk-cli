## movidesk-cli telephony made-call-link

POST /setMadeCallLink — attach a recording link to an outbound call

```
movidesk-cli telephony made-call-link [flags]
```

### Options

```
  -f, --file string   path to JSON body
  -h, --help          help for made-call-link
      --set strings   override fields
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

* [movidesk-cli telephony](movidesk-cli_telephony.md)	 - Dispatch Movidesk call events (asterisk_*)

