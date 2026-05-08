package cli

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/tickets"
)

func newTicketsCustomFieldsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "customfields",
		Aliases: []string{"cf"},
		Short:   "Read and write ticket custom fields (with read-merge-patch safety)",
		Long: `Movidesk's PATCH /tickets deletes any customFieldValues entry not
present in the body. This subcommand uses read-merge-patch internally so
you only describe the change you want, never the whole list.

A local catalog at ~/.movidesk/<tenant>/customfields.yaml maps human-friendly
labels to numeric field IDs and types so you can use --field-label "Severidade"
instead of --field 125529.`,
	}
	cmd.AddCommand(
		newTicketsCFShowCmd(),
		newTicketsCFSetCmd(),
		newTicketsCFClearCmd(),
		newTicketsCFCatalogCmd(),
	)
	return cmd
}

func newTicketsCFShowCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "show <ticket-id>",
		Short: "List a ticket's customFieldValues",
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
			vs, err := tickets.New(r.client).ListCustomFieldValues(cmd.Context(), id)
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
					"items":             summarizeItems(v.Items),
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

func newTicketsCFSetCmd() *cobra.Command {
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
		Use:   "set <ticket-id>",
		Short: "Set a custom field value (read-merge-patch)",
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
			cat, _ := loadCatalog(r.tenant.Name)
			entry, err := resolveField(cat, fieldID, fieldLabel, ruleID)
			if err != nil {
				return err
			}
			ln := line
			if ln == 0 {
				ln = 1
			}

			cfv := tickets.CustomFieldValue{
				CustomFieldID:     entry.id,
				CustomFieldRuleID: entry.ruleID,
				Line:              ln,
				Value:             value,
			}
			for _, it := range items {
				cfv.Items = append(cfv.Items, tickets.CustomFieldItem{CustomFieldItem: it})
			}
			for _, p := range itemPersons {
				cfv.Items = append(cfv.Items, tickets.CustomFieldItem{PersonID: p})
			}
			for _, c := range itemClients {
				cfv.Items = append(cfv.Items, tickets.CustomFieldItem{ClientID: c})
			}
			for _, tm := range itemTeams {
				cfv.Items = append(cfv.Items, tickets.CustomFieldItem{Team: tm})
			}
			if cfv.Value == "" && len(cfv.Items) == 0 {
				return errors.New("provide --value, --item, --item-person, --item-client or --item-team")
			}

			raw, err := tickets.New(r.client).SetCustomFieldValue(cmd.Context(), id, cfv)
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
	cmd.Flags().IntVar(&fieldID, "field", 0, "numeric custom field id (or use --field-label)")
	cmd.Flags().StringVar(&fieldLabel, "field-label", "", "label registered in the catalog")
	cmd.Flags().IntVar(&ruleID, "rule", 0, "rule id (taken from catalog if omitted)")
	cmd.Flags().IntVar(&line, "line", 0, "row number (default 1)")
	cmd.Flags().StringVar(&value, "value", "", "value for text/numeric/date/etc. types")
	cmd.Flags().StringSliceVar(&items, "item", nil, "list-of-values item label (repeatable)")
	cmd.Flags().StringSliceVar(&itemPersons, "item-person", nil, "person id (repeatable)")
	cmd.Flags().StringSliceVar(&itemClients, "item-client", nil, "client id (repeatable)")
	cmd.Flags().StringSliceVar(&itemTeams, "item-team", nil, "team name (repeatable)")
	return cmd
}

func newTicketsCFClearCmd() *cobra.Command {
	var (
		fieldID    int
		fieldLabel string
		ruleID     int
		line       int
	)
	cmd := &cobra.Command{
		Use:   "clear <ticket-id>",
		Short: "Remove a custom field value (read-merge-patch)",
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
			cat, _ := loadCatalog(r.tenant.Name)
			entry, err := resolveField(cat, fieldID, fieldLabel, ruleID)
			if err != nil {
				return err
			}
			raw, err := tickets.New(r.client).ClearCustomFieldValue(cmd.Context(), id, entry.id, entry.ruleID, line)
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
	cmd.Flags().IntVar(&fieldID, "field", 0, "numeric custom field id")
	cmd.Flags().StringVar(&fieldLabel, "field-label", "", "label from the catalog")
	cmd.Flags().IntVar(&ruleID, "rule", 0, "rule id (omit with catalog)")
	cmd.Flags().IntVar(&line, "line", 0, "specific line; omit to clear every line")
	return cmd
}

type resolvedField struct {
	id     int
	ruleID int
	entry  CatalogEntry
}

func resolveField(cat *Catalog, fieldID int, fieldLabel string, ruleID int) (resolvedField, error) {
	if fieldID == 0 && fieldLabel == "" {
		return resolvedField{}, errors.New("provide --field or --field-label")
	}
	if fieldLabel != "" {
		if cat == nil {
			return resolvedField{}, fmt.Errorf("--field-label %q requires a catalog (run `tickets customfields catalog add`)", fieldLabel)
		}
		entry, ok := cat.Fields[fieldLabel]
		if !ok {
			return resolvedField{}, fmt.Errorf("no catalog entry for %q", fieldLabel)
		}
		rid := ruleID
		if rid == 0 {
			rid = entry.RuleID
		}
		return resolvedField{id: entry.ID, ruleID: rid, entry: entry}, nil
	}
	if ruleID == 0 {
		return resolvedField{}, errors.New("--rule is required when --field is given without a catalog")
	}
	return resolvedField{id: fieldID, ruleID: ruleID}, nil
}

// summarizeItems renders an items[] slice as a short string for tables.
func summarizeItems(items []tickets.CustomFieldItem) string {
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

func joinShort(parts []string, sep string, max int) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
		if len(out) > max {
			return out[:max] + "…"
		}
	}
	return out
}

