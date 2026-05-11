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
		Short:   "Lê e escreve campos personalizados de chamados (com read-merge-patch seguro)",
		Long: `O PATCH /tickets do Movidesk apaga qualquer entrada de customFieldValues
ausente no corpo. Este subcomando usa read-merge-patch internamente, então
você só descreve a alteração desejada, nunca a lista completa.

Um catálogo local em ~/.movidesk/<tenant>/customfields.yaml mapeia rótulos
legíveis para os ids numéricos e tipos dos campos, permitindo usar
--field-label "Severidade" no lugar de --field 125529.`,
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
		Short: "Lista os customFieldValues de um chamado",
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
		Short: "Define o valor de um campo personalizado (read-merge-patch)",
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
				return errors.New("informe --value, --item, --item-person, --item-client ou --item-team")
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
	cmd.Flags().IntVar(&fieldID, "field", 0, "id numérico do campo personalizado (ou use --field-label)")
	cmd.Flags().StringVar(&fieldLabel, "field-label", "", "rótulo registrado no catálogo")
	cmd.Flags().IntVar(&ruleID, "rule", 0, "id da regra (vem do catálogo se omitido)")
	cmd.Flags().IntVar(&line, "line", 0, "número da linha (padrão 1)")
	cmd.Flags().StringVar(&value, "value", "", "valor para tipos texto/numérico/data/etc.")
	cmd.Flags().StringSliceVar(&items, "item", nil, "rótulo de item da lista de valores (repetível)")
	cmd.Flags().StringSliceVar(&itemPersons, "item-person", nil, "id da pessoa (repetível)")
	cmd.Flags().StringSliceVar(&itemClients, "item-client", nil, "id do cliente (repetível)")
	cmd.Flags().StringSliceVar(&itemTeams, "item-team", nil, "nome da equipe (repetível)")
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
		Short: "Remove o valor de um campo personalizado (read-merge-patch)",
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
	cmd.Flags().IntVar(&fieldID, "field", 0, "id numérico do campo personalizado")
	cmd.Flags().StringVar(&fieldLabel, "field-label", "", "rótulo do catálogo")
	cmd.Flags().IntVar(&ruleID, "rule", 0, "id da regra (omita se usar catálogo)")
	cmd.Flags().IntVar(&line, "line", 0, "linha específica; omita para limpar todas")
	return cmd
}

type resolvedField struct {
	id     int
	ruleID int
	entry  CatalogEntry
}

func resolveField(cat *Catalog, fieldID int, fieldLabel string, ruleID int) (resolvedField, error) {
	if fieldID == 0 && fieldLabel == "" {
		return resolvedField{}, errors.New("informe --field ou --field-label")
	}
	if fieldLabel != "" {
		if cat == nil {
			return resolvedField{}, fmt.Errorf("--field-label %q exige um catálogo (execute `tickets customfields catalog add`)", fieldLabel)
		}
		entry, ok := cat.Fields[fieldLabel]
		if !ok {
			return resolvedField{}, fmt.Errorf("nenhuma entrada no catálogo para %q", fieldLabel)
		}
		rid := ruleID
		if rid == 0 {
			rid = entry.RuleID
		}
		return resolvedField{id: entry.ID, ruleID: rid, entry: entry}, nil
	}
	if ruleID == 0 {
		return resolvedField{}, errors.New("--rule é obrigatório quando --field é usado sem catálogo")
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
		Short: "Gerencia o catálogo local de campos personalizados por tenant",
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
		Short: "Lista as entradas do catálogo do tenant atual",
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
		Short: "Registra um campo personalizado no catálogo local",
		RunE: func(cmd *cobra.Command, args []string) error {
			if label == "" || id == 0 || ruleID == 0 || ftype == "" {
				return errors.New("--label, --field, --rule e --type são obrigatórios")
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
				return fmt.Errorf("--type desconhecido %q (veja `tickets customfields help`)", ftype)
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
			fmt.Fprintf(cmd.OutOrStdout(), "%q salvo no catálogo do tenant %q\n", label, r.tenant.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&label, "label", "", "rótulo legível, ex.: \"Severidade\" (obrigatório)")
	cmd.Flags().IntVar(&id, "field", 0, "customFieldId numérico do Movidesk (obrigatório)")
	cmd.Flags().IntVar(&ruleID, "rule", 0, "customFieldRuleId numérico do Movidesk (obrigatório)")
	cmd.Flags().StringVar(&ftype, "type", "", "tipo do campo (obrigatório)")
	cmd.Flags().StringSliceVar(&options, "options", nil, "opções permitidas para tipos de lista")
	return cmd
}

func newTicketsCFCatalogRemoveCmd() *cobra.Command {
	var label string
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove um rótulo do catálogo",
		RunE: func(cmd *cobra.Command, args []string) error {
			if label == "" {
				return errors.New("--label é obrigatório")
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
				return fmt.Errorf("nenhuma entrada no catálogo para %q", label)
			}
			delete(cat.Fields, label)
			if err := saveCatalog(r.tenant.Name, cat); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%q removido do catálogo do tenant %q\n", label, r.tenant.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&label, "label", "", "rótulo a remover (obrigatório)")
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
