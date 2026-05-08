package cli

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// newQueryCmd is an escape hatch for any OData-capable Movidesk endpoint not
// covered by a typed subcommand. It does not parse responses — JSON only.
func newQueryCmd() *cobra.Command {
	var (
		of     odataFlags
		method string
		params []string
	)
	cmd := &cobra.Command{
		Use:   "query <path>",
		Short: "Raw HTTP call against any Movidesk endpoint",
		Long: `Raw escape hatch. Path is relative to the API base, e.g. "tickets",
"/tickets", "/persons". Token is injected automatically.

Examples:
  movidesk-cli query /tickets --filter "id eq 1"
  movidesk-cli query /persons --select id,businessName --top 5
  movidesk-cli query /persons --method GET --param id=abc

Output is always JSON (use jq to slice further).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			path := args[0]
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}

			v := url.Values{}
			of.query().Apply(v)
			for _, kv := range params {
				k, val, ok := strings.Cut(kv, "=")
				if !ok {
					return fmt.Errorf("--param must be key=value, got %q", kv)
				}
				v.Set(k, val)
			}

			method = strings.ToUpper(method)
			if method == "" {
				method = "GET"
			}
			switch method {
			case "GET", "POST", "PATCH", "DELETE":
			default:
				return fmt.Errorf("unsupported method %q (allowed: GET, POST, PATCH, DELETE)", method)
			}
			if method != "GET" && method != "DELETE" {
				return errors.New("query only supports GET/DELETE; use the typed subcommand for write operations")
			}

			if of.all {
				if method != "GET" {
					return errors.New("--all only works with GET")
				}
				q := of.query()
				if q.Top == 0 {
					q.Top = 100
				}
				rows, err := paginateGeneric(cmd.Context(), r.client, path, q, of.max, params)
				if err != nil {
					return err
				}
				return renderRows(cmd.OutOrStdout(), rows, r.output, "", nil)
			}

			raw, err := r.client.Do(cmd.Context(), method, path, v, nil)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "", nil)
		},
	}
	of.bind(cmd)
	cmd.Flags().StringVar(&method, "method", "GET", "HTTP method (GET or DELETE only)")
	cmd.Flags().StringSliceVar(&params, "param", nil, "extra query param key=value (repeatable)")
	return cmd
}
