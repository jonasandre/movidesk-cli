package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseRows(t *testing.T, body string) []map[string]any {
	t.Helper()
	var raw any
	require.NoError(t, json.Unmarshal([]byte(body), &raw))
	rows := asRows(raw)
	require.NotNil(t, rows)
	return rows
}

func TestJSON_Pretty(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, FormatJSON, map[string]any{"id": 1, "subject": "x"}, Options{})
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "\n  ")
	assert.Contains(t, out, `"subject": "x"`)
}

func TestJSON_Compact(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, FormatJSON, map[string]any{"id": 1}, Options{Compact: true})
	require.NoError(t, err)
	assert.Equal(t, "{\"id\":1}\n", buf.String())
}

func TestTable_DefaultColumnsPickIdAndSubject(t *testing.T) {
	rows := parseRows(t, `[{"id":1,"subject":"alpha","status":"Novo","other":"x"},{"id":2,"subject":"beta","status":"Em atendimento","other":"y"}]`)
	var buf bytes.Buffer
	err := Render(&buf, FormatTable, rows, Options{})
	require.NoError(t, err)
	out := strings.ToLower(buf.String())
	assert.Contains(t, out, "id")
	assert.Contains(t, out, "subject")
	assert.Contains(t, out, "alpha")
	assert.Contains(t, out, "beta")
}

func TestTable_ExplicitColumns(t *testing.T) {
	rows := parseRows(t, `[{"id":1,"subject":"alpha","owner":{"businessName":"Joe"}}]`)
	var buf bytes.Buffer
	err := Render(&buf, FormatTable, rows, Options{Columns: []string{"id", "owner.businessName"}})
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "Joe")
	assert.NotContains(t, out, "alpha")
}

func TestTable_EmptyShowsNoRows(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, FormatTable, []any{}, Options{})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "sem linhas")
}

func TestCSV_HeaderAndRows(t *testing.T) {
	rows := parseRows(t, `[{"id":1,"subject":"alpha"},{"id":2,"subject":"beta"}]`)
	var buf bytes.Buffer
	err := Render(&buf, FormatCSV, rows, Options{Columns: []string{"id", "subject"}})
	require.NoError(t, err)
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	require.Len(t, lines, 3)
	assert.Equal(t, "id,subject", lines[0])
	assert.Equal(t, "1,alpha", lines[1])
	assert.Equal(t, "2,beta", lines[2])
}

func TestCSV_NestedDotPath(t *testing.T) {
	rows := parseRows(t, `[{"id":1,"owner":{"name":"Joe"}}]`)
	var buf bytes.Buffer
	err := Render(&buf, FormatCSV, rows, Options{Columns: []string{"id", "owner.name"}})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Joe")
}

func TestGet_UnknownFormat(t *testing.T) {
	_, err := Get("yaml")
	require.Error(t, err)
}

func TestStringify(t *testing.T) {
	assert.Equal(t, "", stringify(nil))
	assert.Equal(t, "true", stringify(true))
	assert.Equal(t, "42", stringify(float64(42)))
	assert.Equal(t, "3.14", stringify(3.14))
	assert.Equal(t, "a,b", stringify([]any{"a", "b"}))
}

func TestAsRows_RawMessageSlice(t *testing.T) {
	rm := []json.RawMessage{
		json.RawMessage(`{"id":1,"subject":"a"}`),
		json.RawMessage(`{"id":2,"subject":"b"}`),
	}
	rows := asRows(rm)
	require.Len(t, rows, 2)
	assert.Equal(t, float64(1), rows[0]["id"])
	assert.Equal(t, "b", rows[1]["subject"])
}

func TestAsRows_SingleRawMessage(t *testing.T) {
	rm := json.RawMessage(`{"id":7}`)
	rows := asRows(rm)
	require.Len(t, rows, 1)
	assert.Equal(t, float64(7), rows[0]["id"])
}

func TestTable_HandlesRawMessageSlice(t *testing.T) {
	rm := []json.RawMessage{
		json.RawMessage(`{"id":1,"subject":"alpha"}`),
		json.RawMessage(`{"id":2,"subject":"beta"}`),
	}
	var buf bytes.Buffer
	err := Render(&buf, FormatTable, rm, Options{})
	require.NoError(t, err)
	out := strings.ToLower(buf.String())
	assert.Contains(t, out, "alpha")
	assert.Contains(t, out, "beta")
	assert.NotContains(t, out, "sem linhas")
}

func TestCSV_HandlesRawMessageSlice(t *testing.T) {
	rm := []json.RawMessage{
		json.RawMessage(`{"id":1,"subject":"alpha"}`),
		json.RawMessage(`{"id":2,"subject":"beta"}`),
	}
	var buf bytes.Buffer
	err := Render(&buf, FormatCSV, rm, Options{Columns: []string{"id", "subject"}})
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 3)
	assert.Equal(t, "id,subject", lines[0])
	assert.Equal(t, "1,alpha", lines[1])
}

func TestDig(t *testing.T) {
	v := map[string]any{"a": map[string]any{"b": "c"}}
	assert.Equal(t, "c", dig(v, "a.b"))
	assert.Nil(t, dig(v, "a.missing"))
	assert.Nil(t, dig(v, "missing.path"))
}
