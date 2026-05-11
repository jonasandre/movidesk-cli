package cli

import (
	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/surveys"
)

func newSurveysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "surveys",
		Short: "Lê dados de pesquisas de satisfação do Movidesk (/survey/...)",
	}
	cmd.AddCommand(newSurveysQuestionsCmd(), newSurveysResponsesCmd())
	return cmd
}

func newSurveysQuestionsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "questions", Short: "Perguntas de pesquisas"}
	cmd.AddCommand(newSurveysQuestionsListCmd(), newSurveysQuestionsGetCmd())
	return cmd
}

func newSurveysQuestionsListCmd() *cobra.Command {
	var (
		typeFilter int
		cf         columnsFlag
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista perguntas de pesquisas",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			body, err := surveys.New(r.client).ListQuestions(cmd.Context(), typeFilter)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), body, r.output, "surveys.questions", cf.cols)
		},
	}
	cmd.Flags().IntVar(&typeFilter, "type", 0, "filtra por tipo da pergunta (1=satisfação, 2=carinhas, 3=NPS, 4=sim/não)")
	cf.bind(cmd)
	return cmd
}

func newSurveysQuestionsGetCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Obtém uma única pergunta de pesquisa pelo id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			body, err := surveys.New(r.client).GetQuestion(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), body, r.output, "surveys.questions", cf.cols)
		},
	}
	cf.bind(cmd)
	return cmd
}

func newSurveysResponsesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "responses", Short: "Respostas de pesquisas"}
	cmd.AddCommand(newSurveysResponsesListCmd())
	return cmd
}

func newSurveysResponsesListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		all           bool
		max           int
		cf            columnsFlag
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista respostas de pesquisas (paginação por cursor)",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			svc := surveys.New(r.client)
			if all {
				rows, err := svc.ListAllResponses(cmd.Context(), max)
				if err != nil {
					return err
				}
				return renderRows(cmd.OutOrStdout(), rows, r.output, "surveys.responses", cf.cols)
			}
			page, err := svc.ListResponsesPage(cmd.Context(), limit, startingAfter)
			if err != nil {
				return err
			}
			return renderRows(cmd.OutOrStdout(), page, r.output, "surveys.responses", cf.cols)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 0, "tamanho da página (1..100, padrão 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "cursor (id do último item da página anterior)")
	cmd.Flags().BoolVar(&all, "all", false, "percorre todas as páginas")
	cmd.Flags().IntVar(&max, "max", 0, "com --all, interrompe após este número de registros")
	cf.bind(cmd)
	return cmd
}
