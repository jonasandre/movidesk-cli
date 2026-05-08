// Package contracts covers Movidesk's /timeAgreement and
// /timeAgreementConsumption endpoints (cadastro de contrato de horas e consumo).
package contracts

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
)

const (
	pathAgreement   = "/timeAgreement"
	pathConsumption = "/timeAgreementConsumption"
)

// Person is the minimal client/organization projection used by contract APIs.
type Person struct {
	ID           string `json:"id,omitempty"`
	BusinessName string `json:"businessName,omitempty"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	PersonType   int    `json:"personType,omitempty"`
	ProfileType  int    `json:"profileType,omitempty"`
	IsDeleted    bool   `json:"isDeleted,omitempty"`
}

// Client is timeAgreement.clients[n] entry.
type Client struct {
	Person
	Organization *Person `json:"organization,omitempty"`
}

// TypeActivity is timeAgreement.typeActivities[n].
type TypeActivity struct {
	ID                  int     `json:"id,omitempty"`
	WorkingTimeType     string  `json:"workingTimeType,omitempty"`
	Activity            string  `json:"activity,omitempty"`
	Franchise           int     `json:"franchise,omitempty"`
	Value               float64 `json:"value,omitempty"`
	ValueExceededHour   float64 `json:"valueExceededHour,omitempty"`
	ShootdownContract   bool    `json:"shootdownContract,omitempty"`
	AllowHoursExcedent  bool    `json:"allowHoursExcedent,omitempty"`
	ActivityID          int     `json:"activityId,omitempty"`
}

// TimeTypeConsumption is timeAgreement.timeTypeConsumption[n].
type TimeTypeConsumption struct {
	ID                       int     `json:"id,omitempty"`
	WorkingTimeType          string  `json:"workingTimeType,omitempty"`
	Value                    float64 `json:"value,omitempty"`
	WorkingTimeAgreementID   int     `json:"workingTimeAgreementId,omitempty"`
}

// EmailDestination is timeAgreement.emailDestinations[n].
type EmailDestination struct {
	Type  string `json:"type,omitempty"`
	Email string `json:"email,omitempty"`
}

// Agreement is the /timeAgreement record.
type Agreement struct {
	ID                            int                   `json:"id,omitempty"`
	Name                          string                `json:"name,omitempty"`
	EmailSubject                  string                `json:"emailSubject,omitempty"`
	EmailDescription              string                `json:"emailDescription,omitempty"`
	EmailAccount                  string                `json:"emailAccount,omitempty"`
	IsActive                      bool                  `json:"isActive"`
	DifferentiateHoursFranchise   bool                  `json:"differentiateHoursFranchise"`
	DifferentiateHoursConsumption bool                  `json:"differentiateHoursConsumption"`
	AccumulateUnusedHours         bool                  `json:"accumulateUnusedHours"`
	RenewalDay                    int                   `json:"renewalDay,omitempty"`
	ContractedHours               int                   `json:"contractedHours,omitempty"`
	ConsumptionDeadline           int                   `json:"consumptionDeadline,omitempty"`
	EmailSendDay                  int                   `json:"emailSendDay,omitempty"`
	BaseAmount                    float64               `json:"baseAmount,omitempty"`
	CreatedOn                     string                `json:"createdOn,omitempty"`
	DisabledDate                  string                `json:"disabledDate,omitempty"`
	TypeActivities                []TypeActivity        `json:"typeActivities,omitempty"`
	TimeTypeConsumption           []TimeTypeConsumption `json:"timeTypeConsumption,omitempty"`
	EmailDestinations             []EmailDestination    `json:"emailDestinations,omitempty"`
	Clients                       []Client              `json:"clients,omitempty"`

	Extra json.RawMessage `json:"-"`
}

// UnmarshalJSON populates Agreement and stores raw bytes in Extra.
func (a *Agreement) UnmarshalJSON(data []byte) error {
	type alias Agreement
	var x alias
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	*a = Agreement(x)
	a.Extra = append(json.RawMessage(nil), data...)
	return nil
}

// API binds /timeAgreement endpoints.
type API struct {
	C *movidesk.Client
}

func New(c *movidesk.Client) *API { return &API{C: c} }

// List queries /timeAgreement with the supplied OData query.
func (a *API) List(ctx context.Context, q odata.Query) ([]byte, error) {
	return a.C.Get(ctx, pathAgreement, q, nil)
}

// Get fetches a single agreement by id.
func (a *API) Get(ctx context.Context, id int) ([]byte, error) {
	v := url.Values{}
	v.Set("id", strconv.Itoa(id))
	return a.C.Do(ctx, "GET", pathAgreement, v, nil)
}

// Create posts a new agreement.
func (a *API) Create(ctx context.Context, body any, returnAllProperties bool) ([]byte, error) {
	v := url.Values{}
	if returnAllProperties {
		v.Set("returnAllProperties", "true")
	}
	return a.C.Post(ctx, pathAgreement, v, body)
}

// Update patches an existing agreement.
func (a *API) Update(ctx context.Context, id int, body any) ([]byte, error) {
	v := url.Values{}
	v.Set("id", strconv.Itoa(id))
	return a.C.Patch(ctx, pathAgreement, v, body)
}

// Delete removes an agreement by id.
func (a *API) Delete(ctx context.Context, id int) ([]byte, error) {
	v := url.Values{}
	v.Set("id", strconv.Itoa(id))
	return a.C.Delete(ctx, pathAgreement, v)
}

// Paginate walks /timeAgreement pages.
func (a *API) Paginate(ctx context.Context, q odata.Query, pageSize, max int) ([]json.RawMessage, error) {
	return movidesk.Paginate(ctx, q, func(ctx context.Context, q odata.Query) ([]byte, error) {
		return a.List(ctx, q)
	}, pageSize, max)
}

// ListConsumption queries /timeAgreementConsumption. Per docs, name is required
// when filtering by startPeriod/endPeriod, and $select cannot be combined with
// startPeriod/endPeriod/period filters.
func (a *API) ListConsumption(ctx context.Context, q odata.Query) ([]byte, error) {
	return a.C.Get(ctx, pathConsumption, q, nil)
}

// PaginateConsumption walks /timeAgreementConsumption pages.
func (a *API) PaginateConsumption(ctx context.Context, q odata.Query, pageSize, max int) ([]json.RawMessage, error) {
	return movidesk.Paginate(ctx, q, func(ctx context.Context, q odata.Query) ([]byte, error) {
		return a.ListConsumption(ctx, q)
	}, pageSize, max)
}
