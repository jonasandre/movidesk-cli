package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaults_Tickets(t *testing.T) {
	cols := Defaults("tickets")
	require.NotEmpty(t, cols)
	assert.Contains(t, cols, "id")
	assert.Contains(t, cols, "subject")
	assert.Contains(t, cols, "owner.businessName")
	assert.Contains(t, cols, "slaSolutionDate")
}

func TestDefaults_UnknownResource(t *testing.T) {
	assert.Nil(t, Defaults("ghost"))
}

func TestTable_UsesResourceDefaults(t *testing.T) {
	rows := parseRows(t, `[{"id":1,"subject":"x","status":"Novo","owner":{"businessName":"Joe"},"lastUpdate":"2026-01-01"}]`)
	var buf bytes.Buffer
	err := Render(&buf, FormatTable, rows, Options{Resource: "tickets"})
	require.NoError(t, err)
	out := strings.ToLower(buf.String())
	assert.Contains(t, out, "owner.businessname")
	assert.Contains(t, out, "joe")
}

func TestCSV_UsesResourceDefaults(t *testing.T) {
	rows := parseRows(t, `[{"id":1,"subject":"x","status":"Novo","owner":{"businessName":"Joe"},"lastUpdate":"2026-01-01"}]`)
	var buf bytes.Buffer
	err := Render(&buf, FormatCSV, rows, Options{Resource: "tickets"})
	require.NoError(t, err)
	first := strings.SplitN(buf.String(), "\n", 2)[0]
	assert.Contains(t, first, "id,")
	assert.Contains(t, first, "subject,")
	assert.Contains(t, first, "owner.businessName")
}
