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
		Short: "Chamada HTTP bruta contra qualquer endpoint do Movidesk",
		Long: `Escape hatch genérico. O caminho é relativo à base da API, ex.: "tickets",
"/tickets", "/persons". O token é injetado automaticamente.

Sintaxe completa de --filter / --select / --orderby em:
  movidesk-cli topics filters

O formato de saída segue --output (json|table|csv); use -o json com jq pra fatiar mais.`,
		Example: `  movidesk-cli query /tickets --filter "id eq 1"
  movidesk-cli query /persons --select id,businessName --top 5
  movidesk-cli query /persons --method GET --param id=abc`,
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
					return fmt.Errorf("--param deve ser chave=valor, recebido %q", kv)
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
				return fmt.Errorf("método não suportado %q (permitidos: GET, POST, PATCH, DELETE)", method)
			}
			if method != "GET" && method != "DELETE" {
				return errors.New("query só suporta GET/DELETE; use o subcomando tipado pra escritas")
			}

			if of.all {
				if method != "GET" {
					return errors.New("--all só funciona com GET")
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
	cmd.Flags().StringVar(&method, "method", "GET", "método HTTP (apenas GET ou DELETE)")
	cmd.Flags().StringSliceVar(&params, "param", nil, "query param extra chave=valor (repetível)")
	return cmd
}
