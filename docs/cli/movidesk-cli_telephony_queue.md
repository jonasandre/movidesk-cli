## movidesk-cli telephony queue

POST a queue-controlled call event (--event receivedCall|transferedCall|completedCall|lostCall|canceledCall)

```
movidesk-cli telephony queue [flags]
```

### Options

```
      --event string   event name (receivedCall, transferedCall, completedCall, lostCall, canceledCall)
  -f, --file string    path to JSON body
  -h, --help           help for queue
      --set strings    override fields, e.g. --set id=abc --set queueId=1
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

