## movidesk-cli tickets

Gerencia chamados (tickets) do Movidesk

### Synopsis

Gerencia chamados do Movidesk via os endpoints /tickets, /tickets/past,
/tickets/merged e /tickets/htmldescription.

Os verbos list/get aceitam parâmetros de consulta OData; create/update aceitam
um corpo JSON via --file, um template salvo via --from-template, ou substituições
inline via --set chave=valor.

### Options

```
  -h, --help   help for tickets
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
* [movidesk-cli tickets actions](movidesk-cli_tickets_actions.md)	 - Inspeciona e modifica ações de chamados
* [movidesk-cli tickets assets](movidesk-cli_tickets_assets.md)	 - Inspeciona ativos vinculados ao chamado
* [movidesk-cli tickets attach](movidesk-cli_tickets_attach.md)	 - Envia um arquivo para uma ação de chamado via /ticketFileUpload
* [movidesk-cli tickets bulk-close](movidesk-cli_tickets_bulk-close.md)	 - Encerra vários chamados em lote, registrando uma ação com a mensagem
* [movidesk-cli tickets bulk-update](movidesk-cli_tickets_bulk-update.md)	 - Aplica o mesmo PATCH a vários chamados (com seletor interativo)
* [movidesk-cli tickets clients](movidesk-cli_tickets_clients.md)	 - Inspeciona clientes de um chamado
* [movidesk-cli tickets create](movidesk-cli_tickets_create.md)	 - Cria um chamado a partir de corpo JSON, template ou substituições --set
* [movidesk-cli tickets customfields](movidesk-cli_tickets_customfields.md)	 - Lê e escreve campos personalizados de chamados (com read-merge-patch seguro)
* [movidesk-cli tickets get](movidesk-cli_tickets_get.md)	 - Obtém um chamado por id (posicional) ou protocolo (--protocol)
* [movidesk-cli tickets histories](movidesk-cli_tickets_histories.md)	 - Históricos de responsável e status
* [movidesk-cli tickets html](movidesk-cli_tickets_html.md)	 - Obtém o corpo HTML de um chamado (ou de uma de suas ações)
* [movidesk-cli tickets list](movidesk-cli_tickets_list.md)	 - Lista chamados (últimos 90 dias; mais antigos em `tickets past list`)
* [movidesk-cli tickets merged](movidesk-cli_tickets_merged.md)	 - Inspeciona chamados mesclados (/tickets/merged)
* [movidesk-cli tickets past](movidesk-cli_tickets_past.md)	 - Gerencia chamados com mais de 90 dias (/tickets/past)
* [movidesk-cli tickets relations](movidesk-cli_tickets_relations.md)	 - Lista chamados pais e filhos
* [movidesk-cli tickets timeline](movidesk-cli_tickets_timeline.md)	 - Mescla cronológica de ações, mudanças de status e responsável
* [movidesk-cli tickets update](movidesk-cli_tickets_update.md)	 - Aplica patch em um chamado por id

