// Package odata builds query strings using the OData conventions Movidesk
// supports: $filter, $select, $expand, $orderby, $top, $skip, $count.
package odata

import (
	"net/url"
	"strconv"
	"strings"
)

type Query struct {
	Filter  string
	Select  []string
	Expand  []string
	OrderBy []string
	Top     int
	Skip    int
	Count   bool
}

// Apply writes the OData params into v. It does not touch the token param.
func (q Query) Apply(v url.Values) {
	if q.Filter != "" {
		v.Set("$filter", q.Filter)
	}
	if len(q.Select) > 0 {
		v.Set("$select", strings.Join(q.Select, ","))
	}
	if len(q.Expand) > 0 {
		v.Set("$expand", strings.Join(q.Expand, ","))
	}
	if len(q.OrderBy) > 0 {
		v.Set("$orderby", strings.Join(q.OrderBy, ","))
	}
	if q.Top > 0 {
		v.Set("$top", strconv.Itoa(q.Top))
	}
	if q.Skip > 0 {
		v.Set("$skip", strconv.Itoa(q.Skip))
	}
	if q.Count {
		v.Set("$count", "true")
	}
}

// Encode returns the encoded query string, sans token.
func (q Query) Encode() string {
	v := url.Values{}
	q.Apply(v)
	return v.Encode()
}
