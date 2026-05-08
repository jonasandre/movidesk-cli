package tickets

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetCustomFieldValue_ReadMergePatch_AppendsNew(t *testing.T) {
	var patchBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte(`{"customFieldValues":[{"customFieldId":1,"customFieldRuleId":1,"line":1,"value":"existing"}]}`))
		case "PATCH":
			patchBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(200)
		}
	}))
	defer srv.Close()
	svc, _ := newSvcWith(t, nil)
	svc.C.BaseURL = srv.URL // overwrite

	_, err := svc.SetCustomFieldValue(context.Background(), 1, CustomFieldValue{
		CustomFieldID: 2, CustomFieldRuleID: 1, Line: 1, Value: "added",
	})
	require.NoError(t, err)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(patchBody, &sent))
	cfvs, _ := sent["customFieldValues"].([]any)
	require.Len(t, cfvs, 2)
	// Existing preserved + new appended.
	assert.Contains(t, string(patchBody), `"value":"existing"`)
	assert.Contains(t, string(patchBody), `"value":"added"`)
}

func TestSetCustomFieldValue_ReplacesMatchingKey(t *testing.T) {
	var patchBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte(`{"customFieldValues":[{"customFieldId":1,"customFieldRuleId":1,"line":1,"value":"old"},{"customFieldId":2,"customFieldRuleId":1,"line":1,"value":"keep"}]}`))
		case "PATCH":
			patchBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(200)
		}
	}))
	defer srv.Close()
	svc, _ := newSvcWith(t, nil)
	svc.C.BaseURL = srv.URL

	_, err := svc.SetCustomFieldValue(context.Background(), 1, CustomFieldValue{
		CustomFieldID: 1, CustomFieldRuleID: 1, Line: 1, Value: "new",
	})
	require.NoError(t, err)
	assert.NotContains(t, string(patchBody), `"value":"old"`)
	assert.Contains(t, string(patchBody), `"value":"new"`)
	assert.Contains(t, string(patchBody), `"value":"keep"`)
}

func TestClearCustomFieldValue_RemovesMatchingKey(t *testing.T) {
	var patchBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte(`{"customFieldValues":[{"customFieldId":1,"customFieldRuleId":1,"line":1,"value":"a"},{"customFieldId":2,"customFieldRuleId":1,"line":1,"value":"keep"}]}`))
		case "PATCH":
			patchBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(200)
		}
	}))
	defer srv.Close()
	svc, _ := newSvcWith(t, nil)
	svc.C.BaseURL = srv.URL

	_, err := svc.ClearCustomFieldValue(context.Background(), 1, 1, 0, 0)
	require.NoError(t, err)
	assert.NotContains(t, string(patchBody), `"value":"a"`)
	assert.Contains(t, string(patchBody), `"value":"keep"`)
}

func TestClearCustomFieldValue_FilterByLine(t *testing.T) {
	var patchBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte(`{"customFieldValues":[{"customFieldId":1,"customFieldRuleId":1,"line":1,"value":"a"},{"customFieldId":1,"customFieldRuleId":1,"line":2,"value":"b"}]}`))
		case "PATCH":
			patchBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(200)
		}
	}))
	defer srv.Close()
	svc, _ := newSvcWith(t, nil)
	svc.C.BaseURL = srv.URL

	_, err := svc.ClearCustomFieldValue(context.Background(), 1, 1, 0, 1)
	require.NoError(t, err)
	assert.NotContains(t, string(patchBody), `"value":"a"`)
	assert.Contains(t, string(patchBody), `"value":"b"`)
}

func TestSetCustomFieldValue_ValidatesInput(t *testing.T) {
	svc, srv := newSvcWith(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()
	_, err := svc.SetCustomFieldValue(context.Background(), 1, CustomFieldValue{})
	require.Error(t, err)
	_, err = svc.SetCustomFieldValue(context.Background(), 1, CustomFieldValue{CustomFieldID: 1})
	require.Error(t, err)
}

func TestMergeCustomFields_PreservesOrder(t *testing.T) {
	cur := []CustomFieldValue{
		{CustomFieldID: 1, CustomFieldRuleID: 1, Line: 1, Value: "x"},
		{CustomFieldID: 2, CustomFieldRuleID: 1, Line: 1, Value: "y"},
	}
	got := mergeCustomFields(cur, CustomFieldValue{CustomFieldID: 1, CustomFieldRuleID: 1, Line: 1, Value: "z"})
	require.Len(t, got, 2)
	assert.Equal(t, "z", got[0].Value)
	assert.Equal(t, "y", got[1].Value)
}
