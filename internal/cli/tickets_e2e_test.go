package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonasandre/movidesk-cli/internal/auth"
	"github.com/jonasandre/movidesk-cli/internal/config"
)

// setupTenant creates a temporary MOVIDESK_HOME with a single tenant pointing
// at the supplied test server. Returns the cleanup-aware home dir.
func setupTenant(t *testing.T, srvURL string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("MOVIDESK_HOME", dir)
	t.Setenv(auth.EnvPassphrase, "e2e")
	t.Setenv(auth.EnvToken, "tok")

	cfg := &config.Config{
		CurrentTenant: "test",
		Tenants: map[string]*config.Tenant{
			"test": {Name: "test", Label: "Test", BaseURL: srvURL},
		},
	}
	require.NoError(t, cfg.Save())
}

func runCmd(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	root := newRootCmd()
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs(args)
	root.SetContext(context.Background())
	// Reset persistent flag state across tests in the same process.
	flags = globalFlags{}
	err := root.Execute()
	return out.String(), errBuf.String(), err
}

func TestE2E_TicketsList_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tickets", r.URL.Path)
		assert.Equal(t, "tok", r.URL.Query().Get("token"))
		assert.Equal(t, "id eq 1", r.URL.Query().Get("$filter"))
		w.Write([]byte(`[{"id":1,"subject":"hello","status":"Novo"}]`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "list", "--filter", "id eq 1", "--output", "json", "--compact")
	require.NoError(t, err)

	var v []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &v))
	require.Len(t, v, 1)
	assert.Equal(t, float64(1), v[0]["id"])
	assert.Equal(t, "hello", v[0]["subject"])
}

func TestE2E_TicketsList_TableUsesDefaultColumns(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":1,"subject":"hello","status":"Novo","owner":{"businessName":"Joe"},"lastUpdate":"2026-01-01"}]`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "list", "--output", "table")
	require.NoError(t, err)
	low := strings.ToLower(out)
	assert.Contains(t, low, "owner.businessname")
	assert.Contains(t, low, "joe")
}

func TestE2E_TicketsList_AllPaginates_TableOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("$skip") {
		case "", "0":
			w.Write([]byte(`[{"id":1,"subject":"first"},{"id":2,"subject":"second"}]`))
		case "2":
			w.Write([]byte(`[{"id":3,"subject":"third"}]`))
		default:
			w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "list", "--all", "--top", "2", "--output", "table")
	require.NoError(t, err)
	low := strings.ToLower(out)
	assert.NotContains(t, low, "no rows")
	assert.Contains(t, low, "first")
	assert.Contains(t, low, "second")
	assert.Contains(t, low, "third")
}

func TestE2E_TicketsList_AllPaginates(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch r.URL.Query().Get("$skip") {
		case "", "0":
			w.Write([]byte(`[{"id":1},{"id":2}]`))
		case "2":
			w.Write([]byte(`[{"id":3}]`))
		default:
			w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "list", "--all", "--top", "2", "--output", "json", "--compact")
	require.NoError(t, err)
	var v []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &v))
	assert.Len(t, v, 3)
}

func TestE2E_TicketsGet_ByID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tickets", r.URL.Path)
		assert.Equal(t, "42", r.URL.Query().Get("id"))
		w.Write([]byte(`{"id":42,"subject":"foo"}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "get", "42", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, `"id":42`)
}

func TestE2E_TicketsGet_ByProtocol(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "MOVI202109000001", r.URL.Query().Get("protocol"))
		w.Write([]byte(`{"id":1,"protocol":"MOVI202109000001"}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "get", "--protocol", "MOVI202109000001", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "MOVI202109000001")
}

func TestE2E_TicketsCreate_Set(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Hello", got["subject"])
		assert.Equal(t, float64(2), got["type"])
		w.WriteHeader(201)
		w.Write([]byte(`{"id":99}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t,
		"tickets", "create",
		"--set", "type=2",
		"--set", "subject=Hello",
		"--output", "json", "--compact",
	)
	require.NoError(t, err)
	assert.Contains(t, out, `"id":99`)
}

func TestE2E_TicketsCreate_Template(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), `"subject":"From template"`)
		assert.Contains(t, string(body), `"category":"Suporte"`)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	dir, err := config.Dir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "templates"), 0o700))
	tmpl := []byte(`{"type":2,"subject":"From template","category":"Suporte"}`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "templates", "support.json"), tmpl, 0o600))

	_, _, err = runCmd(t, "tickets", "create", "--from-template", "support", "--output", "json", "--compact")
	require.NoError(t, err)
}

func TestE2E_TicketsUpdate_PatchSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "5", r.URL.Query().Get("id"))
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), `"subject":"Updated"`)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "update", "5", "--set", "subject=Updated")
	require.NoError(t, err)
	assert.Contains(t, out, "OK")
}

func TestE2E_TicketsHTML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tickets/htmldescription", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("id"))
		assert.Equal(t, "3", r.URL.Query().Get("actionId"))
		w.Write([]byte(`{"id":1,"description":"<p>hi</p>"}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "html", "1", "--action-id", "3", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "<p>hi</p>")
}

func TestE2E_TicketsAttach(t *testing.T) {
	var ct string
	var ticket, action string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/ticketFileUpload", r.URL.Path)
		ct = r.Header.Get("Content-Type")
		ticket = r.URL.Query().Get("id")
		action = r.URL.Query().Get("actionId")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	tmpFile := filepath.Join(t.TempDir(), "report.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("hello"), 0o600))

	_, _, err := runCmd(t, "tickets", "attach", "12", "--action-id", "34", "--file", tmpFile, "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Equal(t, "12", ticket)
	assert.Equal(t, "34", action)
	assert.Contains(t, ct, "multipart/form-data; boundary=")
}

func TestE2E_TicketsPastList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tickets/past", r.URL.Path)
		w.Write([]byte(`[{"id":99,"subject":"old"}]`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "past", "list", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "old")
}

func TestE2E_TicketsMergedList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tickets/merged", r.URL.Path)
		w.Write([]byte(`[{"id":7,"subject":"merged"}]`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "merged", "list", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "merged")
}
