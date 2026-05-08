package persons

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
)

func newSvc(t *testing.T, h http.HandlerFunc) (*Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	c := movidesk.New(srv.URL, "tok")
	c.Limiter = movidesk.NewLimiter(1000, 0)
	return New(c), srv
}

func TestList_BuildsODataQuery(t *testing.T) {
	var seen string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		w.Write([]byte(`[]`))
	})
	defer srv.Close()

	q := odata.Query{Filter: "isActive eq true", Top: 5}
	_, err := svc.List(context.Background(), q)
	require.NoError(t, err)
	assert.Contains(t, seen, "%24filter=isActive+eq+true")
	assert.Contains(t, seen, "%24top=5")
	assert.Contains(t, seen, "token=tok")
}

func TestGet(t *testing.T) {
	var seen string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		w.Write([]byte(`{"id":"42","businessName":"Joe","isActive":true,"personType":1,"profileType":2}`))
	})
	defer srv.Close()

	body, err := svc.Get(context.Background(), "42")
	require.NoError(t, err)
	assert.Contains(t, seen, "id=42")
	var p Person
	require.NoError(t, json.Unmarshal(body, &p))
	assert.Equal(t, "Joe", p.BusinessName)
	assert.True(t, p.IsActive)
}

func TestCreate(t *testing.T) {
	var got string
	var seen string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		seen = r.URL.RawQuery
		buf, _ := io.ReadAll(r.Body)
		got = string(buf)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":"abc"}`))
	})
	defer srv.Close()

	body, err := svc.Create(context.Background(), map[string]any{"businessName": "x", "isActive": true}, true)
	require.NoError(t, err)
	assert.Contains(t, got, `"businessName":"x"`)
	assert.Contains(t, got, `"isActive":true`)
	assert.Contains(t, seen, "returnAllProperties=true")
	assert.Contains(t, string(body), `"id":"abc"`)
}

func TestUpdate_PatchByID(t *testing.T) {
	var seen, method string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		seen = r.URL.RawQuery
		w.WriteHeader(200)
	})
	defer srv.Close()

	_, err := svc.Update(context.Background(), "5", map[string]any{"businessName": "y"})
	require.NoError(t, err)
	assert.Equal(t, "PATCH", method)
	assert.Contains(t, seen, "id=5")
}

func TestDelete(t *testing.T) {
	var seen, method string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		seen = r.URL.RawQuery
		w.WriteHeader(200)
	})
	defer srv.Close()

	_, err := svc.Delete(context.Background(), "5")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", method)
	assert.Contains(t, seen, "id=5")
}

func TestPaginate_AggregatesPages(t *testing.T) {
	calls := 0
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch calls {
		case 1:
			w.Write([]byte(`[{"id":"1"},{"id":"2"},{"id":"3"}]`))
		case 2:
			w.Write([]byte(`[{"id":"4"}]`))
		default:
			w.Write([]byte(`[]`))
		}
	})
	defer srv.Close()

	rows, err := svc.Paginate(context.Background(), odata.Query{}, 3, 0)
	require.NoError(t, err)
	assert.Len(t, rows, 4)
}

func TestSetCustomFieldValue_ReadMergePatch(t *testing.T) {
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
	svc, _ := newSvc(t, nil)
	svc.C.BaseURL = srv.URL

	_, err := svc.SetCustomFieldValue(context.Background(), "1", CustomFieldValue{
		CustomFieldID: 2, CustomFieldRuleID: 1, Line: 1, Value: "added",
	})
	require.NoError(t, err)
	assert.Contains(t, string(patchBody), `"value":"keep"`)
	assert.Contains(t, string(patchBody), `"value":"added"`)
}

func TestClearCustomFieldValue(t *testing.T) {
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
	svc, _ := newSvc(t, nil)
	svc.C.BaseURL = srv.URL

	_, err := svc.ClearCustomFieldValue(context.Background(), "1", 1, 0, 0)
	require.NoError(t, err)
	assert.NotContains(t, string(patchBody), `"value":"goaway"`)
	assert.Contains(t, string(patchBody), `"value":"keep"`)
}

func TestPerson_UnmarshalCapturesExtra(t *testing.T) {
	var p Person
	require.NoError(t, json.Unmarshal([]byte(`{"id":"1","businessName":"X","futureFlag":true,"isActive":true}`), &p))
	assert.Equal(t, "1", p.ID)
	assert.True(t, p.IsActive)
	assert.Contains(t, string(p.Extra), `"futureFlag":true`)
}
