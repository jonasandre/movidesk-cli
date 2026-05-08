package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/contracts"
)

func newContractsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contracts",
		Short: "Manage Movidesk hour contracts (/timeAgreement)",
	}
	cmd.AddCommand(
		newContractsListCmd(),
		newContractsGetCmd(),
		newContractsCreateCmd(),
		newContractsUpdateCmd(),
		newContractsDeleteCmd(),
		newContractsConsumptionCmd(),
	)
	return cmd
}

func newContractsListCmd() *cobra.Command {
	var (
		of odataFlags
		cf columnsFlag
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time agreements",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			a := contracts.New(r.client)
			q := of.query()
			out := cmd.OutOrStdout()
			if of.all {
				if q.Top == 0 {
					q.Top = 100
				}
				rows, err := a.Paginate(cmd.Context(), q, q.Top, of.max)
				if err != nil {
					return err
				}
				return renderRows(out, rows, r.output, "contracts", cf.cols)
			}
			body, err := a.List(cmd.Context(), q)
			if err != nil {
				return err
			}
			return renderJSON(out, body, r.output, "contracts", cf.cols)
		},
	}
	of.bind(cmd)
	cf.bind(cmd)
	return cmd
}

func newContractsGetCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get one contract by id",
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
			body, err := contracts.New(r.client).Get(cmd.Context(), id)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), body, r.output, "contracts", cf.cols)
		},
	}
	cf.bind(cmd)
	return cmd
}

func newContractsCreateCmd() *cobra.Command {
	var (
		file                string
		template            string
		templateFile        string
		sets                []string
		returnAllProperties bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a contract",
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := loadBody(file, template, templateFile, sets)
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return errors.New("no body fields supplied")
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := contracts.New(r.client).Create(cmd.Context(), body, returnAllProperties)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "contracts", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to JSON body")
	cmd.Flags().StringVar(&template, "from-template", "", "")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "")
	cmd.Flags().BoolVar(&returnAllProperties, "return-all", false, "ask Movidesk to return the full contract")
	return cmd
}

func newContractsUpdateCmd() *cobra.Command {
	var (
		file         string
		template     string
		templateFile string
		sets         []string
	)
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Patch a contract by id",
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
				return errors.New("no fields to update")
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := contracts.New(r.client).Update(cmd.Context(), id, body)
			if err != nil {
				return err
			}
			if len(raw) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "OK")
				return nil
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "contracts", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "")
	cmd.Flags().StringVar(&template, "from-template", "", "")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "")
	return cmd
}

func newContractsDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q", args[0])
			}
			if !force {
				if err := confirm(cmd, fmt.Sprintf("Delete contract %d? This cannot be undone.", id)); err != nil {
					return err
				}
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			if _, err := contracts.New(r.client).Delete(cmd.Context(), id); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation prompt")
	return cmd
}

func newContractsConsumptionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consumption",
		Short: "Read /timeAgreementConsumption",
		Long: `When filtering by startPeriod/endPeriod, Movidesk requires the contract
name in the OData $filter (e.g. --filter "name eq 'Default' and period gt
2026-01-01T00:00:00Z"). Avoid combining $select with period filters.`,
	}
	cmd.AddCommand(newContractsConsumptionListCmd())
	return cmd
}

func newContractsConsumptionListCmd() *cobra.Command {
	var (
		of odataFlags
		cf columnsFlag
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List consumption rows",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			a := contracts.New(r.client)
			q := of.query()
			out := cmd.OutOrStdout()
			if of.all {
				if q.Top == 0 {
					q.Top = 100
				}
				rows, err := a.PaginateConsumption(cmd.Context(), q, q.Top, of.max)
				if err != nil {
					return err
				}
				return renderRows(out, rows, r.output, "contracts.consumption", cf.cols)
			}
			body, err := a.ListConsumption(cmd.Context(), q)
			if err != nil {
				return err
			}
			return renderJSON(out, body, r.output, "contracts.consumption", cf.cols)
		},
	}
	of.bind(cmd)
	cf.bind(cmd)
	return cmd
}
