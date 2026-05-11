package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/activities"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/contracts"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/knowledgebase"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/persons"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/services"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/surveys"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/tickets"
)

// registerTools wires every read-only MCP tool exposed by the v1 server.
func registerTools(s *mcpsdk.Server, c *movidesk.Client) {
	registerTickets(s, c)
	registerPersons(s, c)
	registerServices(s, c)
	registerContracts(s, c)
	registerKB(s, c)
	registerActivities(s, c)
	registerSurveys(s, c)
	registerQuery(s, c)
}

// ---------- tickets ----------

type ticketsListArgs struct {
	ODataArgs
	IncludeDeleted bool `json:"include_deleted,omitempty" jsonschema:"Inclui relações/registros marcados como excluídos."`
}

type ticketGetArgs struct {
	ID             int    `json:"id,omitempty" jsonschema:"ID numérico do ticket. Use ID xou protocol."`
	Protocol       string `json:"protocol,omitempty" jsonschema:"Protocolo do ticket (string). Use ID xou protocol."`
	IncludeDeleted bool   `json:"include_deleted,omitempty"`
}

type ticketHTMLArgs struct {
	ID       int `json:"id" jsonschema:"ID do ticket."`
	ActionID int `json:"action_id,omitempty" jsonschema:"ID da ação. 0 = corpo do ticket."`
}

type ticketIDArg struct {
	TicketID int `json:"ticket_id" jsonschema:"ID do ticket."`
}

type ticketsPastListArgs struct {
	ODataArgs
}

func registerTickets(s *mcpsdk.Server, c *movidesk.Client) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name: "tickets_list",
		Description: `Lista tickets dos últimos 90 dias suportando OData $filter/$select/$expand/$orderby.
$select é OBRIGATÓRIO (mín. ["id"]); top ≤ 250 (clamp implícito acima disso).
Campos comuns: id, protocol, subject, status, baseStatus ('New','InAttendance','Stopped','Resolved','Closed','Canceled'),
ownerTeam (STRING — use "ownerTeam eq 'Qlik'", nunca "ownerTeam/name eq ..."), createdDate, lastUpdate.
Datas devem ser UTC com sufixo 'Z'. Para tickets >90d use tickets_past_list. Veja resource movidesk://odata-filter-syntax.
Exemplo: {"select": ["id","protocol","subject"], "filter": "createdDate ge 2026-05-01T00:00:00Z", "top": 100}.`,
		InputSchema: buildInputSchema[ticketsListArgs](),
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a ticketsListArgs) (*mcpsdk.CallToolResult, any, error) {
		svc := tickets.New(c)
		q := a.Query()
		var warning string
		q.Top, warning = clampTicketsTop(q.Top)
		maxRows := applyDefaultMax(a.All, a.Max)
		if a.All {
			pageSize := q.Top
			q.Top = 0
			rows, err := svc.Paginate(ctx, q, a.IncludeDeleted, pageSize, maxRows)
			if err != nil {
				return nil, nil, wrapAPIError(err)
			}
			return withWarning(rowsResult(rows), warning), nil, nil
		}
		raw, err := svc.List(ctx, q, a.IncludeDeleted)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return withWarning(rawResult(raw), warning), nil, nil
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name: "tickets_past_list",
		Description: `Lista tickets arquivados (lastUpdate > 90 dias). Mesma sintaxe OData do tickets_list, sem include_deleted.
$select é OBRIGATÓRIO (mín. ["id"]); top ≤ 250 (clamp implícito).
Exemplo: {"select": ["id","protocol","createdDate"], "filter": "lastUpdate lt 2026-02-10T00:00:00Z", "top": 100}.`,
		InputSchema: buildInputSchema[ticketsPastListArgs](),
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a ticketsPastListArgs) (*mcpsdk.CallToolResult, any, error) {
		svc := tickets.New(c)
		q := a.Query()
		var warning string
		q.Top, warning = clampTicketsTop(q.Top)
		maxRows := applyDefaultMax(a.All, a.Max)
		if a.All {
			pageSize := q.Top
			q.Top = 0
			rows, err := svc.PaginatePast(ctx, q, pageSize, maxRows)
			if err != nil {
				return nil, nil, wrapAPIError(err)
			}
			return withWarning(rowsResult(rows), warning), nil, nil
		}
		raw, err := svc.Past(ctx, q)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return withWarning(rawResult(raw), warning), nil, nil
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "tickets_get",
		Description: "Busca um ticket por id OU protocol. Forneça exatamente um dos dois.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a ticketGetArgs) (*mcpsdk.CallToolResult, any, error) {
		if (a.ID == 0) == (a.Protocol == "") {
			return nil, nil, errors.New("informe id OU protocol (exatamente um)")
		}
		svc := tickets.New(c)
		var (
			raw []byte
			err error
		)
		if a.ID != 0 {
			raw, err = svc.GetByID(ctx, a.ID, a.IncludeDeleted)
		} else {
			raw, err = svc.GetByProtocol(ctx, a.Protocol, a.IncludeDeleted)
		}
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "tickets_html_description",
		Description: "Retorna o corpo HTML de um ticket (action_id=0) ou de uma de suas ações.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a ticketHTMLArgs) (*mcpsdk.CallToolResult, any, error) {
		raw, err := tickets.New(c).HTMLDescription(ctx, a.ID, a.ActionID)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "tickets_actions_list",
		Description: "Lista as ações (mensagens, anotações) de um ticket via $expand=actions.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a ticketIDArg) (*mcpsdk.CallToolResult, any, error) {
		actions, err := tickets.New(c).ListActions(ctx, a.TicketID)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return marshalResult(actions)
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "tickets_timeline",
		Description: "Linha do tempo cronológica de um ticket: ações + histórico de status + histórico de owner.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a ticketIDArg) (*mcpsdk.CallToolResult, any, error) {
		entries, err := tickets.New(c).Timeline(ctx, a.TicketID)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return marshalResult(entries)
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "tickets_customfields_show",
		Description: "Lista os customFieldValues atuais de um ticket.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a ticketIDArg) (*mcpsdk.CallToolResult, any, error) {
		cfs, err := tickets.New(c).ListCustomFieldValues(ctx, a.TicketID)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return marshalResult(cfs)
	})
}

