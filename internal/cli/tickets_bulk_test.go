package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bulkServer is a configurable httptest backend for the bulk commands. It
// counts PATCH calls per ticket id, optionally returns a status for ids in
// failIDs, and ignores anything else with a 404.
type bulkServer struct {
	patched  map[int]int
	bodies   map[int][]byte
	failIDs  map[int]int // id -> status code
	listBody []byte
	pastBody []byte
	calls    int32
}

func newBulkServer(listBody []byte) *bulkServer {
	return &bulkServer{
		patched: map[int]int{},
		bodies:  map[int][]byte{},
		failIDs: map[int]int{},
		listBody: func() []byte {
			if listBody == nil {
				return []byte("[]")
			}
			return listBody
		}(),
		pastBody: []byte("[]"),
	}
}

func (b *bulkServer) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&b.calls, 1)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/tickets" && r.URL.Query().Get("id") == "":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(b.listBody)
		case r.Method == http.MethodGet && r.URL.Path == "/tickets/past":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(b.pastBody)
		case r.Method == http.MethodPatch && r.URL.Path == "/tickets":
			idStr := r.URL.Query().Get("id")
			if idStr == "" {
				http.Error(w, "missing id", http.StatusBadRequest)
				return
			}
			var id int
			_, _ = jsonNumberDecode(idStr, &id)
			if code, bad := b.failIDs[id]; bad {
				http.Error(w, "forced failure", code)
				return
			}
			body, _ := io.ReadAll(r.Body)
			b.patched[id]++
			b.bodies[id] = body
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}
}

func jsonNumberDecode(s string, out *int) (int, error) {
	var n int
	_, err := jsonScanInt(s, &n)
	if err != nil {
		return 0, err
	}
	*out = n
	return n, nil
}

// jsonScanInt parses a numeric query-param value without depending on
// strconv.Atoi to keep this helper self-contained.
func jsonScanInt(s string, out *int) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, &json.SyntaxError{}
		}
		n = n*10 + int(c-'0')
	}
	*out = n
	return n, nil
}

// listFixture returns a small payload mirroring what tickets list returns
// for the columns the bulk selector projects via $select.
func listFixture() []byte {
	return []byte(`[
		{"id":1,"subject":"Erro de login","status":"Parado","baseStatus":"Stopped","ownerTeam":"Suporte"},
		{"id":2,"subject":"Lentidão no portal","status":"Parado","baseStatus":"Stopped","ownerTeam":"Suporte"},
		{"id":3,"subject":"Solicitação de acesso","status":"Aberto","baseStatus":"New","ownerTeam":"Suporte"}
	]`)
}

func TestE2E_TicketsBulkUpdate_ByIDsHappyPath(t *testing.T) {
	be := newBulkServer(listFixture())
	srv := httptest.NewServer(be.handler())
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, errOut, err := runCmd(t,
		"tickets", "bulk-update",
		"--ids", "1,2",
		"--set", "status=Resolvido",
		"--force",
	)
	require.NoError(t, err, out, errOut)
	assert.Equal(t, 1, be.patched[1])
	assert.Equal(t, 1, be.patched[2])
	assert.Equal(t, 0, be.patched[3])

	// Body contains the patch field.
	var body map[string]any
	require.NoError(t, json.Unmarshal(be.bodies[1], &body))
	assert.Equal(t, "Resolvido", body["status"])
}

func TestE2E_TicketsBulkUpdate_DryRunMakesNoCalls(t *testing.T) {
	be := newBulkServer(listFixture())
	srv := httptest.NewServer(be.handler())
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, errOut, err := runCmd(t,
		"tickets", "bulk-update",
		"--ids", "1,2",
		"--set", "status=Resolvido",
		"--dry-run",
	)
	require.NoError(t, err)
	assert.Empty(t, be.patched, "dry-run must not call PATCH")
	assert.Contains(t, errOut, "[dry-run]")
}

func TestE2E_TicketsBulkUpdate_RequiresBody(t *testing.T) {
	be := newBulkServer(nil)
	srv := httptest.NewServer(be.handler())
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "tickets", "bulk-update", "--ids", "1", "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nenhum campo")
}

func TestE2E_TicketsBulkUpdate_FilterRequiredWhenNoIDs(t *testing.T) {
	be := newBulkServer(nil)
	srv := httptest.NewServer(be.handler())
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t,
		"tickets", "bulk-update",
		"--set", "status=Resolvido", "--force",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--filter")
}

func TestE2E_TicketsBulkUpdate_FilterAllFromFilter(t *testing.T) {
	be := newBulkServer(listFixture())
	srv := httptest.NewServer(be.handler())
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, errOut, err := runCmd(t,
		"tickets", "bulk-update",
		"--filter", "baseStatus eq 'Stopped'",
		"--all-from-filter",
		"--set", "status=Resolvido",
		"--force",
	)
	require.NoError(t, err, errOut)
	// All three rows in fixture get patched (filter applied server-side; here it's just a label).
	assert.Equal(t, 3, len(be.patched))
}

func TestE2E_TicketsBulkUpdate_ContinueOnError(t *testing.T) {
	be := newBulkServer(listFixture())
	be.failIDs[2] = http.StatusForbidden
	srv := httptest.NewServer(be.handler())
	defer srv.Close()
	setupTenant(t, srv.URL)

	reportPath := filepath.Join(t.TempDir(), "report.jsonl")
	_, errOut, err := runCmd(t,
		"tickets", "bulk-update",
		"--ids", "1,2,3",
		"--set", "status=Resolvido",
		"--force",
		"--continue-on-error",
		"--report", reportPath,
	)
	require.Error(t, err, errOut) // surfaces "1 falha"
	assert.Equal(t, 1, be.patched[1])
	assert.Equal(t, 0, be.patched[2], "id 2 returned 403; nothing recorded")
	assert.Equal(t, 1, be.patched[3])

	raw, rerr := os.ReadFile(reportPath)
	require.NoError(t, rerr)
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	require.Len(t, lines, 3)
	// id 2 must report ok=false
	var entry struct {
		ID int  `json:"id"`
		OK bool `json:"ok"`
	}
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &entry))
	assert.Equal(t, 2, entry.ID)
	assert.False(t, entry.OK)
}

