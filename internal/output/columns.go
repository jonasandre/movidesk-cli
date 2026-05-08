package output

// Default table columns per resource. The CLI passes the resource name to
// Render via Options.Resource so the table formatter picks sensible defaults
// without each command duplicating the list.
//
// Keys are deliberately short and well-known.
var defaultColumns = map[string][]string{
	"tickets":        {"id", "subject", "status", "owner.businessName", "lastUpdate"},
	"tickets.past":   {"id", "subject", "status", "createdDate", "lastUpdate"},
	"tickets.merged": {"id", "subject", "status", "lastUpdate"},
	"persons":        {"id", "businessName", "personType", "profileType", "isActive"},
	"services":       {"id", "name", "isActive"},
	"activities":     {"id", "subject", "createdDate"},
	"contracts":      {"id", "name", "isActive"},
	"surveys":        {"id", "name", "isActive"},
	"articles":       {"id", "title", "status", "createdDate"},
}

// Defaults returns the registered default columns for a resource, or nil.
func Defaults(resource string) []string {
	cols, ok := defaultColumns[resource]
	if !ok {
		return nil
	}
	return append([]string(nil), cols...)
}
