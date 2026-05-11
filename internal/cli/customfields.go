package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/customfields"
)

func newCustomFieldsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "customfields",
		Short: "Gerencia o conjunto de opções de campos personalizados do tipo lista (por tenant)",
		Long: `Estes comandos encapsulam os endpoints /ticketCustomFieldValue/{InsertValues,UpdateValues,DeleteValues},
que gerenciam o CONJUNTO DE OPÇÕES dos campos personalizados do tipo lista —
os valores que aparecem no dropdown pros agentes selecionarem.

Pra definir um valor em um CHAMADO ou PESSOA específica, use:
  tickets customfields set ...
  persons customfields set ...
`,
	}
	cmd.AddCommand(newCustomFieldsOptionsCmd())
	return cmd
}

func newCustomFieldsOptionsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "options", Short: "Adiciona/renomeia/remove valores do conjunto de opções"}
	cmd.AddCommand(
		newCustomFieldsOptionsAddCmd(),
		newCustomFieldsOptionsRenameCmd(),
		newCustomFieldsOptionsRemoveCmd(),
	)
	return cmd
}

func newCustomFieldsOptionsAddCmd() *cobra.Command {
	var (
		fieldID string
		values  []string
	)
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Insere valores de opção no conjunto de um campo tipo lista",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fieldID == "" || len(values) == 0 {
				return errors.New("--field e --value (repetível) são obrigatórios")
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := customfields.New(r.client).AddOptions(cmd.Context(), fieldID, values)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "", nil)
		},
	}
	cmd.Flags().StringVar(&fieldID, "field", "", "customFieldId numérico (obrigatório)")
	cmd.Flags().StringSliceVar(&values, "value", nil, "valor de opção (repetível)")
	return cmd
}

func newCustomFieldsOptionsRenameCmd() *cobra.Command {
	var (
		fieldID string
		pairs   []string
	)
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Renomeia valores existentes via --pair ANTIGO=NOVO (repetível)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fieldID == "" || len(pairs) == 0 {
				return errors.New("--field e --pair (repetível) são obrigatórios")
			}
			parsed := make([]customfields.UpdatePair, 0, len(pairs))
			for _, p := range pairs {
				old, neu, ok := strings.Cut(p, "=")
				if !ok {
					return fmt.Errorf("--pair deve ser ANTIGO=NOVO, recebido %q", p)
				}
				parsed = append(parsed, customfields.UpdatePair{OldName: old, NewName: neu})
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := customfields.New(r.client).RenameOptions(cmd.Context(), fieldID, parsed)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "", nil)
		},
	}
	cmd.Flags().StringVar(&fieldID, "field", "", "customFieldId numérico (obrigatório)")
	cmd.Flags().StringSliceVar(&pairs, "pair", nil, "ANTIGO=NOVO (repetível)")
	return cmd
}

func newCustomFieldsOptionsRemoveCmd() *cobra.Command {
	var (
		fieldID string
		values  []string
	)
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove valores de opção do conjunto de um campo tipo lista",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fieldID == "" || len(values) == 0 {
				return errors.New("--field e --value (repetível) são obrigatórios")
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := customfields.New(r.client).RemoveOptions(cmd.Context(), fieldID, values)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "", nil)
		},
	}
	cmd.Flags().StringVar(&fieldID, "field", "", "customFieldId numérico (obrigatório)")
	cmd.Flags().StringSliceVar(&values, "value", nil, "valor de opção (repetível)")
	return cmd
}
