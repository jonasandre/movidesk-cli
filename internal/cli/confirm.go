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
		return errors.New("ação destrutiva exige --force quando stdin não é um TTY")
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "%s [sim/N]: ", prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return errors.New("sem entrada")
	}
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if answer != "sim" && answer != "s" && answer != "yes" && answer != "y" {
		return errors.New("cancelado")
	}
	return nil
}
