package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/tickets"
)

func newTicketsActionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "actions",
		Short: "Inspeciona e modifica ações de chamados",
	}
	cmd.AddCommand(
		newTicketsActionsListCmd(),
		newTicketsActionsGetCmd(),
		newTicketsActionsAddCmd(),
		newTicketsActionsUpdateCmd(),
		newTicketsActionsDeleteCmd(),
	)
	return cmd
}

func newTicketsActionsListCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "list <ticket-id>",
		Short: "Lista as ações de um chamado",
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
			actions, err := tickets.New(r.client).ListActions(cmd.Context(), id)
			if err != nil {
				return err
			}
			return renderRows(cmd.OutOrStdout(), actions, r.output, "tickets.actions", cf.cols)
		},
	}
	cf.bind(cmd)
	return cmd
}

func newTicketsActionsGetCmd() *cobra.Command {
	var actionID int
	cmd := &cobra.Command{
		Use:   "get <ticket-id> --action-id N",
		Short: "Obtém uma ação pelo id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id do chamado inválido %q", args[0])
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			a, err := tickets.New(r.client).GetAction(cmd.Context(), id, actionID)
			if err != nil {
				return err
			}
			return renderRows(cmd.OutOrStdout(), a, r.output, "", nil)
		},
	}
	cmd.Flags().IntVar(&actionID, "action-id", 0, "id da ação (obrigatório)")
	_ = cmd.MarkFlagRequired("action-id")
	return cmd
}

func newTicketsActionsAddCmd() *cobra.Command {
	var (
		atype          int
		descr          string
		descrFile      string
		public         bool
		internal       bool
		tags           []string
		justification  string
		statusOverride string
	)
	cmd := &cobra.Command{
		Use:   "add <ticket-id>",
		Short: "Adiciona uma nova ação a um chamado",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id do chamado inválido %q", args[0])
			}
			if public && internal {
				return errors.New("--public e --internal são mutuamente exclusivos")
			}
			if internal {
				atype = 1
			} else if public {
				atype = 2
			}
			if atype != 1 && atype != 2 {
				return errors.New("informe --type 1 (interno) ou --type 2 (público), ou use --internal/--public")
			}
			body := descr
			if descrFile != "" {
				if descr != "" {
					return errors.New("--description e --description-file são mutuamente exclusivos")
				}
				raw, err := os.ReadFile(descrFile)
				if err != nil {
					return err
				}
				body = string(raw)
			}
			if body == "" {
				return errors.New("--description (ou --description-file) é obrigatório")
			}
			a := tickets.Action{
				Type:          atype,
				Description:   body,
				Tags:          tags,
				Justification: justification,
				Status:        statusOverride,
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			if r.userID != "" && a.CreatedBy == nil {
				a.CreatedBy = &tickets.Person{ID: r.userID}
			}
			raw, err := tickets.New(r.client).AddAction(cmd.Context(), id, a)
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
	cmd.Flags().IntVar(&atype, "type", 0, "tipo da ação: 1=interna, 2=pública")
	cmd.Flags().StringVar(&descr, "description", "", "corpo da ação (HTML em escritas)")
	cmd.Flags().StringVar(&descrFile, "description-file", "", "lê o corpo de um arquivo")
	cmd.Flags().BoolVar(&public, "public", false, "alias para --type 2")
	cmd.Flags().BoolVar(&internal, "internal", false, "alias para --type 1")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "tag (repetível)")
	cmd.Flags().StringVar(&justification, "justification", "", "justificativa do status")
	cmd.Flags().StringVar(&statusOverride, "status", "", "transiciona o chamado para um status")
	return cmd
}

