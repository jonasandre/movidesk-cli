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
		Short: "Lê artigos da base de conhecimento do Movidesk (/article/:id)",
		Long: `A API pública de base de conhecimento do Movidesk expõe apenas
leitura de artigo único (GET /article/:id). Não existe endpoint público
de listagem — é preciso conhecer o id do artigo previamente.`,
	}
	cmd.AddCommand(newKBArticlesCmd())
	return cmd
}

func newKBArticlesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "articles", Short: "Artigos"}
	cmd.AddCommand(newKBArticlesGetCmd())
	return cmd
}

func newKBArticlesGetCmd() *cobra.Command {
	var cf columnsFlag
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Obtém um artigo de base de conhecimento pelo id",
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
