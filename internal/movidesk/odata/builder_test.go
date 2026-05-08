package odata

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApply_AllFields(t *testing.T) {
	q := Query{
		Filter:  "createdDate gt 2026-01-01T00:00:00Z",
		Select:  []string{"id", "subject"},
		Expand:  []string{"owner", "actions($select=id)"},
		OrderBy: []string{"id desc"},
		Top:     50,
		Skip:    100,
		Count:   true,
	}
	v := url.Values{}
	q.Apply(v)

	assert.Equal(t, "createdDate gt 2026-01-01T00:00:00Z", v.Get("$filter"))
	assert.Equal(t, "id,subject", v.Get("$select"))
	assert.Equal(t, "owner,actions($select=id)", v.Get("$expand"))
	assert.Equal(t, "id desc", v.Get("$orderby"))
	assert.Equal(t, "50", v.Get("$top"))
	assert.Equal(t, "100", v.Get("$skip"))
	assert.Equal(t, "true", v.Get("$count"))
}

func TestApply_EmptyOmitsAll(t *testing.T) {
	v := url.Values{}
	(Query{}).Apply(v)
	assert.Empty(t, v)
}

func TestEncode_DoesNotIncludeToken(t *testing.T) {
	q := Query{Filter: "id eq 1", Top: 1}
	enc := q.Encode()
	assert.Contains(t, enc, "%24filter=id+eq+1")
	assert.Contains(t, enc, "%24top=1")
	assert.NotContains(t, enc, "token")
}
