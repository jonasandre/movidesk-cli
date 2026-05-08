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

func TestE2E_ServicesList_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/services", r.URL.Path)
		w.Write([]byte(`[{"id":1,"name":"Suporte","isActive":true,"allowAllCategories":true}]`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "services", "list", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "Suporte")
}

func TestE2E_ServicesList_Table(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":1,"name":"Suporte","isActive":true,"allowAllCategories":true}]`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "services", "list", "--output", "table")
	require.NoError(t, err)
	low := strings.ToLower(out)
	assert.Contains(t, low, "name")
	assert.Contains(t, low, "suporte")
}

func TestE2E_ServicesGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "5712", r.URL.Query().Get("id"))
		w.Write([]byte(`{"id":5712,"name":"TI","isActive":true}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "services", "get", "5712", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, `"id":5712`)
}

func TestE2E_ServicesCreate(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":99}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t,
		"services", "create",
		"--set", `name=NovoServico`,
		"--set", "isActive=true",
		"--set", "serviceForTicketType=2",
		"--set", "isVisible=3",
		"--set", "allowSelection=3",
		"--set", "allowFinishTicket=true",
		"--set", "allowAllCategories=true",
		"--output", "json", "--compact",
	)
	require.NoError(t, err)
	assert.Contains(t, captured, `"name":"NovoServico"`)
	assert.Contains(t, out, `"id":99`)
}

func TestE2E_ServicesUpdate(t *testing.T) {
	var captured, seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		seen = r.URL.RawQuery
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "services", "update", "1", "--set", "name=Renamed")
	require.NoError(t, err)
	assert.Contains(t, captured, `"name":"Renamed"`)
	assert.Contains(t, seen, "id=1")
	assert.Contains(t, out, "OK")
}

func TestE2E_ServicesDelete_Force(t *testing.T) {
	var method, seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		seen = r.URL.RawQuery
		w.WriteHeader(200)
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "services", "delete", "1", "--force")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", method)
	assert.Contains(t, seen, "id=1")
	assert.Contains(t, out, "OK")
}

func TestE2E_ServicesList_AllPaginates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("$skip") {
		case "", "0":
			w.Write([]byte(`[{"id":1,"name":"a"},{"id":2,"name":"b"}]`))
		case "2":
			w.Write([]byte(`[{"id":3,"name":"c"}]`))
		default:
			w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "services", "list", "--all", "--top", "2", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, `"name":"c"`)
}
