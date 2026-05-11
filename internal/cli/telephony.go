package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/telephony"
)

func newTelephonyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "telephony",
		Short: "Dispara eventos de chamada do Movidesk (asterisk_*)",
		Long: `Estes comandos disparam eventos de telefonia no Movidesk para que uma
integração de telefonia possa vincular chamadas a chamados. Duas variantes:

  queue       POST /asterisk_<evento>      (com controle de fila)
  nonqueue    GET  /asterisk_<evento>      (sem controle de fila)
`,
	}
	cmd.AddCommand(
		newTelephonyQueueCmd(),
		newTelephonyNonQueueCmd(),
		newTelephonyMadeCallLinkCmd(),
	)
	return cmd
}

func newTelephonyQueueCmd() *cobra.Command {
	var (
		event string
		file  string
		sets  []string
	)
	cmd := &cobra.Command{
		Use:   "queue",
		Short: "POST de evento de chamada com controle de fila (--event receivedCall|transferedCall|completedCall|lostCall|canceledCall)",
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := buildTelephonyBody(file, sets)
			if err != nil {
				return err
			}
			if event == "" {
				return errors.New("--event é obrigatório")
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := telephony.New(r.client).QueuePost(cmd.Context(), event, body)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "", nil)
		},
	}
	cmd.Flags().StringVar(&event, "event", "", "nome do evento (receivedCall, transferedCall, completedCall, lostCall, canceledCall)")
	cmd.Flags().StringVarP(&file, "file", "f", "", "caminho do corpo JSON")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "sobrescreve campos, ex.: --set id=abc --set queueId=1")
	return cmd
}

func newTelephonyNonQueueCmd() *cobra.Command {
	var (
		event  string
		params []string
	)
	cmd := &cobra.Command{
		Use:   "nonqueue",
		Short: "GET de evento de chamada sem controle de fila (--event startTransferedCall|completedCall|startCanceledCall)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if event == "" {
				return errors.New("--event é obrigatório")
			}
			v := url.Values{}
			for _, kv := range params {
				k, val, ok := strings.Cut(kv, "=")
				if !ok {
					return fmt.Errorf("--param deve ser chave=valor, recebido %q", kv)
				}
				v.Set(k, val)
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := telephony.New(r.client).NonQueueGet(cmd.Context(), event, v)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "", nil)
		},
	}
	cmd.Flags().StringVar(&event, "event", "", "nome do evento (startTransferedCall, completedCall, startCanceledCall)")
	cmd.Flags().StringSliceVar(&params, "param", nil, "query param chave=valor (repetível), ex.: --param id=abc --param branchLine=1001")
	return cmd
}

func newTelephonyMadeCallLinkCmd() *cobra.Command {
	var (
		file string
		sets []string
	)
	cmd := &cobra.Command{
		Use:   "made-call-link",
		Short: "POST /setMadeCallLink — vincula um link de gravação a uma chamada de saída",
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := buildTelephonyBody(file, sets)
			if err != nil {
				return err
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := telephony.New(r.client).SetMadeCallLink(cmd.Context(), body)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "", nil)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "caminho do corpo JSON")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "sobrescreve campos")
	return cmd
}

func buildTelephonyBody(file string, sets []string) (map[string]any, error) {
	body := map[string]any{}
	if file != "" {
		raw, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, fmt.Errorf("interpretar corpo: %w", err)
		}
	}
	for _, kv := range sets {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return nil, fmt.Errorf("--set deve ser chave=valor, recebido %q", kv)
		}
		body[k] = parseSetValue(v)
	}
	return body, nil
}
