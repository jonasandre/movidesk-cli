package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_TicketsActions_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("$expand"), "actions")
		w.Write([]byte(`{"id":1,"actions":[{"id":1,"type":1,"description":"first"},{"id":2,"type":2,"description":"second"}]}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "actions", "list", "1", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "first")
	assert.Contains(t, out, "second")
}

func TestE2E_TicketsActions_Add(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "1", r.URL.Query().Get("id"))
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "actions", "add", "1", "--internal", "--description", "Note from CLI")
	require.NoError(t, err)
	assert.Contains(t, captured, `"actions"`)
	assert.Contains(t, captured, `"type":1`)
	assert.Contains(t, captured, `"description":"Note from CLI"`)
	assert.Contains(t, out, "OK")
}

func TestE2E_TicketsActions_Update(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "tickets", "actions", "update", "1", "--action-id", "5", "--description", "edited")
	require.NoError(t, err)
	assert.Contains(t, captured, `"id":5`)
	assert.Contains(t, captured, `"description":"edited"`)
}

func TestE2E_TicketsActions_Delete(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "tickets", "actions", "delete", "1", "--action-id", "5")
	require.NoError(t, err)
	assert.Contains(t, captured, `"isDeleted":true`)
}

func TestE2E_TicketsClients_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("$expand"), "clients")
		w.Write([]byte(`{"id":1,"clients":[{"id":"c-1","businessName":"Acme","personType":2}]}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "clients", "list", "1", "--output", "table")
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(out), "acme")
}

func TestE2E_TicketsRelations(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"parentTickets":[{"id":10}],"childrenTickets":[{"id":20},{"id":21}]}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "relations", "1", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "Pais:")
	assert.Contains(t, out, "Filhos:")
	assert.Contains(t, out, `"id":10`)
	assert.Contains(t, out, `"id":21`)
}

func TestE2E_TicketsTimeline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"actions":[{"id":1,"type":2,"description":"reply","createdDate":"2026-04-01T13:30:00Z"}],
			"statusHistories":[{"status":"Novo","changedDate":"2026-04-01T13:00:00Z"}],
			"ownerHistories":[{"ownerTeam":"T1","owner":{"businessName":"Mike"},"changedDate":"2026-04-01T13:10:00Z"}]
		}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "timeline", "1", "--output", "table")
	require.NoError(t, err)
	low := strings.ToLower(out)
	assert.Contains(t, low, "novo")
	assert.Contains(t, low, "mike")
	assert.Contains(t, low, "reply")
}

func TestE2E_TicketsAssetsList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("$expand"), "assets")
		w.Write([]byte(`{"assets":[{"id":"a-1","name":"Switch","label":"SW-01"}]}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "assets", "list", "1", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "SW-01")
}

func TestE2E_TicketsHistoriesList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ownerHistories":[{"ownerTeam":"T1","owner":{"businessName":"Mike"},"changedDate":"2026-04-01T13:00:00Z"}],"statusHistories":[{"status":"Novo","changedDate":"2026-04-01T13:00:00Z"}]}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "histories", "list", "1", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "Histórico de responsável:")
	assert.Contains(t, out, "Histórico de status:")
	assert.Contains(t, out, "Mike")
	assert.Contains(t, out, "Novo")
}
