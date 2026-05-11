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

func TestE2E_CFShow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("$expand"), "customFieldValues")
		w.Write([]byte(`{"customFieldValues":[{"customFieldId":1,"customFieldRuleId":1,"line":1,"value":"high"},{"customFieldId":2,"customFieldRuleId":1,"line":1,"items":[{"customFieldItem":"Alta"}]}]}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "customfields", "show", "1", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, `"value":"high"`)
	assert.Contains(t, out, "Alta")
}

func TestE2E_CFSet_ReadMergePatch(t *testing.T) {
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
		"tickets", "customfields", "set", "1",
		"--field", "2",
		"--rule", "1",
		"--value", "added",
	)
	require.NoError(t, err)
	assert.Contains(t, out, "OK")
	assert.Contains(t, string(patchBody), `"value":"keep"`)
	assert.Contains(t, string(patchBody), `"value":"added"`)
}

func TestE2E_CFSet_WithLabel(t *testing.T) {
	var patchBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte(`{"customFieldValues":[]}`))
		case "PATCH":
			patchBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(200)
		}
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	// Register catalog entry first.
	_, _, err := runCmd(t,
		"tickets", "customfields", "catalog", "add",
		"--label", "Severidade",
		"--field", "125529",
		"--rule", "7",
		"--type", "list-of-values",
		"--options", "Baixa,Alta",
	)
	require.NoError(t, err)

	_, _, err = runCmd(t,
		"tickets", "customfields", "set", "1",
		"--field-label", "Severidade",
		"--item", "Alta",
	)
	require.NoError(t, err)
	assert.Contains(t, string(patchBody), `"customFieldId":125529`)
	assert.Contains(t, string(patchBody), `"customFieldRuleId":7`)
	assert.Contains(t, string(patchBody), `"customFieldItem":"Alta"`)
}

func TestE2E_CFClear(t *testing.T) {
	var patchBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte(`{"customFieldValues":[{"customFieldId":1,"customFieldRuleId":1,"line":1,"value":"goaway"},{"customFieldId":2,"customFieldRuleId":1,"line":1,"value":"keep"}]}`))
		case "PATCH":
			patchBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(200)
		}
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "tickets", "customfields", "clear", "1", "--field", "1", "--rule", "1")
	require.NoError(t, err)
	assert.Contains(t, out, "OK")
	assert.NotContains(t, string(patchBody), `"value":"goaway"`)
	assert.Contains(t, string(patchBody), `"value":"keep"`)
}

func TestE2E_CFCatalog_AddListRemove(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t,
		"tickets", "customfields", "catalog", "add",
		"--label", "Squad", "--field", "300", "--rule", "9", "--type", "single-select",
		"--options", "Platform,Growth,Mobile",
	)
	require.NoError(t, err)
	assert.Contains(t, out, "salvo")

	listOut, _, err := runCmd(t, "tickets", "customfields", "catalog", "list", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(listOut), "squad")
	assert.Contains(t, listOut, "300")

	removeOut, _, err := runCmd(t, "tickets", "customfields", "catalog", "remove", "--label", "Squad")
	require.NoError(t, err)
	assert.Contains(t, removeOut, "removido")

	listOut2, _, err := runCmd(t, "tickets", "customfields", "catalog", "list", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.NotContains(t, strings.ToLower(listOut2), "squad")
}

func TestE2E_CFSet_ValidatesInputs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "tickets", "customfields", "set", "1", "--field", "1")
	require.Error(t, err)
}
