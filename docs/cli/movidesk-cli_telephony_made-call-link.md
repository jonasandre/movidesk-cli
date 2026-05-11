## movidesk-cli telephony made-call-link

POST /setMadeCallLink — vincula um link de gravação a uma chamada de saída

```
movidesk-cli telephony made-call-link [flags]
```

### Options

```
  -f, --file string   caminho do corpo JSON
  -h, --help          help for made-call-link
      --set strings   sobrescreve campos
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

* [movidesk-cli telephony](movidesk-cli_telephony.md)	 - Dispara eventos de chamada do Movidesk (asterisk_*)

