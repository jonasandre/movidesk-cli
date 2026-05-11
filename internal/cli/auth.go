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
		Short: "Gerencia tokens e tenants do Movidesk",
	}
	cmd.AddCommand(
		newAuthLoginCmd(),
		newAuthListCmd(),
		newAuthSwitchCmd(),
		newAuthStatusCmd(),
		newAuthLogoutCmd(),
		newAuthTokenCmd(),
		newAuthSetUserCmd(),
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
		tenant         string
		label          string
		baseURL        string
		makeDefault    bool
		skipVerify     bool
		skipVerifyUser bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Adiciona ou atualiza um tenant e armazena seu token de API",
		Long: `Adiciona ou atualiza um tenant do Movidesk. O token é lido de um prompt
oculto (ou da stdin quando não há TTY) e salvo no chaveiro do sistema operacional
quando disponível; caso contrário, em arquivo criptografado em ~/.movidesk.

Por padrão, login valida o token fazendo GET /persons?$top=1 contra a base URL
configurada. Use --skip-verify para pular essa verificação.

Após validar o token, login pode solicitar interativamente um usuário padrão
(Cod. Ref.) que será injetado automaticamente como createdBy nas escritas que
exigem atribuição. Passe --user <id> para definir sem prompt, ou deixe vazio
para pular. Use --skip-verify-user para pular a checagem de existência (útil
quando o token não tem permissão de ler a API de pessoas).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tenant == "" {
				return errors.New("--tenant é obrigatório")
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

			token, err := readToken("Token do Movidesk para o tenant " + tenant + ": ")
			if err != nil {
				return fmt.Errorf("ler token: %w", err)
			}
			if token == "" {
				return errors.New("token não pode ser vazio")
			}

			if !skipVerify {
				if err := validateToken(cmd.Context(), tn.EffectiveBaseURL(), token); err != nil {
					if movidesk.IsUnauthorized(err) {
						return fmt.Errorf("token rejeitado pelo Movidesk (401/403). Verifique o token e tente novamente")
					}
					return fmt.Errorf("validar token: %w", err)
				}
			}

			// Resolve default user: --user from persistent flag, else interactive prompt.
			userID := strings.TrimSpace(flags.user)
			if userID == "" && term.IsTerminal(int(os.Stdin.Fd())) {
				userID, err = readLine(cmd.ErrOrStderr(), "Usuário padrão (Cod. Ref.) [opcional, enter para pular]: ")
				if err != nil {
					return fmt.Errorf("ler usuário: %w", err)
				}
			}
			if userID != "" && !skipVerifyUser {
				name, err := validateUser(cmd.Context(), tn.EffectiveBaseURL(), token, userID)
				if err != nil {
					return fmt.Errorf("validar usuário %q: %w (use --skip-verify-user para pular)", userID, err)
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "Usuário padrão: %s (%s)\n", userID, strOrDash(name))
			}
			tn.DefaultUser = userID

			store := auth.New()
			if err := store.Set(tenant, token); err != nil {
				return fmt.Errorf("armazenar token: %w", err)
			}

			cfg.Set(tn)
			if makeDefault || cfg.CurrentTenant == "" {
				cfg.CurrentTenant = tenant
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Tenant %q salvo (atual: %s)\n", tenant, cfg.CurrentTenant)
			return nil
		},
	}
	cmd.Flags().StringVar(&tenant, "tenant", "", "nome do tenant (obrigatório)")
	cmd.Flags().StringVar(&label, "label", "", "rótulo legível, ex.: \"Acme Prod\"")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "sobrepõe a base URL da API (sandbox)")
	cmd.Flags().BoolVar(&makeDefault, "make-default", false, "define este tenant como o atual")
	cmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "não valida o token contra a API")
	cmd.Flags().BoolVar(&skipVerifyUser, "skip-verify-user", false, "pula a checagem de existência do usuário padrão")
	_ = cmd.MarkFlagRequired("tenant")
	return cmd
}

// readLine reads one line from stdin without echoing; used for non-secret
// prompts like the default user id.
func readLine(out interface{ Write([]byte) (int, error) }, prompt string) (string, error) {
	if _, err := out.Write([]byte(prompt)); err != nil {
		return "", err
	}
	var line string
	_, err := fmt.Fscanln(os.Stdin, &line)
	if err != nil {
		// Fscanln treats empty input as an error; treat that as "skip".
		if err.Error() == "unexpected newline" {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func newAuthListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lista os tenants configurados",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if len(cfg.Tenants) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Nenhum tenant configurado. Execute `movidesk-cli auth login --tenant <nome>`.")
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
				user := t.DefaultUser
				if user == "" {
					user = "—"
				}
				fmt.Fprintf(out, "%s%s\t%s\t%s\tuser=%s\n", marker, n, label, t.EffectiveBaseURL(), user)
			}
			return nil
		},
	}
}

func newAuthSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <tenant>",
		Short: "Troca o tenant atual",
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
			fmt.Fprintf(cmd.OutOrStdout(), "Tenant atual: %s\n", args[0])
			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	var tenantOverride string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Valida o token do tenant atual (ou do informado)",
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
			fmt.Fprintf(out, "rótulo:   %s\n", strOrDash(tn.Label))
			fmt.Fprintf(out, "base_url: %s\n", tn.EffectiveBaseURL())
			fmt.Fprintf(out, "token:    %s\n", auth.EncodePeek(tok))
			if tn.DefaultUser != "" {
				name, vErr := validateUser(cmd.Context(), tn.EffectiveBaseURL(), tok, tn.DefaultUser)
				if vErr != nil {
					fmt.Fprintf(out, "usuário:  %s (ERRO — %s)\n", tn.DefaultUser, vErr)
				} else {
					fmt.Fprintf(out, "usuário:  %s (%s)\n", tn.DefaultUser, strOrDash(name))
				}
			} else {
				fmt.Fprintln(out, "usuário:  —")
			}
			if err != nil {
				fmt.Fprintf(out, "status:   ERRO — %s\n", err)
				return err
			}
			fmt.Fprintln(out, "status:   OK")
			return nil
		},
	}
	cmd.Flags().StringVar(&tenantOverride, "tenant", "", "tenant a verificar (padrão: atual)")
	return cmd
}

func newAuthSetUserCmd() *cobra.Command {
	var (
		tenantOverride string
		clear          bool
		skipVerifyUser bool
	)
	cmd := &cobra.Command{
		Use:   "set-user [<id>]",
		Short: "Define ou remove o usuário padrão (Cod. Ref.) do tenant atual",
		Long: `Define o usuário padrão que o CLI injeta como createdBy nas escritas que
exigem atribuição (ex.: tickets create, tickets actions add). Sobreponha por
comando com --user <id>.

Use --clear para remover o padrão configurado.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if clear && len(args) > 0 {
				return errors.New("--clear não pode ser combinado com um id posicional")
			}
			if !clear && len(args) != 1 {
				return errors.New("informe um id de usuário, ou use --clear para remover")
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			tn, err := cfg.Resolve(tenantOverride)
			if err != nil {
				return err
			}

			if clear {
				tn.DefaultUser = ""
				cfg.Set(tn)
				if err := cfg.Save(); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Usuário padrão removido do tenant %q\n", tn.Name)
				return nil
			}

			id := args[0]
			if !skipVerifyUser {
				tok, err := auth.ResolveToken(auth.New(), tn.Name)
				if err != nil {
					return err
				}
				name, err := validateUser(cmd.Context(), tn.EffectiveBaseURL(), tok, id)
				if err != nil {
					return fmt.Errorf("validar usuário %q: %w (use --skip-verify-user para pular)", id, err)
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "Usuário padrão: %s (%s)\n", id, strOrDash(name))
			}
			tn.DefaultUser = id
			cfg.Set(tn)
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Usuário padrão %q salvo no tenant %q\n", id, tn.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&tenantOverride, "tenant", "", "tenant a atualizar (padrão: atual)")
	cmd.Flags().BoolVar(&clear, "clear", false, "remove o usuário padrão configurado")
	cmd.Flags().BoolVar(&skipVerifyUser, "skip-verify-user", false, "pula a checagem de existência")
	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	var (
		tenantOverride string
		all            bool
	)
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove o token armazenado de um tenant (e opcionalmente o registro do tenant)",
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
					return errors.New("nenhum tenant informado e nenhum tenant atual definido")
				}
				targets = []string{name}
			}
			for _, n := range targets {
				if err := store.Delete(n); err != nil && !errors.Is(err, auth.ErrNotFound) {
					return fmt.Errorf("excluir token de %s: %w", n, err)
				}
				cfg.Delete(n)
				fmt.Fprintf(cmd.OutOrStdout(), "Logout do tenant %q realizado\n", n)
			}
			return cfg.Save()
		},
	}
	cmd.Flags().StringVar(&tenantOverride, "tenant", "", "tenant para fazer logout (padrão: atual)")
	cmd.Flags().BoolVar(&all, "all", false, "faz logout de todos os tenants configurados")
	return cmd
}

func newAuthTokenCmd() *cobra.Command {
	var tenantOverride string
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Imprime o token de um tenant no stdout (use com cuidado; para piping)",
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
	cmd.Flags().StringVar(&tenantOverride, "tenant", "", "tenant (padrão: atual)")
	return cmd
}

func strOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
