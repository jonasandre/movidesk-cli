package services

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

func newAPI(t *testing.T, h http.HandlerFunc) (*API, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	c := movidesk.New(srv.URL, "tok")
	c.Limiter = movidesk.NewLimiter(1000, 0)
	return New(c), srv
}

func TestList_BuildsODataQuery(t *testing.T) {
	var seen string
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		w.Write([]byte(`[]`))
	})
	defer srv.Close()

	q := odata.Query{Filter: "isActive eq false", OrderBy: []string{"id desc"}, Top: 100}
	_, err := a.List(context.Background(), q)
	require.NoError(t, err)
	assert.Contains(t, seen, "%24filter=isActive+eq+false")
	assert.Contains(t, seen, "%24orderby=id+desc")
}

func TestGet(t *testing.T) {
	var seen string
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		w.Write([]byte(`{"id":5712,"name":"TI","isActive":true,"allowAllCategories":true,"serviceForTicketType":2,"isVisible":3,"allowSelection":3,"allowFinishTicket":true}`))
	})
	defer srv.Close()

	body, err := a.Get(context.Background(), 5712)
	require.NoError(t, err)
	assert.Contains(t, seen, "id=5712")
	var s Service
	require.NoError(t, json.Unmarshal(body, &s))
	assert.Equal(t, "TI", s.Name)
}

func TestCreate(t *testing.T) {
	var got string
	var seen string
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		seen = r.URL.RawQuery
		buf, _ := io.ReadAll(r.Body)
		got = string(buf)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":42}`))
	})
	defer srv.Close()

	body, err := a.Create(context.Background(), map[string]any{
		"name":                 "Teste",
		"isActive":             true,
		"serviceForTicketType": 1,
		"isVisible":            3,
		"allowSelection":       3,
		"allowFinishTicket":    true,
		"allowAllCategories":   true,
	}, true)
	require.NoError(t, err)
	assert.Contains(t, got, `"name":"Teste"`)
	assert.Contains(t, seen, "returnAllProperties=true")
	assert.Contains(t, string(body), `"id":42`)
}

func TestUpdate(t *testing.T) {
	var seen, method string
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		seen = r.URL.RawQuery
		w.WriteHeader(200)
	})
	defer srv.Close()

	_, err := a.Update(context.Background(), 1, map[string]any{"name": "novo"})
	require.NoError(t, err)
	assert.Equal(t, "PATCH", method)
	assert.Contains(t, seen, "id=1")
}

func TestDelete(t *testing.T) {
	var seen, method string
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		seen = r.URL.RawQuery
		w.WriteHeader(200)
	})
	defer srv.Close()

	_, err := a.Delete(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "DELETE", method)
	assert.Contains(t, seen, "id=1")
}

func TestPaginate(t *testing.T) {
	calls := 0
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch calls {
		case 1:
			w.Write([]byte(`[{"id":1,"name":"a"},{"id":2,"name":"b"}]`))
		case 2:
			w.Write([]byte(`[{"id":3,"name":"c"}]`))
		default:
			w.Write([]byte(`[]`))
		}
	})
	defer srv.Close()

	rows, err := a.Paginate(context.Background(), odata.Query{}, 2, 0)
	require.NoError(t, err)
	assert.Len(t, rows, 3)
}

func TestService_UnmarshalCapturesExtra(t *testing.T) {
	var s Service
	require.NoError(t, json.Unmarshal([]byte(`{"id":1,"name":"x","futureFlag":true}`), &s))
	assert.Contains(t, string(s.Extra), `"futureFlag":true`)
}
