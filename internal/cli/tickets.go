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
	cmd.Flags().StringVar(&f.filter, "filter", "", "OData $filter expression")
	cmd.Flags().StringSliceVar(&f.selectF, "select", nil, "comma-separated $select fields")
	cmd.Flags().StringSliceVar(&f.expand, "expand", nil, "comma-separated $expand expressions")
	cmd.Flags().StringSliceVar(&f.orderBy, "orderby", nil, "comma-separated $orderby clauses (e.g. \"id desc\")")
	cmd.Flags().IntVar(&f.top, "top", 0, "$top: page size or single-page limit")
	cmd.Flags().IntVar(&f.skip, "skip", 0, "$skip: server-side offset")
	cmd.Flags().BoolVar(&f.all, "all", false, "fetch every page (auto-paginate)")
	cmd.Flags().IntVar(&f.max, "max", 0, "with --all, stop after this many records")
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
	cmd.Flags().StringSliceVar(&c.cols, "columns", nil, "comma-separated columns for table/csv output (dot-paths supported)")
}

func newTicketsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tickets",
		Short: "Manage Movidesk tickets",
		Long: `Manage Movidesk tickets via the /tickets, /tickets/past, /tickets/merged
and /tickets/htmldescription endpoints.

The list/get verbs accept OData query parameters; create/update accept a JSON
body via --file, a saved template via --from-template, or inline overrides
via --set key=value.`,
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
		Short: "List tickets (last 90 days; older tickets in `tickets past list`)",
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
	cmd.Flags().BoolVar(&includeDeleted, "include-deleted", false, "include deleted actions/clients/parents/children")
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
		Short: "Get one ticket by id (positional) or protocol (--protocol)",
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
					return fmt.Errorf("invalid id %q", args[0])
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
				return errors.New("provide either an id (positional) or --protocol")
			}
			return renderJSON(cmd.OutOrStdout(), body, r.output, "tickets", cf.cols)
		},
	}
	cmd.Flags().StringVar(&protocol, "protocol", "", "ticket protocol (e.g. MOVI202109000001)")
	cmd.Flags().BoolVar(&includeDeleted, "include-deleted", false, "include deleted children")
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
		Short: "Create a ticket from a JSON body, template, or --set overrides",
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := loadBody(file, template, templateFile, sets)
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return errors.New("no body fields supplied; pass --file, --from-template[-file], or --set key=value")
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			svc := tickets.New(r.client)
			raw, err := svc.Create(cmd.Context(), body, returnAllProperties)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "tickets", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to JSON body")
	cmd.Flags().StringVar(&template, "from-template", "", "load ~/.movidesk/templates/<name>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "load template from a specific path")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "override fields, e.g. --set type=2 --set subject=Hello")
	cmd.Flags().BoolVar(&returnAllProperties, "return-all", false, "ask Movidesk to return the full ticket")
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
		Short: "Patch a ticket by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q", args[0])
			}
			body, err := loadBody(file, template, templateFile, sets)
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return errors.New("no fields to update; pass --file, --from-template[-file], or --set key=value")
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
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to JSON patch body")
	cmd.Flags().StringVar(&template, "from-template", "", "load ~/.movidesk/templates/<name>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "load template from a specific path")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "override fields inline")
	return cmd
}

func newTicketsHTMLCmd() *cobra.Command {
	var actionID int
	cmd := &cobra.Command{
		Use:   "html <id>",
		Short: "Get the HTML body of a ticket (or one of its actions)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q", args[0])
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
	cmd.Flags().IntVar(&actionID, "action-id", 0, "specific action id (default: ticket description)")
	return cmd
}

func newTicketsPastCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "past",
		Short: "Manage tickets older than 90 days (/tickets/past)",
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
		Short: "List archived tickets (older than 90 days)",
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
		Short: "Inspect merged tickets (/tickets/merged)",
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
		Short: "List merged tickets",
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
		Short: "Upload a file to a ticket action via /ticketFileUpload",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ticketID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ticket id %q", args[0])
			}
			if file == "" {
				return errors.New("--file is required")
			}
			if actionID <= 0 {
				return errors.New("--action-id is required (and must be > 0)")
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
	cmd.Flags().IntVar(&actionID, "action-id", 0, "action id to attach the file to (required)")
	cmd.Flags().StringVar(&file, "file", "", "path to the local file (required)")
	cmd.Flags().StringVar(&name, "name", "", "filename to record on Movidesk (default: basename of --file)")
	_ = cmd.MarkFlagRequired("file")
	_ = cmd.MarkFlagRequired("action-id")
	return cmd
}
