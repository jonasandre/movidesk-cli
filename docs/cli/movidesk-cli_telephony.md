## movidesk-cli telephony

Dispatch Movidesk call events (asterisk_*)

### Synopsis

These commands fire telephony events at Movidesk so a phone system
integration can attach calls to tickets. Two flavors:

  queue       POST /asterisk_<event>      (with queue control)
  nonqueue    GET  /asterisk_<event>      (without queue control)


### Options

```
  -h, --help   help for telephony
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
* [movidesk-cli telephony made-call-link](movidesk-cli_telephony_made-call-link.md)	 - POST /setMadeCallLink — attach a recording link to an outbound call
* [movidesk-cli telephony nonqueue](movidesk-cli_telephony_nonqueue.md)	 - GET a no-queue call event (--event startTransferedCall|completedCall|startCanceledCall)
* [movidesk-cli telephony queue](movidesk-cli_telephony_queue.md)	 - POST a queue-controlled call event (--event receivedCall|transferedCall|completedCall|lostCall|canceledCall)

