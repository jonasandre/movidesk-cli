package attachments

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

func TestUpload_PostsMultipartWithIDs(t *testing.T) {
	var (
		gotTicket   string
		gotAction   string
		gotName     string
		gotContents string
		gotCT       string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTicket = r.URL.Query().Get("id")
		gotAction = r.URL.Query().Get("actionId")
		gotCT = r.Header.Get("Content-Type")
		mt, params, err := mime.ParseMediaType(gotCT)
		require.NoError(t, err)
		require.Equal(t, "multipart/form-data", mt)

		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			gotName = p.FileName()
			b, _ := io.ReadAll(p)
			gotContents = string(b)
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := movidesk.New(srv.URL, "tok")
	c.Limiter = movidesk.NewLimiter(1000, 0)
	s := New(c)

	_, err := s.Upload(context.Background(), 12, 34, "report.pdf", strings.NewReader("hello"))
	require.NoError(t, err)
	assert.Equal(t, "12", gotTicket)
	assert.Equal(t, "34", gotAction)
	assert.Equal(t, "report.pdf", gotName)
	assert.Equal(t, "hello", gotContents)
	assert.Contains(t, gotCT, "multipart/form-data; boundary=")
}

func TestUpload_RejectsBadInput(t *testing.T) {
	c := movidesk.New("http://example", "tok")
	s := New(c)
	_, err := s.Upload(context.Background(), 0, 1, "x", strings.NewReader(""))
	require.Error(t, err)
	_, err = s.Upload(context.Background(), 1, 0, "x", strings.NewReader(""))
	require.Error(t, err)
	_, err = s.Upload(context.Background(), 1, 1, "", strings.NewReader(""))
	require.Error(t, err)
}
