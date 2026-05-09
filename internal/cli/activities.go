package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/activities"
)

func newActivitiesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activities",
		Short: "Manage Movidesk activities (/activity)",
		Long: `Activities use cursor-based pagination (limit/startingAfter) — not OData.
Use --name to substring-filter, --all to walk every page.`,
	}
	cmd.AddCommand(
		newActivitiesListCmd(),
		newActivitiesGetCmd(),
		newActivitiesCreateCmd(),
		newActivitiesUpdateCmd(),
		newActivitiesDeleteCmd(),
		newActivitiesAddTeamsCmd(),
	)
	return cmd
}

func newActivitiesListCmd() *cobra.Command {
	var (
		nameFilter    string
		limit         int
		startingAfter string
		all           bool
		max           int
		cf            columnsFlag
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List activities (cursor pagination)",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			svc := activities.New(r.client)
			if all {
				rows, err := svc.ListAll(cmd.Context(), nameFilter, max)
				if err != nil {
					return err
				}
				return renderRows(cmd.OutOrStdout(), rows, r.output, "activities", cf.cols)
			}
			page, err := svc.ListPage(cmd.Context(), limit, startingAfter, nameFilter)
			if err != nil {
				return err
			}
			return renderRows(cmd.OutOrStdout(), page, r.output, "activities", cf.cols)
		},
	}
	cmd.Flags().StringVar(&nameFilter, "name", "", "filter by substring on activity name")
	cmd.Flags().IntVar(&limit, "limit", 0, "page size (1..100, default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "cursor (last id of previous page)")
	cmd.Flags().BoolVar(&all, "all", false, "walk every page")
	cmd.Flags().IntVar(&max, "max", 0, "with --all, stop after this many records")
	cf.bind(cmd)
	return cmd
}

func newActivitiesGetCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get one activity by id",
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
			body, err := activities.New(r.client).Get(cmd.Context(), id)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), body, r.output, "activities", cf.cols)
		},
	}
	cf.bind(cmd)
	return cmd
}

func newActivitiesCreateCmd() *cobra.Command {
	var (
		file         string
		template     string
		templateFile string
		sets         []string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an activity",
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
			raw, err := activities.New(r.client).Create(cmd.Context(), body)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "activities", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to JSON body")
	cmd.Flags().StringVar(&template, "from-template", "", "load ~/.movidesk/templates/<name>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "load template from a path")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "override fields, e.g. --set name=\"Atividade\" --set isActive=true")
	return cmd
}

func newActivitiesUpdateCmd() *cobra.Command {
	var (
		file         string
		template     string
		templateFile string
		sets         []string
	)
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Patch an activity by id",
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
			raw, err := activities.New(r.client).Update(cmd.Context(), id, body)
			if err != nil {
				return err
			}
			if len(raw) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "OK")
				return nil
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "activities", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to JSON patch body")
	cmd.Flags().StringVar(&template, "from-template", "", "load ~/.movidesk/templates/<name>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "load template from a specific path")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "override fields inline, e.g. --set name=Foo")
	return cmd
}

func newActivitiesDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an activity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q", args[0])
			}
			if !force {
				if err := confirm(cmd, fmt.Sprintf("Delete activity %d? This cannot be undone.", id)); err != nil {
					return err
				}
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			if _, err := activities.New(r.client).Delete(cmd.Context(), id); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation prompt")
	return cmd
}

func newActivitiesAddTeamsCmd() *cobra.Command {
	var teams []string
	cmd := &cobra.Command{
		Use:   "add-teams <activity-id>",
		Short: "Append teams to an activity (POST /addTeamsToActivity)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid activity id %q", args[0])
			}
			if len(teams) == 0 {
				return errors.New("--team is required (repeatable)")
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := activities.New(r.client).AddTeams(cmd.Context(), id, teams)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "", nil)
		},
	}
	cmd.Flags().StringSliceVar(&teams, "team", nil, "team name to append (repeatable)")
	return cmd
}