// ---------- persons ----------

type personGetArgs struct {
	ID string `json:"id" jsonschema:"Cod. Ref. (id) da pessoa, agente ou empresa."`
}

type personsListArgs struct {
	ODataArgs
}

func registerPersons(s *mcpsdk.Server, c *movidesk.Client) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name: "persons_list",
		Description: `Lista pessoas, clientes, agentes e empresas com filtros OData.
Campos comuns: id, businessName, corporateName, personType (1=Pessoa,2=Empresa,4=Departamento),
profileType (1=Agente,2=Cliente,3=Ambos), isActive, userName.
Exemplo: {"select": ["id","businessName","personType"], "filter": "personType eq 2 and isActive eq true", "top": 100}.`,
		InputSchema: buildInputSchema[personsListArgs](),
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a personsListArgs) (*mcpsdk.CallToolResult, any, error) {
		svc := persons.New(c)
		q := a.Query()
		maxRows := applyDefaultMax(a.All, a.Max)
		if a.All {
			pageSize := q.Top
			q.Top = 0
			rows, err := svc.Paginate(ctx, q, pageSize, maxRows)
			if err != nil {
				return nil, nil, wrapAPIError(err)
			}
			return rowsResult(rows), nil, nil
		}
		raw, err := svc.List(ctx, q)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "persons_get",
		Description: "Busca uma pessoa/cliente/agente/empresa pelo id (Cod. Ref.).",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a personGetArgs) (*mcpsdk.CallToolResult, any, error) {
		raw, err := persons.New(c).Get(ctx, a.ID)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "persons_customfields_show",
		Description: "Lista os customFieldValues atuais de uma pessoa.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a personGetArgs) (*mcpsdk.CallToolResult, any, error) {
		cfs, err := persons.New(c).ListCustomFieldValues(ctx, a.ID)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return marshalResult(cfs)
	})
}

// ---------- services ----------

type intIDArg struct {
	ID int `json:"id" jsonschema:"ID numérico."`
}

type servicesListArgs struct {
	ODataArgs
}

func registerServices(s *mcpsdk.Server, c *movidesk.Client) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name: "services_list",
		Description: `Lista o catálogo de serviços do Movidesk. Campos: id, name, parentServiceId, isVisible, allowFinalUser.
Exemplo: {"select": ["id","name","isVisible"], "filter": "isVisible eq true", "top": 100}.`,
		InputSchema: buildInputSchema[servicesListArgs](),
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a servicesListArgs) (*mcpsdk.CallToolResult, any, error) {
		api := services.New(c)
		q := a.Query()
		maxRows := applyDefaultMax(a.All, a.Max)
		if a.All {
			pageSize := q.Top
			q.Top = 0
			rows, err := api.Paginate(ctx, q, pageSize, maxRows)
			if err != nil {
				return nil, nil, wrapAPIError(err)
			}
			return rowsResult(rows), nil, nil
		}
		raw, err := api.List(ctx, q)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "services_get",
		Description: "Busca um serviço do catálogo por id.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a intIDArg) (*mcpsdk.CallToolResult, any, error) {
		raw, err := services.New(c).Get(ctx, a.ID)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})
}

