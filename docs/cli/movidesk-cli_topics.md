## movidesk-cli topics

Tópicos de ajuda detalhados (sintaxe, convenções, armadilhas)

### Synopsis

Coleção de páginas de ajuda longas que não cabem no --help de um
comando específico. Cada subcomando aqui imprime uma referência pronta
para consulta. Hoje há:

  topics filters    Sintaxe OData aceita pelos comandos list (--filter, --select, …)

```
movidesk-cli topics [flags]
```

### Options

```
  -h, --help   help for topics
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
* [movidesk-cli topics filters](movidesk-cli_topics_filters.md)	 - Sintaxe OData aceita pelo --filter (operadores, tipos, campos, exemplos)

