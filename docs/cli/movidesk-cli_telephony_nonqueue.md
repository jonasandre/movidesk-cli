## movidesk-cli telephony nonqueue

GET a no-queue call event (--event startTransferedCall|completedCall|startCanceledCall)

```
movidesk-cli telephony nonqueue [flags]
```

### Options

```
      --event string    event name (startTransferedCall, completedCall, startCanceledCall)
  -h, --help            help for nonqueue
      --param strings   query param key=value (repeatable), e.g. --param id=abc --param branchLine=1001
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

