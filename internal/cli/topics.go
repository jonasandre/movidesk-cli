package cli

import (
	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
)

func newTopicsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topics",
		Short: "Tópicos de ajuda detalhados (sintaxe, convenções, armadilhas)",
		Long: `Coleção de páginas de ajuda longas que não cabem no --help de um
comando específico. Cada subcomando aqui imprime uma referência pronta
para consulta. Hoje há:

  topics filters    Sintaxe OData aceita pelos comandos list (--filter, --select, …)`,
		Run: func(c *cobra.Command, _ []string) { _ = c.Help() },
	}
	cmd.AddCommand(newTopicFiltersCmd())
	return cmd
}

func newTopicFiltersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "filters",
		Short: "Sintaxe OData aceita pelo --filter (operadores, tipos, campos, exemplos)",
		Run:   func(c *cobra.Command, _ []string) { _ = c.Help() },
		Long:  odata.FilterTopic,
	}
}
