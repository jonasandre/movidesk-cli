## movidesk-cli

CLI para a API REST do Movidesk

### Synopsis

movidesk-cli é uma interface de linha de comando para a API REST do Movidesk.

Suporta múltiplos tenants (cada um com seu próprio token armazenado no chaveiro
do sistema operacional), formata a saída como JSON, tabela ou CSV, e respeita o
limite de 10 requisições por minuto do Movidesk com retentativa automática em 429.

### Options

```
      --compact         JSON compacto (sem indentação)
  -h, --help            help for movidesk-cli
      --no-color        desativa cores na saída
      --no-retry        desativa retentativa automática em 429/5xx
  -o, --output string   formato de saída: json|table|csv (padrão: do tenant ou 'json')
      --tenant string   nome do tenant (sobrepõe o tenant atual; env: MOVIDESK_TENANT)
      --user string     id do usuário padrão (Cod. Ref.) usado em createdBy nas escritas; sobrepõe a configuração do tenant; env: MOVIDESK_USER
  -v, --verbose         log detalhado em stderr
```

### SEE ALSO

* [movidesk-cli activities](movidesk-cli_activities.md)	 - Gerencia atividades do Movidesk (/activity)
* [movidesk-cli auth](movidesk-cli_auth.md)	 - Gerencia tokens e tenants do Movidesk
* [movidesk-cli contracts](movidesk-cli_contracts.md)	 - Gerencia contratos de horas do Movidesk (/timeAgreement)
* [movidesk-cli customfields](movidesk-cli_customfields.md)	 - Gerencia o conjunto de opções de campos personalizados do tipo lista (por tenant)
* [movidesk-cli kb](movidesk-cli_kb.md)	 - Lê artigos da base de conhecimento do Movidesk (/article/:id)
* [movidesk-cli persons](movidesk-cli_persons.md)	 - Gerencia pessoas do Movidesk (/persons)
* [movidesk-cli query](movidesk-cli_query.md)	 - Chamada HTTP bruta contra qualquer endpoint do Movidesk
* [movidesk-cli services](movidesk-cli_services.md)	 - Gerencia o catálogo de serviços do Movidesk (/services)
* [movidesk-cli surveys](movidesk-cli_surveys.md)	 - Lê dados de pesquisas de satisfação do Movidesk (/survey/...)
* [movidesk-cli telephony](movidesk-cli_telephony.md)	 - Dispara eventos de chamada do Movidesk (asterisk_*)
* [movidesk-cli tickets](movidesk-cli_tickets.md)	 - Gerencia chamados (tickets) do Movidesk
* [movidesk-cli topics](movidesk-cli_topics.md)	 - Tópicos de ajuda detalhados (sintaxe, convenções, armadilhas)

