package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/persons"
)

func newPersonsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "persons",
		Short: "Manage Movidesk persons (/persons)",
		Long: `Manage Movidesk persons. The /persons endpoint serves agents,
clients, companies, and departments — disambiguated by personType
(1=Pessoa, 2=Empresa, 4=Departamento) and profileType (1=Agente, 2=Cliente,
3=Both).

OData filters and projection apply on list. Custom field values follow the
same read-merge-patch semantics as tickets to avoid Movidesk's "delete
missing entries" trap.`,
	}
	cmd.AddCommand(
		newPersonsListCmd(),
		newPersonsGetCmd(),
		newPersonsCreateCmd(),
		newPersonsUpdateCmd(),
		newPersonsDeleteCmd(),
		newPersonsCustomFieldsCmd(),
	)
	return cmd
}

func newPersonsListCmd() *cobra.Command {
	var (
		of odataFlags
		cf columnsFlag
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List persons",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			svc := persons.New(r.client)
			q := of.query()
			out := cmd.OutOrStdout()
			if of.all {
				if q.Top == 0 {
					q.Top = 100
				}
				rows, err := svc.Paginate(cmd.Context(), q, q.Top, of.max)
				if err != nil {
					return err
				}
				return renderRows(out, rows, r.output, "persons", cf.cols)
			}
			body, err := svc.List(cmd.Context(), q)
			if err != nil {
				return err
			}
			return renderJSON(out, body, r.output, "persons", cf.cols)
		},
	}
	of.bind(cmd)
	cf.bind(cmd)
	return cmd
}

func newPersonsGetCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get one person by id (Cod. Ref.)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			body, err := persons.New(r.client).Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), body, r.output, "persons", cf.cols)
		},
	}
	cf.bind(cmd)
	return cmd
}

func newPersonsCreateCmd() *cobra.Command {
	var (
		file                string
		template            string
		templateFile        string
		sets                []string
		returnAllProperties bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a person from a JSON body, template, or --set overrides",
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
			raw, err := persons.New(r.client).Create(cmd.Context(), body, returnAllProperties)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "persons", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to JSON body")
	cmd.Flags().StringVar(&template, "from-template", "", "load ~/.movidesk/templates/<name>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "load template from a specific path")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "override fields, e.g. --set personType=1 --set businessName=\"Joe\"")
	cmd.Flags().BoolVar(&returnAllProperties, "return-all", false, "ask Movidesk to return the full person")
	return cmd
}

func newPersonsUpdateCmd() *cobra.Command {
	var (
		file         string
		template     string
		templateFile string
		sets         []string
	)
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Patch a person by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			raw, err := persons.New(r.client).Update(cmd.Context(), args[0], body)
			if err != nil {
				return err
			}
			if len(raw) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "OK")
				return nil
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "persons", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to JSON patch body")
	cmd.Flags().StringVar(&template, "from-template", "", "load ~/.movidesk/templates/<name>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "load template from a specific path")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "override fields inline")
	return cmd
}

func newPersonsDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Permanently delete a person (DELETE /persons?id=)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				if err := confirm(cmd, fmt.Sprintf("Delete person %q? This cannot be undone.", args[0])); err != nil {
					return err
				}
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			if _, err := persons.New(r.client).Delete(cmd.Context(), args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation prompt")
	return cmd
}

func newPersonsCustomFieldsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "customfields",
		Aliases: []string{"cf"},
		Short:   "Read and write person custom fields (read-merge-patch)",
	}
	cmd.AddCommand(
		newPersonsCFShowCmd(),
		newPersonsCFSetCmd(),
		newPersonsCFClearCmd(),
	)
	return cmd
}

func newPersonsCFShowCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "show <person-id>",
		Short: "List a person's customFieldValues",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			vs, err := persons.New(r.client).ListCustomFieldValues(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			cat, _ := loadCatalog(r.tenant.Name)
			rows := make([]map[string]any, 0, len(vs))
			for _, v := range vs {
				row := map[string]any{
					"customFieldId":     v.CustomFieldID,
					"customFieldRuleId": v.CustomFieldRuleID,
					"line":              v.Line,
					"value":             v.Value,
					"items":             summarizePersonItems(v.Items),
				}
				if cat != nil {
					row["label"] = cat.labelForID(v.CustomFieldID)
				}
				rows = append(rows, row)
			}
			return renderRows(cmd.OutOrStdout(), rows, r.output, "tickets.customfields", cf.cols)
		},
	}
	cf.bind(cmd)
	return cmd
}

