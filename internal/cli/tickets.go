package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/attachments"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/tickets"
)

// odataFlags is the shared flag set for any list command.
type odataFlags struct {
	filter  string
	selectF []string
	expand  []string
	orderBy []string
	top     int
	skip    int
	all     bool
	max     int
}

func (f *odataFlags) bind(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.filter, "filter", "", "expressão OData $filter")
	cmd.Flags().StringSliceVar(&f.selectF, "select", nil, "campos $select separados por vírgula")
	cmd.Flags().StringSliceVar(&f.expand, "expand", nil, "expressões $expand separadas por vírgula")
	cmd.Flags().StringSliceVar(&f.orderBy, "orderby", nil, "cláusulas $orderby separadas por vírgula (ex.: \"id desc\")")
	cmd.Flags().IntVar(&f.top, "top", 0, "$top: tamanho da página ou limite de uma única página")
	cmd.Flags().IntVar(&f.skip, "skip", 0, "$skip: offset no servidor")
	cmd.Flags().BoolVar(&f.all, "all", false, "busca todas as páginas (auto-paginação)")
	cmd.Flags().IntVar(&f.max, "max", 0, "com --all, interrompe após este número de registros")
}

func (f *odataFlags) query() odata.Query {
	return odata.Query{
		Filter:  f.filter,
		Select:  f.selectF,
		Expand:  f.expand,
		OrderBy: f.orderBy,
		Top:     f.top,
		Skip:    f.skip,
	}
}

// columnsFlag binds --columns once, owned by the parent command.
type columnsFlag struct{ cols []string }

func (c *columnsFlag) bind(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&c.cols, "columns", nil, "colunas separadas por vírgula para saída table/csv (suporta dot-paths)")
}

func newTicketsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tickets",
		Short: "Gerencia chamados (tickets) do Movidesk",
		Long: `Gerencia chamados do Movidesk via os endpoints /tickets, /tickets/past,
/tickets/merged e /tickets/htmldescription.

Os verbos list/get aceitam parâmetros de consulta OData; create/update aceitam
um corpo JSON via --file, um template salvo via --from-template, ou substituições
inline via --set chave=valor.`,
	}
	cmd.AddCommand(
		newTicketsListCmd(),
		newTicketsGetCmd(),
		newTicketsCreateCmd(),
		newTicketsUpdateCmd(),
		newTicketsHTMLCmd(),
		newTicketsPastCmd(),
		newTicketsMergedCmd(),
		newTicketsAttachCmd(),
		newTicketsActionsCmd(),
		newTicketsClientsCmd(),
		newTicketsRelationsCmd(),
		newTicketsTimelineCmd(),
		newTicketsAssetsCmd(),
		newTicketsHistoriesCmd(),
		newTicketsCustomFieldsCmd(),
	)
	return cmd
}

func newTicketsListCmd() *cobra.Command {
	var (
		of             odataFlags
		cf             columnsFlag
		includeDeleted bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista chamados (últimos 90 dias; mais antigos em `tickets past list`)",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			svc := tickets.New(r.client)
			q := of.query()

			out := cmd.OutOrStdout()
			if of.all {
				if q.Top == 0 {
					q.Top = 100
				}
				rows, err := svc.Paginate(cmd.Context(), q, includeDeleted, q.Top, of.max)
				if err != nil {
					return err
				}
				return renderRows(out, rows, r.output, "tickets", cf.cols)
			}
			body, err := svc.List(cmd.Context(), q, includeDeleted)
			if err != nil {
				return err
			}
			return renderJSON(out, body, r.output, "tickets", cf.cols)
		},
	}
	of.bind(cmd)
	cf.bind(cmd)
	cmd.Flags().BoolVar(&includeDeleted, "include-deleted", false, "inclui ações/clientes/relações excluídas")
	return cmd
}

func newTicketsGetCmd() *cobra.Command {
	var (
		protocol       string
		includeDeleted bool
		cf             columnsFlag
	)
	cmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Obtém um chamado por id (posicional) ou protocolo (--protocol)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			svc := tickets.New(r.client)

			var body []byte
			switch {
			case len(args) == 1:
				id, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("id inválido %q", args[0])
				}
				body, err = svc.GetByID(cmd.Context(), id, includeDeleted)
				if err != nil {
					return err
				}
			case protocol != "":
				body, err = svc.GetByProtocol(cmd.Context(), protocol, includeDeleted)
				if err != nil {
					return err
				}
			default:
				return errors.New("informe um id (posicional) ou --protocol")
			}
			return renderJSON(cmd.OutOrStdout(), body, r.output, "tickets", cf.cols)
		},
	}
	cmd.Flags().StringVar(&protocol, "protocol", "", "protocolo do chamado (ex.: MOVI202109000001)")
	cmd.Flags().BoolVar(&includeDeleted, "include-deleted", false, "inclui filhos excluídos")
	cf.bind(cmd)
	return cmd
}