func TestE2E_TicketsBulkUpdate_AbortOnFirstFailureByDefault(t *testing.T) {
	be := newBulkServer(listFixture())
	be.failIDs[1] = http.StatusForbidden
	srv := httptest.NewServer(be.handler())
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t,
		"tickets", "bulk-update",
		"--ids", "1,2,3",
		"--set", "status=Resolvido",
		"--force",
	)
	require.Error(t, err)
	assert.Equal(t, 0, be.patched[2], "should not have continued past id 1")
}

func TestE2E_TicketsBulkClose_HappyPath(t *testing.T) {
	be := newBulkServer(listFixture())
	srv := httptest.NewServer(be.handler())
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, errOut, err := runCmd(t,
		"tickets", "bulk-close",
		"--ids", "1,2",
		"--message", "Fechado por inatividade",
		"--force",
	)
	require.NoError(t, err, errOut)
	assert.Equal(t, 1, be.patched[1])

	var body map[string]any
	require.NoError(t, json.Unmarshal(be.bodies[1], &body))
	assert.Equal(t, "Resolvido", body["status"])
	just, hasJust := body["justification"]
	assert.True(t, hasJust, "justification deve sempre estar presente no body (Movidesk exige o campo)")
	assert.Equal(t, "", just, "default vazio quando --justification não é informado")
	actions, ok := body["actions"].([]any)
	require.True(t, ok)
	require.Len(t, actions, 1)
	act := actions[0].(map[string]any)
	assert.Equal(t, "Fechado por inatividade", act["description"])
	assert.EqualValues(t, 2, act["type"]) // internal by default
	assert.EqualValues(t, 9, act["origin"])
}

func TestE2E_TicketsBulkClose_PublicAndCustomStatus(t *testing.T) {
	be := newBulkServer(listFixture())
	srv := httptest.NewServer(be.handler())
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t,
		"tickets", "bulk-close",
		"--ids", "9",
		"--message", "Concluído",
		"--public",
		"--status", "Fechado",
		"--justification", "Encerrado conforme acordo",
		"--force",
	)
	require.NoError(t, err)
	var body map[string]any
	require.NoError(t, json.Unmarshal(be.bodies[9], &body))
	assert.Equal(t, "Fechado", body["status"])
	assert.Equal(t, "Encerrado conforme acordo", body["justification"])
	actions := body["actions"].([]any)
	act := actions[0].(map[string]any)
	assert.EqualValues(t, 1, act["type"]) // public
}

func TestE2E_TicketsBulkUpdate_SourcePast(t *testing.T) {
	be := newBulkServer(listFixture())
	// Past holds an old, stalled ticket the live API would never return.
	be.pastBody = []byte(`[{"id":777,"subject":"Parado há 150 dias","status":"Parado","baseStatus":"Stopped"}]`)
	srv := httptest.NewServer(be.handler())
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, errOut, err := runCmd(t,
		"tickets", "bulk-update",
		"--source", "past",
		"--filter", "createdDate lt 2025-12-01T00:00:00.000Z",
		"--all-from-filter",
		"--set", "status=Resolvido",
		"--force",
	)
	require.NoError(t, err, errOut)
	assert.Equal(t, 1, be.patched[777])
	assert.Equal(t, 0, be.patched[1], "live results must not leak when --source=past")
}

func TestE2E_TicketsBulkUpdate_SourceBothDedupes(t *testing.T) {
	be := newBulkServer(listFixture())
	// Past returns id 3 (also in live) + a unique 777 — must dedup to 4 rows total.
	be.pastBody = []byte(`[
		{"id":3,"subject":"Solicitação de acesso","status":"Aberto","baseStatus":"New"},
		{"id":777,"subject":"Parado há 150 dias","status":"Parado","baseStatus":"Stopped"}
	]`)
	srv := httptest.NewServer(be.handler())
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, errOut, err := runCmd(t,
		"tickets", "bulk-update",
		"--source", "both",
		"--filter", "id ne 0",
		"--all-from-filter",
		"--set", "status=Resolvido",
		"--force",
	)
	require.NoError(t, err, errOut)
	// 3 from live (1,2,3) + 777 unique from past = 4 patches.
	assert.Equal(t, 1, be.patched[1])
	assert.Equal(t, 1, be.patched[2])
	assert.Equal(t, 1, be.patched[3], "id 3 must be patched exactly once even if both APIs returned it")
	assert.Equal(t, 1, be.patched[777])
	assert.Equal(t, 4, len(be.patched))
}

func TestE2E_TicketsBulkUpdate_InvalidSource(t *testing.T) {
	be := newBulkServer(nil)
	srv := httptest.NewServer(be.handler())
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t,
		"tickets", "bulk-update",
		"--filter", "id eq 1",
		"--source", "archived",
		"--set", "status=Resolvido",
		"--force",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--source")
}

func TestE2E_TicketsBulkClose_RequiresMessage(t *testing.T) {
	be := newBulkServer(nil)
	srv := httptest.NewServer(be.handler())
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "tickets", "bulk-close", "--ids", "1", "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "message")
}
