package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/customfields"
)

func newCustomFieldsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "customfields",
		Short: "Manage list-type custom field option pools (tenant-wide)",
		Long: `These commands wrap the /ticketCustomFieldValue/{InsertValues,UpdateValues,DeleteValues}
endpoints, which manage the OPTION POOL of list-type custom fields — the set
of values agents can pick from in the dropdown.

To set a value on a SPECIFIC ticket or person, use:
  tickets customfields set ...
  persons customfields set ...
`,
	}
	cmd.AddCommand(newCustomFieldsOptionsCmd())
	return cmd
}

func newCustomFieldsOptionsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "options", Short: "Add/rename/remove option-pool values"}
	cmd.AddCommand(
		newCustomFieldsOptionsAddCmd(),
		newCustomFieldsOptionsRenameCmd(),
		newCustomFieldsOptionsRemoveCmd(),
	)
	return cmd
}

func newCustomFieldsOptionsAddCmd() *cobra.Command {
	var (
		fieldID string
		values  []string
	)
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Insert option values into a list-type field's pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fieldID == "" || len(values) == 0 {
				return errors.New("--field and --value (repeatable) are required")
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := customfields.New(r.client).AddOptions(cmd.Context(), fieldID, values)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "", nil)
		},
	}
	cmd.Flags().StringVar(&fieldID, "field", "", "numeric customFieldId (required)")
	cmd.Flags().StringSliceVar(&values, "value", nil, "option value (repeatable)")
	return cmd
}

func newCustomFieldsOptionsRenameCmd() *cobra.Command {
	var (
		fieldID string
		pairs   []string
	)
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename existing option values via --pair OLD=NEW (repeatable)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fieldID == "" || len(pairs) == 0 {
				return errors.New("--field and --pair (repeatable) are required")
			}
			parsed := make([]customfields.UpdatePair, 0, len(pairs))
			for _, p := range pairs {
				old, neu, ok := strings.Cut(p, "=")
				if !ok {
					return fmt.Errorf("--pair must be OLD=NEW, got %q", p)
				}
				parsed = append(parsed, customfields.UpdatePair{OldName: old, NewName: neu})
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := customfields.New(r.client).RenameOptions(cmd.Context(), fieldID, parsed)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "", nil)
		},
	}
	cmd.Flags().StringVar(&fieldID, "field", "", "numeric customFieldId (required)")
	cmd.Flags().StringSliceVar(&pairs, "pair", nil, "OLD=NEW (repeatable)")
	return cmd
}

func newCustomFieldsOptionsRemoveCmd() *cobra.Command {
	var (
		fieldID string
		values  []string
	)
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove option values from a list-type field's pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fieldID == "" || len(values) == 0 {
				return errors.New("--field and --value (repeatable) are required")
			}
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}
			raw, err := customfields.New(r.client).RemoveOptions(cmd.Context(), fieldID, values)
			if err != nil {
				return err
			}
			return renderJSON(cmd.OutOrStdout(), raw, r.output, "", nil)
		},
	}
	cmd.Flags().StringVar(&fieldID, "field", "", "numeric customFieldId (required)")
	cmd.Flags().StringSliceVar(&values, "value", nil, "option value (repeatable)")
	return cmd
}
