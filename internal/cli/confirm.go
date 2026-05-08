package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// confirm asks the user to type "yes" before a destructive action. When stdin
// is not a TTY, refuses unless --force was passed.
func confirm(cmd *cobra.Command, prompt string) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("destructive action requires --force when stdin is not a TTY")
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "%s [yes/N]: ", prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return errors.New("no input")
	}
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if answer != "yes" && answer != "y" {
		return errors.New("aborted")
	}
	return nil
}
