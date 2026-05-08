package tickets

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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

	q := odata.Query{Filter: "id eq 1", Select: []string{"id", "subject"}, Top: 10}
	_, err := svc.List(context.Background(), q, false)
	require.NoError(t, err)
	assert.Contains(t, seen, "%24filter=id+eq+1")
	assert.Contains(t, seen, "%24select=id%2Csubject")
	assert.Contains(t, seen, "%24top=10")
	assert.Contains(t, seen, "token=tok")
}

func TestList_IncludeDeleted(t *testing.T) {
	var seen string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		w.Write([]byte(`[]`))
	})
	defer srv.Close()

	_, err := svc.List(context.Background(), odata.Query{}, true)
	require.NoError(t, err)
	assert.Contains(t, seen, "includeDeletedItems=true")
}

func TestGetByID(t *testing.T) {
	var seen string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		w.Write([]byte(`{"id":42,"subject":"x"}`))
	})
	defer srv.Close()

	body, err := svc.GetByID(context.Background(), 42, false)
	require.NoError(t, err)
	assert.Contains(t, seen, "id=42")
	assert.Contains(t, string(body), `"id":42`)
}

func TestGetByProtocol(t *testing.T) {
	var seen string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		w.Write([]byte(`{}`))
	})
	defer srv.Close()

	_, err := svc.GetByProtocol(context.Background(), "MOVI202109000001", false)
	require.NoError(t, err)
	assert.Contains(t, seen, "protocol=MOVI202109000001")
}

func TestCreate_PostJSON(t *testing.T) {
	var got, seen string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		got = string(buf)
		seen = r.URL.RawQuery
		w.WriteHeader(201)
		w.Write([]byte(`{"id":7}`))
	})
	defer srv.Close()

	body, err := svc.Create(context.Background(), map[string]any{"subject": "hi"}, true)
	require.NoError(t, err)
	assert.Contains(t, got, `"subject":"hi"`)
	assert.Contains(t, seen, "returnAllProperties=true")
	assert.Contains(t, string(body), `"id":7`)
}

func TestUpdate_PatchByID(t *testing.T) {
	var seen, method string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		seen = r.URL.RawQuery
		w.WriteHeader(200)
	})
	defer srv.Close()

	_, err := svc.Update(context.Background(), 5, map[string]any{"subject": "y"})
	require.NoError(t, err)
	assert.Equal(t, "PATCH", method)
	assert.Contains(t, seen, "id=5")
}

func TestHTMLDescription_WithActionID(t *testing.T) {
	var seen string
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Path + "?" + r.URL.RawQuery
		w.Write([]byte(`{"id":1,"description":"<p>x</p>"}`))
	})
	defer srv.Close()

	_, err := svc.HTMLDescription(context.Background(), 1, 3)
	require.NoError(t, err)
	assert.Contains(t, seen, "/tickets/htmldescription")
	assert.Contains(t, seen, "id=1")
	assert.Contains(t, seen, "actionId=3")
}

func TestPaginate_AggregatesPages(t *testing.T) {
	calls := 0
	svc, srv := newSvc(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch calls {
		case 1:
			w.Write([]byte(`[{"id":1},{"id":2},{"id":3}]`))
		case 2:
			w.Write([]byte(`[{"id":4},{"id":5}]`))
		default:
			w.Write([]byte(`[]`))
		}
	})
	defer srv.Close()

	rows, err := svc.Paginate(context.Background(), odata.Query{}, false, 3, 0)
	require.NoError(t, err)
	assert.Len(t, rows, 5)
}

func TestDecodeList(t *testing.T) {
	body := []byte(`[{"id":1,"subject":"a","owner":{"id":"x","businessName":"Joe"}}]`)
	got, err := DecodeList(body)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 1, got[0].ID)
	require.NotNil(t, got[0].Owner)
	assert.Equal(t, "Joe", got[0].Owner.BusinessName)
	assert.True(t, strings.HasPrefix(string(got[0].Extra), `{`))
}
