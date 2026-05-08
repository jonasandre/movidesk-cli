package tickets

import "encoding/json"

// Person describes the people-shaped objects embedded in tickets (owner,
// createdBy, slaSolutionChangedBy, ownerHistories[].owner, action.createdBy,
// timeAppointments[].accountable). Movidesk uses the same minimal projection.
type Person struct {
	ID           string `json:"id,omitempty"`
	BusinessName string `json:"businessName,omitempty"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	PersonType   int    `json:"personType,omitempty"`
	ProfileType  int    `json:"profileType,omitempty"`
}

// Client is a ticket.clients[n].
type Client struct {
	ID           string  `json:"id,omitempty"`
	BusinessName string  `json:"businessName,omitempty"`
	Email        string  `json:"email,omitempty"`
	Phone        string  `json:"phone,omitempty"`
	PersonType   int     `json:"personType,omitempty"`
	ProfileType  int     `json:"profileType,omitempty"`
	IsDeleted    bool    `json:"isDeleted,omitempty"`
	Organization *Person `json:"organization,omitempty"`
}

// Attachment is an action.attachments[n] entry.
type Attachment struct {
	ID          int    `json:"id,omitempty"`
	FileName    string `json:"fileName,omitempty"`
	Path        string `json:"path,omitempty"`
	CreatedDate string `json:"createdDate,omitempty"`
	CreatedBy   string `json:"createdBy,omitempty"`
	IsDeleted   bool   `json:"isDeleted,omitempty"`
}

// TimeAppointment is an action.timeAppointments[n] entry.
type TimeAppointment struct {
	ID                     int     `json:"id,omitempty"`
	Activity               string  `json:"activity,omitempty"`
	Date                   string  `json:"date,omitempty"`
	PeriodStart            string  `json:"periodStart,omitempty"`
	PeriodEnd              string  `json:"periodEnd,omitempty"`
	WorkTime               string  `json:"workTime,omitempty"`
	WorkTypeName           string  `json:"workTypeName,omitempty"`
	Accountable            *Person `json:"accountable,omitempty"`
	IsDeleted              bool    `json:"isDeleted,omitempty"`
	CreatedBy              *Person `json:"createdBy,omitempty"`
	CreatedDate            string  `json:"createdDate,omitempty"`
	WorkTypeID             int     `json:"workTypeId,omitempty"`
}

// Expense is an action.expenses[n] entry.
type Expense struct {
	ID            int     `json:"id,omitempty"`
	Type          string  `json:"type,omitempty"`
	Value         string  `json:"value,omitempty"`
	ServiceReport string  `json:"serviceReport,omitempty"`
	CreatedBy     *Person `json:"createdBy,omitempty"`
	CreatedDate   string  `json:"createdDate,omitempty"`
	IsDeleted     bool    `json:"isDeleted,omitempty"`
}

// Action is a ticket.actions[n] entry.
type Action struct {
	ID                int               `json:"id,omitempty"`
	Type              int               `json:"type,omitempty"`
	Origin            int               `json:"origin,omitempty"`
	Description       string            `json:"description,omitempty"`
	HTMLDescription   string            `json:"htmlDescription,omitempty"`
	Status            string            `json:"status,omitempty"`
	Justification     string            `json:"justification,omitempty"`
	CreatedDate       string            `json:"createdDate,omitempty"`
	CreatedBy         *Person           `json:"createdBy,omitempty"`
	IsDeleted         bool              `json:"isDeleted,omitempty"`
	Tags              []string          `json:"tags,omitempty"`
	TimeAppointments  []TimeAppointment `json:"timeAppointments,omitempty"`
	Expenses          []Expense         `json:"expenses,omitempty"`
	Attachments       []Attachment      `json:"attachments,omitempty"`

	Extra json.RawMessage `json:"-"`
}

// ParentChild is a ticket.parentTickets[n] / childrenTickets[n] entry.
type ParentChild struct {
	ID            int    `json:"id"`
	Subject       string `json:"subject,omitempty"`
	IsDeleted     bool   `json:"isDeleted,omitempty"`
	IsClosed      bool   `json:"isClosed,omitempty"`
	BaseStatus    string `json:"baseStatus,omitempty"`
	Status        string `json:"status,omitempty"`
}

// OwnerHistory is a ticket.ownerHistories[n] entry.
type OwnerHistory struct {
	OwnerTeam              string  `json:"ownerTeam,omitempty"`
	Owner                  *Person `json:"owner,omitempty"`
	ChangedBy              *Person `json:"changedBy,omitempty"`
	ChangedDate            string  `json:"changedDate,omitempty"`
	PermanencyTime         int     `json:"permanencyTime,omitempty"`
	PermanencyTimeFullTime int     `json:"permanencyTimeFullTime,omitempty"`
	PermanencyTimeWorkingTime int  `json:"permanencyTimeWorkingTime,omitempty"`
}

// StatusHistory is a ticket.statusHistories[n] entry.
type StatusHistory struct {
	Status                 string  `json:"status,omitempty"`
	Justification          string  `json:"justification,omitempty"`
	ChangedBy              *Person `json:"changedBy,omitempty"`
	ChangedDate            string  `json:"changedDate,omitempty"`
	PermanencyTime         int     `json:"permanencyTime,omitempty"`
	PermanencyTimeFullTime int     `json:"permanencyTimeFullTime,omitempty"`
	PermanencyTimeWorkingTime int  `json:"permanencyTimeWorkingTime,omitempty"`
}

// CustomFieldItem is a ticket.customFieldValues[n].items[n].
//
// Movidesk reuses the same shape across multiple field types, where exactly
// one of CustomFieldItem / PersonID / ClientID / Team / TeamID is meaningful
// for the field's type.
type CustomFieldItem struct {
	ID              int    `json:"id,omitempty"`
	CustomFieldItem string `json:"customFieldItem,omitempty"`
	PersonID        string `json:"personId,omitempty"`
	ClientID        string `json:"clientId,omitempty"`
	Team            string `json:"team,omitempty"`
	TeamID          int    `json:"teamId,omitempty"`
	Storage         string `json:"storage,omitempty"`
}

// CustomFieldValue is a ticket.customFieldValues[n].
type CustomFieldValue struct {
	CustomFieldID     int               `json:"customFieldId"`
	CustomFieldRuleID int               `json:"customFieldRuleId"`
	Line              int               `json:"line"`
	Column            int               `json:"column,omitempty"`
	Value             string            `json:"value,omitempty"`
	Items             []CustomFieldItem `json:"items,omitempty"`

	Extra json.RawMessage `json:"-"`
}

// Asset is a ticket.assets[n] entry.
type Asset struct {
	ID                  string   `json:"id,omitempty"`
	Name                string   `json:"name,omitempty"`
	Label               string   `json:"label,omitempty"`
	SerialNumber        string   `json:"serialNumber,omitempty"`
	CategoryFull        []string `json:"categoryFull,omitempty"`
	CategoryFirstLevel  string   `json:"categoryFirstLevel,omitempty"`
	CategorySecondLevel string   `json:"categorySecondLevel,omitempty"`
	CategoryThirdLevel  string   `json:"categoryThirdLevel,omitempty"`
	IsDeleted           bool     `json:"isDeleted,omitempty"`
}
