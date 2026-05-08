package activities

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

func newSvc(t *testing.T, h http.HandlerFunc) (*Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	c := movidesk.New(srv.URL, "tok")
	c.Limiter = movidesk.NewLimiter(1000, 0)
	return New(c), srv
}

func TestGet(t *testing.T) {
	var seen string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/activity", r.URL.Path)
		seen = r.URL.RawQuery
		w.Write([]byte(`{"id":1,"name":"x","isActive":true,"isAllowsAllTeams":false,"teams":[{"name":"A"}]}`))
	})
	defer srv.Close()

	body, err := svc.Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Contains(t, seen, "id=1")
	assert.Contains(t, string(body), `"name":"x"`)
}

func TestListPage_AppliesLimitAndCursor(t *testing.T) {
	var seen string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		w.Write([]byte(`{"hasMore":false,"items":[{"id":1,"name":"a"}]}`))
	})
	defer srv.Close()

	p, err := svc.ListPage(context.Background(), 5, "100", "needle")
	require.NoError(t, err)
	assert.Contains(t, seen, "limit=5")
	assert.Contains(t, seen, "startingAfter=100")
	assert.Contains(t, seen, "name=needle")
	require.NotNil(t, p)
	assert.Len(t, p.Items, 1)
	assert.False(t, p.HasMore)
}

func TestListAll_FollowsCursor(t *testing.T) {
	var calls int
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch r.URL.Query().Get("startingAfter") {
		case "":
			w.Write([]byte(`{"hasMore":true,"items":[{"id":1},{"id":2},{"id":3}]}`))
		case "3":
			w.Write([]byte(`{"hasMore":false,"items":[{"id":4}]}`))
		default:
			t.Fatalf("unexpected startingAfter %q", r.URL.Query().Get("startingAfter"))
		}
	})
	defer srv.Close()

	out, err := svc.ListAll(context.Background(), "", 0)
	require.NoError(t, err)
	assert.Len(t, out, 4)
	assert.Equal(t, 2, calls)
}

func TestListAll_RespectsMax(t *testing.T) {
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"hasMore":true,"items":[{"id":1},{"id":2},{"id":3}]}`))
	})
	defer srv.Close()
	out, err := svc.ListAll(context.Background(), "", 2)
	require.NoError(t, err)
	assert.Len(t, out, 2)
}

func TestCreate(t *testing.T) {
	var got string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		buf, _ := io.ReadAll(r.Body)
		got = string(buf)
		w.WriteHeader(200)
		w.Write([]byte(`12`))
	})
	defer srv.Close()
	_, err := svc.Create(context.Background(), map[string]any{"name": "x", "isActive": true})
	require.NoError(t, err)
	assert.Contains(t, got, `"name":"x"`)
}

func TestUpdate(t *testing.T) {
	var seen, method string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		seen = r.URL.RawQuery
		w.WriteHeader(200)
	})
	defer srv.Close()
	_, err := svc.Update(context.Background(), 5, map[string]any{"name": "y"})
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
	_, err := svc.Delete(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, "DELETE", method)
	assert.Contains(t, seen, "id=5")
}

func TestAddTeams(t *testing.T) {
	var got, seen string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/addTeamsToActivity", r.URL.Path)
		seen = r.URL.RawQuery
		buf, _ := io.ReadAll(r.Body)
		got = string(buf)
		w.WriteHeader(200)
		w.Write([]byte(`["Suporte"]`))
	})
	defer srv.Close()

	_, err := svc.AddTeams(context.Background(), 7, []string{"Suporte"})
	require.NoError(t, err)
	assert.Contains(t, seen, "activityId=7")
	assert.Contains(t, got, `"Suporte"`)
}
