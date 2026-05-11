// Package tickets covers Movidesk's /tickets, /tickets/past, /tickets/merged
// and /tickets/htmldescription endpoints.
//
// Field coverage is intentionally partial: the high-traffic fields are typed
// for safety and table rendering, while the long tail flows through as
// json.RawMessage in Extra so callers don't lose data.
package tickets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
)

const (
	pathTickets = "/tickets"
	pathPast    = "/tickets/past"
	pathMerged  = "/tickets/merged"
	pathHTML    = "/tickets/htmldescription"
)

// Ticket models a /tickets resource. Field-level docs follow the public
// Movidesk schema; see types.go for nested types.
//
// Round-trip: when Ticket is decoded with UnmarshalJSON it captures the
// original payload in Extra so unknown fields survive a re-marshal.
type Ticket struct {
	// Identity
	ID                 int    `json:"id,omitempty"`
	Protocol           string `json:"protocol,omitempty"`
	Type               int    `json:"type,omitempty"`
	Subject            string `json:"subject,omitempty"`
	Category           string `json:"category,omitempty"`
	Urgency            string `json:"urgency,omitempty"`
	Status             string `json:"status,omitempty"`
	BaseStatus         string `json:"baseStatus,omitempty"`
	Justification      string `json:"justification,omitempty"`
	Origin             int    `json:"origin,omitempty"`
	CreatedDate        string `json:"createdDate,omitempty"`
	OriginEmailAccount string `json:"originEmailAccount,omitempty"`
	Sequence           int    `json:"sequence,omitempty"`
	CC                 string `json:"cc,omitempty"`
	ContactForm        string `json:"contactForm,omitempty"`

	// Ownership
	Owner     *Person `json:"owner,omitempty"`
	OwnerTeam string  `json:"ownerTeam,omitempty"`
	CreatedBy *Person `json:"createdBy,omitempty"`

	// Service catalog
	ServiceFull         []string `json:"serviceFull,omitempty"`
	ServiceFirstLevelID int      `json:"serviceFirstLevelId,omitempty"`
	ServiceFirstLevel   string   `json:"serviceFirstLevel,omitempty"`
	ServiceSecondLevel  string   `json:"serviceSecondLevel,omitempty"`
	ServiceThirdLevel   string   `json:"serviceThirdLevel,omitempty"`

	// Tagging
	Tags []string `json:"tags,omitempty"`

	// Lifecycle dates
	ResolvedIn     string `json:"resolvedIn,omitempty"`
	ReopenedIn     string `json:"reopenedIn,omitempty"`
	ClosedIn       string `json:"closedIn,omitempty"`
	LastActionDate string `json:"lastActionDate,omitempty"`
	LastUpdate     string `json:"lastUpdate,omitempty"`
	ActionCount    int    `json:"actionCount,omitempty"`

	// Working time
	LifetimeWorkingTime    int `json:"lifetimeWorkingTime,omitempty"`
	StoppedTime            int `json:"stoppedTime,omitempty"`
	StoppedTimeWorkingTime int `json:"stoppedTimeWorkingTime,omitempty"`

	// Chat
	ResolvedInFirstCall bool   `json:"resolvedInFirstCall,omitempty"`
	ChatWidget          string `json:"chatWidget,omitempty"`
	ChatGroup           string `json:"chatGroup,omitempty"`
	ChatTalkTime        int    `json:"chatTalkTime,omitempty"`
	ChatWaitingTime     int    `json:"chatWaitingTime,omitempty"`

	// SLA
	SLAAgreement             string  `json:"slaAgreement,omitempty"`
	SLAAgreementRule         string  `json:"slaAgreementRule,omitempty"`
	SLASolutionTime          int     `json:"slaSolutionTime,omitempty"`
	SLAResponseTime          int     `json:"slaResponseTime,omitempty"`
	SLASolutionChangedByUser bool    `json:"slaSolutionChangedByUser,omitempty"`
	SLASolutionChangedBy     *Person `json:"slaSolutionChangedBy,omitempty"`
	SLASolutionDate          string  `json:"slaSolutionDate,omitempty"`
	SLASolutionDateIsPaused  bool    `json:"slaSolutionDateIsPaused,omitempty"`
	SLAResponseDate          string  `json:"slaResponseDate,omitempty"`
	SLARealResponseDate      string  `json:"slaRealResponseDate,omitempty"`

	// External integrations
	JiraIssueKey   string `json:"jiraIssueKey,omitempty"`
	RedmineIssueID int    `json:"redmineIssueId,omitempty"`

	// Collections (only populated with $expand)
	Clients           []Client           `json:"clients,omitempty"`
	Actions           []Action           `json:"actions,omitempty"`
	ParentTickets     []ParentChild      `json:"parentTickets,omitempty"`
	ChildrenTickets   []ParentChild      `json:"childrenTickets,omitempty"`
	OwnerHistories    []OwnerHistory     `json:"ownerHistories,omitempty"`
	StatusHistories   []StatusHistory    `json:"statusHistories,omitempty"`
	CustomFieldValues []CustomFieldValue `json:"customFieldValues,omitempty"`
	Assets            []Asset            `json:"assets,omitempty"`

	// Extra preserves the raw payload for round-trip safety.
	Extra json.RawMessage `json:"-"`
}

