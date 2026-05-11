package odata

// FilterTopic is the long help text that documents the OData filter syntax
// accepted by every Movidesk list/query endpoint. It is shared by the
// `topics filters` CLI page and the MCP server's `movidesk://odata-filter-syntax`
// resource so both surfaces stay in sync from a single source.
const FilterTopic = `Sintaxe de filtros aceita pelo Movidesk
=======================================

O valor passado em --filter (e os demais $select/$expand/$orderby) é
encaminhado literalmente para a API do Movidesk como o parâmetro OData
$filter. Nada é validado no cliente: se a expressão estiver errada o
Movidesk responde HTTP 400 com a mensagem original.

Referência oficial:
  https://atendimento.movidesk.com/kb/pt-br/article/256/movidesk-ticket-api

Operadores de comparação
------------------------
  eq   igual           ne   diferente
  lt   menor           le   menor ou igual
  gt   maior           ge   maior ou igual

Operadores lógicos
------------------
  and   or   not   (...)   parênteses controlam precedência

Funções de string
-----------------
  startswith(campo, 'texto')
  endswith(campo, 'texto')
  contains(campo, 'texto')
  substringof('texto', campo)     -- variante OData v3 ainda aceita
  tolower(campo)
  toupper(campo)

Literais
--------
  string   aspas simples:   'Qlik'
           apóstrofo interno é duplicado:   'O''Brien'
  número   sem aspas:       42         3.14
  data     ISO 8601 UTC com Z:   2026-04-01T00:00:00.000Z
           ('+00:00' costuma ser rejeitado, use sempre 'Z')
  boolean  true  /  false
  nulo     null

Campos úteis por recurso (lista pragmática, não exaustiva)
----------------------------------------------------------
  tickets (movidesk-cli tickets list)
    id, protocol, subject, type (1=interno, 2=externo),
    status, baseStatus ('New','InAttendance','Stopped','Resolved','Closed','Canceled'),
    justification, ownerTeam (STRING, não navegação),
    category, urgency, urgencyId,
    createdDate, lastUpdate, lastActionDate, slaSolutionDate,
    createdBy/id, owner/id, owner/businessName

  persons (movidesk-cli persons list)
    id, businessName, corporateName, codeReferenceAdditional,
    personType (1=Pessoa, 2=Empresa, 4=Departamento),
    profileType (1=Agente, 2=Cliente, 3=Ambos),
    isActive, userName

  services (movidesk-cli services list)
    id, name, parentServiceId, isVisible, allowFinalUser

  contracts (movidesk-cli contracts list)
    id, name, isActive, beginDate, endDate

Armadilhas frequentes
---------------------
  * ownerTeam é uma STRING, não uma navigation property.
      certo:    --filter "ownerTeam eq 'Qlik'"
      errado:   --filter "ownerTeam/name eq 'Qlik'"
                  → HTTP 400 "segment ... isn't Navigation/Structural/Complex/Collections"

  * No shell, use aspas duplas externas para preservar as aspas simples
    da expressão OData:
      --filter "id eq 1 and ownerTeam eq 'Qlik'"

  * --select/--expand afetam apenas a resposta. Para filtrar por um
    campo aninhado o --filter referencia o caminho direto
    (ex.: createdBy/id eq 'abc'), independente de você expandir ou não.

  * Datas precisam ser UTC com sufixo 'Z'. Fuso explícito ('+00:00')
    ou sem timezone costuma ser rejeitado.

  * tickets list cobre apenas os últimos 90 dias. Para chamados mais
    antigos use 'tickets past list' (mesmos filtros).

Exemplos
--------
  # tickets criados em abril/2026 do time "Qlik" (caso que motivou o tópico)
  movidesk-cli tickets list --all \
    --filter "createdDate ge 2026-04-01T00:00:00.000Z and createdDate lt 2026-05-01T00:00:00.000Z and ownerTeam eq 'Qlik'" \
    --select "id,protocol,subject,createdDate"

  # empresas ativas
  movidesk-cli persons list \
    --filter "personType eq 2 and isActive eq true" \
    --select "id,businessName"

  # tickets cujo assunto começa com "Falha"
  movidesk-cli tickets list \
    --filter "startswith(subject, 'Falha')" \
    --top 10

  # tickets em atendimento, ordenando do mais recente
  movidesk-cli tickets list \
    --filter "baseStatus eq 'InAttendance'" \
    --orderby "createdDate desc" --top 20

  # escape hatch: qualquer endpoint OData via query
  movidesk-cli query /tickets --filter "id eq 12345" --select id,subject
`
