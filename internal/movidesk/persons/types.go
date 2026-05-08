// Package persons covers Movidesk's /persons endpoint (GET/POST/PATCH/DELETE).
package persons

import "encoding/json"

// PersonType values: 1=Pessoa, 2=Empresa, 4=Departamento.
// ProfileType values: 1=Agente, 2=Cliente, 3=Agente+Cliente.

// Address is a person.addresses[n] entry.
type Address struct {
	AddressType string `json:"addressType,omitempty"`
	Country     string `json:"country,omitempty"`
	PostalCode  string `json:"postalCode,omitempty"`
	State       string `json:"state,omitempty"`
	City        string `json:"city,omitempty"`
	District    string `json:"district,omitempty"`
	Street      string `json:"street,omitempty"`
	Number      string `json:"number,omitempty"`
	Complement  string `json:"complement,omitempty"`
	Reference   string `json:"reference,omitempty"`
	IsDefault   bool   `json:"isDefault,omitempty"`
}

// Contact is a person.contacts[n] entry.
type Contact struct {
	ContactType string `json:"contactType,omitempty"`
	Contact     string `json:"contact,omitempty"`
	IsDefault   bool   `json:"isDefault,omitempty"`
}

// Email is a person.emails[n] entry.
type Email struct {
	EmailType string `json:"emailType,omitempty"`
	Email     string `json:"email,omitempty"`
	IsDefault bool   `json:"isDefault,omitempty"`
}

// RelationshipService is a person.relationships[n].services[n] entry.
type RelationshipService struct {
	ID             int    `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	CopyToChildren bool   `json:"copyToChildren,omitempty"`
}

// Relationship is a person.relationships[n] entry.
type Relationship struct {
	ID                               string                `json:"id,omitempty"`
	Name                             string                `json:"name,omitempty"`
	SLAAgreement                     string                `json:"slaAgreement,omitempty"`
	ForceChildrenToHaveSomeAgreement bool                  `json:"forceChildrenToHaveSomeAgreement,omitempty"`
	AllowAllServices                 bool                  `json:"allowAllServices,omitempty"`
	IncludeInParents                 bool                  `json:"includeInParents,omitempty"`
	LoadChildOrganizations           bool                  `json:"loadChildOrganizations,omitempty"`
	Services                         []RelationshipService `json:"services,omitempty"`
}

// CustomFieldItem mirrors the ticket-side type. The shape is identical
// across both APIs.
type CustomFieldItem struct {
	PersonID        string `json:"personId,omitempty"`
	ClientID        string `json:"clientId,omitempty"`
	Team            string `json:"team,omitempty"`
	CustomFieldItem string `json:"customFieldItem,omitempty"`
}

// CustomFieldValue is a person.customFieldValues[n] entry.
type CustomFieldValue struct {
	CustomFieldID     int               `json:"customFieldId"`
	CustomFieldRuleID int               `json:"customFieldRuleId"`
	Line              int               `json:"line"`
	Value             string            `json:"value,omitempty"`
	Items             []CustomFieldItem `json:"items,omitempty"`

	Extra json.RawMessage `json:"-"`
}

// Asset is a person.atAssets[n] entry. Note the JSON key uses an uppercase Id.
type Asset struct {
	ID    string `json:"Id,omitempty"`
	Name  string `json:"name,omitempty"`
	Label string `json:"label,omitempty"`
}

// Person models /persons.
type Person struct {
	ID                      string             `json:"id,omitempty"`
	CodeReferenceAdditional string             `json:"codeReferenceAdditional,omitempty"`
	IsActive                bool               `json:"isActive"`
	PersonType              int                `json:"personType,omitempty"`
	ProfileType             int                `json:"profileType,omitempty"`
	AccessProfile           string             `json:"accessProfile,omitempty"`
	CorporateName           string             `json:"corporateName,omitempty"`
	BusinessName            string             `json:"businessName,omitempty"`
	BusinessBranch          string             `json:"businessBranch,omitempty"`
	CPFCNPJ                 string             `json:"cpfCnpj,omitempty"`
	UserName                string             `json:"userName,omitempty"`
	Password                string             `json:"password,omitempty"`
	Role                    string             `json:"role,omitempty"`
	BossID                  string             `json:"bossId,omitempty"`
	BossName                string             `json:"bossName,omitempty"`
	Classification          string             `json:"classification,omitempty"`
	CultureID               string             `json:"cultureId,omitempty"`
	TimeZoneID              string             `json:"timeZoneId,omitempty"`
	AuthenticateOn          string             `json:"authenticateOn,omitempty"`
	CreatedDate             string             `json:"createdDate,omitempty"`
	CreatedBy               string             `json:"createdBy,omitempty"`
	ChangedDate             string             `json:"changedDate,omitempty"`
	ChangedBy               string             `json:"changedBy,omitempty"`
	Observations            string             `json:"observations,omitempty"`
	Addresses               []Address          `json:"addresses,omitempty"`
	Contacts                []Contact          `json:"contacts,omitempty"`
	Emails                  []Email            `json:"emails,omitempty"`
	Teams                   []string           `json:"teams,omitempty"`
	Relationships           []Relationship     `json:"relationships,omitempty"`
	CustomFieldValues       []CustomFieldValue `json:"customFieldValues,omitempty"`
	AtAssets                []Asset            `json:"atAssets,omitempty"`

	Extra json.RawMessage `json:"-"`
}

// UnmarshalJSON populates Person and stores the source bytes in Extra so
// re-marshaling preserves any field added to the API after this build.
func (p *Person) UnmarshalJSON(data []byte) error {
	type alias Person
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*p = Person(a)
	p.Extra = append(json.RawMessage(nil), data...)
	return nil
}

// UnmarshalJSON for CustomFieldValue mirrors the tickets behavior.
func (c *CustomFieldValue) UnmarshalJSON(data []byte) error {
	type alias CustomFieldValue
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*c = CustomFieldValue(a)
	c.Extra = append(json.RawMessage(nil), data...)
	return nil
}
