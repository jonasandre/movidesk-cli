package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPCommand_RegisteredOnRoot ensures the `mcp` subcommand stays wired
// into the root cobra tree. The actual stdio loop is exercised end-to-end in
// internal/mcp/server_test.go via in-memory transports; this test only guards
// against accidentally removing the cobra registration.
func TestMCPCommand_RegisteredOnRoot(t *testing.T) {
	root := newRootCmd()
	cmd, _, err := root.Find([]string{"mcp"})
	require.NoError(t, err)
	require.NotNil(t, cmd)
	assert.Equal(t, "mcp", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}
