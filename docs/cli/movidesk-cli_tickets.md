## movidesk-cli tickets

Manage Movidesk tickets

### Synopsis

Manage Movidesk tickets via the /tickets, /tickets/past, /tickets/merged
and /tickets/htmldescription endpoints.

The list/get verbs accept OData query parameters; create/update accept a JSON
body via --file, a saved template via --from-template, or inline overrides
via --set key=value.

### Options

```
  -h, --help   help for tickets
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
* [movidesk-cli tickets actions](movidesk-cli_tickets_actions.md)	 - Inspect and modify ticket actions
* [movidesk-cli tickets assets](movidesk-cli_tickets_assets.md)	 - Inspect ticket assets
* [movidesk-cli tickets attach](movidesk-cli_tickets_attach.md)	 - Upload a file to a ticket action via /ticketFileUpload
* [movidesk-cli tickets clients](movidesk-cli_tickets_clients.md)	 - Inspect ticket clients
* [movidesk-cli tickets create](movidesk-cli_tickets_create.md)	 - Create a ticket from a JSON body, template, or --set overrides
* [movidesk-cli tickets customfields](movidesk-cli_tickets_customfields.md)	 - Read and write ticket custom fields (with read-merge-patch safety)
* [movidesk-cli tickets get](movidesk-cli_tickets_get.md)	 - Get one ticket by id (positional) or protocol (--protocol)
* [movidesk-cli tickets histories](movidesk-cli_tickets_histories.md)	 - Owner and status histories
* [movidesk-cli tickets html](movidesk-cli_tickets_html.md)	 - Get the HTML body of a ticket (or one of its actions)
* [movidesk-cli tickets list](movidesk-cli_tickets_list.md)	 - List tickets (last 90 days; older tickets in `tickets past list`)
* [movidesk-cli tickets merged](movidesk-cli_tickets_merged.md)	 - Inspect merged tickets (/tickets/merged)
* [movidesk-cli tickets past](movidesk-cli_tickets_past.md)	 - Manage tickets older than 90 days (/tickets/past)
* [movidesk-cli tickets relations](movidesk-cli_tickets_relations.md)	 - List parent and child tickets
* [movidesk-cli tickets timeline](movidesk-cli_tickets_timeline.md)	 - Chronological merge of actions, status and owner changes
* [movidesk-cli tickets update](movidesk-cli_tickets_update.md)	 - Patch a ticket by id

