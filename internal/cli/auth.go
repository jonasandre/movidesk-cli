package cli

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jonasandre/movidesk-cli/internal/auth"
	"github.com/jonasandre/movidesk-cli/internal/config"
	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Movidesk tokens and tenants",
	}
	cmd.AddCommand(
		newAuthLoginCmd(),
		newAuthListCmd(),
		newAuthSwitchCmd(),
		newAuthStatusCmd(),
		newAuthLogoutCmd(),
		newAuthTokenCmd(),
	)
	return cmd
}

// readToken reads the token from stdin without echoing it. If stdin is not a
// terminal (CI pipe), it reads a single line normally.
func readToken(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	var line string
	if _, err := fmt.Fscanln(os.Stdin, &line); err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// validateToken makes a tiny request to confirm the token is accepted.
func validateToken(ctx context.Context, baseURL, token string) error {
	c := movidesk.New(baseURL, token)
	c.HTTP.Timeout = 15 * time.Second
	c.Retry.MaxAttempts = 1 // fast feedback during login
	v := url.Values{}
	v.Set("$top", "1")
	v.Set("$select", "id")
	if _, err := c.Do(ctx, "GET", "/persons", v, nil); err != nil {
		return err
	}
	return nil
}

func newAuthLoginCmd() *cobra.Command {
	var (
		tenant      string
		label       string
		baseURL     string
		makeDefault bool
		skipVerify  bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Add or update a tenant and store its API token",
		Long: `Add or update a Movidesk tenant. The token is read from a hidden prompt
(or stdin when not a TTY) and saved to the OS keychain when available, otherwise
to an encrypted file under ~/.movidesk.

By default, login validates the token by issuing GET /persons?$top=1 against
the configured base URL. Use --skip-verify to bypass.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tenant == "" {
				return errors.New("--tenant is required")
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			tn := cfg.Tenants[tenant]
			if tn == nil {
				tn = &config.Tenant{Name: tenant}
			}
			if label != "" {
				tn.Label = label
			}
			if baseURL != "" {
				tn.BaseURL = baseURL
			}

			token, err := readToken("Enter Movidesk token for tenant " + tenant + ": ")
			if err != nil {
				return fmt.Errorf("read token: %w", err)
			}
			if token == "" {
				return errors.New("token cannot be empty")
			}

			if !skipVerify {
				if err := validateToken(cmd.Context(), tn.EffectiveBaseURL(), token); err != nil {
					if movidesk.IsUnauthorized(err) {
						return fmt.Errorf("token rejected by Movidesk (401/403). Check the token and try again")
					}
					return fmt.Errorf("validate token: %w", err)
				}
			}

			store := auth.New()
			if err := store.Set(tenant, token); err != nil {
				return fmt.Errorf("store token: %w", err)
			}

			cfg.Set(tn)
			if makeDefault || cfg.CurrentTenant == "" {
				cfg.CurrentTenant = tenant
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Saved tenant %q (current: %s)\n", tenant, cfg.CurrentTenant)
			return nil
		},
	}
	cmd.Flags().StringVar(&tenant, "tenant", "", "tenant name (required)")
	cmd.Flags().StringVar(&label, "label", "", "human label, e.g. \"Acme Prod\"")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "override API base URL (sandbox)")
	cmd.Flags().BoolVar(&makeDefault, "make-default", false, "set this tenant as the current one")
	cmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "do not validate the token against the API")
	_ = cmd.MarkFlagRequired("tenant")
	return cmd
}

func newAuthListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured tenants",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if len(cfg.Tenants) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No tenants configured. Run `movidesk-cli auth login --tenant <name>`.")
				return nil
			}
			names := make([]string, 0, len(cfg.Tenants))
			for n := range cfg.Tenants {
				names = append(names, n)
			}
			sort.Strings(names)
			out := cmd.OutOrStdout()
			for _, n := range names {
				t := cfg.Tenants[n]
				marker := "  "
				if n == cfg.CurrentTenant {
					marker = "* "
				}
				label := t.Label
				if label == "" {
					label = "—"
				}
				fmt.Fprintf(out, "%s%s\t%s\t%s\n", marker, n, label, t.EffectiveBaseURL())
			}
			return nil
		},
	}
}

func newAuthSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <tenant>",
		Short: "Switch the current tenant",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if _, err := cfg.Get(args[0]); err != nil {
				return err
			}
			cfg.CurrentTenant = args[0]
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Current tenant: %s\n", args[0])
			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	var tenantOverride string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Validate the current (or specified) tenant token",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			tn, err := cfg.Resolve(tenantOverride)
			if err != nil {
				return err
			}
			tok, err := auth.ResolveToken(auth.New(), tn.Name)
			if err != nil {
				return err
			}
			err = validateToken(cmd.Context(), tn.EffectiveBaseURL(), tok)
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "tenant:   %s\n", tn.Name)
			fmt.Fprintf(out, "label:    %s\n", strOrDash(tn.Label))
			fmt.Fprintf(out, "base_url: %s\n", tn.EffectiveBaseURL())
			fmt.Fprintf(out, "token:    %s\n", auth.EncodePeek(tok))
			if err != nil {
				fmt.Fprintf(out, "status:   ERROR — %s\n", err)
				return err
			}
			fmt.Fprintln(out, "status:   OK")
			return nil
		},
	}
	cmd.Flags().StringVar(&tenantOverride, "tenant", "", "tenant to check (default: current)")
	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	var (
		tenantOverride string
		all            bool
	)
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove a tenant's stored token (and optionally the tenant entry)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			store := auth.New()
			targets := []string{}
			if all {
				for n := range cfg.Tenants {
					targets = append(targets, n)
				}
			} else {
				name := tenantOverride
				if name == "" {
					name = cfg.CurrentTenant
				}
				if name == "" {
					return errors.New("no tenant specified and no current tenant set")
				}
				targets = []string{name}
			}
			for _, n := range targets {
				if err := store.Delete(n); err != nil && !errors.Is(err, auth.ErrNotFound) {
					return fmt.Errorf("delete token for %s: %w", n, err)
				}
				cfg.Delete(n)
				fmt.Fprintf(cmd.OutOrStdout(), "Logged out tenant %q\n", n)
			}
			return cfg.Save()
		},
	}
	cmd.Flags().StringVar(&tenantOverride, "tenant", "", "tenant to log out (default: current)")
	cmd.Flags().BoolVar(&all, "all", false, "log out every configured tenant")
	return cmd
}

func newAuthTokenCmd() *cobra.Command {
	var tenantOverride string
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Print a tenant's token to stdout (use with care; for piping)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			tn, err := cfg.Resolve(tenantOverride)
			if err != nil {
				return err
			}
			tok, err := auth.ResolveToken(auth.New(), tn.Name)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), tok)
			return nil
		},
	}
	cmd.Flags().StringVar(&tenantOverride, "tenant", "", "tenant (default: current)")
	return cmd
}

func strOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
