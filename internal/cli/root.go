package cli

import (
	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/version"
)

type globalFlags struct {
	tenant  string
	output  string
	noColor bool
	verbose bool
	noRetry bool
	compact bool
	user    string
}

var flags globalFlags

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "movidesk-cli",
		Short: "CLI para a API REST do Movidesk",
		Long: `movidesk-cli é uma interface de linha de comando para a API REST do Movidesk.

Suporta múltiplos tenants (cada um com seu próprio token armazenado no chaveiro
do sistema operacional), formata a saída como JSON, tabela ou CSV, e respeita o
limite de 10 requisições por minuto do Movidesk com retentativa automática em 429.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.String(),
	}

	cmd.PersistentFlags().StringVar(&flags.tenant, "tenant", "", "nome do tenant (sobrepõe o tenant atual; env: MOVIDESK_TENANT)")
	cmd.PersistentFlags().StringVarP(&flags.output, "output", "o", "", "formato de saída: json|table|csv (padrão: do tenant ou 'json')")
	cmd.PersistentFlags().BoolVar(&flags.noColor, "no-color", false, "desativa cores na saída")
	cmd.PersistentFlags().BoolVarP(&flags.verbose, "verbose", "v", false, "log detalhado em stderr")
	cmd.PersistentFlags().BoolVar(&flags.noRetry, "no-retry", false, "desativa retentativa automática em 429/5xx")
	cmd.PersistentFlags().BoolVar(&flags.compact, "compact", false, "JSON compacto (sem indentação)")
	cmd.PersistentFlags().StringVar(&flags.user, "user", "", "id do usuário padrão (Cod. Ref.) usado em createdBy nas escritas; sobrepõe a configuração do tenant; env: MOVIDESK_USER")

	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newTicketsCmd())
	cmd.AddCommand(newPersonsCmd())
	cmd.AddCommand(newServicesCmd())
	cmd.AddCommand(newActivitiesCmd())
	cmd.AddCommand(newContractsCmd())
	cmd.AddCommand(newSurveysCmd())
	cmd.AddCommand(newKBCmd())
	cmd.AddCommand(newTelephonyCmd())
	cmd.AddCommand(newCustomFieldsCmd())
	cmd.AddCommand(newQueryCmd())
	_ = flags // referenced by subcommand packages

	return cmd
}

func Execute() error {
	return newRootCmd().Execute()
}

// NewRootForDocs is exported so tools (notably cmd/gen-docs) can walk the
// command tree and emit Markdown reference pages without invoking commands.
func NewRootForDocs() *cobra.Command { return newRootCmd() }