// ----- catalog subcommands -----

func newTicketsCFCatalogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "Manage the local catalog of custom fields per tenant",
	}
	cmd.AddCommand(
		newTicketsCFCatalogListCmd(),
		newTicketsCFCatalogAddCmd(),
		newTicketsCFCatalogRemoveCmd(),
	)
	return cmd
}

func newTicketsCFCatalogListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List catalog entries for the current tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			cat, err := loadCatalog(r.tenant.Name)
			if err != nil {
				return err
			}
			rows := make([]map[string]any, 0, len(cat.Fields))
			labels := cat.sortedLabels()
			for _, label := range labels {
				e := cat.Fields[label]
				rows = append(rows, map[string]any{
					"label":   label,
					"id":      e.ID,
					"rule_id": e.RuleID,
					"type":    string(e.Type),
					"options": joinShort(e.Options, ", ", 60),
				})
			}
			return renderRows(cmd.OutOrStdout(), rows, r.output, "", []string{"label", "id", "rule_id", "type", "options"})
		},
	}
	return cmd
}

func newTicketsCFCatalogAddCmd() *cobra.Command {
	var (
		label   string
		id      int
		ruleID  int
		ftype   string
		options []string
	)
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Register a custom field in the local catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
			if label == "" || id == 0 || ruleID == 0 || ftype == "" {
				return errors.New("--label, --field, --rule and --type are required")
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			cat, err := loadCatalog(r.tenant.Name)
			if err != nil {
				return err
			}
			if !knownFieldType(ftype) {
				return fmt.Errorf("unknown --type %q (see `tickets customfields help`)", ftype)
			}
			cat.Fields[label] = CatalogEntry{
				ID:      id,
				RuleID:  ruleID,
				Type:    FieldType(ftype),
				Options: options,
			}
			if err := saveCatalog(r.tenant.Name, cat); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Saved %q in tenant %q catalog\n", label, r.tenant.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&label, "label", "", "human label, e.g. \"Severidade\" (required)")
	cmd.Flags().IntVar(&id, "field", 0, "numeric customFieldId from Movidesk (required)")
	cmd.Flags().IntVar(&ruleID, "rule", 0, "numeric customFieldRuleId from Movidesk (required)")
	cmd.Flags().StringVar(&ftype, "type", "", "field type (required)")
	cmd.Flags().StringSliceVar(&options, "options", nil, "allowed options for list types")
	return cmd
}

func newTicketsCFCatalogRemoveCmd() *cobra.Command {
	var label string
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a label from the catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
			if label == "" {
				return errors.New("--label is required")
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			cat, err := loadCatalog(r.tenant.Name)
			if err != nil {
				return err
			}
			if _, ok := cat.Fields[label]; !ok {
				return fmt.Errorf("no catalog entry for %q", label)
			}
			delete(cat.Fields, label)
			if err := saveCatalog(r.tenant.Name, cat); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed %q from tenant %q catalog\n", label, r.tenant.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&label, "label", "", "label to remove (required)")
	return cmd
}

func knownFieldType(t string) bool {
	known := []string{
		string(FieldText), string(FieldMultilineText), string(FieldHTML), string(FieldRegex),
		string(FieldNumber), string(FieldDate), string(FieldTime), string(FieldDateTime),
		string(FieldEmail), string(FieldPhone), string(FieldURL),
		string(FieldListOfValues), string(FieldListOfPersons), string(FieldListOfClients),
		string(FieldListOfAgents), string(FieldSingleSelect), string(FieldMultiSelect),
	}
	sort.Strings(known)
	idx := sort.SearchStrings(known, t)
	return idx < len(known) && known[idx] == t
}
