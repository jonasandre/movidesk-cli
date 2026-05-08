// gen-docs walks every Cobra command and writes Markdown reference pages.
// Output goes under docs/cli/ to keep the public README focused.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra/doc"

	"github.com/jonasandre/movidesk-cli/internal/cli"
)

func main() {
	out := "docs/cli"
	if v := os.Getenv("DOCS_OUT"); v != "" {
		out = v
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		fail(err)
	}
	root := cli.NewRootForDocs()
	root.DisableAutoGenTag = true
	if err := doc.GenMarkdownTree(root, out); err != nil {
		fail(err)
	}
	abs, _ := filepath.Abs(out)
	fmt.Fprintf(os.Stderr, "wrote docs to %s\n", abs)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gen-docs:", err)
	os.Exit(1)
}
