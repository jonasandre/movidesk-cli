package output

// Default table columns per resource. The CLI passes the resource name to
// Render via Options.Resource so the table formatter picks sensible defaults
// without each command duplicating the list.
//
// Keys are deliberately short and well-known.
var defaultColumns = map[string][]string{
	"tickets":               {"id", "protocol", "subject", "status", "urgency", "owner.businessName", "ownerTeam", "slaSolutionDate", "lastUpdate"},
	"tickets.past":          {"id", "subject", "status", "createdDate", "lastUpdate"},
	"tickets.merged":        {"id", "subject", "status", "lastUpdate"},
	"tickets.actions":       {"id", "type", "origin", "createdDate", "createdBy.businessName", "isDeleted"},
	"tickets.clients":       {"id", "businessName", "email", "personType", "profileType", "isDeleted"},
	"tickets.relations":     {"id", "subject", "isDeleted"},
	"tickets.timeline":      {"when", "kind", "actor", "summary"},
	"tickets.assets":        {"id", "name", "label", "categoryFirstLevel"},
	"tickets.histories":     {"changedDate", "ownerTeam", "owner.businessName", "status", "permanencyTime"},
	"tickets.customfields":  {"label", "customFieldId", "customFieldRuleId", "line", "value", "items"},
	"persons":               {"id", "businessName", "corporateName", "personType", "profileType", "accessProfile", "isActive"},
	"services":              {"id", "name", "parentServiceId", "serviceForTicketType", "isVisible", "allowSelection", "isActive"},
	"activities":            {"id", "name", "isActive", "isAllowsAllTeams"},
	"contracts":             {"id", "name", "renewalDay", "contractedHours", "baseAmount", "isActive"},
	"contracts.consumption": {"id", "name", "period", "consumedHours", "baseAmount", "exceededHourAmount"},
	"surveys.questions":     {"id", "type", "isActive"},
	"surveys.responses":     {"id", "questionId", "type", "ticketId", "value", "responseDate"},
	"articles":              {"id", "title", "articleStatus", "createdDate", "updatedDate"},
}

// Defaults returns the registered default columns for a resource, or nil.
func Defaults(resource string) []string {
	cols, ok := defaultColumns[resource]
	if !ok {
		return nil
	}
	return append([]string(nil), cols...)
}
