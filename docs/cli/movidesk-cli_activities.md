## movidesk-cli activities

Manage Movidesk activities (/activity)

### Synopsis

Activities use cursor-based pagination (limit/startingAfter) — not OData.
Use --name to substring-filter, --all to walk every page.

### Options

```
  -h, --help   help for activities
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
* [movidesk-cli activities add-teams](movidesk-cli_activities_add-teams.md)	 - Append teams to an activity (POST /addTeamsToActivity)
* [movidesk-cli activities create](movidesk-cli_activities_create.md)	 - Create an activity
* [movidesk-cli activities delete](movidesk-cli_activities_delete.md)	 - Delete an activity
* [movidesk-cli activities get](movidesk-cli_activities_get.md)	 - Get one activity by id
* [movidesk-cli activities list](movidesk-cli_activities_list.md)	 - List activities (cursor pagination)
* [movidesk-cli activities update](movidesk-cli_activities_update.md)	 - Patch an activity by id