func newTicketsCreateCmd() *cobra.Command {
	var (
		file                string
		template            string
		templateFile        string
		sets                []string
		returnAllProperties bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Cria um chamado a partir de corpo JSON, template ou substituições --set",
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := loadBody(file, template, templateFile, sets)
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return errors.New("nenhum campo informado; passe --file, --from-template[-file] ou --set chave=valor")
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			injectCreatedBy(body, r.userID)
			svc := tickets.New(r.client)
			raw, err := svc.Create(cmd.Context(), body, returnAllProperties)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "tickets", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "caminho do corpo JSON")
	cmd.Flags().StringVar(&template, "from-template", "", "carrega ~/.movidesk/templates/<nome>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "carrega template de um caminho específico")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "sobrescreve campos, ex.: --set type=2 --set subject=Olá")
	cmd.Flags().BoolVar(&returnAllProperties, "return-all", false, "pede ao Movidesk pra retornar o chamado completo")
	return cmd
}

func newTicketsUpdateCmd() *cobra.Command {
	var (
		file         string
		template     string
		templateFile string
		sets         []string
	)
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Aplica patch em um chamado por id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id inválido %q", args[0])
			}
			body, err := loadBody(file, template, templateFile, sets)
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return errors.New("nenhum campo para atualizar; passe --file, --from-template[-file] ou --set chave=valor")
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			svc := tickets.New(r.client)
			raw, err := svc.Update(cmd.Context(), id, body)
			if err != nil {
				return err
			}
			if len(raw) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "OK")
				return nil
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "tickets", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "caminho do corpo JSON de patch")
	cmd.Flags().StringVar(&template, "from-template", "", "carrega ~/.movidesk/templates/<nome>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "carrega template de um caminho específico")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "sobrescreve campos inline")
	return cmd
}

func newTicketsHTMLCmd() *cobra.Command {
	var actionID int
	cmd := &cobra.Command{
		Use:   "html <id>",
		Short: "Obtém o corpo HTML de um chamado (ou de uma de suas ações)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id inválido %q", args[0])
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			svc := tickets.New(r.client)
			raw, err := svc.HTMLDescription(cmd.Context(), id, actionID)
			if err != nil {
				return err
			}
			// HTML endpoint returns {id, description}. Show description directly when
			// stdout is a TTY (--output table), otherwise the JSON.
			if r.output == "table" {
				var v struct {
					ID          int    `json:"id"`
					Description string `json:"description"`
				}
				if err := json.Unmarshal(raw, &v); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), v.Description)
				return nil
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "", nil)
		},
	}
	cmd.Flags().IntVar(&actionID, "action-id", 0, "id de ação específica (padrão: descrição do chamado)")
	return cmd
}

func newTicketsPastCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "past",
		Short: "Gerencia chamados com mais de 90 dias (/tickets/past)",
	}
	cmd.AddCommand(newTicketsPastListCmd())
	return cmd
}

func newTicketsPastListCmd() *cobra.Command {
	var (
		of odataFlags
		cf columnsFlag
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista chamados arquivados (mais de 90 dias)",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			svc := tickets.New(r.client)
			q := of.query()
			out := cmd.OutOrStdout()
			if of.all {
				if q.Top == 0 {
					q.Top = 100
				}
				rows, err := svc.PaginatePast(cmd.Context(), q, q.Top, of.max)
				if err != nil {
					return err
				}
				return renderRows(out, rows, r.output, "tickets.past", cf.cols)
			}
			body, err := svc.Past(cmd.Context(), q)
			if err != nil {
				return err
			}
			return renderJSON(out, body, r.output, "tickets.past", cf.cols)
		},
	}
	of.bind(cmd)
	cf.bind(cmd)
	return cmd
}

func newTicketsMergedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merged",
		Short: "Inspeciona chamados mesclados (/tickets/merged)",
	}
	cmd.AddCommand(newTicketsMergedListCmd())
	return cmd
}

func newTicketsMergedListCmd() *cobra.Command {
	var (
		of odataFlags
		cf columnsFlag
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista chamados mesclados",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			svc := tickets.New(r.client)
			q := of.query()
			out := cmd.OutOrStdout()
			if of.all {
				if q.Top == 0 {
					q.Top = 100
				}
				rows, err := svc.PaginateMerged(cmd.Context(), q, q.Top, of.max)
				if err != nil {
					return err
				}
				return renderRows(out, rows, r.output, "tickets.merged", cf.cols)
			}
			body, err := svc.Merged(cmd.Context(), q)
			if err != nil {
				return err
			}
			return renderJSON(out, body, r.output, "tickets.merged", cf.cols)
		},
	}
	of.bind(cmd)
	cf.bind(cmd)
	return cmd
}

func newTicketsAttachCmd() *cobra.Command {
	var (
		actionID int
		file     string
		name     string
	)
	cmd := &cobra.Command{
		Use:   "attach <ticket-id>",
		Short: "Envia um arquivo para uma ação de chamado via /ticketFileUpload",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ticketID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id do chamado inválido %q", args[0])
			}
			if file == "" {
				return errors.New("--file é obrigatório")
			}
			if actionID <= 0 {
				return errors.New("--action-id é obrigatório (e deve ser > 0)")
			}

			fh, err := os.Open(file)
			if err != nil {
				return err
			}
			defer fh.Close()

			n := name
			if n == "" {
				n = filepath.Base(file)
			}

			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			svc := attachments.New(r.client)
			body, err := svc.Upload(cmd.Context(), ticketID, actionID, n, fh)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), body, r.output, "", nil)
		},
	}
	cmd.Flags().IntVar(&actionID, "action-id", 0, "id da ação onde anexar o arquivo (obrigatório)")
	cmd.Flags().StringVar(&file, "file", "", "caminho do arquivo local (obrigatório)")
	cmd.Flags().StringVar(&name, "name", "", "nome do arquivo a registrar no Movidesk (padrão: basename de --file)")
	_ = cmd.MarkFlagRequired("file")
	_ = cmd.MarkFlagRequired("action-id")
	return cmd
}
