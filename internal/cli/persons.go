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
		Short: "Gerencia pessoas do Movidesk (/persons)",
		Long: `Gerencia pessoas do Movidesk. O endpoint /persons serve agentes,
clientes, empresas e departamentos — diferenciados por personType
(1=Pessoa, 2=Empresa, 4=Departamento) e profileType (1=Agente, 2=Cliente,
3=Ambos).

Filtros e projeções OData aplicam-se em list. Os valores de campos
personalizados seguem o mesmo read-merge-patch dos chamados para evitar
a armadilha "apaga entradas ausentes" do Movidesk.`,
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
		Short: "Lista pessoas",
		Long: `Lista pessoas via GET /persons. Aceita filtros OData; campos comuns:
id, businessName, personType (1=Pessoa, 2=Empresa, 4=Departamento),
profileType (1=Agente, 2=Cliente, 3=Ambos), isActive.

Sintaxe completa em: movidesk-cli topics filters`,
		Example: `  # empresas ativas
  movidesk-cli persons list --filter "personType eq 2 and isActive eq true" --select "id,businessName"

  # agentes cujo nome começa com "Ana"
  movidesk-cli persons list --filter "profileType ne 2 and startswith(businessName, 'Ana')"`,
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
		Short: "Obtém uma pessoa pelo id (Cod. Ref.)",
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
		Short: "Cria uma pessoa a partir de corpo JSON, template ou substituições --set",
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
			raw, err := persons.New(r.client).Create(cmd.Context(), body, returnAllProperties)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "persons", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "caminho do corpo JSON")
	cmd.Flags().StringVar(&template, "from-template", "", "carrega ~/.movidesk/templates/<nome>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "carrega template de um caminho específico")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "sobrescreve campos, ex.: --set personType=1 --set businessName=\"Joe\"")
	cmd.Flags().BoolVar(&returnAllProperties, "return-all", false, "pede ao Movidesk pra retornar a pessoa completa")
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
		Short: "Aplica patch em uma pessoa por id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
	cmd.Flags().StringVarP(&file, "file", "f", "", "caminho do corpo JSON de patch")
	cmd.Flags().StringVar(&template, "from-template", "", "carrega ~/.movidesk/templates/<nome>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "carrega template de um caminho específico")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "sobrescreve campos inline")
	return cmd
}

func newPersonsDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Exclui uma pessoa de forma permanente (DELETE /persons?id=)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				if err := confirm(cmd, fmt.Sprintf("Excluir a pessoa %q? Esta ação é irreversível.", args[0])); err != nil {
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
	cmd.Flags().BoolVar(&force, "force", false, "pula o prompt de confirmação")
	return cmd
}

func newPersonsCustomFieldsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "customfields",
		Aliases: []string{"cf"},
		Short:   "Lê e escreve campos personalizados de pessoa (read-merge-patch)",
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
		Short: "Lista os customFieldValues de uma pessoa",
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
		Short: "Define o valor de um campo personalizado de pessoa (read-merge-patch)",
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
				return errors.New("informe --value, --item, --item-person, --item-client ou --item-team")
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
	cmd.Flags().IntVar(&fieldID, "field", 0, "id numérico do campo personalizado (ou use --field-label)")
	cmd.Flags().StringVar(&fieldLabel, "field-label", "", "rótulo registrado no catálogo")
	cmd.Flags().IntVar(&ruleID, "rule", 0, "id da regra (vem do catálogo se omitido)")
	cmd.Flags().IntVar(&line, "line", 0, "número da linha (padrão 1)")
	cmd.Flags().StringVar(&value, "value", "", "valor para tipos texto/numérico/data")
	cmd.Flags().StringSliceVar(&items, "item", nil, "rótulo de item da lista de valores (repetível)")
	cmd.Flags().StringSliceVar(&itemPersons, "item-person", nil, "id da pessoa (repetível)")
	cmd.Flags().StringSliceVar(&itemClients, "item-client", nil, "id do cliente (repetível)")
	cmd.Flags().StringSliceVar(&itemTeams, "item-team", nil, "nome da equipe (repetível)")
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
		Short: "Remove o valor de um campo personalizado de pessoa (read-merge-patch)",
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
	cmd.Flags().IntVar(&fieldID, "field", 0, "id numérico do campo personalizado")
	cmd.Flags().StringVar(&fieldLabel, "field-label", "", "rótulo do catálogo")
	cmd.Flags().IntVar(&ruleID, "rule", 0, "id da regra (omita se usar catálogo)")
	cmd.Flags().IntVar(&line, "line", 0, "linha específica; omita para limpar todas")
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