// ---------- contracts ----------

type contractsListArgs struct {
	ODataArgs
}

type contractsConsumptionListArgs struct {
	ODataArgs
}

func registerContracts(s *mcpsdk.Server, c *movidesk.Client) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name: "contracts_list",
		Description: `Lista contratos de tempo (timeAgreement). Campos: id, name, isActive, beginDate, endDate.
Exemplo: {"select": ["id","name","isActive"], "filter": "isActive eq true", "top": 100}.`,
		InputSchema: buildInputSchema[contractsListArgs](),
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a contractsListArgs) (*mcpsdk.CallToolResult, any, error) {
		api := contracts.New(c)
		q := a.Query()
		maxRows := applyDefaultMax(a.All, a.Max)
		if a.All {
			pageSize := q.Top
			q.Top = 0
			rows, err := api.Paginate(ctx, q, pageSize, maxRows)
			if err != nil {
				return nil, nil, wrapAPIError(err)
			}
			return rowsResult(rows), nil, nil
		}
		raw, err := api.List(ctx, q)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "contracts_get",
		Description: "Busca um contrato por id.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a intIDArg) (*mcpsdk.CallToolResult, any, error) {
		raw, err := contracts.New(c).Get(ctx, a.ID)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name: "contracts_consumption_list",
		Description: `Lista lançamentos de consumo (/timeAgreementConsumption) com filtros OData.
Exemplo: {"select": ["id","ticketId","timeSpent","date"], "filter": "date ge 2026-05-01T00:00:00Z", "top": 100}.`,
		InputSchema: buildInputSchema[contractsConsumptionListArgs](),
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a contractsConsumptionListArgs) (*mcpsdk.CallToolResult, any, error) {
		api := contracts.New(c)
		q := a.Query()
		maxRows := applyDefaultMax(a.All, a.Max)
		if a.All {
			pageSize := q.Top
			q.Top = 0
			rows, err := api.PaginateConsumption(ctx, q, pageSize, maxRows)
			if err != nil {
				return nil, nil, wrapAPIError(err)
			}
			return rowsResult(rows), nil, nil
		}
		raw, err := api.ListConsumption(ctx, q)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})
}

// ---------- knowledge base ----------

func registerKB(s *mcpsdk.Server, c *movidesk.Client) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "kb_article_get",
		Description: "Busca um artigo da base de conhecimento pelo id.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a intIDArg) (*mcpsdk.CallToolResult, any, error) {
		raw, err := knowledgebase.New(c).Get(ctx, a.ID)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})
}

// ---------- activities ----------

type activitiesListArgs struct {
	NameFilter string `json:"name_filter,omitempty" jsonschema:"Substring (case-insensitive) aplicada ao campo 'name' do tipo de atividade."`
	Max        int    `json:"max,omitempty" jsonschema:"Limite máximo de linhas. 0 aplica o padrão (500)."`
}

func registerActivities(s *mcpsdk.Server, c *movidesk.Client) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "activities_list",
		Description: "Lista os tipos de atividade configurados. Paginação por cursor; retorna JSON array consolidado.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a activitiesListArgs) (*mcpsdk.CallToolResult, any, error) {
		maxRows := a.Max
		if maxRows == 0 {
			maxRows = defaultMaxRows
		}
		rows, err := activities.New(c).ListAll(ctx, a.NameFilter, maxRows)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rowsResult(rows), nil, nil
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "activities_get",
		Description: "Busca um tipo de atividade pelo id.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a intIDArg) (*mcpsdk.CallToolResult, any, error) {
		raw, err := activities.New(c).Get(ctx, a.ID)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})
}

// ---------- surveys ----------

type surveysQuestionsListArgs struct {
	Type int `json:"type,omitempty" jsonschema:"Tipo de pesquisa (filtro typeFilter da API). 0 = todos."`
}

