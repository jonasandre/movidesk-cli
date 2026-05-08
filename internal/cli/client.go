package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/auth"
	"github.com/jonasandre/movidesk-cli/internal/config"
	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

// resolved bundles a ready-to-use client and the resolved tenant/output info.
type resolved struct {
	cfg    *config.Config
	tenant *config.Tenant
	client *movidesk.Client
	output string
}

// resolveClient loads config, resolves the active tenant, fetches its token,
// and returns a configured movidesk.Client. Honors the global --tenant,
// --no-retry, --output, --verbose, --no-color flags.
func resolveClient(cmd *cobra.Command) (*resolved, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	tn, err := cfg.Resolve(flags.tenant)
	if err != nil {
		return nil, err
	}
	tok, err := auth.ResolveToken(auth.New(), tn.Name)
	if err != nil {
		return nil, fmt.Errorf("resolve token for tenant %q: %w", tn.Name, err)
	}

	c := movidesk.New(tn.EffectiveBaseURL(), tok)
	if flags.noRetry {
		c.Retry.Disabled = true
	}
	if flags.verbose {
		c.OnRequest = func(method, url string) {
			fmt.Fprintf(os.Stderr, "→ %s %s\n", method, url)
		}
		c.OnResponse = func(status int, ms int64) {
			fmt.Fprintf(os.Stderr, "← HTTP %d (%dms)\n", status, ms)
		}
	}

	return &resolved{
		cfg:    cfg,
		tenant: tn,
		client: c,
		output: cfg.EffectiveOutput(tn, flags.output),
	}, nil
}
