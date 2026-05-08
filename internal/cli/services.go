package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/services"
)

func newServicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Manage Movidesk service catalog (/services)",
		Long: `Manage entries from the Movidesk service catalog.

Note: PATCH replaces array-valued fields like "categories" — when updating,
send the complete list you want to keep, not just the additions.`,
	}
	cmd.AddCommand(
		newServicesListCmd(),
		newServicesGetCmd(),
		newServicesCreateCmd(),
		newServicesUpdateCmd(),
		newServicesDeleteCmd(),
	)
	return cmd
}

func newServicesListCmd() *cobra.Command {
	var (
		of odataFlags
		cf columnsFlag
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List services",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			api := services.New(r.client)
			q := of.query()
			out := cmd.OutOrStdout()
			if of.all {
				if q.Top == 0 {
					q.Top = 100
				}
				rows, err := api.Paginate(cmd.Context(), q, q.Top, of.max)
				if err != nil {
					return err
				}
				return renderRows(out, rows, r.output, "services", cf.cols)
			}
			body, err := api.List(cmd.Context(), q)
			if err != nil {
				return err
			}
			return renderJSON(out, body, r.output, "services", cf.cols)
		},
	}
	of.bind(cmd)
	cf.bind(cmd)
	return cmd
}

func newServicesGetCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get one service by id",
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
			body, err := services.New(r.client).Get(cmd.Context(), id)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), body, r.output, "services", cf.cols)
		},
	}
	cf.bind(cmd)
	return cmd
}

func newServicesCreateCmd() *cobra.Command {
	var (
		file                string
		template            string
		templateFile        string
		sets                []string
		returnAllProperties bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a service from a JSON body, template, or --set overrides",
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
			raw, err := services.New(r.client).Create(cmd.Context(), body, returnAllProperties)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "services", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to JSON body")
	cmd.Flags().StringVar(&template, "from-template", "", "load ~/.movidesk/templates/<name>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "load template from a specific path")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "override fields, e.g. --set name=\"Suporte\" --set isActive=true")
	cmd.Flags().BoolVar(&returnAllProperties, "return-all", false, "ask Movidesk to return the full service")
	return cmd
}

func newServicesUpdateCmd() *cobra.Command {
	var (
		file         string
		template     string
		templateFile string
		sets         []string
	)
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Patch a service by id",
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
			raw, err := services.New(r.client).Update(cmd.Context(), id, body)
			if err != nil {
				return err
			}
			if len(raw) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "OK")
				return nil
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "services", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to JSON patch body")
	cmd.Flags().StringVar(&template, "from-template", "", "load ~/.movidesk/templates/<name>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "load template from a specific path")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "override fields inline")
	return cmd
}

func newServicesDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Permanently delete a service (DELETE /services?id=)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q", args[0])
			}
			if !force {
				if err := confirm(cmd, fmt.Sprintf("Delete service %d? This cannot be undone.", id)); err != nil {
					return err
				}
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			if _, err := services.New(r.client).Delete(cmd.Context(), id); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation prompt")
	return cmd
}
