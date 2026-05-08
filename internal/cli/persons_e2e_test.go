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

func TestE2E_PersonsList_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/persons", r.URL.Path)
		w.Write([]byte(`[{"id":"1","businessName":"Joe","isActive":true,"personType":1,"profileType":2}]`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "persons", "list", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "Joe")
}

func TestE2E_PersonsList_TableUsesDefaultColumns(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":"1","businessName":"Joe","personType":1,"profileType":2,"isActive":true}]`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "persons", "list", "--output", "table")
	require.NoError(t, err)
	low := strings.ToLower(out)
	assert.Contains(t, low, "businessname")
	assert.Contains(t, low, "joe")
}

func TestE2E_PersonsGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "abc", r.URL.Query().Get("id"))
		w.Write([]byte(`{"id":"abc","businessName":"X","isActive":true}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "persons", "get", "abc", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, `"id":"abc"`)
}

func TestE2E_PersonsCreate_Set(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), `"businessName":"Acme"`)
		assert.Contains(t, string(body), `"personType":2`)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":"acme-1"}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t,
		"persons", "create",
		"--set", "personType=2",
		"--set", "profileType=2",
		"--set", "isActive=true",
		"--set", `businessName=Acme`,
		"--output", "json", "--compact",
	)
	require.NoError(t, err)
	assert.Contains(t, out, `"id":"acme-1"`)
}

func TestE2E_PersonsUpdate_PatchSet(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "abc", r.URL.Query().Get("id"))
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "persons", "update", "abc", "--set", "businessName=Updated")
	require.NoError(t, err)
	assert.Contains(t, captured, `"businessName":"Updated"`)
	assert.Contains(t, out, "OK")
}

func TestE2E_PersonsDelete_Force(t *testing.T) {
	var method, seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		seen = r.URL.RawQuery
		w.WriteHeader(200)
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "persons", "delete", "abc", "--force")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", method)
	assert.Contains(t, seen, "id=abc")
	assert.Contains(t, out, "OK")
}

func TestE2E_PersonsDelete_RefusesNonTTYWithoutForce(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be hit")
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "persons", "delete", "abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--force")
}

func TestE2E_PersonsList_AllPaginates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("$skip") {
		case "", "0":
			w.Write([]byte(`[{"id":"1"},{"id":"2"}]`))
		case "2":
			w.Write([]byte(`[{"id":"3"}]`))
		default:
			w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "persons", "list", "--all", "--top", "2", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, `"id":"3"`)
}

func TestE2E_PersonsCFSet_ReadMergePatch(t *testing.T) {
	var patchBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte(`{"customFieldValues":[{"customFieldId":1,"customFieldRuleId":1,"line":1,"value":"keep"}]}`))
		case "PATCH":
			patchBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(200)
		}
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t,
		"persons", "customfields", "set", "abc",
		"--field", "2",
		"--rule", "1",
		"--value", "added",
	)
	require.NoError(t, err)
	assert.Contains(t, out, "OK")
	assert.Contains(t, string(patchBody), `"value":"keep"`)
	assert.Contains(t, string(patchBody), `"value":"added"`)
}

func TestE2E_PersonsCFShow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("$expand"), "customFieldValues")
		w.Write([]byte(`{"customFieldValues":[{"customFieldId":1,"customFieldRuleId":1,"line":1,"value":"x"}]}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "persons", "customfields", "show", "abc", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, `"value":"x"`)
}
