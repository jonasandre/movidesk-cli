// Package knowledgebase covers Movidesk's /article endpoint.
//
// The public API only exposes single-article reads (GET /article/:id).
// There is no public list endpoint; you must already know the article id.
package knowledgebase

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

const path = "/article"

// Attachment is article.attachments[n].
type Attachment struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Hash string `json:"hash,omitempty"`
}

// Category is article.categories[n].
type Category struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Menu is article.menu.
type Menu struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Related is article.relateds[n].
type Related struct {
	ID    int    `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	Path  string `json:"path,omitempty"`
}

// Service describes article.services[n].
type Service struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Article is /article/:id.
//
// articleStatus: 1=Publicado, 2=Suspenso.
type Article struct {
	ID            int          `json:"id,omitempty"`
	ArticleStatus int          `json:"articleStatus,omitempty"`
	Attachments   []Attachment `json:"attachments,omitempty"`
	Categories    []Category   `json:"categories,omitempty"`
	ContentHTML   string       `json:"contentHtml,omitempty"`
	ContentText   string       `json:"contentText,omitempty"`
	CreatedDate   string       `json:"createdDate,omitempty"`
	UpdatedDate   string       `json:"updatedDate,omitempty"`
	Menu          *Menu        `json:"menu,omitempty"`
	Relateds      []Related    `json:"relateds,omitempty"`
	RevisionID    int          `json:"revisionId,omitempty"`
	Services      []Service    `json:"services,omitempty"`
	Slug          string       `json:"slug,omitempty"`
	Summary       string       `json:"summary,omitempty"`
	Tags          []string     `json:"tags,omitempty"`
	Title         string       `json:"title,omitempty"`
	ReadingTime   string       `json:"readingTime,omitempty"`

	Extra json.RawMessage `json:"-"`
}

// UnmarshalJSON populates Article and stores raw bytes in Extra.
func (a *Article) UnmarshalJSON(data []byte) error {
	type alias Article
	var x alias
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	*a = Article(x)
	a.Extra = append(json.RawMessage(nil), data...)
	return nil
}

// API binds /article to a Movidesk client.
type API struct {
	C *movidesk.Client
}

func New(c *movidesk.Client) *API { return &API{C: c} }

// Get fetches a single article by id. Movidesk uses a path parameter for
// /article/:id rather than the typical ?id= query parameter.
func (a *API) Get(ctx context.Context, id int) ([]byte, error) {
	if id <= 0 {
		return nil, errInvalidID
	}
	return a.C.Do(ctx, "GET", path+"/"+url.PathEscape(strconv.Itoa(id)), nil, nil)
}

// errInvalidID surfaces a non-positive id without importing fmt at the call site.
var errInvalidID = invalidIDError{}

type invalidIDError struct{}

func (invalidIDError) Error() string { return "article id must be positive" }
