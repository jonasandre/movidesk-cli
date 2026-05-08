## movidesk-cli kb

Read Movidesk knowledge base articles (/article/:id)

### Synopsis

Movidesk's public KB API exposes only single-article reads (GET
/article/:id). There is no public list endpoint — you must already know
the article id.

### Options

```
  -h, --help   help for kb
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
* [movidesk-cli kb articles](movidesk-cli_kb_articles.md)	 - Articles

