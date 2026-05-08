package surveys

import (
	"context"
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

func TestListQuestions_NoFilter(t *testing.T) {
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/survey/questions", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("type"))
		w.Write([]byte(`[{"id":"a","isActive":true,"type":3}]`))
	})
	defer srv.Close()
	body, err := svc.ListQuestions(context.Background(), 0)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"id":"a"`)
}

func TestListQuestions_WithFilter(t *testing.T) {
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "3", r.URL.Query().Get("type"))
		w.Write([]byte(`[]`))
	})
	defer srv.Close()
	_, err := svc.ListQuestions(context.Background(), 3)
	require.NoError(t, err)
}

func TestGetQuestion(t *testing.T) {
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/survey/questions/QWMv", r.URL.Path)
		w.Write([]byte(`{"id":"QWMv","isActive":true,"type":3}`))
	})
	defer srv.Close()
	body, err := svc.GetQuestion(context.Background(), "QWMv")
	require.NoError(t, err)
	assert.Contains(t, string(body), `"id":"QWMv"`)
}

func TestListResponsesPage(t *testing.T) {
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/survey/responses", r.URL.Path)
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		assert.Equal(t, "cursor1", r.URL.Query().Get("startingAfter"))
		w.Write([]byte(`{"hasMore":false,"items":[{"id":"x","value":1}]}`))
	})
	defer srv.Close()
	p, err := svc.ListResponsesPage(context.Background(), 10, "cursor1")
	require.NoError(t, err)
	assert.False(t, p.HasMore)
	assert.Len(t, p.Items, 1)
}

func TestListAllResponses_FollowsCursor(t *testing.T) {
	calls := 0
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch r.URL.Query().Get("startingAfter") {
		case "":
			w.Write([]byte(`{"hasMore":true,"items":[{"id":"a"},{"id":"b"}]}`))
		case "b":
			w.Write([]byte(`{"hasMore":false,"items":[{"id":"c"}]}`))
		default:
			t.Fatalf("unexpected cursor")
		}
	})
	defer srv.Close()
	out, err := svc.ListAllResponses(context.Background(), 0)
	require.NoError(t, err)
	assert.Len(t, out, 3)
	assert.Equal(t, 2, calls)
}
