## movidesk-cli activities list

Lista atividades (paginação por cursor)

```
movidesk-cli activities list [flags]
```

### Options

```
      --all                     percorre todas as páginas
      --columns strings         colunas separadas por vírgula para saída table/csv (suporta dot-paths)
  -h, --help                    help for list
      --limit int               tamanho da página (1..100, padrão 100)
      --max int                 com --all, interrompe após este número de registros
      --name string             filtra por substring no nome da atividade
      --starting-after string   cursor (último id da página anterior)
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

* [movidesk-cli activities](movidesk-cli_activities.md)	 - Gerencia atividades do Movidesk (/activity)

