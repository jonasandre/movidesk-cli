## movidesk-cli kb

Lê artigos da base de conhecimento do Movidesk (/article/:id)

### Synopsis

A API pública de base de conhecimento do Movidesk expõe apenas
leitura de artigo único (GET /article/:id). Não existe endpoint público
de listagem — é preciso conhecer o id do artigo previamente.

### Options

```
  -h, --help   help for kb
```

### Options inherited from parent commands

```
      --compact         JSON compacto (sem indentação)
      --no-color        desativa cores na saída
      --no-retry        desativa retentativa automática em 429/5xx
  -o, --output string   formato de saída: json|table|csv (padrão: do tenant ou 'json')
      --tenant string   nome do tenant (sobrepõe o tenant atual; env: MOVIDESK_TENANT)
      --user string     id do usuário padrão (Cod. Ref.) usado em createdBy nas escritas; sobrepõe a configuração do tenant; env: MOVIDESK_USER
  -v, --verbose         log detalhado em stderr
```

### SEE ALSO

* [movidesk-cli](movidesk-cli.md)	 - CLI para a API REST do Movidesk
* [movidesk-cli kb articles](movidesk-cli_kb_articles.md)	 - Artigos