type surveysQuestionGetArgs struct {
	ID string `json:"id" jsonschema:"ID da pergunta de pesquisa."`
}

type surveysResponsesListArgs struct {
	Max int `json:"max,omitempty" jsonschema:"Limite máximo de linhas. 0 aplica o padrão (500)."`
}

func registerSurveys(s *mcpsdk.Server, c *movidesk.Client) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "surveys_questions_list",
		Description: "Lista perguntas das pesquisas de satisfação. Filtre por type quando aplicável.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a surveysQuestionsListArgs) (*mcpsdk.CallToolResult, any, error) {
		raw, err := surveys.New(c).ListQuestions(ctx, a.Type)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "surveys_questions_get",
		Description: "Busca uma pergunta de pesquisa pelo id.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a surveysQuestionGetArgs) (*mcpsdk.CallToolResult, any, error) {
		raw, err := surveys.New(c).GetQuestion(ctx, a.ID)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "surveys_responses_list",
		Description: "Lista respostas (cursor paginado, consolidado em JSON array). Capacidade limitada por 'max'.",
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a surveysResponsesListArgs) (*mcpsdk.CallToolResult, any, error) {
		maxRows := a.Max
		if maxRows == 0 {
			maxRows = defaultMaxRows
		}
		rows, err := surveys.New(c).ListAllResponses(ctx, maxRows)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rowsResult(rows), nil, nil
	})
}

// ---------- generic GET escape hatch ----------

type queryArgs struct {
	ODataArgs
	Path   string            `json:"path" jsonschema:"Caminho relativo do endpoint OData (deve iniciar com '/'). Ex.: /tickets, /services."`
	Params map[string]string `json:"params,omitempty" jsonschema:"Parâmetros adicionais de query string (token é injetado automaticamente)."`
}

func registerQuery(s *mcpsdk.Server, c *movidesk.Client) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name: "query",
		Description: `Escape hatch GET para qualquer endpoint OData do Movidesk não coberto por uma tool dedicada. Apenas leitura.
'path' deve ser um endpoint REAL da API Movidesk (ex.: /tickets, /persons, /services, /contracts, /activities, /articles).
Não invente sub-paths: /tickets/$count, /tickets/count e variantes NÃO existem — para contar use tickets_list com top:1.`,
		InputSchema: buildInputSchema[queryArgs](),
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, a queryArgs) (*mcpsdk.CallToolResult, any, error) {
		if !strings.HasPrefix(a.Path, "/") {
			return nil, nil, errors.New("path deve iniciar com '/'")
		}
		q := a.Query()
		extra := url.Values{}
		for k, v := range a.Params {
			extra.Set(k, v)
		}
		maxRows := applyDefaultMax(a.All, a.Max)
		if a.All {
			pageSize := q.Top
			q.Top = 0
			rows, err := paginateGeneric(ctx, c, a.Path, q, extra, pageSize, maxRows)
			if err != nil {
				return nil, nil, wrapAPIError(err)
			}
			return rowsResult(rows), nil, nil
		}
		raw, err := c.Get(ctx, a.Path, q, extra)
		if err != nil {
			return nil, nil, wrapAPIError(err)
		}
		return rawResult(raw), nil, nil
	})
}

// marshalResult serializes a typed Go value to JSON and wraps it as a tool
// result. Used by handlers whose underlying service method returns []T instead
// of raw bytes (timeline, custom fields, actions).
func marshalResult(v any) (*mcpsdk.CallToolResult, any, error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, nil, fmt.Errorf("serializar resultado: %w", err)
	}
	return rawResult(buf), nil, nil
}

// paginateGeneric drives /:path with the base OData query plus any extra
// query-string params, walking pages until exhaustion or the row cap. It is
// the engine behind the `query` tool's all=true mode.
func paginateGeneric(ctx context.Context, c *movidesk.Client, path string, base odata.Query, extra url.Values, pageSize, maxRows int) ([]json.RawMessage, error) {
	if pageSize <= 0 {
		pageSize = 100
	}
	out := []json.RawMessage{}
	q := base
	q.Top = pageSize
	for {
		body, err := c.Get(ctx, path, q, extra)
		if err != nil {
			return nil, err
		}
		var page []json.RawMessage
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		if len(page) == 0 {
			return out, nil
		}
		out = append(out, page...)
		if maxRows > 0 && len(out) >= maxRows {
			return out[:maxRows], nil
		}
		if len(page) < pageSize {
			return out, nil
		}
		q.Skip += pageSize
	}
}
