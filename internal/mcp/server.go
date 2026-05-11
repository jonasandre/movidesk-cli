package mcp

import (
	"context"
	"io"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
	"github.com/jonasandre/movidesk-cli/internal/version"
)

// Config carries the metadata the MCP server needs that cannot be derived from
// the *movidesk.Client itself. It is populated by internal/cli/mcp.go from the
// same tenant/auth resolution path every other subcommand uses.
type Config struct {
	// Tenant is the active tenant name, surfaced via the server-info resource
	// so an LLM client can self-identify which Movidesk instance it is on.
	Tenant string

	// CustomFields is the already-serialized JSON of the per-tenant custom
	// field catalog (~/.movidesk/<tenant>/customfields.yaml). nil disables the
	// movidesk://customfields-catalog resource.
	CustomFields []byte
}

// Run starts the MCP server on the given streams and blocks until the client
// disconnects or ctx is cancelled. The stdin/stdout pair is used by chat-app
// MCP clients; the in/out arguments exist so tests can plug io.Pipe streams.
func Run(ctx context.Context, c *movidesk.Client, cfg Config, in io.Reader, out io.Writer) error {
	s := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "movidesk-cli",
		Version: version.String(),
	}, nil)

	registerTools(s, c)
	registerResources(s, c, cfg)

	return s.Run(ctx, &mcpsdk.IOTransport{
		Reader: readCloser(in),
		Writer: writeCloser(out),
	})
}

// readCloser adapts a plain io.Reader into an io.ReadCloser (the SDK requires
// closability, but the lifecycle is owned by the host process so a no-op Close
// is the right behavior).
func readCloser(r io.Reader) io.ReadCloser {
	if rc, ok := r.(io.ReadCloser); ok {
		return rc
	}
	return io.NopCloser(r)
}

// writeCloser adapts a plain io.Writer into an io.WriteCloser with a no-op
// Close, mirroring readCloser.
func writeCloser(w io.Writer) io.WriteCloser {
	if wc, ok := w.(io.WriteCloser); ok {
		return wc
	}
	return nopWriteCloser{w}
}

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }
