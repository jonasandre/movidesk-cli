package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

// startServer spins up the MCP server against an in-memory transport pair and
// returns a ready-to-use ClientSession plus the *movidesk.Client driving the
// upstream stub. The caller is responsible for closing the session and the
// upstream test server.
func startServer(t *testing.T, upstream http.Handler, cfg Config) *mcpsdk.ClientSession {
	t.Helper()

	stub := httptest.NewServer(upstream)
	t.Cleanup(stub.Close)

	client := movidesk.New(stub.URL, "test-token")
	// Tests should not be slowed down by the 10 req/min rate limit.
	client.Limiter = movidesk.NewLimiter(1000, time.Second)
	client.Retry.Disabled = true

	st, ct := mcpsdk.NewInMemoryTransports()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	srv := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "movidesk-cli", Version: "test"}, nil)
	registerTools(srv, client)
	registerResources(srv, client, cfg)
	srvSession, err := srv.Connect(ctx, st, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = srvSession.Close() })

	mc := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "0"}, nil)
	cs, err := mc.Connect(ctx, ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cs.Close() })

	return cs
}

func TestToolsList_ExposesEntireV1Catalog(t *testing.T) {
	cs := startServer(t, http.NewServeMux(), Config{Tenant: "demo"})

	res, err := cs.ListTools(context.Background(), nil)
	require.NoError(t, err)

	names := map[string]bool{}
	for _, tl := range res.Tools {
		names[tl.Name] = true
	}

	expected := []string{
		"tickets_list", "tickets_past_list", "tickets_get", "tickets_html_description",
		"tickets_actions_list", "tickets_timeline", "tickets_customfields_show",
		"persons_list", "persons_get", "persons_customfields_show",
		"services_list", "services_get",
		"contracts_list", "contracts_get", "contracts_consumption_list",
		"kb_article_get",
		"activities_list", "activities_get",
		"surveys_questions_list", "surveys_questions_get", "surveys_responses_list",
		"query",
	}
	for _, want := range expected {
		assert.True(t, names[want], "tool %q missing from tools/list", want)
	}
	assert.Len(t, res.Tools, len(expected), "tool count drift; update expected list intentionally")
}

func TestToolsCall_TicketsList_RawJSONRoundTrip(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tickets", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "id eq 1", r.URL.Query().Get("$filter"))
		assert.Equal(t, "test-token", r.URL.Query().Get("token"))
		_, _ = w.Write([]byte(`[{"id":1,"subject":"hi"}]`))
	})
	cs := startServer(t, mux, Config{Tenant: "demo"})

	res, err := cs.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "tickets_list",
		Arguments: map[string]any{"filter": "id eq 1"},
	})
	require.NoError(t, err)
	require.False(t, res.IsError, "unexpected error: %s", textOf(res))

	assert.JSONEq(t, `[{"id":1,"subject":"hi"}]`, textOf(res))
}

func TestToolsCall_RateLimit429_SurfacedAsTranslatedError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tickets", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate_limit"}`))
	})
	cs := startServer(t, mux, Config{Tenant: "demo"})

	res, err := cs.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "tickets_list",
	})
	require.NoError(t, err)
	require.True(t, res.IsError, "expected IsError=true on 429")
	assert.Contains(t, textOf(res), "rate limit")
}

func TestToolsCall_DefaultMax_AppliedOnAllWithoutMax(t *testing.T) {
	var calls int
	mux := http.NewServeMux()
	mux.HandleFunc("/tickets", func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(makePage(100)))
	})
	cs := startServer(t, mux, Config{Tenant: "demo"})

	res, err := cs.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "tickets_list",
		Arguments: map[string]any{"all": true},
	})
	require.NoError(t, err)
	require.False(t, res.IsError, "unexpected error: %s", textOf(res))

	var rows []json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(textOf(res)), &rows))
	assert.Len(t, rows, 500, "default cap should stop at 500 rows")
	assert.Equal(t, 5, calls, "5 pages of 100 to reach cap")
}

func TestToolsCall_QueryEscapeHatch_PropagatesODataAndExtras(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/persons", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "id,businessName", r.URL.Query().Get("$select"))
		assert.Equal(t, "personType eq 2", r.URL.Query().Get("$filter"))
		assert.Equal(t, "true", r.URL.Query().Get("includeDeletedItems"))
		_, _ = w.Write([]byte(`[{"id":"abc"}]`))
	})
	cs := startServer(t, mux, Config{Tenant: "demo"})

	res, err := cs.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "query",
		Arguments: map[string]any{
			"path":   "/persons",
			"filter": "personType eq 2",
			"select": []string{"id", "businessName"},
			"params": map[string]string{"includeDeletedItems": "true"},
		},
	})
	require.NoError(t, err)
	require.False(t, res.IsError, "unexpected error: %s", textOf(res))
	assert.JSONEq(t, `[{"id":"abc"}]`, textOf(res))
}

func TestToolsCall_TicketsGet_RequiresIDXorProtocol(t *testing.T) {
	cs := startServer(t, http.NewServeMux(), Config{Tenant: "demo"})

	res, err := cs.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "tickets_get",
	})
	require.NoError(t, err)
	assert.True(t, res.IsError)

	res, err = cs.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "tickets_get",
		Arguments: map[string]any{"id": 1, "protocol": "P1"},
	})
	require.NoError(t, err)
	assert.True(t, res.IsError)
}

func TestResources_ODataFilterSyntax_AndCustomFieldsCatalog(t *testing.T) {
	catalog := []byte(`{"Severidade":{"id":42,"rule_id":7,"type":"single-select","options":["A","B"]}}`)
	cs := startServer(t, http.NewServeMux(), Config{Tenant: "demo", CustomFields: catalog})

	list, err := cs.ListResources(context.Background(), nil)
	require.NoError(t, err)
	uris := map[string]string{}
	for _, r := range list.Resources {
		uris[r.URI] = r.MIMEType
	}
	assert.Equal(t, "text/markdown", uris["movidesk://odata-filter-syntax"])
	assert.Equal(t, "application/json", uris["movidesk://server-info"])
	assert.Equal(t, "application/json", uris["movidesk://customfields-catalog"])

	got, err := cs.ReadResource(context.Background(), &mcpsdk.ReadResourceParams{
		URI: "movidesk://odata-filter-syntax",
	})
	require.NoError(t, err)
	require.NotEmpty(t, got.Contents)
	assert.Contains(t, got.Contents[0].Text, "ownerTeam")
	assert.Contains(t, got.Contents[0].Text, "startswith")

	got, err = cs.ReadResource(context.Background(), &mcpsdk.ReadResourceParams{
		URI: "movidesk://customfields-catalog",
	})
	require.NoError(t, err)
	require.NotEmpty(t, got.Contents)
	assert.JSONEq(t, string(catalog), got.Contents[0].Text)
}

func TestResources_CustomFieldsCatalogOmittedWhenAbsent(t *testing.T) {
	cs := startServer(t, http.NewServeMux(), Config{Tenant: "demo"})

	list, err := cs.ListResources(context.Background(), nil)
	require.NoError(t, err)
	for _, r := range list.Resources {
		assert.NotEqual(t, "movidesk://customfields-catalog", r.URI,
			"catalog resource should not be advertised without a payload")
	}
}

// textOf concatenates the text-content blocks of a CallToolResult into a
// single string, which is how every tool in this package emits its payload.
func textOf(r *mcpsdk.CallToolResult) string {
	var b strings.Builder
	for _, c := range r.Content {
		if tc, ok := c.(*mcpsdk.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

func TestStringList_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want StringList
	}{
		{"canonical array", `["id","name"]`, StringList{"id", "name"}},
		{"json-stringified array", `"[\"id\",\"name\"]"`, StringList{"id", "name"}},
		{"comma-string", `"id,name"`, StringList{"id", "name"}},
		{"comma-string with spaces", `"id, name , protocol"`, StringList{"id", "name", "protocol"}},
		{"single element string", `"id"`, StringList{"id"}},
		{"null", `null`, nil},
		{"empty string", `""`, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got StringList
			require.NoError(t, json.Unmarshal([]byte(tc.in), &got))
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("rejects number", func(t *testing.T) {
		var got StringList
		assert.Error(t, json.Unmarshal([]byte(`42`), &got))
	})
}

func TestClampTicketsTop(t *testing.T) {
	cases := []struct {
		in          int
		wantTop     int
		wantWarning bool
	}{
		{0, 0, false},
		{100, 100, false},
		{250, 250, false},
		{251, 250, true},
		{1000, 250, true},
	}
	for _, tc := range cases {
		got, warning := clampTicketsTop(tc.in)
		assert.Equal(t, tc.wantTop, got, "top for input %d", tc.in)
		assert.Equal(t, tc.wantWarning, warning != "", "warning for input %d", tc.in)
	}
}

func TestToolsCall_TicketsList_AcceptsSelectVariants(t *testing.T) {
	cases := []struct {
		name    string
		select_ any
	}{
		{"array", []string{"id", "protocol"}},
		{"json-string", `["id","protocol"]`},
		{"comma-string", "id,protocol"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/tickets", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "id,protocol", r.URL.Query().Get("$select"))
				_, _ = w.Write([]byte(`[{"id":1}]`))
			})
			cs := startServer(t, mux, Config{Tenant: "demo"})

			res, err := cs.CallTool(context.Background(), &mcpsdk.CallToolParams{
				Name: "tickets_list",
				Arguments: map[string]any{
					"select": tc.select_,
				},
			})
			require.NoError(t, err)
			require.False(t, res.IsError, "unexpected error: %s", textOf(res))
			assert.Contains(t, textOf(res), `"id":1`)
		})
	}
}

func TestToolsCall_TicketsList_ClampsLargeTop(t *testing.T) {
	var seenTop string
	mux := http.NewServeMux()
	mux.HandleFunc("/tickets", func(w http.ResponseWriter, r *http.Request) {
		seenTop = r.URL.Query().Get("$top")
		_, _ = w.Write([]byte(`[{"id":1}]`))
	})
	cs := startServer(t, mux, Config{Tenant: "demo"})

	res, err := cs.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "tickets_list",
		Arguments: map[string]any{"top": 1000},
	})
	require.NoError(t, err)
	require.False(t, res.IsError, "unexpected error: %s", textOf(res))
	assert.Equal(t, "250", seenTop, "$top should be clamped to 250 before reaching API")
	assert.Contains(t, textOf(res), "reduzido para 250", "user-visible warning should be attached")
}

// makePage builds a JSON array of n placeholder objects matching the shape the
// pagination cap test expects.
func makePage(n int) string {
	var b strings.Builder
	b.WriteByte('[')
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(`{"id":`)
		b.WriteString(strconv.Itoa(i))
		b.WriteByte('}')
	}
	b.WriteByte(']')
	return b.String()
}