// UnmarshalJSON populates Ticket and stores the source bytes in Extra so
// re-marshaling preserves any field Movidesk adds after this build was cut.
func (t *Ticket) UnmarshalJSON(data []byte) error {
	type alias Ticket
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*t = Ticket(a)
	t.Extra = append(json.RawMessage(nil), data...)
	return nil
}

// UnmarshalJSON populates Action and stores raw bytes in Extra.
func (a *Action) UnmarshalJSON(data []byte) error {
	type alias Action
	var x alias
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	*a = Action(x)
	a.Extra = append(json.RawMessage(nil), data...)
	return nil
}

// UnmarshalJSON populates CustomFieldValue and stores raw bytes in Extra.
func (c *CustomFieldValue) UnmarshalJSON(data []byte) error {
	type alias CustomFieldValue
	var x alias
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	*c = CustomFieldValue(x)
	c.Extra = append(json.RawMessage(nil), data...)
	return nil
}

// HTMLBody is returned by /tickets/htmldescription.
type HTMLBody struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

// Service binds tickets endpoints to a Movidesk client.
type Service struct {
	C *movidesk.Client
}

func New(c *movidesk.Client) *Service { return &Service{C: c} }

// List queries /tickets with the supplied OData query.
//
// Returns the raw JSON to keep all fields intact for output rendering.
func (s *Service) List(ctx context.Context, q odata.Query, includeDeleted bool) ([]byte, error) {
	extra := url.Values{}
	if includeDeleted {
		extra.Set("includeDeletedItems", "true")
	}
	return s.C.Get(ctx, pathTickets, q, extra)
}

// Past queries /tickets/past (lastUpdate older than 90 days).
func (s *Service) Past(ctx context.Context, q odata.Query) ([]byte, error) {
	return s.C.Get(ctx, pathPast, q, nil)
}

// Merged queries /tickets/merged.
func (s *Service) Merged(ctx context.Context, q odata.Query) ([]byte, error) {
	return s.C.Get(ctx, pathMerged, q, nil)
}

// GetByID fetches a single ticket by id.
func (s *Service) GetByID(ctx context.Context, id int, includeDeleted bool) ([]byte, error) {
	v := url.Values{}
	v.Set("id", strconv.Itoa(id))
	if includeDeleted {
		v.Set("includeDeletedItems", "true")
	}
	return s.C.Do(ctx, "GET", pathTickets, v, nil)
}

// GetByProtocol fetches a single ticket by protocol.
func (s *Service) GetByProtocol(ctx context.Context, protocol string, includeDeleted bool) ([]byte, error) {
	v := url.Values{}
	v.Set("protocol", protocol)
	if includeDeleted {
		v.Set("includeDeletedItems", "true")
	}
	return s.C.Do(ctx, "GET", pathTickets, v, nil)
}

// HTMLDescription returns the HTML body of a ticket (or one of its actions).
func (s *Service) HTMLDescription(ctx context.Context, id, actionID int) ([]byte, error) {
	v := url.Values{}
	v.Set("id", strconv.Itoa(id))
	if actionID > 0 {
		v.Set("actionId", strconv.Itoa(actionID))
	}
	return s.C.Do(ctx, "GET", pathHTML, v, nil)
}

// Create posts a new ticket. Body is anything json.Marshal can encode.
func (s *Service) Create(ctx context.Context, body any, returnAllProperties bool) ([]byte, error) {
	v := url.Values{}
	if returnAllProperties {
		v.Set("returnAllProperties", "true")
	}
	return s.C.Post(ctx, pathTickets, v, body)
}

// Update patches a ticket by id.
func (s *Service) Update(ctx context.Context, id int, body any) ([]byte, error) {
	v := url.Values{}
	v.Set("id", strconv.Itoa(id))
	return s.C.Patch(ctx, pathTickets, v, body)
}

// Paginate walks /tickets pages with the given base query.
func (s *Service) Paginate(ctx context.Context, q odata.Query, includeDeleted bool, pageSize, max int) ([]json.RawMessage, error) {
	fetch := func(ctx context.Context, q odata.Query) ([]byte, error) {
		return s.List(ctx, q, includeDeleted)
	}
	return movidesk.Paginate(ctx, q, fetch, pageSize, max)
}

// PaginatePast walks /tickets/past.
func (s *Service) PaginatePast(ctx context.Context, q odata.Query, pageSize, max int) ([]json.RawMessage, error) {
	return movidesk.Paginate(ctx, q, func(ctx context.Context, q odata.Query) ([]byte, error) { return s.Past(ctx, q) }, pageSize, max)
}

// PaginateMerged walks /tickets/merged.
func (s *Service) PaginateMerged(ctx context.Context, q odata.Query, pageSize, max int) ([]json.RawMessage, error) {
	return movidesk.Paginate(ctx, q, func(ctx context.Context, q odata.Query) ([]byte, error) { return s.Merged(ctx, q) }, pageSize, max)
}

// DecodeList decodes a list response into typed Tickets, preserving original JSON in Extra.
func DecodeList(raw []byte) ([]Ticket, error) {
	var rms []json.RawMessage
	if err := json.Unmarshal(raw, &rms); err != nil {
		return nil, fmt.Errorf("decodificar lista de chamados: %w", err)
	}
	out := make([]Ticket, 0, len(rms))
	for _, rm := range rms {
		var t Ticket
		if err := json.Unmarshal(rm, &t); err != nil {
			return nil, fmt.Errorf("decodificar chamado: %w", err)
		}
		t.Extra = rm
		out = append(out, t)
	}
	return out, nil
}
