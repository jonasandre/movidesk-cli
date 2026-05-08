package customfields

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

func newAPI(t *testing.T, h http.HandlerFunc) (*API, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	c := movidesk.New(srv.URL, "tok")
	c.Limiter = movidesk.NewLimiter(1000, 0)
	return New(c), srv
}

func TestAddOptions(t *testing.T) {
	var seen, body string
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		seen = r.URL.Path
		buf, _ := io.ReadAll(r.Body)
		body = string(buf)
		w.Write([]byte(`{"values":[{"name":"A","success":true,"message":""}]}`))
	})
	defer srv.Close()

	_, err := a.AddOptions(context.Background(), "125529", []string{"A", "B"})
	require.NoError(t, err)
	assert.Equal(t, "/ticketCustomFieldValue/InsertValues", seen)
	assert.Contains(t, body, `"customfieldid":"125529"`)
	assert.Contains(t, body, `"customfieldvalues":["A","B"]`)
}

func TestRenameOptions(t *testing.T) {
	var seen, body string
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Path
		buf, _ := io.ReadAll(r.Body)
		body = string(buf)
		w.Write([]byte(`{"values":[]}`))
	})
	defer srv.Close()

	_, err := a.RenameOptions(context.Background(), "125529", []UpdatePair{
		{OldName: "OLD", NewName: "NEW"},
	})
	require.NoError(t, err)
	assert.Equal(t, "/ticketCustomFieldValue/UpdateValues", seen)
	assert.Contains(t, body, `"oldname":"OLD"`)
	assert.Contains(t, body, `"newname":"NEW"`)
}

func TestRemoveOptions(t *testing.T) {
	var seen, body string
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Path
		buf, _ := io.ReadAll(r.Body)
		body = string(buf)
		w.Write([]byte(`{"values":[]}`))
	})
	defer srv.Close()

	_, err := a.RemoveOptions(context.Background(), "125529", []string{"X"})
	require.NoError(t, err)
	assert.Equal(t, "/ticketCustomFieldValue/DeleteValues", seen)
	assert.Contains(t, body, `"customfieldvalues":["X"]`)
}

func TestValidatesInputs(t *testing.T) {
	a, srv := newAPI(t, func(w http.ResponseWriter, r *http.Request) { t.Fatalf("should not call") })
	defer srv.Close()
	_, err := a.AddOptions(context.Background(), "", []string{"A"})
	require.Error(t, err)
	_, err = a.AddOptions(context.Background(), "1", nil)
	require.Error(t, err)
	_, err = a.RenameOptions(context.Background(), "", nil)
	require.Error(t, err)
	_, err = a.RemoveOptions(context.Background(), "1", nil)
	require.Error(t, err)
}
