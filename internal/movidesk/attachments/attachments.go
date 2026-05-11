// Package attachments uploads files to Movidesk via /ticketFileUpload.
//
// Movidesk requires multipart/form-data with two query parameters: the ticket
// id and the action id. The file contents are buffered in memory before
// sending; consider chunked streaming for very large files.
package attachments

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

const path = "/ticketFileUpload"

type Service struct {
	C *movidesk.Client
}

func New(c *movidesk.Client) *Service { return &Service{C: c} }

// Upload posts a single file to the given ticket/action.
//
// `name` is the filename Movidesk will record; it does not have to match the
// source path.
func (s *Service) Upload(ctx context.Context, ticketID, actionID int, name string, r io.Reader) ([]byte, error) {
	if ticketID <= 0 {
		return nil, fmt.Errorf("id do chamado obrigatório")
	}
	if actionID <= 0 {
		return nil, fmt.Errorf("id da ação obrigatório")
	}
	if name == "" {
		return nil, fmt.Errorf("nome do arquivo obrigatório")
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", filepath.Base(name))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, r); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("id", strconv.Itoa(ticketID))
	v.Set("actionId", strconv.Itoa(actionID))

	body, err := s.C.Do(ctx, "POST", path, v, multipartBody{ct: mw.FormDataContentType(), buf: &buf})
	if err != nil {
		return nil, err
	}
	return body, nil
}

// multipartBody signals the client to use a custom Content-Type while still
// accepting an io.Reader.
type multipartBody struct {
	ct  string
	buf *bytes.Buffer
}

func (m multipartBody) Read(p []byte) (int, error) { return m.buf.Read(p) }

// Hooks the client uses to detect a multipart body. We piggyback on type
// assertion in client.do, so this satisfies io.Reader and exposes the CT.
func (m multipartBody) ContentType() string { return m.ct }