func newPersonsCFSetCmd() *cobra.Command {
	var (
		fieldID     int
		fieldLabel  string
		ruleID      int
		line        int
		value       string
		items       []string
		itemPersons []string
		itemClients []string
		itemTeams   []string
	)
	cmd := &cobra.Command{
		Use:   "set <person-id>",
		Short: "Set a person custom field value (read-merge-patch)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			cat, _ := loadCatalog(r.tenant.Name)
			entry, err := resolveField(cat, fieldID, fieldLabel, ruleID)
			if err != nil {
				return err
			}
			ln := line
			if ln == 0 {
				ln = 1
			}
			cfv := persons.CustomFieldValue{
				CustomFieldID:     entry.id,
				CustomFieldRuleID: entry.ruleID,
				Line:              ln,
				Value:             value,
			}
			for _, it := range items {
				cfv.Items = append(cfv.Items, persons.CustomFieldItem{CustomFieldItem: it})
			}
			for _, p := range itemPersons {
				cfv.Items = append(cfv.Items, persons.CustomFieldItem{PersonID: p})
			}
			for _, c := range itemClients {
				cfv.Items = append(cfv.Items, persons.CustomFieldItem{ClientID: c})
			}
			for _, tm := range itemTeams {
				cfv.Items = append(cfv.Items, persons.CustomFieldItem{Team: tm})
			}
			if cfv.Value == "" && len(cfv.Items) == 0 {
				return errors.New("provide --value, --item, --item-person, --item-client or --item-team")
			}
			raw, err := persons.New(r.client).SetCustomFieldValue(cmd.Context(), args[0], cfv)
			if err != nil {
				return err
			}
			if len(raw) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "OK")
				return nil
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "persons", nil)
		},
	}
	cmd.Flags().IntVar(&fieldID, "field", 0, "numeric custom field id (or use --field-label)")
	cmd.Flags().StringVar(&fieldLabel, "field-label", "", "label registered in the catalog")
	cmd.Flags().IntVar(&ruleID, "rule", 0, "rule id (taken from catalog if omitted)")
	cmd.Flags().IntVar(&line, "line", 0, "row number (default 1)")
	cmd.Flags().StringVar(&value, "value", "", "value for text/numeric/date types")
	cmd.Flags().StringSliceVar(&items, "item", nil, "list-of-values item label (repeatable)")
	cmd.Flags().StringSliceVar(&itemPersons, "item-person", nil, "person id (repeatable)")
	cmd.Flags().StringSliceVar(&itemClients, "item-client", nil, "client id (repeatable)")
	cmd.Flags().StringSliceVar(&itemTeams, "item-team", nil, "team name (repeatable)")
	return cmd
}

func newPersonsCFClearCmd() *cobra.Command {
	var (
		fieldID    int
		fieldLabel string
		ruleID     int
		line       int
	)
	cmd := &cobra.Command{
		Use:   "clear <person-id>",
		Short: "Remove a person custom field value (read-merge-patch)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			cat, _ := loadCatalog(r.tenant.Name)
			entry, err := resolveField(cat, fieldID, fieldLabel, ruleID)
			if err != nil {
				return err
			}
			raw, err := persons.New(r.client).ClearCustomFieldValue(cmd.Context(), args[0], entry.id, entry.ruleID, line)
			if err != nil {
				return err
			}
			if len(raw) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "OK")
				return nil
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "persons", nil)
		},
	}
	cmd.Flags().IntVar(&fieldID, "field", 0, "numeric custom field id")
	cmd.Flags().StringVar(&fieldLabel, "field-label", "", "label from the catalog")
	cmd.Flags().IntVar(&ruleID, "rule", 0, "rule id (omit with catalog)")
	cmd.Flags().IntVar(&line, "line", 0, "specific line; omit to clear every line")
	return cmd
}

// summarizePersonItems mirrors summarizeItems from tickets but works on the
// persons-package type.
func summarizePersonItems(items []persons.CustomFieldItem) string {
	if len(items) == 0 {
		return ""
	}
	parts := make([]string, 0, len(items))
	for _, it := range items {
		switch {
		case it.CustomFieldItem != "":
			parts = append(parts, it.CustomFieldItem)
		case it.PersonID != "":
			parts = append(parts, "person:"+it.PersonID)
		case it.ClientID != "":
			parts = append(parts, "client:"+it.ClientID)
		case it.Team != "":
			parts = append(parts, "team:"+it.Team)
		}
	}
	return joinShort(parts, ", ", 60)
}
