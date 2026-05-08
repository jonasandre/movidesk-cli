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
		Short: "Inspect and modify ticket actions",
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
		Short: "List actions of a ticket",
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
		Short: "Get one action by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ticket id %q", args[0])
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
	cmd.Flags().IntVar(&actionID, "action-id", 0, "action id (required)")
	_ = cmd.MarkFlagRequired("action-id")
	return cmd
}

func newTicketsActionsAddCmd() *cobra.Command {
	var (
		atype           int
		descr           string
		descrFile       string
		public          bool
		internal        bool
		tags            []string
		justification   string
		statusOverride  string
	)
	cmd := &cobra.Command{
		Use:   "add <ticket-id>",
		Short: "Append a new action to a ticket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ticket id %q", args[0])
			}
			if public && internal {
				return errors.New("--public and --internal are mutually exclusive")
			}
			if internal {
				atype = 1
			} else if public {
				atype = 2
			}
			if atype != 1 && atype != 2 {
				return errors.New("specify --type 1 (internal) or --type 2 (public), or use --internal/--public")
			}
			body := descr
			if descrFile != "" {
				if descr != "" {
					return errors.New("--description and --description-file are mutually exclusive")
				}
				raw, err := os.ReadFile(descrFile)
				if err != nil {
					return err
				}
				body = string(raw)
			}
			if body == "" {
				return errors.New("--description (or --description-file) is required")
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
	cmd.Flags().IntVar(&atype, "type", 0, "action type: 1=internal, 2=public")
	cmd.Flags().StringVar(&descr, "description", "", "action body (HTML on writes)")
	cmd.Flags().StringVar(&descrFile, "description-file", "", "read body from a file")
	cmd.Flags().BoolVar(&public, "public", false, "alias for --type 2")
	cmd.Flags().BoolVar(&internal, "internal", false, "alias for --type 1")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "tag (repeatable)")
	cmd.Flags().StringVar(&justification, "justification", "", "status justification")
	cmd.Flags().StringVar(&statusOverride, "status", "", "transition the ticket to a status")
	return cmd
}

func newTicketsActionsUpdateCmd() *cobra.Command {
	var (
		actionID int
		descr    string
	)
	cmd := &cobra.Command{
		Use:   "update <ticket-id>",
		Short: "Edit an existing action by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ticket id %q", args[0])
			}
			if descr == "" {
				return errors.New("--description is required")
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
	cmd.Flags().IntVar(&actionID, "action-id", 0, "action id (required)")
	cmd.Flags().StringVar(&descr, "description", "", "new description")
	_ = cmd.MarkFlagRequired("action-id")
	return cmd
}

func newTicketsActionsDeleteCmd() *cobra.Command {
	var actionID int
	cmd := &cobra.Command{
		Use:   "delete <ticket-id>",
		Short: "Soft-delete an action (sets isDeleted: true)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ticket id %q", args[0])
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
	cmd.Flags().IntVar(&actionID, "action-id", 0, "action id (required)")
	_ = cmd.MarkFlagRequired("action-id")
	return cmd
}

func newTicketsClientsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "clients", Short: "Inspect ticket clients"}
	cmd.AddCommand(newTicketsClientsListCmd())
	return cmd
}

func newTicketsClientsListCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "list <ticket-id>",
		Short: "List clients of a ticket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ticket id %q", args[0])
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
		Short: "List parent and child tickets",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ticket id %q", args[0])
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
			fmt.Fprintln(out, "Parents:")
			if err := renderRows(out, parents, r.output, "tickets.relations", nil); err != nil {
				return err
			}
			fmt.Fprintln(out, "Children:")
			return renderRows(out, children, r.output, "tickets.relations", nil)
		},
	}
	return cmd
}

func newTicketsTimelineCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "timeline <ticket-id>",
		Short: "Chronological merge of actions, status and owner changes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ticket id %q", args[0])
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
	cmd := &cobra.Command{Use: "assets", Short: "Inspect ticket assets"}
	cmd.AddCommand(newTicketsAssetsListCmd())
	return cmd
}

func newTicketsAssetsListCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "list <ticket-id>",
		Short: "List assets attached to a ticket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ticket id %q", args[0])
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
	cmd := &cobra.Command{Use: "histories", Short: "Owner and status histories"}
	cmd.AddCommand(newTicketsHistoriesListCmd())
	return cmd
}

func newTicketsHistoriesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <ticket-id>",
		Short: "List owner and status histories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ticket id %q", args[0])
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
			fmt.Fprintln(out, "Owner history:")
			if err := renderRows(out, owners, r.output, "tickets.histories", nil); err != nil {
				return err
			}
			fmt.Fprintln(out, "Status history:")
			return renderRows(out, statuses, r.output, "tickets.histories", nil)
		},
	}
	return cmd
}
