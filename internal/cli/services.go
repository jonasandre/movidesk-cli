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
		Short: "Gerencia o catálogo de serviços do Movidesk (/services)",
		Long: `Gerencia entradas do catálogo de serviços do Movidesk.

Atenção: PATCH substitui campos com valor de array como "categories" — ao
atualizar, envie a lista completa que deseja manter, não apenas os adicionais.`,
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
		Short: "Lista serviços",
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
		Short: "Obtém um serviço pelo id",
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
		Short: "Cria um serviço a partir de corpo JSON, template ou substituições --set",
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
			raw, err := services.New(r.client).Create(cmd.Context(), body, returnAllProperties)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "services", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "caminho do corpo JSON")
	cmd.Flags().StringVar(&template, "from-template", "", "carrega ~/.movidesk/templates/<nome>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "carrega template de um caminho específico")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "sobrescreve campos, ex.: --set name=\"Suporte\" --set isActive=true")
	cmd.Flags().BoolVar(&returnAllProperties, "return-all", false, "pede ao Movidesk pra retornar o serviço completo")
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
		Short: "Aplica patch em um serviço por id",
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
				return errors.New("nenhum campo para atualizar; passe --file, --from-template[-file] ou --set chave=valor")
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
	cmd.Flags().StringVarP(&file, "file", "f", "", "caminho do corpo JSON de patch")
	cmd.Flags().StringVar(&template, "from-template", "", "carrega ~/.movidesk/templates/<nome>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "carrega template de um caminho específico")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "sobrescreve campos inline")
	return cmd
}

func newServicesDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Exclui um serviço de forma permanente (DELETE /services?id=)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id inválido %q", args[0])
			}
			if !force {
				if err := confirm(cmd, fmt.Sprintf("Excluir o serviço %d? Esta ação é irreversível.", id)); err != nil {
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
	cmd.Flags().BoolVar(&force, "force", false, "pula o prompt de confirmação")
	return cmd
}
