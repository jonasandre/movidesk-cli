package telephony

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestQueuePost_KnownEvents(t *testing.T) {
	cases := map[string]string{
		"receivedCall":   "/asterisk_receivedCall",
		"transferedCall": "/asterisk_transferedCall",
		"completedCall":  "/asterisk_completedCall",
		"lostCall":       "/asterisk_lostCall",
		"canceledCall":   "/asterisk_canceledCall",
	}
	for event, path := range cases {
		t.Run(event, func(t *testing.T) {
			var seen string
			svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				seen = r.URL.Path
				buf, _ := io.ReadAll(r.Body)
				assert.Contains(t, string(buf), `"id":"x"`)
				w.WriteHeader(200)
			})
			defer srv.Close()

			_, err := svc.QueuePost(context.Background(), event, map[string]any{"id": "x"})
			require.NoError(t, err)
			assert.Equal(t, path, seen)
		})
	}
}

func TestQueuePost_UnknownEvent(t *testing.T) {
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) { t.Fatalf("should not call") })
	defer srv.Close()
	_, err := svc.QueuePost(context.Background(), "ghost", nil)
	require.Error(t, err)
}

func TestNonQueueGet(t *testing.T) {
	var seen, query string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		seen = r.URL.Path
		query = r.URL.RawQuery
		w.WriteHeader(200)
	})
	defer srv.Close()

	v := url.Values{}
	v.Set("queueId", "1")
	v.Set("clientNumber", "555")
	v.Set("id", "abc")
	_, err := svc.NonQueueGet(context.Background(), "startTransferedCall", v)
	require.NoError(t, err)
	assert.Equal(t, "/asterisk_startTransferedCall", seen)
	assert.Contains(t, query, "queueId=1")
	assert.Contains(t, query, "clientNumber=555")
}

func TestNonQueueGet_UnknownEvent(t *testing.T) {
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) { t.Fatalf("should not call") })
	defer srv.Close()
	_, err := svc.NonQueueGet(context.Background(), "ghost", nil)
	require.Error(t, err)
}

func TestSetMadeCallLink(t *testing.T) {
	var seen string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		seen = r.URL.Path
		w.WriteHeader(200)
	})
	defer srv.Close()
	_, err := svc.SetMadeCallLink(context.Background(), map[string]any{"id": "x", "link": "https://recordings/x"})
	require.NoError(t, err)
	assert.Equal(t, "/setMadeCallLink", seen)
}
