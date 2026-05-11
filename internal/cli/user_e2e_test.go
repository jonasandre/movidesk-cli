package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonasandre/movidesk-cli/internal/config"
)

// withDefaultUser writes the tenant config to give it a default_user without
// going through `auth login` (which requires TTY for the prompt).
func withDefaultUser(t *testing.T, srvURL, userID string) {
	t.Helper()
	setupTenant(t, srvURL)
	cfg, err := config.Load()
	require.NoError(t, err)
	cfg.Tenants["test"].DefaultUser = userID
	require.NoError(t, cfg.Save())
}

func TestE2E_TicketsCreate_AutoInjectsCreatedBy(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()
	withDefaultUser(t, srv.URL, "u-bot")

	_, _, err := runCmd(t,
		"tickets", "create",
		"--set", "type=2", "--set", "subject=Hello",
		"--output", "json", "--compact",
	)
	require.NoError(t, err)
	assert.Contains(t, captured, `"createdBy":{"id":"u-bot"}`)
}

func TestE2E_TicketsCreate_OverrideUserFlag(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()
	withDefaultUser(t, srv.URL, "u-bot")

	_, _, err := runCmd(t,
		"tickets", "create",
		"--user", "u-other",
		"--set", "type=2", "--set", "subject=Hello",
		"--output", "json", "--compact",
	)
	require.NoError(t, err)
	assert.Contains(t, captured, `"createdBy":{"id":"u-other"}`)
	assert.NotContains(t, captured, "u-bot")
}

func TestE2E_TicketsCreate_ExplicitCreatedByWins(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()
	withDefaultUser(t, srv.URL, "u-bot")

	// Explicit createdBy via JSON body (--set CSV parser can't handle embedded
	// quotes well, but the file flow proves the precedence rule).
	dir := t.TempDir()
	bodyFile := dir + "/body.json"
	require.NoError(t, os.WriteFile(bodyFile, []byte(`{"type":2,"subject":"x","createdBy":{"id":"u-explicit"}}`), 0o600))

	_, _, err := runCmd(t,
		"tickets", "create",
		"--file", bodyFile,
		"--output", "json", "--compact",
	)
	require.NoError(t, err)
	assert.Contains(t, captured, "u-explicit")
	assert.NotContains(t, captured, "u-bot")
}

func TestE2E_TicketsCreate_NoUserWhenUnset(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t,
		"tickets", "create",
		"--set", "type=2", "--set", "subject=Hi",
		"--output", "json", "--compact",
	)
	require.NoError(t, err)
	assert.NotContains(t, captured, "createdBy")
}

func TestE2E_ActionsAdd_AutoInjectsCreatedBy(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	withDefaultUser(t, srv.URL, "u-bot")

	_, _, err := runCmd(t,
		"tickets", "actions", "add", "1",
		"--internal", "--description", "note",
	)
	require.NoError(t, err)
	assert.Contains(t, captured, `"createdBy":{"id":"u-bot"}`)
}

func TestE2E_ActionsUpdate_DoesNotAutoInject(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	withDefaultUser(t, srv.URL, "u-bot")

	_, _, err := runCmd(t,
		"tickets", "actions", "update", "1",
		"--action-id", "5",
		"--description", "edited",
	)
	require.NoError(t, err)
	assert.NotContains(t, captured, "createdBy")
}

func TestE2E_AuthSetUser_ValidatesAndSaves(t *testing.T) {
	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/persons", r.URL.Path)
		seen = r.URL.RawQuery
		w.Write([]byte(`{"id":"u-bot","businessName":"Acme Bot"}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "auth", "set-user", "u-bot")
	require.NoError(t, err)
	assert.Contains(t, seen, "id=u-bot")
	assert.Contains(t, out, `Usuário padrão "u-bot" salvo`)

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "u-bot", cfg.Tenants["test"].DefaultUser)
}

func TestE2E_AuthSetUser_Clear(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be hit on --clear")
	}))
	defer srv.Close()
	withDefaultUser(t, srv.URL, "u-bot")

	out, _, err := runCmd(t, "auth", "set-user", "--clear")
	require.NoError(t, err)
	assert.Contains(t, out, "Usuário padrão removido")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Empty(t, cfg.Tenants["test"].DefaultUser)
}

func TestE2E_AuthSetUser_RejectsMissing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "auth", "set-user", "u-ghost")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "u-ghost")
}

func TestE2E_AuthSetUser_SkipVerify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be hit with --skip-verify-user")
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "auth", "set-user", "u-future", "--skip-verify-user")
	require.NoError(t, err)
}

func TestE2E_AuthList_ShowsUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	withDefaultUser(t, srv.URL, "u-bot")

	out, _, err := runCmd(t, "auth", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "user=u-bot")
}

func TestE2E_EnvUser_OverridesTenant(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()
	withDefaultUser(t, srv.URL, "u-tenant-default")
	t.Setenv(EnvUser, "u-from-env")

	_, _, err := runCmd(t,
		"tickets", "create",
		"--set", "type=2", "--set", "subject=Hi",
		"--output", "json", "--compact",
	)
	require.NoError(t, err)
	assert.Contains(t, captured, "u-from-env")
	assert.NotContains(t, captured, "u-tenant-default")
}