func newTicketsActionsUpdateCmd() *cobra.Command {
	var (
		actionID int
		descr    string
	)
	cmd := &cobra.Command{
		Use:   "update <ticket-id>",
		Short: "Edita uma ação existente pelo id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id do chamado inválido %q", args[0])
			}
			if descr == "" {
				return errors.New("--description é obrigatório")
			}
			a := tickets.Action{ID: actionID, Description: descr}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := tickets.New(r.client).UpdateAction(cmd.Context(), id, a)
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
	cmd.Flags().IntVar(&actionID, "action-id", 0, "id da ação (obrigatório)")
	cmd.Flags().StringVar(&descr, "description", "", "nova descrição")
	_ = cmd.MarkFlagRequired("action-id")
	return cmd
}

func newTicketsActionsDeleteCmd() *cobra.Command {
	var actionID int
	cmd := &cobra.Command{
		Use:   "delete <ticket-id>",
		Short: "Marca uma ação como excluída (soft delete: isDeleted: true)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id do chamado inválido %q", args[0])
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := tickets.New(r.client).DeleteAction(cmd.Context(), id, actionID)
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
	cmd.Flags().IntVar(&actionID, "action-id", 0, "id da ação (obrigatório)")
	_ = cmd.MarkFlagRequired("action-id")
	return cmd
}

func newTicketsClientsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "clients", Short: "Inspeciona clientes de um chamado"}
	cmd.AddCommand(newTicketsClientsListCmd())
	return cmd
}

func newTicketsClientsListCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "list <ticket-id>",
		Short: "Lista os clientes de um chamado",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id do chamado inválido %q", args[0])
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			cs, err := tickets.New(r.client).ListClients(cmd.Context(), id)
			if err != nil {
				return err
			}
			return renderRows(cmd.OutOrStdout(), cs, r.output, "tickets.clients", cf.cols)
		},
	}
	cf.bind(cmd)
	return cmd
}

func newTicketsRelationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relations <ticket-id>",
		Short: "Lista chamados pais e filhos",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id do chamado inválido %q", args[0])
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			parents, children, err := tickets.New(r.client).Relations(cmd.Context(), id)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Pais:")
			if err := renderRows(out, parents, r.output, "tickets.relations", nil); err != nil {
				return err
			}
			fmt.Fprintln(out, "Filhos:")
			return renderRows(out, children, r.output, "tickets.relations", nil)
		},
	}
	return cmd
}

func newTicketsTimelineCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "timeline <ticket-id>",
		Short: "Mescla cronológica de ações, mudanças de status e responsável",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id do chamado inválido %q", args[0])
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			entries, err := tickets.New(r.client).Timeline(cmd.Context(), id)
			if err != nil {
				return err
			}
			return renderRows(cmd.OutOrStdout(), entries, r.output, "tickets.timeline", cf.cols)
		},
	}
	cf.bind(cmd)
	return cmd
}

func newTicketsAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "assets", Short: "Inspeciona ativos vinculados ao chamado"}
	cmd.AddCommand(newTicketsAssetsListCmd())
	return cmd
}

func newTicketsAssetsListCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "list <ticket-id>",
		Short: "Lista ativos vinculados ao chamado",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id do chamado inválido %q", args[0])
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			assets, err := tickets.New(r.client).ListAssets(cmd.Context(), id)
			if err != nil {
				return err
			}
			return renderRows(cmd.OutOrStdout(), assets, r.output, "tickets.assets", cf.cols)
		},
	}
	cf.bind(cmd)
	return cmd
}

func newTicketsHistoriesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "histories", Short: "Históricos de responsável e status"}
	cmd.AddCommand(newTicketsHistoriesListCmd())
	return cmd
}

func newTicketsHistoriesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <ticket-id>",
		Short: "Lista o histórico de responsável e de status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id do chamado inválido %q", args[0])
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			owners, statuses, err := tickets.New(r.client).Histories(cmd.Context(), id)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Histórico de responsável:")
			if err := renderRows(out, owners, r.output, "tickets.histories", nil); err != nil {
				return err
			}
			fmt.Fprintln(out, "Histórico de status:")
			return renderRows(out, statuses, r.output, "tickets.histories", nil)
		},
	}
	return cmd
}
