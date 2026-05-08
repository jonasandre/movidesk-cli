package cli

import (
	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/version"
)

type globalFlags struct {
	tenant   string
	output   string
	noColor  bool
	verbose  bool
	noRetry  bool
	compact  bool
}

var flags globalFlags

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "movidesk-cli",
		Short: "CLI for the Movidesk REST API",
		Long: `movidesk-cli is a command-line interface for the Movidesk REST API.

It supports multiple tenants (each with its own token stored in the OS keychain),
formats output as JSON, table or CSV, and respects Movidesk's rate limit of
10 requests per minute with automatic retry on 429.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.String(),
	}

	cmd.PersistentFlags().StringVar(&flags.tenant, "tenant", "", "tenant name (overrides current tenant; env: MOVIDESK_TENANT)")
	cmd.PersistentFlags().StringVarP(&flags.output, "output", "o", "", "output format: json|table|csv (default: tenant or 'json')")
	cmd.PersistentFlags().BoolVar(&flags.noColor, "no-color", false, "disable colored output")
	cmd.PersistentFlags().BoolVarP(&flags.verbose, "verbose", "v", false, "verbose logging to stderr")
	cmd.PersistentFlags().BoolVar(&flags.noRetry, "no-retry", false, "disable automatic retry on 429/5xx")
	cmd.PersistentFlags().BoolVar(&flags.compact, "compact", false, "compact JSON output (no indentation)")

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
