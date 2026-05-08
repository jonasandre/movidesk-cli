package knowledgebase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

func newAPI(t *testing.T, h http.HandlerFunc) (*API, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	c := movidesk.New(srv.URL, "tok")
	c.Limiter = movidesk.NewLimiter(1000, 0)
	return New(c), srv
}

func TestGet_UsesPathParam(t *testing.T) {
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/article/19040", r.URL.Path)
		assert.Equal(t, "tok", r.URL.Query().Get("token"))
		w.Write([]byte(`{"id":19040,"title":"X","articleStatus":1}`))
	})
	defer srv.Close()

	body, err := a.Get(context.Background(), 19040)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"id":19040`)
}

func TestGet_RejectsInvalidID(t *testing.T) {
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be hit")
	})
	defer srv.Close()
	_, err := a.Get(context.Background(), 0)
	require.Error(t, err)
}

func TestArticle_UnmarshalCapturesExtra(t *testing.T) {
	var ar Article
	require.NoError(t, ar.UnmarshalJSON([]byte(`{"id":1,"title":"x","futureFlag":true}`)))
	assert.Contains(t, string(ar.Extra), `"futureFlag":true`)
}
