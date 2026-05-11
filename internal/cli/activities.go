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
		Short: "Gerencia atividades do Movidesk (/activity)",
		Long: `Atividades usam paginação por cursor (limit/startingAfter) — não OData.
Use --name para filtrar por substring e --all para percorrer todas as páginas.`,
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
		Short: "Lista atividades (paginação por cursor)",
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
	cmd.Flags().StringVar(&nameFilter, "name", "", "filtra por substring no nome da atividade")
	cmd.Flags().IntVar(&limit, "limit", 0, "tamanho da página (1..100, padrão 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "cursor (último id da página anterior)")
	cmd.Flags().BoolVar(&all, "all", false, "percorre todas as páginas")
	cmd.Flags().IntVar(&max, "max", 0, "com --all, interrompe após este número de registros")
	cf.bind(cmd)
	return cmd
}

func newActivitiesGetCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Obtém uma atividade pelo id",
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
		Short: "Cria uma atividade",
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
			raw, err := activities.New(r.client).Create(cmd.Context(), body)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "activities", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "caminho do corpo JSON")
	cmd.Flags().StringVar(&template, "from-template", "", "carrega ~/.movidesk/templates/<nome>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "carrega template de um caminho específico")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "sobrescreve campos, ex.: --set name=\"Atividade\" --set isActive=true")
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
		Short: "Aplica patch em uma atividade por id",
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
	cmd.Flags().StringVarP(&file, "file", "f", "", "caminho do corpo JSON de patch")
	cmd.Flags().StringVar(&template, "from-template", "", "carrega ~/.movidesk/templates/<nome>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "carrega template de um caminho específico")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "sobrescreve campos inline, ex.: --set name=Foo")
	return cmd
}

func newActivitiesDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Exclui uma atividade",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id inválido %q", args[0])
			}
			if !force {
				if err := confirm(cmd, fmt.Sprintf("Excluir a atividade %d? Esta ação é irreversível.", id)); err != nil {
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
	cmd.Flags().BoolVar(&force, "force", false, "pula o prompt de confirmação")
	return cmd
}

func newActivitiesAddTeamsCmd() *cobra.Command {
	var teams []string
	cmd := &cobra.Command{
		Use:   "add-teams <activity-id>",
		Short: "Adiciona equipes a uma atividade (POST /addTeamsToActivity)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("id da atividade inválido %q", args[0])
			}
			if len(teams) == 0 {
				return errors.New("--team é obrigatório (repetível)")
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
	cmd.Flags().StringSliceVar(&teams, "team", nil, "nome da equipe a adicionar (repetível)")
	return cmd
}
