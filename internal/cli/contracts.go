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
		Short: "Gerencia contratos de horas do Movidesk (/timeAgreement)",
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
		Short: "Lista contratos de horas",
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
		Short: "Obtém um contrato pelo id",
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
		Short: "Cria um contrato",
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := loadBody(file, template, templateFile, sets)
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return errors.New("nenhum campo informado")
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
	cmd.Flags().StringVarP(&file, "file", "f", "", "caminho do corpo JSON")
	cmd.Flags().StringVar(&template, "from-template", "", "carrega ~/.movidesk/templates/<nome>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "carrega template de um caminho específico")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "sobrescreve campos inline, ex.: --set status=2")
	cmd.Flags().BoolVar(&returnAllProperties, "return-all", false, "pede ao Movidesk pra retornar o contrato completo")
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
		Short: "Aplica patch em um contrato por id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id inválido %q", args[0])
			}
			body, err := loadBody(file, template, templateFile, sets)
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return errors.New("nenhum campo para atualizar")
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
	cmd.Flags().StringVarP(&file, "file", "f", "", "caminho do corpo JSON de patch")
	cmd.Flags().StringVar(&template, "from-template", "", "carrega ~/.movidesk/templates/<nome>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "carrega template de um caminho específico")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "sobrescreve campos inline, ex.: --set status=2")
	return cmd
}

func newContractsDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Exclui um contrato",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id inválido %q", args[0])
			}
			if !force {
				if err := confirm(cmd, fmt.Sprintf("Excluir o contrato %d? Esta ação é irreversível.", id)); err != nil {
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
	cmd.Flags().BoolVar(&force, "force", false, "pula o prompt de confirmação")
	return cmd
}

func newContractsConsumptionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consumption",
		Short: "Lê /timeAgreementConsumption",
		Long: `Ao filtrar por startPeriod/endPeriod, o Movidesk exige o nome do
contrato no $filter OData (ex.: --filter "name eq 'Default' and period gt
2026-01-01T00:00:00Z"). Evite combinar $select com filtros de período.`,
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
		Short: "Lista linhas de consumo",
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
