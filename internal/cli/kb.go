package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/knowledgebase"
)

func newKBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kb",
		Short: "Read Movidesk knowledge base articles (/article/:id)",
		Long: `Movidesk's public KB API exposes only single-article reads (GET
/article/:id). There is no public list endpoint — you must already know
the article id.`,
	}
	cmd.AddCommand(newKBArticlesCmd())
	return cmd
}

func newKBArticlesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "articles", Short: "Articles"}
	cmd.AddCommand(newKBArticlesGetCmd())
	return cmd
}

func newKBArticlesGetCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get one KB article by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q", args[0])
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			body, err := knowledgebase.New(r.client).Get(cmd.Context(), id)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), body, r.output, "articles", cf.cols)
		},
	}
	cf.bind(cmd)
	return cmd
}
