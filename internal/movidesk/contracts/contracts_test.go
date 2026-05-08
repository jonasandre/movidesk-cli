package contracts

import (
	"context"
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

func TestList(t *testing.T) {
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/timeAgreement", r.URL.Path)
		w.Write([]byte(`[{"id":1,"name":"Default","isActive":true}]`))
	})
	defer srv.Close()
	body, err := a.List(context.Background(), odata.Query{Top: 5})
	require.NoError(t, err)
	assert.Contains(t, string(body), "Default")
}

func TestGet(t *testing.T) {
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "1", r.URL.Query().Get("id"))
		w.Write([]byte(`{"id":1,"name":"X"}`))
	})
	defer srv.Close()
	_, err := a.Get(context.Background(), 1)
	require.NoError(t, err)
}

func TestCreate(t *testing.T) {
	var got string
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		buf, _ := io.ReadAll(r.Body)
		got = string(buf)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":7}`))
	})
	defer srv.Close()
	_, err := a.Create(context.Background(), map[string]any{"name": "x", "isActive": true}, true)
	require.NoError(t, err)
	assert.Contains(t, got, `"name":"x"`)
}

func TestUpdateDelete(t *testing.T) {
	var method, seen string
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		seen = r.URL.RawQuery
		w.WriteHeader(200)
	})
	defer srv.Close()

	_, err := a.Update(context.Background(), 1, map[string]any{"name": "y"})
	require.NoError(t, err)
	assert.Equal(t, "PATCH", method)
	assert.Contains(t, seen, "id=1")

	_, err = a.Delete(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "DELETE", method)
}

func TestListConsumption(t *testing.T) {
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/timeAgreementConsumption", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "%24filter=name+eq+%27X%27")
		w.Write([]byte(`[{"id":"abc","name":"X"}]`))
	})
	defer srv.Close()
	body, err := a.ListConsumption(context.Background(), odata.Query{Filter: "name eq 'X'"})
	require.NoError(t, err)
	assert.Contains(t, string(body), "X")
}

func TestAgreement_UnmarshalCapturesExtra(t *testing.T) {
	var ag Agreement
	require.NoError(t, ag.UnmarshalJSON([]byte(`{"id":1,"name":"x","futureFlag":true}`)))
	assert.Contains(t, string(ag.Extra), `"futureFlag":true`)
}
