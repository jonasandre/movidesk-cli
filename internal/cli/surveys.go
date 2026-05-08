package cli

import (
	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/surveys"
)

func newSurveysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "surveys",
		Short: "Read Movidesk satisfaction survey data (/survey/...)",
	}
	cmd.AddCommand(newSurveysQuestionsCmd(), newSurveysResponsesCmd())
	return cmd
}

func newSurveysQuestionsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "questions", Short: "Survey questions"}
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
		Short: "List survey questions",
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
	cmd.Flags().IntVar(&typeFilter, "type", 0, "filter by question type (1=satisf, 2=faces, 3=NPS, 4=yes/no)")
	cf.bind(cmd)
	return cmd
}

func newSurveysQuestionsGetCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single survey question by id",
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
	cmd := &cobra.Command{Use: "responses", Short: "Survey responses"}
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
		Short: "List survey responses (cursor pagination)",
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
	cmd.Flags().IntVar(&limit, "limit", 0, "page size (1..100, default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "cursor (id of last item from previous page)")
	cmd.Flags().BoolVar(&all, "all", false, "walk every page")
	cmd.Flags().IntVar(&max, "max", 0, "with --all, stop after this many records")
	cf.bind(cmd)
	return cmd
}
