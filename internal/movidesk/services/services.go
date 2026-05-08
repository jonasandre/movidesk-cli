// Package services covers Movidesk's /services endpoint (GET/POST/PATCH/DELETE).
package services

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
)

const path = "/services"

// Service is a /services row.
//
// serviceForTicketType: 0=Público, 1=Interno, 2=Públicos+Internos
// isVisible / allowSelection: 1=Agente, 2=Cliente, 3=Both
type Service struct {
	ID                   int      `json:"id,omitempty"`
	Name                 string   `json:"name,omitempty"`
	Description          string   `json:"description,omitempty"`
	ParentServiceID      *int     `json:"parentServiceId,omitempty"`
	ServiceForTicketType int      `json:"serviceForTicketType"`
	IsVisible            int      `json:"isVisible"`
	AllowSelection       int      `json:"allowSelection"`
	AllowFinishTicket    bool     `json:"allowFinishTicket"`
	IsActive             bool     `json:"isActive"`
	AutomationMacro      string   `json:"automationMacro,omitempty"`
	DefaultCategory      string   `json:"defaultCategory,omitempty"`
	DefaultUrgency       string   `json:"defaultUrgency,omitempty"`
	AllowAllCategories   bool     `json:"allowAllCategories"`
	Categories           []string `json:"categories,omitempty"`

	Extra json.RawMessage `json:"-"`
}

// UnmarshalJSON populates Service and stores raw bytes in Extra so re-marshal
// preserves any field added to the API after this build.
func (s *Service) UnmarshalJSON(data []byte) error {
	type alias Service
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*s = Service(a)
	s.Extra = append(json.RawMessage(nil), data...)
	return nil
}

// API binds /services endpoints to a Movidesk client.
type API struct {
	C *movidesk.Client
}

func New(c *movidesk.Client) *API { return &API{C: c} }

// List queries /services with the supplied OData query.
func (a *API) List(ctx context.Context, q odata.Query) ([]byte, error) {
	return a.C.Get(ctx, path, q, nil)
}

// Get fetches a single service by id.
func (a *API) Get(ctx context.Context, id int) ([]byte, error) {
	v := url.Values{}
	v.Set("id", strconv.Itoa(id))
	return a.C.Do(ctx, "GET", path, v, nil)
}

// Create posts a new service.
func (a *API) Create(ctx context.Context, body any, returnAllProperties bool) ([]byte, error) {
	v := url.Values{}
	if returnAllProperties {
		v.Set("returnAllProperties", "true")
	}
	return a.C.Post(ctx, path, v, body)
}

// Update patches an existing service. Note: array fields (categories) are
// fully replaced by what you send.
func (a *API) Update(ctx context.Context, id int, body any) ([]byte, error) {
	v := url.Values{}
	v.Set("id", strconv.Itoa(id))
	return a.C.Patch(ctx, path, v, body)
}

// Delete removes a service.
func (a *API) Delete(ctx context.Context, id int) ([]byte, error) {
	v := url.Values{}
	v.Set("id", strconv.Itoa(id))
	return a.C.Delete(ctx, path, v)
}

// Paginate walks /services pages with the given base query.
func (a *API) Paginate(ctx context.Context, q odata.Query, pageSize, max int) ([]json.RawMessage, error) {
	return movidesk.Paginate(ctx, q, func(ctx context.Context, q odata.Query) ([]byte, error) {
		return a.List(ctx, q)
	}, pageSize, max)
}
