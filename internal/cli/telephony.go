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
		Short: "Dispatch Movidesk call events (asterisk_*)",
		Long: `These commands fire telephony events at Movidesk so a phone system
integration can attach calls to tickets. Two flavors:

  queue       POST /asterisk_<event>      (with queue control)
  nonqueue    GET  /asterisk_<event>      (without queue control)
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
		Short: "POST a queue-controlled call event (--event receivedCall|transferedCall|completedCall|lostCall|canceledCall)",
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := buildTelephonyBody(file, sets)
			if err != nil {
				return err
			}
			if event == "" {
				return errors.New("--event is required")
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
	cmd.Flags().StringVar(&event, "event", "", "event name (receivedCall, transferedCall, completedCall, lostCall, canceledCall)")
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to JSON body")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "override fields, e.g. --set id=abc --set queueId=1")
	return cmd
}

func newTelephonyNonQueueCmd() *cobra.Command {
	var (
		event  string
		params []string
	)
	cmd := &cobra.Command{
		Use:   "nonqueue",
		Short: "GET a no-queue call event (--event startTransferedCall|completedCall|startCanceledCall)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if event == "" {
				return errors.New("--event is required")
			}
			v := url.Values{}
			for _, kv := range params {
				k, val, ok := strings.Cut(kv, "=")
				if !ok {
					return fmt.Errorf("--param must be key=value, got %q", kv)
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
	cmd.Flags().StringVar(&event, "event", "", "event name (startTransferedCall, completedCall, startCanceledCall)")
	cmd.Flags().StringSliceVar(&params, "param", nil, "query param key=value (repeatable), e.g. --param id=abc --param branchLine=1001")
	return cmd
}

func newTelephonyMadeCallLinkCmd() *cobra.Command {
	var (
		file string
		sets []string
	)
	cmd := &cobra.Command{
		Use:   "made-call-link",
		Short: "POST /setMadeCallLink — attach a recording link to an outbound call",
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
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to JSON body")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "override fields")
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
			return nil, fmt.Errorf("parse body: %w", err)
		}
	}
	for _, kv := range sets {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return nil, fmt.Errorf("--set must be key=value, got %q", kv)
		}
		body[k] = parseSetValue(v)
	}
	return body, nil
}
