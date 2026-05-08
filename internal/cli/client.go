package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/auth"
	"github.com/jonasandre/movidesk-cli/internal/config"
	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

var _ = os.Getenv // keep os.Getenv import live regardless of OnRequest hook usage

// resolved bundles a ready-to-use client and the resolved tenant/output info.
type resolved struct {
	cfg    *config.Config
	tenant *config.Tenant
	client *movidesk.Client
	output string
	userID string
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
		userID: resolveUser(tn),
	}, nil
}

// resolveUser picks the active user id for createdBy injection, in priority:
// CLI flag > env > tenant.DefaultUser. Returns "" when none is set.
func resolveUser(tn *config.Tenant) string {
	if flags.user != "" {
		return flags.user
	}
	if v := os.Getenv(EnvUser); v != "" {
		return v
	}
	if tn != nil {
		return tn.DefaultUser
	}
	return ""
}
