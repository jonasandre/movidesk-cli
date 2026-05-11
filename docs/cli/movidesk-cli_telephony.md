## movidesk-cli telephony

Dispara eventos de chamada do Movidesk (asterisk_*)

### Synopsis

Estes comandos disparam eventos de telefonia no Movidesk para que uma
integração de telefonia possa vincular chamadas a chamados. Duas variantes:

  queue       POST /asterisk_<evento>      (com controle de fila)
  nonqueue    GET  /asterisk_<evento>      (sem controle de fila)


### Options

```
  -h, --help   help for telephony
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
* [movidesk-cli telephony made-call-link](movidesk-cli_telephony_made-call-link.md)	 - POST /setMadeCallLink — vincula um link de gravação a uma chamada de saída
* [movidesk-cli telephony nonqueue](movidesk-cli_telephony_nonqueue.md)	 - GET de evento de chamada sem controle de fila (--event startTransferedCall|completedCall|startCanceledCall)
* [movidesk-cli telephony queue](movidesk-cli_telephony_queue.md)	 - POST de evento de chamada com controle de fila (--event receivedCall|transferedCall|completedCall|lostCall|canceledCall)

